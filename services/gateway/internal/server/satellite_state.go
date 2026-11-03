package server

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	satellitev1 "github.com/satellite-communications/gateway/gen/satellite/v1"
)

type SatelliteState struct {
	LastReceivedAt   time.Time
	PreviousSequence uint64
	ExpectedInterval time.Duration
}

type SatelliteTrackers struct {
	mu         sync.Mutex
	satellites map[string]SatelliteState
}

func NewSatelliteTracker() *SatelliteTrackers {
	return &SatelliteTrackers{
		satellites: make(map[string]SatelliteState),
	}
}

func PrintTelemetry(req *satellitev1.SendTelemetryRequest) {
	satellite := req.GetSatellite()
	timestamp := req.GetObservedAt().AsTime().Local()

	fmt.Println("Satellite:", satellite.GetSatelliteId())
	fmt.Println("Name: ", satellite.GetName())
	fmt.Println("Battery:", req.GetBatteryPercent())
	fmt.Println("Sequence:", req.GetSequenceNumber())
	fmt.Println("Timestamp:", timestamp.Format("2006-01-02 03:04:05 PM MST"))
}

func (s *SatelliteTrackers) RecordTelemetry(
	req *satellitev1.SendTelemetryRequest,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	satellites := s.satellites
	satellite := req.GetSatellite()
	satelliteID := satellite.GetSatelliteId()
	sat, exists := satellites[satelliteID]
	currentSequence := req.GetSequenceNumber()

	if exists {
		previousSequence := sat.PreviousSequence
		deltaSequence := int64(currentSequence) - int64(previousSequence)

		if deltaSequence > 1 { //missed telemetry(s)
			fmt.Println(
				getMissedTelemetryMsg(currentSequence, previousSequence),
			)
		}

		sat.LastReceivedAt = time.Now()
		sat.PreviousSequence = currentSequence
		satellites[satelliteID] = sat
	} else {
		duration := req.GetExpectedIntervalSeconds()

		if duration == nil {
			fmt.Println("Error. Expected Interval is nil.")
			return
		}

		satellites[satelliteID] = SatelliteState{
			LastReceivedAt:   time.Now(),
			PreviousSequence: currentSequence,
			ExpectedInterval: duration.AsDuration(),
		}
	}
}

func getMissedTelemetryMsg(sequence uint64, previousSequence uint64) string {
	missedTel := sequence - previousSequence
	msg := "Missed telemetry: " + strconv.FormatUint((sequence-1), 10)
	if missedTel > 2 {
		msg = "Missed telemetries: " + strconv.FormatUint((previousSequence+1), 10) + " - " + strconv.FormatUint((sequence-1), 10) + "."
	}
	return msg
}

func checkInterval(sequence uint64, previousSequence uint64, interval int) {

}
