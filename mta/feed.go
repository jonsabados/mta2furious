package mta

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/jonsabados/mta2furious/mta/wire"
	"google.golang.org/protobuf/proto"
)

type Direction string

const (
	DirectionNorth Direction = "NORTH"
	DirectionSouth Direction = "SOUTH"
	DirectionEast  Direction = "EAST"
	DirectionWest  Direction = "WEST"
)

type TimeRange struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

type TripReplacementPeriod struct {
	RouteId string `json:"routeId"`
	// The start time is omitted, the end time is currently now + 30 minutes for
	// all routes of the A division
	ReplacementPeriod TimeRange `json:"replacementPeriod,omitempty"`
}

type FeedHeader struct {
	GtfsRealtimeVersion   string                  `json:"gtfsRealtimeVersion,omitempty"`
	Timestamp             *time.Time              `json:"timestamp,omitempty"`
	NyctSubwayVersion     string                  `json:"nyctSubwayVersion,omitempty"`
	TripReplacementPeriod []TripReplacementPeriod `json:"tripReplacementPeriod,omitempty"`
}

type StopTimeUpdate struct {
	// Must be the same as in stops.txt in the corresponding GTFS feed.
	StopID    string     `json:"stopID,omitempty"`
	Arrival   *time.Time `json:"arrival,omitempty"`
	Departure *time.Time `json:"departure,omitempty"`
	// Provides the planned station arrival track. The following is the Manhattan
	// track configurations:
	// 1: southbound local
	// 2: southbound express
	// 3: northbound express
	// 4: northbound local
	//
	// In the Bronx (except Dyre Ave line)
	// M: bi-directional express (in the AM express to Manhattan, in the PM
	// express away).
	//
	// The Dyre Ave line is configured:
	// 1: southbound
	// 2: northbound
	// 3: bi-directional
	ScheduledTrack *string `json:"scheduledTrack,omitempty"`
	// This is the actual track that the train is operating on and can be used to
	// determine if a train is operating according to its current schedule
	// (plan).
	//
	// The actual track is known only shortly before the train reaches a station,
	// typically not before it leaves the previous station. Therefore, the NYCT
	// feed sets this field only for the first station of the remaining trip.
	//
	// Different actual and scheduled track is the result of manually rerouting a
	// train off it scheduled path.  When this occurs, prediction data may become
	// unreliable since the train is no longer operating in accordance to its
	// schedule.  The rules engine for the 'countdown' clocks will remove this
	// train from all schedule stations.
	ActualTrack *string `json:"actualTrack,omitempty"`
	// IsComplete represents if this TripUpdate has completed (in practice this becomes True when an assigned record drops off the feed)
	IsComplete bool `json:"isComplete"`
}

type TripUpdate struct {
	TripId string `json:"tripId,omitempty"`
	// The route_id from the GTFS that this selector refers to.
	RouteId string `json:"routeId,omitempty"`
	TrainId string `json:"trainId,omitempty"`
	// This trip has been assigned to a physical train. If true, this trip is
	// already underway or most likely will depart shortly.
	//
	// Train Assignment is a function of the Automatic Train Supervision (ATS)
	// office system used by NYCT Rail Operations to monitor and track train
	// movements. ATS provides the ability to "assign" the nyct_train_id
	// attribute when a physical train is at its origin terminal. These assigned
	// trips have the is_assigned field set in the TripDescriptor.
	//
	// When a train is at a terminal but has not been given a work program it is
	// declared unassigned and is tagged as such. Unassigned trains can be moved
	// to a storage location or assigned a nyct_train_id when a determination for
	// service is made.
	IsAssigned bool `json:"isAssigned,omitempty"`
	// Uptown and Bronx-bound trains are moving NORTH.
	// Times Square Shuttle to Grand Central is also northbound.
	//
	// Downtown and Brooklyn-bound trains are moving SOUTH.
	// Times Square Shuttle to Times Square is also southbound.
	//
	// EAST and WEST are not used currently.
	Direction *Direction `json:"direction,omitempty"`

	StopTimeUpdate []StopTimeUpdate `json:"stopTimeUpdate,omitempty"`
}

