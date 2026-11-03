"""Lifecycle scaffold; transport and simulation behavior are intentionally absent."""

import logging
from dataclasses import dataclass
from threading import Event

from satellite_simulator.config import Config

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class Satellite:
    satellite_id: str
    name: str


class SimulatorService:
    def __init__(self, config: Config) -> None:
        self.config = config
        self.satellite = Satellite(config.satellite_id, "Demo satellite")

    def generate_telemetry(self) -> None:
        """Placeholder: future telemetry generation."""

    def connect_to_gateway(self) -> None:
        """Placeholder: future gRPC connection; opens no network connection."""

    def simulate_outage(self) -> None:
        """Placeholder: future simulated communication outage."""

    def run(self, stop: Event) -> None:
        logger.info("Simulator ready for %s", self.satellite.satellite_id)
        while not stop.wait(self.config.interval_seconds):
            logger.debug("Simulator idle; telemetry generation is planned")
        logger.info("Simulator stopped")
