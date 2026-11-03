import grpc
import logging
import random
from dataclasses import dataclass
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
    charging: bool = False
    communication_outage: bool = False
    outage_ticks_remaining: int = 0


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
        if self.satellite.battery_percent < 10 and not self.satellite.charging:
            self.satellite.charging = True
            logger.info("Satellite %s is now charging", self.satellite.satellite_id)
        elif self.satellite.battery_percent == 100 and self.satellite.charging:
            self.satellite.charging = False
            logger.info("Satellite %s is fully charged", self.satellite.satellite_id)
        battery_change = random.uniform(0.1, 0.3) * (-1 if not self.satellite.charging else 1)
        new_battery_percent = self.satellite.battery_percent + battery_change
        self.satellite.battery_percent = round(max(0, min(100, new_battery_percent)), 2)
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
        if self.satellite.communication_outage:
            self.satellite.outage_ticks_remaining -= 1
            if self.satellite.outage_ticks_remaining <= 0:
                self.satellite.communication_outage = False
                logger.info("Satellite %s communication restored", self.satellite.satellite_id)
        else:
            if random.random() < 0.05:
                self.satellite.communication_outage = True
                self.satellite.outage_ticks_remaining = random.randint(1, 5)
                logger.warning(
                    "Satellite %s communication outage for %d ticks",
                    self.satellite.satellite_id,
                    self.satellite.outage_ticks_remaining,
                )

    def run(self, stop: Event) -> None:
        logger.info("Simulator ready for %s", self.satellite.satellite_id)
        while not stop.wait(self.config.interval_seconds):
            telemetry = self.generate_telemetry()
            self.simulate_outage()

            if self.satellite.communication_outage:
                logger.warning(
                    "Satellite %s communication outage, skipping telemetry #%s",
                    self.satellite.satellite_id,
                    telemetry.sequence_number,
                )
                continue

            try:
                accepted = self.client.send_telemetry(telemetry)
                logger.info(
                    "Telemetry #%s accepted: %s",
                    telemetry.sequence_number,
                    accepted,
                )

            except grpc.RpcError as error:
                logger.warning(
                    "Failed to send telemetry #%s: %s",
                    telemetry.sequence_number,
                    error.code(),
                )
        logger.info("Simulator stopped")
