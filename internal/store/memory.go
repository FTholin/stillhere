// Package store persists switches. The Store interface hides whether they
// live in memory or in PostgreSQL
package store

import (
	"context"
	"sync"
	"time"

	"github.com/FTholin/stillhere/internal/switches"
)

type Store interface {
	Create(ctx context.Context, s *switches.Switch) error
	Get(ctx context.Context, id string) (*switches.Switch, error)
	ByCheckInToken(ctx context.Context, token string) (*switches.Switch, error)
	ByRevealToken(ctx context.Context, token string) (*switches.Switch, error)
	Save(ctx context.Context, s *switches.Switch) error
	Overdue(ctx context.Context, now time.Time) ([]*switches.Switch, error)
}

type Memory struct {
	mu        sync.RWMutex
	data      map[string]*switches.Switch
	byCheckIn map[string]string // token → id
	byReveal  map[string]string // token → id
}

func NewMemory() *Memory {
	return &Memory{
		data:      make(map[string]*switches.Switch),
		byCheckIn: make(map[string]string),
		byReveal:  make(map[string]string),
	}
}

func (m *Memory) ByCheckInToken(ctx context.Context, token string) (*switches.Switch, error) {
	if token == "" {
		return nil, switches.ErrNotFound
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	id, ok := m.byCheckIn[token]
	if !ok {
		return nil, switches.ErrNotFound
	}
	sw, ok := m.data[id]
	if !ok {
		return nil, switches.ErrNotFound // index out of sync: treat as absent
	}

	cp := *sw
	return &cp, nil
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

func (m *Memory) Create(ctx context.Context, sw *switches.Switch) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[sw.ID]; exists {
		return switches.ErrAlreadyExists
	}

	cp := *sw
	m.data[sw.ID] = &cp
	m.byCheckIn[sw.CheckInToken] = sw.ID
	m.byReveal[sw.RevealToken] = sw.ID
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
