package mta

import (
	"context"

	"github.com/rs/zerolog"
)

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
	dropCount := 0
	for _, f := range t.feeds {
		status, err := f.Feed(ctx)
		if err != nil {
			return nil, err
		}
		for _, tu := range status.TripUpdates {
			zerolog.Ctx(ctx).Trace().Interface("trip", tu).Msg("trip observed")
			if tu.IsAssigned {
				ret = append(ret, tu)
			} else {
				dropCount++
			}
		}
	}
	if dropCount > 0 {
		zerolog.Ctx(ctx).Info().Int("dropCount", dropCount).Msg("filtered out unassigned trips")
	}
	return ret, nil
}
