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
			ping := payload.PingResponse

			fmt.Printf(
				"Ping response from %s: healthy=%t\n",
				ping.SatelliteId,
				ping.Healthy,
			)
			fmt.Println()

			if !ping.Healthy {
				fmt.Printf(
					"Attempting recovery for satellite %s\n",
					ping.SatelliteId,
				)

				if err := s.communications.SendRecoveryCommand(ping.SatelliteId); err != nil {
					fmt.Printf("Recovery command failed: %v\n", err)
				}
			}
		case *satellitev1.SatelliteMessage_CommandAcknowledgement:
			ack := payload.CommandAcknowledgement

			fmt.Printf(
				"Command acknowledged by %s\n",
				ack.SatelliteId,
				ack.Accepted,
				ack.Detail,
			)
			fmt.Println()
		default:
			fmt.Println("Unknown satellite message")
		}
	}
}