type TripStatus struct {
	Header      FeedHeader
	TripUpdates []TripUpdate `json:"tripUpdates"`
}

type LiveFeed struct {
	endpoint string
	apiKey   string
}

func NewLiveFeed(endpoint string, apikey string) *LiveFeed {
	return &LiveFeed{
		endpoint: endpoint,
		apiKey:   apikey,
	}
}

func (f *LiveFeed) Feed(ctx context.Context) (TripStatus, error) {
	// A, C, E lines
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.endpoint, nil)
	if err != nil {
		return TripStatus{}, err
	}
	req.Header.Add("x-api-key", f.apiKey)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return TripStatus{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return TripStatus{}, err
	}
	var raw wire.FeedMessage
	err = proto.Unmarshal(body, &raw)
	if err != nil {
		return TripStatus{}, err
	}
	return TripStatus{
		Header:      convertHeader(raw.Header),
		TripUpdates: convertEntities(raw.Entity),
	}, nil
}

func convertHeader(header *wire.FeedHeader) FeedHeader {
	extension := proto.GetExtension(header, wire.E_NyctFeedHeader).(*wire.NyctFeedHeader)
	replacementPeriod := make([]TripReplacementPeriod, len(extension.TripReplacementPeriod))
	for i, p := range extension.TripReplacementPeriod {
		replacementPeriod[i] = TripReplacementPeriod{
			RouteId: *p.RouteId,
			ReplacementPeriod: TimeRange{
				Start: convertTime(p.ReplacementPeriod.Start),
				End:   convertTime(p.ReplacementPeriod.End),
			},
		}
	}
	return FeedHeader{
		GtfsRealtimeVersion:   *header.GtfsRealtimeVersion,
		Timestamp:             convertTime(header.Timestamp),
		NyctSubwayVersion:     *extension.NyctSubwayVersion,
		TripReplacementPeriod: replacementPeriod,
	}
}

func convertEntities(raw []*wire.FeedEntity) []TripUpdate {
	ret := make([]TripUpdate, 0, len(raw))
	for _, r := range raw {
		// we get records for vehicle and trip updates, but we really only care about trip updates
		if r.TripUpdate == nil {
			continue
		}
		ret = append(ret, convertEntity(r.TripUpdate))
	}
	return ret
}

func convertEntity(raw *wire.TripUpdate) TripUpdate {
	nytTrip := proto.GetExtension(raw.Trip, wire.E_NyctTripDescriptor).(*wire.NyctTripDescriptor)
	isAssigned := false
	if nytTrip.IsAssigned != nil {
		isAssigned = *nytTrip.IsAssigned
	}
	return TripUpdate{
		TripId:         *raw.Trip.TripId,
		RouteId:        *raw.Trip.RouteId,
		TrainId:        *nytTrip.TrainId,
		IsAssigned:     isAssigned,
		Direction:      convertDirection(nytTrip.Direction),
		StopTimeUpdate: convertStopTimeUpdates(raw.StopTimeUpdate),
	}
}

func convertDirection(raw *wire.NyctTripDescriptor_Direction) *Direction {
	if raw == nil {
		return nil
	}
	ret := Direction(raw.String())
	return &ret
}

func convertStopTimeUpdates(raw []*wire.TripUpdate_StopTimeUpdate) []StopTimeUpdate {
	ret := make([]StopTimeUpdate, len(raw))
	for i, r := range raw {
		nytUpdate := proto.GetExtension(r, wire.E_NyctStopTimeUpdate).(*wire.NyctStopTimeUpdate)
		ret[i] = StopTimeUpdate{
			StopID:         *r.StopId,
			Arrival:        extractTime(r.Arrival),
			Departure:      extractTime(r.Departure),
			ScheduledTrack: nytUpdate.ScheduledTrack,
			ActualTrack:    nytUpdate.ActualTrack,
		}
	}
	return ret
}

func convertTime(t *uint64) *time.Time {
	if t == nil {
		return nil
	}
	ret := time.Unix(int64(*t), 0)
	return &ret
}

func extractTime(raw *wire.TripUpdate_StopTimeEvent) *time.Time {
	if raw == nil {
		return nil
	}
	ret := time.Unix(*raw.Time, 0)
	return &ret
}
