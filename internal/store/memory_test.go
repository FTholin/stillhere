package store

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/FTholin/stillhere/internal/switches"
)

func newSwitch(t *testing.T, id string, lastCheckIn time.Time) *switches.Switch {
	t.Helper()
	return &switches.Switch{
		ID:           id,
		Label:        "switch " + id,
		Secret:       []byte("secret-" + id),
		Recipient:    id + "@example.test",
		Interval:     24 * time.Hour,
		LastCheckIn:  lastCheckIn,
		State:        switches.StateArmed,
		CheckInToken: "checkin-" + id,
		RevealToken:  "reveal-" + id,
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		nom     string
		id      string
		wantErr error
	}{
		{nom: "new switch", id: "sw-new", wantErr: nil},
		{nom: "duplicate rejected", id: "sw-existing", wantErr: switches.ErrAlreadyExists},
	}

	for _, tt := range tests {
		t.Run(tt.nom, func(t *testing.T) {
			m := NewMemory()
			if err := m.Create(ctx, newSwitch(t, "sw-existing", time.Now())); err != nil {
				t.Fatalf("setup: %v", err)
			}

			err := m.Create(ctx, newSwitch(t, tt.id, time.Now()))
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Create() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()

	if err := m.Create(ctx, newSwitch(t, "sw-1", time.Now())); err != nil {
		t.Fatalf("préparation : %v", err)
	}

	t.Run("existing switch", func(t *testing.T) {
		got, err := m.Get(ctx, "sw-1")
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if got.ID != "sw-1" {
			t.Errorf("ID = %q, want %q", got.ID, "sw-1")
		}
	})

	t.Run("missing switch", func(t *testing.T) {
		_, err := m.Get(ctx, "unknown")
		if !errors.Is(err, switches.ErrNotFound) {
			t.Errorf("Get() error = %v, want %v", err, switches.ErrNotFound)
		}
	})
}

func TestGetReturnsACopy(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	if err := m.Create(ctx, newSwitch(t, "sw-1", time.Now())); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tampered, err := m.Get(ctx, "sw-1")
	if err != nil {
		t.Fatalf("Get(): %v", err)
	}
	tampered.State = switches.StateDisarmed

	reloaded, err := m.Get(ctx, "sw-1")
	if err != nil {
		t.Fatalf("Get(): %v", err)
	}
	if reloaded.State == switches.StateDisarmed {
		t.Error("Get() leaked its internal pointer: the store was mutated from outside")
	}
}

func TestCreateCopiesOnWrite(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()

	original := newSwitch(t, "sw-1", time.Now())
	if err := m.Create(ctx, original); err != nil {
		t.Fatalf("Create(): %v", err)
	}
	original.State = switches.StateDisarmed

	reloaded, err := m.Get(ctx, "sw-1")
	if err != nil {
		t.Fatalf("Get(): %v", err)
	}

	if reloaded.State == switches.StateDisarmed {
		t.Errorf("Create() stored the caller's pointer instead of a copy")
	}
}

func TestSave(t *testing.T) {
	ctx := context.Background()
	t.Run("existing switch", func(t *testing.T) {
		m := NewMemory()
		if err := m.Create(ctx, newSwitch(t, "sw-1", time.Now())); err != nil {
			t.Fatalf("setup: %v", err)
		}

		updated := newSwitch(t, "sw-1", time.Now())
		updated.Label = "new label"

		if err := m.Save(ctx, updated); err != nil {
			t.Fatalf("Save(): %v", err)
		}

		reloaded, err := m.Get(ctx, "sw-1")
		if err != nil {
			t.Fatalf("Get(): %v", err)
		}

		if reloaded.Label != "new label" {
			t.Errorf("Label = %q, want %q", reloaded.Label, "new label")
		}
	})

	t.Run("missing switch", func(t *testing.T) {
		m := NewMemory()
		err := m.Save(ctx, newSwitch(t, "ghost", time.Now()))
		if !errors.Is(err, switches.ErrNotFound) {
			t.Errorf("Save() error = %v, want %v", err, switches.ErrNotFound)
		}
	})
}

func TestByTokens(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	if err := m.Create(ctx, newSwitch(t, "sw-1", time.Now())); err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Run("check-in token found", func(t *testing.T) {
		got, err := m.ByCheckInToken(ctx, "checkin-sw-1")
		if err != nil {
			t.Fatalf("ByCheckInToken(): %v", err)
		}

		if got.ID != "sw-1" {
			t.Errorf("ID = %q, want %q", got.ID, "sw-1")
		}
	})

	t.Run("check-in token unknown", func(t *testing.T) {
		_, err := m.ByCheckInToken(ctx, "bogus-token")
		if !errors.Is(err, switches.ErrNotFound) {
			t.Errorf("ByCheckInToken() error = %v, want %v", err, switches.ErrNotFound)
		}
	})

	t.Run("reveal token found", func(t *testing.T) {
		got, err := m.ByRevealToken(ctx, "reveal-sw-1")
		if err != nil {
			t.Fatalf("ByRevealToken(): %v", err)
		}
		if got.ID != "sw-1" {
			t.Errorf("ID = %q, want %q", got.ID, "sw-1")
		}
	})

	t.Run("token namespaces do not overlap", func(t *testing.T) {
		if _, err := m.ByCheckInToken(ctx, "reveal-sw-1"); !errors.Is(err, switches.ErrNotFound) {
			t.Error("a reveal token was accepted as a check-in token")
		}

		if _, err := m.ByRevealToken(ctx, "checkin-sw-1"); !errors.Is(err, switches.ErrNotFound) {
			t.Error("a check-in token was accepted as a reveal token")
		}
	})
}

func TestOverdue(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("filters overdue switches", func(t *testing.T) {
		m := NewMemory()
		for _, s := range []*switches.Switch{
			newSwitch(t, "late", now.Add(-48*time.Hour)),
			newSwitch(t, "on-time", now.Add(-1*time.Hour)),
		} {
			if err := m.Create(ctx, s); err != nil {
				t.Fatalf("setup: %v", err)
			}
		}

		due, err := m.Overdue(ctx, now)
		if err != nil {
			t.Fatalf("Overdue(): %v", err)
		}

		if len(due) != 1 {
			t.Fatalf("len(due) = %d, want 1", len(due))
		}
		if due[0].ID != "late" {
			t.Errorf("ID = %q, want %q", due[0].ID, "late")
		}
	})

	t.Run("empty store", func(t *testing.T) {
		m := NewMemory()
		due, err := m.Overdue(ctx, now)
		if err != nil {
			t.Fatalf("Overdue(): %v", err)
		}

		if len(due) != 0 {
			t.Errorf("len(due) = %d, want 0", len(due))
		}
	})

	t.Run("a disarmed switch is never overdue", func(t *testing.T) {
		m := NewMemory()
		s := newSwitch(t, "disarmed", now.Add(-48*time.Hour))
		s.State = switches.StateDisarmed
		if err := m.Create(ctx, s); err != nil {
			t.Fatalf("setup: %v", err)
		}

		due, err := m.Overdue(ctx, now)
		if err != nil {
			t.Fatalf("Overdue(): %v", err)
		}
		if len(due) != 0 {
			t.Errorf("len(due) = %d, want 0: a disarmed switch was reported as overdue ", len(due))
		}
	})
}

func TestMemoryConcurrent(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()

	if err := m.Create(ctx, newSwitch(t, "sw-1", time.Now())); err != nil {
		t.Fatalf("setup: %v", err)
	}

	var wg sync.WaitGroup

	for i := range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			switch i % 4 {
			case 0:
				_, _ = m.Get(ctx, "sw-1")
			case 1:
				_, _ = m.Overdue(ctx, time.Now())
			case 2:
				_, _ = m.ByCheckInToken(ctx, "checkin-sw-1")

			case 3:
				if s, err := m.Get(ctx, "sw-1"); err == nil {
					_ = m.Save(ctx, s)
				}
			}
		}()
	}
	wg.Wait()
}
