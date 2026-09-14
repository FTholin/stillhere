package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FTholin/stillhere/internal/switches"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}
}

func TestVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	Version(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}

	if body := rec.Body.String(); body != `{"version":"0.1.0"}` {
		t.Errorf("body = %q, want %q", body, `{"version":"0.1.0"}`)
	}
}

func TestGetCheckInPageDoesNotCheckIn(t *testing.T) {
	// This is the whole point of jalon 11: mail scanners follow links.
	// A GET must leave the switch exactly as it was.
	srv, _ := newTestServer(t)
	clock := srv.clock.(*switches.FakeClock)

	sw := &switches.Switch{
		ID: "abc", CheckInToken: "tok", State: switches.StateArmed,
		Interval: time.Hour, LastCheckIn: clock.Now(),
	}
	if err := srv.store.Create(context.Background(), sw); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	before := sw.LastCheckIn

	clock.Advance(30 * time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/checkin/tok", nil)
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	got, err := srv.store.Get(context.Background(), "abc")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !got.LastCheckIn.Equal(before) {
		t.Errorf("GET moved LastCheckIn to %v, want it untouched at %v", got.LastCheckIn, before)
	}
}
