package mta

import "context"

type Feed interface {
	Feed(ctx context.Context) (TripStatus, error)
}

type TransitSystem struct {
	feeds []Feed
}

func NewTransitSystem(feeds ...Feed) *TransitSystem {
	return &TransitSystem{
		feeds: feeds,
	}
}

func (t *TransitSystem) CurrentState(ctx context.Context) ([]TripUpdate, error) {
	ret := make([]TripUpdate, 0)
	for _, f := range t.feeds {
		status, err := f.Feed(ctx)
		if err != nil {
			return nil, err
		}
		ret = append(ret, status.TripUpdates...)
	}
	return ret, nil
}
