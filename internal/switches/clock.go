package switches

import "time"

// Clock reports the current time. Production code uses SystemClock;
// test use a FakeClock they can move at will.
type Clock interface {
	Now() time.Time
}

// SystemClock is the real wall clock
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now() }

// FakeClock is a controllable Clock for tests
type FakeClock struct {
	current time.Time
}

func NewFakeClock(t time.Time) *FakeClock {
	return &FakeClock{current: t}
}

func (f *FakeClock) Now() time.Time { return f.current }

// Advance moves the clock forward. Thirty days pass in a nanosecond.
func (f *FakeClock) Advance(d time.Duration) { f.current = f.current.Add(d) }
