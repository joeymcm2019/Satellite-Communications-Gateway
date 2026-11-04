package server

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	satellitev1 "github.com/satellite-communications/gateway/gen/satellite/v1"
)

type SatelliteState struct {
	lastReceivedAt   time.Time
	previousSequence uint64
	expectedInterval time.Duration
}

type SatelliteTrackers struct {
	mu                 sync.Mutex
	satellites         map[string]SatelliteState
	intervalChanged    chan struct{}
	diagnosticRequests chan string
}

func NewSatelliteTracker() *SatelliteTrackers {
	return &SatelliteTrackers{
		satellites:         make(map[string]SatelliteState),
		intervalChanged:    make(chan struct{}, 1),
		diagnosticRequests: make(chan string, 100),
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
	fmt.Println()
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
		previousSequence := sat.previousSequence
		deltaSequence := currentSequence - previousSequence

		if deltaSequence > 1 && currentSequence > previousSequence { //missed telemetry(s)
			fmt.Println(
				satellite.GetName()+":",
				getMissedTelemetryMsg(currentSequence, previousSequence),
			)
			fmt.Println()
		}

		if deltaSequence >= 1 && currentSequence > previousSequence {
			percentOff := getPercentOff(
				currentSequence, previousSequence, sat.expectedInterval.Seconds(), time.Since(sat.lastReceivedAt).Seconds(),
			)

			if percentOff > 30 { //how to define a constant? put this in config?
				fmt.Println(satellite.GetName(), "telemetry is late by", percentOff, "%")
				fmt.Println()
			}
		}

		sat.lastReceivedAt = time.Now()
		sat.previousSequence = currentSequence
		satellites[satelliteID] = sat
	} else {
		duration := req.GetExpectedIntervalSeconds()

		if duration == nil {
			fmt.Println("Error. Expected Interval is nil.")
			return
		}

		interval := duration.AsDuration()

		if interval <= 0 {
			fmt.Println("Error. Expected interval must be positive.")
			return
		}

		satellites[satelliteID] = SatelliteState{
			lastReceivedAt:   time.Now(),
			previousSequence: currentSequence,
			expectedInterval: duration.AsDuration(),
		}

		select {
		case s.intervalChanged <- struct{}{}: //new satellite has arrived
		default: //prevents telemetry handler from blocking if a notification is already waiting
		}
	}
}

func getMissedTelemetryMsg(sequence uint64, previousSequence uint64) string {
	missedTel := sequence - previousSequence
	msg := "Missed telemetry: " + strconv.FormatUint((sequence-1), 10)
	if missedTel > 2 {
		msg = "Missed telemetries: " + strconv.FormatUint((previousSequence+1), 10) + " - " + strconv.FormatUint((sequence-1), 10)
	}
	return msg
}

func getPercentOff(sequence uint64, previousSequence uint64, interval float64, timeDifference float64) uint64 {
	expectedTimeDifference := float64(sequence-previousSequence) * interval

	percentOff := 100 * (timeDifference - expectedTimeDifference) / expectedTimeDifference

	return uint64(math.Floor(math.Max(0, percentOff)))
}

func (s *SatelliteTrackers) getShortest3Cycles() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	shortest := 5 * time.Second //fallback
	first := true

	for _, Satellite := range s.satellites {
		if first || Satellite.expectedInterval < shortest {
			shortest = Satellite.expectedInterval
			first = false
		}
	}

	return shortest * 3
}

func (s *SatelliteTrackers) MonitorSatellites(ctx context.Context) {
	interval := s.getShortest3Cycles()
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			s.checkInactiveSatellites()
			timer.Reset(interval)

		case <-s.intervalChanged:
			interval = s.getShortest3Cycles()
			fmt.Println()

			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}

			timer.Reset(interval)

		case <-ctx.Done():
			timer.Stop()
			return
		}
	}
}

func (s *SatelliteTrackers) checkInactiveSatellites() {
	s.mu.Lock()

	var inactiveSatellites []string

	for id, sat := range s.satellites {
		if sat.expectedInterval.Seconds() == 0 {
			return
		}
		cyclesMissed := int(math.Round(time.Since(sat.lastReceivedAt).Seconds() / sat.expectedInterval.Seconds()))
		if cyclesMissed >= 3 {
			fmt.Printf(
				"Satellite %s is inactive and has missed %d cycles.\n",
				id,
				cyclesMissed,
			)
			fmt.Println()
			inactiveSatellites = append(inactiveSatellites, id)
		}
	}

	s.mu.Unlock()
	for _, id := range inactiveSatellites {
		s.diagnosticRequests <- id
	}
}
