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
	timestamp := req.GetObservedAt().AsTime().Local()
	fmt.Println("Satellite:", req.GetSatellite().GetSatelliteId())
	fmt.Println("Battery:", req.GetBatteryPercent())
	fmt.Println("Sequence:", req.GetSequenceNumber())
	fmt.Println("Timestamp:", timestamp.Format("2006-01-02 03:04:05 PM MST"))

	return &satellitev1.TelemetryReceipt{
		Accepted: true,
	}, nil
}
