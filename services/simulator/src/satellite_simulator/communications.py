import logging
from queue import Queue
from threading import Event

import grpc
from satellite.v1 import satellite_pb2


class CommunicationClient:
    def __init__(self, gateway_client, satellites, satellite_lock):
        self.gateway_client = gateway_client
        self.satellites = {satellite.config.satellite_id: satellite for satellite in satellites}
        self.satellite_lock = satellite_lock

    def outgoing_messages(self, outgoing):
        while True:
            message = outgoing.get()

            if message is None:  ##when does this happen?
                return

            yield message

    def handle_message(self, message, outgoing):
        if message.HasField("ping"):
            ping = message.ping

            logging.info(
                "Ping received for satellite %s",
                ping.satellite_id,
            )

            with self.satellite_lock:
                satellite = self.satellites.get(ping.satellite_id)

                healthy = satellite is not None and not satellite.communication_outage

            response = satellite_pb2.SatelliteMessage(
                ping_response=satellite_pb2.PingResponse(
                    satellite_id=ping.satellite_id,
                    healthy=healthy,
                )
            )

            outgoing.put(response)

        elif message.HasField("command"):
            self.handle_command(message.command, outgoing)

    def handle_command(self, command, outgoing):

        accepted = False
        detail = ""

        with self.satellite_lock:
            satellite = self.satellites.get(command.satellite_id)

            if satellite is None:
                detail = "Satellite not found"

            elif command.action != "RESUME_TELEMETRY":
                detail = "Unsupported command"

            else:
                satellite.communication_outage = False
                satellite.outage_ticks_remaining = 0

                accepted = True
                detail = "Telemetry restored"

        if accepted:
            logging.info(
                "Recovered satellite %s",
                command.satellite_id,
            )

        response = satellite_pb2.SatelliteMessage(
            command_acknowledgment=satellite_pb2.CommandAcknowledgment(
                command_id=command.command_id,
                satellite_id=command.satellite_id,
                accepted=accepted,
                detail=detail,
            )
        )

        outgoing.put(response)

    def run(self, stop: Event):
        retry_delay = 2

        while not stop.is_set():
            outgoing = Queue()

            try:
                logging.info("Connecting to Go gateway...")

                responses = self.gateway_client.communicate(self.outgoing_messages(outgoing))
                ## Live iterator connected to our gRPC stream.

                logging.info("Communication stream started")

                for message in responses:
                    if stop.is_set():
                        return

                    self.handle_message(message, outgoing)

                logging.warning("Communication stream closed")

            except grpc.RpcError as error:
                if stop.is_set():
                    return

                logging.warning(
                    "Communication stream failed: %s",
                    error,
                )

            finally:
                outgoing.put(None)

            if not stop.is_set():
                logging.info(
                    "Reconnecting in %s seconds...",
                    retry_delay,
                )
                stop.wait(retry_delay)
