"""Environment-backed simulator configuration."""

import math
import os
from dataclasses import dataclass
from typing import Self


@dataclass(frozen=True)
class Config:
    gateway_address: str = "localhost:50051"
    interval_seconds: float = 5.0

    @classmethod
    def from_env(cls) -> Self:
        interval = float(
            os.getenv(
                "SIMULATOR_INTERVAL_SECONDS",
                "5",
            )
        )
        if not math.isfinite(interval) or interval <= 0:
            raise ValueError("SIMULATOR_INTERVAL_SECONDS must be finite and positive")
        satellite_id = os.getenv("SATELLITE_ID", "SAT-001").strip()
        address = os.getenv("GATEWAY_GRPC_ADDRESS", "localhost:50051").strip()
        if not satellite_id or not address:
            raise ValueError("Satellite ID and gateway address must not be empty")
        return cls(satellite_id, address, interval)
