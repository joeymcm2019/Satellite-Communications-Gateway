package server

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	satellitev1 "github.com/satellite-communications/gateway/gen/satellite/v1"
	"github.com/satellite-communications/gateway/internal/events"
	"github.com/satellite-communications/gateway/internal/model"
)

type SatelliteState struct {
	name             string
	batteryPercent   float64
	lastReceivedAt   time.Time
	previousSequence uint64
	expectedInterval time.Duration
	charging         bool
}

type SatelliteTrackers struct {
	mu                 sync.Mutex
	satellites         map[string]SatelliteState
	intervalChanged    chan struct{}
	diagnosticRequests chan string
	broadcaster        *events.Broadcaster
}

func NewSatelliteTracker(broadcaster *events.Broadcaster) *SatelliteTrackers {
	return &SatelliteTrackers{
		satellites:         make(map[string]SatelliteState),
		intervalChanged:    make(chan struct{}, 1),
		diagnosticRequests: make(chan string, 100),
		broadcaster:        broadcaster,
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
	fmt.Println("Charging:", req.GetCharging())
	fmt.Println()
}

type BackOnline struct {
	backOnline bool
	missedMsg  string
}

func (s *SatelliteTrackers) RecordTelemetry(
	req *satellitev1.SendTelemetryRequest,
) {

	success, online := s.recordTelemetryState(req)

	if !success {
		return
	}

	satellite := req.GetSatellite()

	satelliteID := satellite.GetSatelliteId()

	if online.backOnline {
		s.broadcaster.Publish(model.GatewayEvent{
			Type:        model.EventOnline,
			SatelliteID: satelliteID,
			Timestamp:   time.Now(),
			Data: map[string]any{
				"msg": online.missedMsg,
			},
		})
	}

	s.broadcaster.Publish(model.GatewayEvent{
		Type:        model.EventTelemetry,
		SatelliteID: satelliteID,
		Timestamp:   time.Now(),
		Data: map[string]any{
			"name":            satellite.GetName(),
			"battery_percent": req.GetBatteryPercent(),
			"sequence_number": req.GetSequenceNumber(),
		},
	})

}

func (s *SatelliteTrackers) recordTelemetryState(
	req *satellitev1.SendTelemetryRequest,
) (bool, BackOnline) {
	s.mu.Lock()
	defer s.mu.Unlock()

	satellites := s.satellites
	satellite := req.GetSatellite()
	satelliteID := satellite.GetSatelliteId()
	sat, exists := satellites[satelliteID]
	currentSequence := req.GetSequenceNumber()
	name := satellite.GetName()
	batteryPercent := req.GetBatteryPercent()
	charging := req.GetCharging()
	backOnline := false
	missedMsg := ""

	online := BackOnline{
		backOnline: backOnline,
		missedMsg:  missedMsg,
	}

	if exists {
		previousSequence := sat.previousSequence
		deltaSequence := currentSequence - previousSequence

		if deltaSequence > 1 && currentSequence > previousSequence { //missed telemetry(s)
			missedMsg = getMissedTelemetryMsg(currentSequence, previousSequence)
			online.missedMsg = missedMsg
			online.backOnline = true

			fmt.Println(
				satellite.GetName()+":",
				missedMsg,
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
		sat.batteryPercent = batteryPercent
		sat.charging = charging
		satellites[satelliteID] = sat
	} else {
		duration := req.GetExpectedIntervalSeconds()

		if duration == nil {
			fmt.Println("Error. Expected Interval is nil.")
			return false, online
		}

		interval := duration.AsDuration()

		if interval <= 0 {
			fmt.Println("Error. Expected interval must be positive.")
			return false, online
		}

		satellites[satelliteID] = SatelliteState{
			name:             name,
			batteryPercent:   batteryPercent,
			lastReceivedAt:   time.Now(),
			previousSequence: currentSequence,
			expectedInterval: duration.AsDuration(),
			charging:         charging,
		}

		select {
		case s.intervalChanged <- struct{}{}: //new satellite has arrived
		default: //prevents telemetry handler from blocking if a notification is already waiting
		}
	}

	return true, online
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
	type InactiveSatellite struct {
		id            string
		name          string
		cycles_missed int
	}

	var inactiveSatellites []InactiveSatellite

	s.mu.Lock()

	for id, sat := range s.satellites {
		if sat.expectedInterval.Seconds() <= 0 {
			continue
		}
		cycles_missed := int(math.Floor(time.Since(sat.lastReceivedAt).Seconds() / sat.expectedInterval.Seconds()))
		if cycles_missed >= 3 {
			inactiveSatellites = append(inactiveSatellites, InactiveSatellite{
				id:            id,
				name:          sat.name,
				cycles_missed: cycles_missed,
			})
		}
	}

	s.mu.Unlock()

	for _, satellite := range inactiveSatellites {
		fmt.Printf(
			"Satellite %s is inactive and has missed %d cycles.\n",
			satellite.id,
			satellite.cycles_missed,
		)
		fmt.Println()

		s.diagnosticRequests <- satellite.id

		s.broadcaster.Publish(model.GatewayEvent{
			Type:        model.EventInactive,
			SatelliteID: satellite.id,
			Timestamp:   time.Now(),
			Data: map[string]any{
				"name":          satellite.name,
				"cycles_missed": satellite.cycles_missed,
			},
		})
	}
}

func (s *SatelliteTrackers) GetSnapshot() []model.SatelliteSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshots := make([]model.SatelliteSnapshot, 0, len(s.satellites))

	for id, sat := range s.satellites {
		status := "online"

		if sat.expectedInterval > 0 &&
			time.Since(sat.lastReceivedAt) >= sat.expectedInterval+time.Second/2 {
			status = "inactive"
		}

		snapshots = append(snapshots, model.SatelliteSnapshot{
			ID:                      id,
			Name:                    sat.name,
			Status:                  status,
			BatteryPercent:          sat.batteryPercent,
			SequenceNumber:          sat.previousSequence,
			LastSeen:                sat.lastReceivedAt,
			ExpectedIntervalSeconds: sat.expectedInterval.Seconds(),
			Charging:                sat.charging,
		})
	}

	return snapshots
}
