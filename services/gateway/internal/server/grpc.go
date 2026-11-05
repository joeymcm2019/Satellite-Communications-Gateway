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

	communications *CommunicationManager
	tracker        *SatelliteTrackers
}

func NewGatewayServer() *GatewayServer {
	tracker := NewSatelliteTracker()

	manager := NewCommunicationManager(
		tracker.diagnosticRequests,
	)

	return &GatewayServer{
		tracker:        tracker,
		communications: manager,
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

func (s *GatewayServer) MonitorCommunications(ctx context.Context) {
	s.communications.Run(ctx)
}

func (s *GatewayServer) Communicate(
	stream grpc.BidiStreamingServer[
		satellitev1.SatelliteMessage,
		satellitev1.GatewayMessage,
	],
) error {
	s.communications.SetStream(stream)
	defer s.communications.ClearStream(stream)
	for {
		message, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		switch payload := message.Payload.(type) {
		case *satellitev1.SatelliteMessage_PingResponse:
			fmt.Printf(
				"Ping response from %s: healthy=%t\n",
				payload.PingResponse.SatelliteId,
				payload.PingResponse.Healthy,
			)
		case *satellitev1.SatelliteMessage_CommandAcknowledgement:
			fmt.Printf(
				"Command acknowledged by %s\n",
				payload.CommandAcknowledgement.SatelliteId,
			)
		default:
			fmt.Println("Unknown satellite message")
		}
	}
}
