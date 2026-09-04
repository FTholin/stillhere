package switches

import (
	"errors"
	"testing"
	"time"
)

func TestCheckIn(t *testing.T) {
	base := time.Date(2026, 8, 26, 14, 0, 0, 0, time.UTC)

	t.Run("armed switch records the new check-in", func(t *testing.T) {
		sw := &Switch{State: StateArmed, Interval: time.Hour, LastCheckIn: base}
		later := base.Add(30 * time.Minute)

		if err := sw.CheckIn(later); err != nil {
			t.Fatalf("CheckIn() = %v, want nil", err)
		}
		if !sw.LastCheckIn.Equal(later) {
			t.Errorf("LastCheckIn = %v, want %v", sw.LastCheckIn, later)
		}
		if !sw.Deadline().Equal(later.Add(time.Hour)) {
			t.Errorf("Deadline = %v, want %v", sw.Deadline(), later.Add(time.Hour))
		}
	})

	t.Run("triggered switch refuses the check-in", func(t *testing.T) {
		sw := &Switch{State: StateTriggered, Interval: time.Hour, LastCheckIn: base}

		err := sw.CheckIn(base.Add(time.Hour))

		if !errors.Is(err, ErrNotArmed) {
			t.Fatalf("CheckIn() = %v, want ErrNotArmed", err)
		}
		if !sw.LastCheckIn.Equal(base) {
			t.Errorf("LastCheckIn was modified, want it untouched")
		}
	})
}

func TestIsOverdue(t *testing.T) {
	base := time.Date(2026, 8, 26, 14, 0, 0, 0, time.UTC)

	cas := []struct {
		nom   string
		state State
		now   time.Time
		want  bool
	}{
		{"armed, before deadline", StateArmed, base.Add(59 * time.Minute), false},
		{"armed, after deadline", StateArmed, base.Add(61 * time.Minute), true},
		{"already triggered", StateTriggered, base.Add(48 * time.Hour), false},
		{"disarmed", StateDisarmed, base.Add(48 * time.Hour), false},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			sw := &Switch{State: c.state, Interval: time.Hour, LastCheckIn: base}

			if got := sw.IsOverdue(c.now); got != c.want {
				t.Errorf("IsOverdue() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestDisarm(t *testing.T) {
	base := time.Date(2026, 8, 26, 14, 0, 0, 0, time.UTC)

	cas := []struct {
		nom     string
		state   State
		wantErr error
		want    State
	}{
		{"armed", StateArmed, nil, StateDisarmed},
		{"triggered", StateTriggered, ErrNotArmed, StateTriggered},
		{"disarmed", StateDisarmed, ErrNotArmed, StateDisarmed},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			sw := &Switch{State: c.state, Interval: time.Hour, LastCheckIn: base}

			err := sw.Disarm()

			if !errors.Is(err, c.wantErr) {
				t.Fatalf("Disarm() = %v, want %v", err, c.wantErr)
			}

			if sw.State != c.want {
				t.Errorf("State = %q, want %q", sw.State, c.want)
			}
		})
	}

}

func TestSwitchFiresAfterInterval(t *testing.T) {
	clock := NewFakeClock(time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC))

	sw := &Switch{
		State:       StateArmed,
		Interval:    30 * 24 * time.Hour,
		LastCheckIn: clock.Now(),
	}

	clock.Advance(29 * 24 * time.Hour)
	if sw.IsOverdue(clock.Now()) {
		t.Errorf("fired after 29 days, want still armed")
	}

	clock.Advance(2 * 24 * time.Hour)

	if !sw.IsOverdue(clock.Now()) {
		t.Error("still armed after 31 days, want fired")
	}
}
