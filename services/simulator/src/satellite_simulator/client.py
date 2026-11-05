import grpc
from satellite.v1 import satellite_pb2, satellite_pb2_grpc


class GatewayClient:
    def __init__(self, address: str) -> None:
        self.channel = grpc.insecure_channel(address)
        self.stub = satellite_pb2_grpc.SatelliteGatewayServiceStub(self.channel)

    def send_telemetry(
        self,
        request: satellite_pb2.SendTelemetryRequest,
    ) -> bool:
        response = self.stub.SendTelemetry(request, timeout=5)
        return response.accepted

    def close(self) -> None:
        self.channel.close()

    def communicate(self, outgoing_messages):
        return self.stub.Communicate(outgoing_messages)
