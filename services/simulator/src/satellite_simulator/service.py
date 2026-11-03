import logging
import random
import time
from datetime import timedelta
from threading import Event

import grpc
from google.protobuf.timestamp_pb2 import Timestamp
from satellite.v1 import satellite_pb2

from satellite_simulator.client import GatewayClient
from satellite_simulator.config import Config
from satellite_simulator.satellite import Satellite

logger = logging.getLogger(__name__)


class SimulatorService:
    def __init__(self, config: Config, client: GatewayClient, satellites: list[Satellite]) -> None:
        self.config = config
        self.client = client
        self.satellites = satellites

    def generate_telemetry(self, satellite: Satellite) -> satellite_pb2.SendTelemetryRequest:
        satellite.sequence_number += 1
        observed_at = Timestamp()
        observed_at.GetCurrentTime()

        if satellite.battery_percent < 10 and not satellite.charging:
            satellite.charging = True
            logger.info("Satellite %s is now charging", satellite.config.satellite_id)
        elif satellite.battery_percent == 100 and satellite.charging:
            satellite.charging = False
            logger.info("Satellite %s is fully charged", satellite.config.satellite_id)

        battery_change = random.uniform(0.1, 0.3) * (-1 if not satellite.charging else 1)
        new_battery_percent = satellite.battery_percent + battery_change
        satellite.battery_percent = round(max(0, min(100, new_battery_percent)), 2)

        return satellite_pb2.SendTelemetryRequest(
            satellite=satellite_pb2.SatelliteIdentity(
                satellite_id=satellite.config.satellite_id,
                name=satellite.config.name,
            ),
            observed_at=observed_at,
            battery_percent=satellite.battery_percent,
            connection_status=satellite_pb2.CONNECTION_STATUS_CONNECTED,
            sequence_number=satellite.sequence_number,
            expected_interval_seconds=timedelta(  ##Todo only want to send this once.
                seconds=satellite.config.telemetry_interval_seconds
            ),
        )

    def connect_to_gateway(self) -> None:
        """Placeholder: future gRPC connection; opens no network connection."""

    def simulate_outage(self, satellite: Satellite) -> None:
        if satellite.communication_outage:
            satellite.outage_ticks_remaining -= 1
            if satellite.outage_ticks_remaining <= 0:
                satellite.communication_outage = False
                logger.info("Satellite %s communication restored", satellite.config.satellite_id)
        else:
            if random.random() < 0.05:
                satellite.communication_outage = True
                satellite.outage_ticks_remaining = random.randint(1, 5)
                logger.warning(
                    "Satellite %s communication outage for %d ticks",
                    satellite.config.satellite_id,
                    satellite.outage_ticks_remaining,
                )

    def run(self, stop: Event) -> None:
        logger.info(
            "Starting simulator with %d satellites",
            len(self.satellites),
        )

        # track the next transmission time for each satellite
        next_transmissions = {
            satellite.config.satellite_id: (
                time.monotonic() + satellite.config.telemetry_interval_seconds
            )
            for satellite in self.satellites
        }

        while not stop.is_set():
            now = time.monotonic()

            for satellite in self.satellites:
                satellite_id = satellite.config.satellite_id

                if now < next_transmissions[satellite_id]:
                    continue

                telemetry = self.generate_telemetry(satellite)
                self.simulate_outage(satellite)

                if satellite.communication_outage:
                    logger.warning(
                        "Satellite %s is in communication outage, skipping telemetry %d",
                        satellite_id,
                        telemetry.sequence_number,
                    )

                else:
                    try:
                        accepted = self.client.send_telemetry(telemetry)
                        if accepted:
                            logger.info(
                                "Telemetry sent for satellite %s, sequence %d",
                                satellite_id,
                                telemetry.sequence_number,
                            )

                    except grpc.RpcError as error:
                        logger.error(
                            "Failed to send telemetry for satellite %s, sequence %d: %s",
                            satellite_id,
                            telemetry.sequence_number,
                            error,
                        )
                next_transmissions[satellite_id] = (
                    time.monotonic() + satellite.config.telemetry_interval_seconds
                )
            if next_transmissions:
                next_due = min(next_transmissions.values())
                wait_seconds = max(0, next_due - time.monotonic())
                stop.wait(wait_seconds)
            else:
                stop.wait()

        logger.info("Simulator stopped")
