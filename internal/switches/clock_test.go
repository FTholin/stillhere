package switches

import (
	"testing"
	"time"
)

func TestFakeClockAdvance(t *testing.T) {
	base := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	c := NewFakeClock(base)

	if !c.Now().Equal(base) {
		t.Fatalf("Now() = %v, want %v", c.Now(), base)
	}

	c.Advance(48 * time.Hour)

	want := base.Add(48 * time.Hour)
	if !c.Now().Equal(want) {
		t.Errorf("after Advance, Now() = %v, want %v", c.Now(), want)
	}
}
