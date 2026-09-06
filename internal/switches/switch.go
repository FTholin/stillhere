// Package switches contains the core domain of the project
//a dead man's switch, its lifecycle, and the rules that govern when it fires.

package switches

import (
	"errors"
	"time"
)

// State is the lifecycle stage of a Switch.
type State string

const (
	StateArmed     State = "armed"     // waiting for check-ins
	StateTriggered State = "triggered" // deadline passed, secret released
	StateDisarmed  State = "disarmed"  // switched off by its owner
)

var (
	ErrNotFound      = errors.New("switch not found")
	ErrNotArmed      = errors.New("switch is not armed")
	ErrEmptySecret   = errors.New("secret must not be empty")
	ErrIntervalRange = errors.New("interval must be between 1 minute and 365 days")
	ErrAlreadyExists = errors.New("switch already exists")
)

const (
	MinInterval = time.Minute
	MaxInterval = 365 * 24 * time.Hour
)

type Switch struct {
	ID           string
	Label        string
	Secret       []byte
	Recipient    string
	Interval     time.Duration
	LastCheckIn  time.Time
	State        State
	CheckInToken string
	RevealToken  string
}

// Deadline is the instant past which the switch fires
func (s *Switch) Deadline() time.Time {
	return s.LastCheckIn.Add(s.Interval)
}

// IsOverdue reports whether the switch should fire at the given instant.
func (s *Switch) IsOverdue(now time.Time) bool {
	return s.State == StateArmed && now.After(s.Deadline())
}

// CheckIn pushes the deadline back. It is a no-op error on a switch that
// has already fired: you cannot un-send a message.
func (s *Switch) CheckIn(now time.Time) error {
	if s.State != StateArmed {
		return ErrNotArmed
	}
	s.LastCheckIn = now
	return nil
}

// Disarm marks the switch as disarmed, cancelling its pending trigger.
// It returns ErrNotArmed if the switch is not currently armed.
func (s *Switch) Disarm() error {
	if s.State != StateArmed {
		return ErrNotArmed
	}
	s.State = StateDisarmed
	return nil
}
