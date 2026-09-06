// Package store persists switches. The Store interface hides whether they
// live in memory or in PostgreSQL
package store

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/FTholin/stillhere/internal/switches"
)

// ErrDuplicate is returned when creating a switch whose ID already exists.
var ErrDuplicate = errors.New("switch already exists")

// Package store persists switches. The Store interface hides whether they
// live in memory or in PostgreSQL.
type Store interface {
	Create(ctx context.Context, s *switches.Switch) error
	Get(ctx context.Context, id string) (*switches.Switch, error)
	ByCheckInToken(ctx context.Context, token string) (*switches.Switch, error)
	ByRevealToken(ctx context.Context, token string) (*switches.Switch, error)
	Save(ctx context.Context, s *switches.Switch) error
	Overdue(ctx context.Context, now time.Time) ([]*switches.Switch, error)
}

type Memory struct {
	mu   sync.RWMutex
	data map[string]*switches.Switch
}

func NewMemory() *Memory {
	return &Memory{data: make(map[string]*switches.Switch)}
}

func (m *Memory) Get(ctx context.Context, id string) (*switches.Switch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.data[id]
	if !ok {
		return nil, switches.ErrNotFound
	}

	cp := *s
	return &cp, nil
}

func (m *Memory) Create(ctx context.Context, s *switches.Switch) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[s.ID]; exists {
		return switches.ErrAlreadyExists
	}

	cp := *s
	m.data[s.ID] = &cp
	return nil
}

func (m *Memory) Overdue(ctx context.Context, now time.Time) ([]*switches.Switch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var due []*switches.Switch
	for _, s := range m.data {
		if s.IsOverdue(now) {
			cp := *s
			due = append(due, &cp)
		}
	}
	return due, nil
}

func (m *Memory) Save(ctx context.Context, s *switches.Switch) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[s.ID]; !exists {
		return switches.ErrNotFound
	}

	cp := *s
	m.data[s.ID] = &cp

	return nil
}

func (m *Memory) ByCheckInToken(ctx context.Context, token string) (*switches.Switch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.data {
		if s.CheckInToken == token {
			cp := *s
			return &cp, nil
		}
	}

	return nil, switches.ErrNotFound
}

func (m *Memory) ByRevealToken(ctx context.Context, token string) (*switches.Switch, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.data {
		if s.RevealToken == token {
			cp := *s
			return &cp, nil
		}
	}

	return nil, switches.ErrNotFound
}
