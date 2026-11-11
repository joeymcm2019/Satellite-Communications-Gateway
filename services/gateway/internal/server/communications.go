package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	satellitev1 "github.com/satellite-communications/gateway/gen/satellite/v1"
	"github.com/satellite-communications/gateway/internal/events"
	"github.com/satellite-communications/gateway/internal/model"
	"google.golang.org/grpc"
)

type CommunicationManager struct {
	diagnosticRequests <-chan string

	mu     sync.Mutex
	stream grpc.BidiStreamingServer[
		satellitev1.SatelliteMessage,
		satellitev1.GatewayMessage,
	]
	broadcaster *events.Broadcaster
}

func NewCommunicationManager(
	diagnosticRequests <-chan string,
	broadcaster *events.Broadcaster,
) *CommunicationManager {
	return &CommunicationManager{
		diagnosticRequests: diagnosticRequests,
		broadcaster:        broadcaster,
	}
}

func (m *CommunicationManager) SetStream(
	stream grpc.BidiStreamingServer[
		satellitev1.SatelliteMessage,
		satellitev1.GatewayMessage,
	],
) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stream = stream
}

func (m *CommunicationManager) Run(ctx context.Context) {
	for {
		select {
		case satelliteID, ok := <-m.diagnosticRequests:
			if !ok { //exit rather than repeatedly receiving empty string
				return
			}

			if err := m.sendPing(satelliteID); err != nil {
				fmt.Printf(
					"Failed to ping satellite %s: %v\n",
					satelliteID,
					err,
				)
				continue
			}

			m.broadcaster.Publish(model.GatewayEvent{
				Type:        model.EventPingSent,
				SatelliteID: satelliteID,
				Timestamp:   time.Now(),
			})

			fmt.Println("Ping sent to:", satelliteID)
			fmt.Println()

		case <-ctx.Done():
			return
		}
	}
}

func (m *CommunicationManager) sendPing(satelliteID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stream == nil {
		return fmt.Errorf("No active satellite communication stream.")
	}

	message := &satellitev1.GatewayMessage{
		Payload: &satellitev1.GatewayMessage_Ping{
			Ping: &satellitev1.PingRequest{
				SatelliteId: satelliteID,
			},
		},
	}

	return m.stream.Send(message)
}

func (m *CommunicationManager) ClearStream(
	stream grpc.BidiStreamingServer[
		satellitev1.SatelliteMessage,
		satellitev1.GatewayMessage,
	],
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	//protect against race condition during reconnect
	if m.stream == stream {
		m.stream = nil
	}
}

func (m *CommunicationManager) SendRecoveryCommand(satelliteID string) error {
	m.mu.Lock()

	if m.stream == nil {
		return fmt.Errorf("No active communication stream")
	}

	message := &satellitev1.GatewayMessage{
		Payload: &satellitev1.GatewayMessage_Command{
			Command: &satellitev1.Command{
				CommandId:   uuid.NewString(),
				SatelliteId: satelliteID,
				Action:      "RESUME_TELEMETRY",
			},
		},
	}

	err := m.stream.Send(message)
	m.mu.Unlock()

	if err != nil {
		return err
	}

	m.broadcaster.Publish(model.GatewayEvent{
		Type:        model.EventRecoveryRequested,
		SatelliteID: satelliteID,
		Timestamp:   time.Now(),
	})

	return nil
}
