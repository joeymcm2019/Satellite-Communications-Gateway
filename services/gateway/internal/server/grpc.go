package server

import (
	"context"
	"fmt"
	satellitev1 "github.com/satellite-communications/gateway/gen/satellite/v1"
)

type GatewayServer struct {
	satellitev1.UnimplementedSatelliteGatewayServiceServer
}

func (s *GatewayServer) SendTelemetry(
	ctx context.Context,
	req *satellitev1.SendTelemetryRequest,
) (*satellitev1.TelemetryReceipt, error) {
	fmt.Println("Satellite:", req.GetSatellite().GetSatelliteId())
	fmt.Println("Battery:", req.GetBatteryPercent())
	fmt.Println("Sequence:", req.GetSequenceNumber())

	return &satellitev1.TelemetryReceipt {
		Accepted: true,
	}, nil
}