package mta

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type StateOracle interface {
	CurrentState(ctx context.Context) ([]TripUpdate, error)
}

type StateStore interface {
	PriorState(ctx context.Context) ([]TripUpdate, error)
	RecordState(ctx context.Context, state []TripUpdate) error
}

type Segment struct {
	FromStation    string
	ToStation      string
	DepartAt       time.Time
	ArriveAt       time.Time
	TripID         string
	RouteID        string
	TrainID        string
	IsAssigned     bool
	ScheduledTrack *string
	ActualTrack    *string
}

type StateUpdateResults struct {
	CompletedSegments []Segment
}

type StateProcessor struct {
	oracle StateOracle
	store  StateStore
}

func NewStateProcessor(oracle StateOracle, store StateStore) *StateProcessor {
	return &StateProcessor{
		oracle: oracle,
		store:  store,
	}
}

func (p *StateProcessor) ProcessUpdates(ctx context.Context) (StateUpdateResults, error) {
	zerolog.Ctx(ctx).Debug().Msg("processing updates")
	priorState, err := p.store.PriorState(ctx)
	if err != nil {
		return StateUpdateResults{}, err
	}
	currentState, err := p.oracle.CurrentState(ctx)
	if err != nil {
		return StateUpdateResults{}, err
	}
	newState := make([]TripUpdate, 0)
	completedSegments := make([]Segment, 0)

	// first update the state of things
	for _, prior := range priorState {
		newVersion, segmentsDone := p.processTrip(ctx, prior, currentState)
		if len(segmentsDone) > 0 {
			completedSegments = append(completedSegments, segmentsDone...)
		}
		if newVersion != nil {
			newState = append(newState, *newVersion)
		}
	}

	// add any trips we haven't seen before to our new durable state
	for _, current := range currentState {
		if locateTrip(current.TripId, priorState) == nil {
			zerolog.Ctx(ctx).Info().Interface("trip", current).Msg("new trip found")
			newState = append(newState, current)
		}
	}

	err = p.store.RecordState(ctx, newState)
	if err != nil {
		return StateUpdateResults{}, err
	}

	return StateUpdateResults{
		CompletedSegments: completedSegments,
	}, nil
}

// processTrip looks for updates to the trip, and returns the new version and completed segments. If the trip has been completed entirely nil is returned
func (p *StateProcessor) processTrip(ctx context.Context, trip TripUpdate, currentState []TripUpdate) (*TripUpdate, []Segment) {
	var rawUpdates []StopTimeUpdate
	rawState := locateTrip(trip.TripId, currentState)

	if rawState == nil {
		// sometimes trips don't appear in some pulls of the feed and then re-appear, if there are some completed segments we should retain it, otherwise drop
		hasCompletedStops := false
		hasCompletedStopsInWindow := false
		for _, update := range trip.StopTimeUpdate {
			if update.IsComplete {
				hasCompletedStops = true
				if !hasCompletedStopsInWindow {
					hasCompletedStopsInWindow = isInWindow(update.Arrival) || isInWindow(update.Departure)
				}
			}
		}
		if hasCompletedStops && hasCompletedStopsInWindow {
			if len(trip.StopTimeUpdate) == 2 && trip.StopTimeUpdate[0].IsComplete {
				zerolog.Ctx(ctx).Debug().Str("tripID", trip.TripId).Msg("trip fell off the radar with a single stop remaining, assuming completion")
				// note - were intentionally not returning, rawUpdates will be left nil which will cause the remaining stop to be picked up as completed later
			} else {
				zerolog.Ctx(ctx).Warn().Interface("trip", trip).Msg("trip with completed stops fell off the radar")
				return &trip, nil
			}
		} else {
			zerolog.Ctx(ctx).Warn().Interface("trip", trip).Bool("hasCompletedStops", hasCompletedStops).Msg("trip feel off the radar and discarding")
			return nil, nil
		}
	} else {
		rawUpdates = rawState.StopTimeUpdate
	}

	updates := make([]StopTimeUpdate, 0)
	for _, stop := range trip.StopTimeUpdate {
		// if the stop is complete already add it to updates for later segment completion check
		if stop.IsComplete {
			updates = append(updates, stop)
			continue
		}
		newVersion := locateStop(stop.StopID, rawUpdates)
		// if the new version is gone then the stop is complete (mta signals completion by dropping it...)
		if newVersion == nil {
			stop.IsComplete = true
			updates = append(updates, stop)
			continue
		}
		// finally drop the updated version in place
		updates = append(updates, *newVersion)
	}

	// next lets find completed segments in our updates, and build the new list of pending items
	stillPending := make([]StopTimeUpdate, 0)
	completed := make([]Segment, 0)
	for i := 0; i < len(updates); i++ {
		leg := updates[i]
		if !leg.IsComplete {
			stillPending = append(stillPending, leg)
			continue
		}
		// if we are complete and were the tail item were done
		if i == len(updates)-1 {
			break
		}
		nextLeg := updates[i+1]
		if !nextLeg.IsComplete {
			// if our next leg isn't complete we need to retain the completed leg
			stillPending = append(stillPending, leg)
		} else {
			completed = append(completed, Segment{
				FromStation:    leg.StopID,
				ToStation:      nextLeg.StopID,
				DepartAt:       *leg.Departure,
				ArriveAt:       *nextLeg.Arrival,
				TripID:         trip.TripId,
				RouteID:        trip.RouteId,
				TrainID:        trip.TrainId,
				IsAssigned:     trip.IsAssigned,
				ScheduledTrack: leg.ScheduledTrack,
				ActualTrack:    leg.ActualTrack,
			})
		}
	}

	if len(stillPending) > 0 {
		return &TripUpdate{
			TripId:         rawState.TripId,
			RouteId:        rawState.RouteId,
			TrainId:        rawState.TrainId,
			IsAssigned:     rawState.IsAssigned,
			Direction:      rawState.Direction,
			StopTimeUpdate: stillPending,
		}, nil
	}
	zerolog.Ctx(ctx).Info().Str("tripID", trip.TripId).Msg("trip complete")
	return nil, completed
}

func locateStop(stopID string, updates []StopTimeUpdate) *StopTimeUpdate {
	var ret *StopTimeUpdate
	for _, u := range updates {
		if u.StopID == stopID {
			return &u
		}
	}
	return ret
}

func locateTrip(tripID string, allTrips []TripUpdate) *TripUpdate {
	var ret *TripUpdate
	for _, t := range allTrips {
		if t.TripId == tripID {
			return &t
		}
	}
	return ret
}

func isInWindow(t *time.Time) bool {
	return t != nil && t.After(time.Now().Add(time.Minute*-30))
}
