"""Run with python -m satellite_simulator."""

import argparse
import logging
import signal
from threading import Event

from satellite_simulator.client import GatewayClient
from satellite_simulator.config import Config
from satellite_simulator.satellite import Satellite, SatelliteConfig
from satellite_simulator.service import SimulatorService


def main() -> None:
    parser = argparse.ArgumentParser(description="Local satellite simulator scaffold")
    parser.add_argument("--once", action="store_true", help="Validate configuration and exit")
    args = parser.parse_args()
    logging.basicConfig(level=logging.INFO, format="%(levelname)s %(message)s")
    config = Config.from_env()
    client = GatewayClient(config.gateway_address)

    satellites = [
        Satellite(SatelliteConfig("SAT-001", "Demo Satellite 1", 5.0)),
        Satellite(SatelliteConfig("SAT-002", "Demo Satellite 2", 10.0)),
        Satellite(SatelliteConfig("SAT-003", "Demo Satellite 3", 15.0)),
    ]

    service = SimulatorService(config, client, satellites)

    if args.once:
        logging.info("Configuration valid for %s", service.satellite.config.satellite_id)
        return
    stop = Event()
    for sig in (signal.SIGINT, signal.SIGTERM):
        signal.signal(sig, lambda _signum, _frame: stop.set())
    try:
        service.run(stop)
    finally:
        client.close()


if __name__ == "__main__":
    main()
