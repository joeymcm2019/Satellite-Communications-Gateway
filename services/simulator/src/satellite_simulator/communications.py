import logging
from queue import Queue
from threading import Event

import grpc
from satellite.v1 import satellite_pb2


class CommunicationClient:
    def __init__(self, gateway_client):
        self.gateway_client = gateway_client
        self.outgoing = Queue()

    def outgoing_messages(self):
        while True:
            message = self.outgoing.get()
            yield message

            if message is None:  ##when does this happen?
                return

            yield message

    def handle_message(self, message):
        if message.HasField("ping"):
            ping = message.ping

            logging.info(
                "Ping received for satellite %s",
                ping.satellite_id,
            )

            response = satellite_pb2.SatelliteMessage(
                ping_response=satellite_pb2.PingResponse(
                    satellite_id=ping.satellite_id,
                    healthy=True,
                )
            )

            self.outgoing.put(response)

    def run(self, stop: Event):
        retry_delay = 2

        while not stop.is_set():
            try:
                logging.info("Connecting to Go gateway...")

                responses = self.gateway_client.communicate(self.outgoing_messages())
                ## Live iterator connected to our gRPC stream.

                logging.info("Communication stream started")

                for message in responses:
                    if stop.is_set():
                        return

                    self.handle_message(message)

                logging.warning("Communication stream closed")

            except grpc.RpcError as error:
                if stop.is_set():
                    return

                logging.warning(
                    "Communication stream failed: %s",
                    error,
                )

            if not stop.is_set():
                logging.info(
                    "Reconnecting in %s seconds...",
                    retry_delay,
                )

            stop.wait(retry_delay)
