import logging
import random
from dataclasses import dataclass
from datetime import datetime
from threading import Event

from google.protobuf.timestamp_pb2 import Timestamp
from satellite.v1 import satellite_pb2

from satellite_simulator.client import GatewayClient
from satellite_simulator.config import Config

logger = logging.getLogger(__name__)


@dataclass
class Satellite:
    satellite_id: str
    name: str
    battery_percent: float


class SimulatorService:
    def __init__(self, config: Config, client: GatewayClient) -> None:
        self.config = config
        self.client = client
        self.satellite = Satellite(config.satellite_id, "Demo satellite", battery_percent=100.0)
        self.sequence_number = 0

    def generate_telemetry(self) -> satellite_pb2.SendTelemetryRequest:
        self.sequence_number += 1
        observed_at = Timestamp()
        observed_at.GetCurrentTime()
        battery_percent = self.satellite.battery_percent - random.uniform(0.1, 0.3)
        self.satellite.battery_percent = round(max(0, battery_percent), 2)
        return satellite_pb2.SendTelemetryRequest(
            satellite=satellite_pb2.SatelliteIdentity(
                satellite_id=self.satellite.satellite_id,
                name=self.satellite.name,
            ),
            observed_at=observed_at,
            battery_percent=self.satellite.battery_percent,
            connection_status=satellite_pb2.CONNECTION_STATUS_CONNECTED,
            sequence_number=self.sequence_number,
        )

    def connect_to_gateway(self) -> None:
        """Placeholder: future gRPC connection; opens no network connection."""

    def simulate_outage(self) -> None:
        """Placeholder: future simulated communication outage."""

    def run(self, stop: Event) -> None:
        logger.info("Simulator ready for %s", self.satellite.satellite_id)
        while not stop.wait(self.config.interval_seconds):
            telemetry = self.generate_telemetry()
            logger.info("Generated telemetry: %s", telemetry)
            self.client.send_telemetry(telemetry)
        logger.info("Simulator stopped")
