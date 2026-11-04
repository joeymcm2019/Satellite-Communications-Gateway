package server

import (
	"context"
	"fmt"
)

type CommunicationManager struct {
	diagnosticRequests <-chan string
}

func (m *CommunicationManager) Run(ctx context.Context) {
	for {
		select {
		case satelliteID := <-m.diagnosticRequests:
			fmt.Println("Ping requested for:", satelliteID)

		case <-ctx.Done():
			return
		}
	}
}
