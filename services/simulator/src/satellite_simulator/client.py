import grpc
from google.protobuf.timestamp_pb2 import Timestamp
from satellite.v1 import satellite_pb2, satellite_pb2_grpc


def main():
    with grpc.insecure_channel("localhost:50051") as channel:
        client = satellite_pb2_grpc.SatelliteGatewayServiceStub(channel)

        observed_at = Timestamp()
        observed_at.GetCurrentTime()

        request = satellite_pb2.SendTelemetryRequest(
            satellite=satellite_pb2.SatelliteIdentity(
                satellite_id="SAT-001",
                name="Test Satellite",
            ),
            observed_at=observed_at,
            battery_percent=87.5,
            connection_status=satellite_pb2.CONNECTION_STATUS_CONNECTED,
            sequence_number=1,
        )

        response = client.SendTelemetry(request, timeout=5)

        print("Telemetry accepted:", response.accepted)


if __name__ == "__main__":
    main()
