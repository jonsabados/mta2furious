package mta

import (
	"context"
	"sync"

	"github.com/rs/zerolog"
)

type MemoryStore struct {
	mutex      sync.RWMutex
	worldState []TripUpdate
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (m *MemoryStore) PriorState(_ context.Context) ([]TripUpdate, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.worldState, nil
}

func (m *MemoryStore) RecordState(ctx context.Context, state []TripUpdate) error {
	zerolog.Ctx(ctx).Debug().Int("trips", len(state)).Msg("persisting state in memory")
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.worldState = state
	return nil
}
