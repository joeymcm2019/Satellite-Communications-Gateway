package model

import "time"

type SatelliteSnapshot struct {
	ID                      string    `json:"id"`
	Name                    string    `json:"name"`
	Status                  string    `json:"status"`
	BatteryPercent          float64   `json:"battery_percent"`
	SequenceNumber          uint64    `json:"sequence_number"`
	LastSeen                time.Time `json:"last_seen"`
	ExpectedIntervalSeconds float64   `json:"expected_interval_seconds"`
	Charging                bool      `json:"charging"`
}

const (
	EventTelemetry           = "satellite.telemetry"
	EventInactive            = "satellite.inactive"
	EventPingSent            = "satellite.ping_sent"
	EventPingResponse        = "satellite.ping_response"
	EventRecoveryRequested   = "satellite.recovery_requested"
	EventCommandAcknowledged = "satellite.command_acknowledged"
	EventOnline              = "satellite.online"
)

type GatewayEvent struct {
	Type        string         `json:"type"`
	SatelliteID string         `json:"satellite_id"`
	Timestamp   time.Time      `json:"timestamp"`
	Data        map[string]any `json:"data,omitempty"`
}
