package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FTholin/stillhere/internal/store"
	"github.com/FTholin/stillhere/internal/switches"
)

func TestCreateSwitch(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{"valid", `{"label": "laptop", "secret": "pin1234", "recipient":"lea@example.com", "interval":"72h"}`, http.StatusCreated},
		{"empty secret", `{"label":"x", "secret":"","recipient":"lea@example.com", "interval":"72h"}`, http.StatusBadRequest},
		{"interval too short", `{"secret":"s", "recipient":"r", "interval":"10s"}`, http.StatusBadRequest},
		{"broken json", `{"secret":}`, http.StatusBadRequest},
		{"empty body", ``, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)
			req := httptest.NewRequest(http.MethodPost, "/switches", strings.NewReader(tt.body))

			rec := httptest.NewRecorder()

			srv.Routes().ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body.String())
			}

			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("content-type=%q, want application/json", ct)
			}

		})
	}
}

func newTestServer(t *testing.T) (*Server, *switches.FakeClock) {
	t.Helper()
	clock := switches.NewFakeClock(time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC))
	return New(store.NewMemory(), clock, slog.New(slog.DiscardHandler)), clock
}

func TestCreateNeverLeaksTheSecret(t *testing.T) {
	const secret = "the pin is 1234"

	srv, _ := newTestServer(t)
	body := fmt.Sprintf(`{"label":"laptop", "secret":%q, "recipient": "lea@eample.com", "interval":"72h"}`, secret)

	req := httptest.NewRequest(http.MethodPost, "/switches", strings.NewReader(body))

	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body:%s)", rec.Code, rec.Body.String())
	}

	if strings.Contains(rec.Body.String(), secret) {
		t.Error("the create response contains the plaintext secret")
	}

	var got createResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// Same check on the public view.
	req = httptest.NewRequest(http.MethodGet, "/switches/"+got.ID, nil)
	rec = httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)

	body2 := rec.Body.String()
	if strings.Contains(body2, secret) || strings.Contains(body2, got.RevealToken) {
		t.Error("the public view leaks the secret or the reveal token")
	}

}

func TestAllRoutesAreWired(t *testing.T) {
	srv, clock := newTestServer(t)

	// Seed one switch so that GET /switches/{id} can succeed: a 404 from
	// the store would be indistinguishable from a 404 from the router.
	sw := &switches.Switch{
		ID: "seeded", CheckInToken: "seeded-token", State: switches.StateArmed,
		Interval: time.Hour, LastCheckIn: clock.Now(),
	}
	if err := srv.store.Create(context.Background(), sw); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/healthz"},
		{http.MethodGet, "/version"},
		{http.MethodPost, "/switches"},
		{http.MethodGet, "/switches/seeded"},
		{http.MethodGet, "/checkin/seeded-token"},
		{http.MethodPost, "/checkin/seeded-token"},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			req := httptest.NewRequest(rt.method, rt.path, nil)
			rec := httptest.NewRecorder()

			srv.Routes().ServeHTTP(rec, req)

			// We do not care what it answers, only that something is wired.

			if rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed {
				t.Errorf("status = %d: no handler registered for this pattern", rec.Code)
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodDelete, "/switches", nil)
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}

	if allow := rec.Header().Get("Allow"); !strings.Contains(allow, http.MethodPost) {
		t.Errorf("Allow = %q, want it to contain POST", allow)
	}
}

func TestGetSwitch(t *testing.T) {
	srv, clock := newTestServer(t)

	sw := &switches.Switch{
		ID: "abc", State: switches.StateArmed, Interval: time.Hour, LastCheckIn: clock.Now(),
	}

	if err := srv.store.Create(context.Background(), sw); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/switches/abc", nil)
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}

	var got switchResponse

	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.State != "armed" {
		t.Errorf("state = %q, want armed", got.State)
	}
}
