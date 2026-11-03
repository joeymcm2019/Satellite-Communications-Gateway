package server

import (
	"context"

	satellitev1 "github.com/satellite-communications/gateway/gen/satellite/v1"
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
