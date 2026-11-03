from dataclasses import dataclass


@dataclass(frozen=True)
class SatelliteConfig:
    satellite_id: str
    name: str
    telemetry_interval_seconds: float = 5.0


@dataclass
class Satellite:
    config: SatelliteConfig
    battery_percent: float = 100.0
    charging: bool = False
    communication_outage: bool = False
    outage_ticks_remaining: int = 0
    sequence_number: int = 0
