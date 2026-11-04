package server

import (
	"context"
	"fmt"
)

type CommunicationManager struct {
	diagnosticRequests <-chan string
}

func NewCommunicationManager(
	diagnosticRequests <-chan string,
) *CommunicationManager {
	return &CommunicationManager{
		diagnosticRequests: diagnosticRequests,
	}
}

func (m *CommunicationManager) Run(ctx context.Context) {
	for {
		select {
		case satelliteID, ok := <-m.diagnosticRequests:
			if !ok { //exit rather than repeatedly receiving empty string
				return
			}

			fmt.Println("Ping requested for:", satelliteID)
			fmt.Println()

		case <-ctx.Done():
			return
		}
	}
}
