package server

import (
	"context"
	"fmt"
	"io"

	satellitev1 "github.com/satellite-communications/gateway/gen/satellite/v1"
	"google.golang.org/grpc"
)

type GatewayServer struct {
	satellitev1.UnimplementedSatelliteGatewayServiceServer

	tracker *SatelliteTrackers
}

func NewGatewayServer() *GatewayServer {
	return &GatewayServer{
		tracker: NewSatelliteTracker(),
	}
}

func (s *GatewayServer) SendTelemetry(
	ctx context.Context,
	req *satellitev1.SendTelemetryRequest,
) (*satellitev1.TelemetryReceipt, error) {

	PrintTelemetry(req)
	s.tracker.RecordTelemetry(req)

	return &satellitev1.TelemetryReceipt{
		Accepted: true,
	}, nil
}

func (s *GatewayServer) MonitorSatellites(ctx context.Context) {
	s.tracker.MonitorSatellites(ctx)
}

func (s *GatewayServer) Communicate(
	stream grpc.BidiStreamingServer[
		satellitev1.SatelliteMessage,
		satellitev1.GatewayMessage,
	],
) error {

	for {
		message, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		fmt.Println("Received satellite message:", message)
	}
}
