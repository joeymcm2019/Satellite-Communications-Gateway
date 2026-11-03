import os
import unittest
from threading import Event
from unittest.mock import patch

from satellite_simulator.config import Config
from satellite_simulator.service import SimulatorService


class SimulatorTests(unittest.TestCase):
    def test_defaults_and_placeholders(self) -> None:
        with patch.dict(os.environ, {}, clear=True):
            service = SimulatorService(Config.from_env())
        self.assertEqual(service.satellite.config.satellite_id, "SAT-001")
        self.assertIsNone(service.generate_telemetry())
        self.assertIsNone(service.connect_to_gateway())
        self.assertIsNone(service.simulate_outage())

    def test_invalid_interval(self) -> None:
        for value in ("0", "-1", "nan", "inf", "invalid"):
            with self.subTest(value=value):
                with patch.dict(os.environ, {"SIMULATOR_INTERVAL_SECONDS": value}):
                    with self.assertRaises(ValueError):
                        Config.from_env()

    def test_stopped_service_exits(self) -> None:
        stop = Event()
        stop.set()
        SimulatorService(Config()).run(stop)
