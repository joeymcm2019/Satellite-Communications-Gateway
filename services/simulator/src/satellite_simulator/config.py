import os
from dataclasses import dataclass
from typing import Self


@dataclass(frozen=True)
class Config:
    gateway_address: str = "localhost:50051"

    @classmethod
    def from_env(cls) -> Self:
        address = os.getenv("GATEWAY_GRPC_ADDRESS", "localhost:50051").strip()
        if not address:
            raise ValueError("Gateway address must not be empty")
        return cls(address)
