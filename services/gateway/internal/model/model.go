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
}
