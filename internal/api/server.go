package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/FTholin/stillhere/internal/store"
	"github.com/FTholin/stillhere/internal/switches"
)

// Server carries everything the handlers need. Nothing is global.
// each test builds its own Server with its own store and its own clock.
type Server struct {
	store store.Store
	clock switches.Clock
	log   *slog.Logger
}

func New(st store.Store, clk switches.Clock, log *slog.Logger) *Server {
	return &Server{store: st, clock: clk, log: log}
}

func (srv *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10) // 64 Kib
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req createRequest
	if err := dec.Decode(&req); err != nil {
		writeDecodeError(w, err)
		return
	}

	interval, err := time.ParseDuration(req.Interval)
	if err != nil {
		writeError(w, `interval must be a duration like "72h" or "30m"`, http.StatusBadRequest)
		return
	}

	sw, err := switches.New(req.Label, req.Recipient, []byte(req.Secret), interval, srv.clock.Now())
	if err != nil {
		srv.writeDomainError(w, err)
		return
	}

	if err := identify(sw); err != nil {
		srv.writeDomainError(w, err)
		return
	}

	if err := srv.store.Create(r.Context(), sw); err != nil {
		srv.writeDomainError(w, err)
		return
	}

	w.Header().Set("Location", "/switches/"+sw.ID)
	writeJSON(w, http.StatusCreated, createResponse{
		ID:           sw.ID,
		CheckInToken: sw.CheckInToken,
		RevealToken:  sw.RevealToken,
		Deadline:     sw.Deadline(),
	})
}

// Routes returns the fully wired handler for the whole API
func (srv *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", Health)  // free function: needs nothing
	mux.HandleFunc("GET /version", Version) // free function: needs nothing
	mux.HandleFunc("POST /switches", srv.handleCreate)
	mux.HandleFunc("GET /switches/{id}", srv.handleGet)
	mux.HandleFunc("GET /checkin/{token}", srv.handleCheckInPage)
	mux.HandleFunc("POST /checkin/{token}", srv.handleCheckIn)
	return mux
}

func (srv *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	sw, err := srv.store.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		srv.writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, switchResponse{
		ID:          sw.ID,
		Label:       sw.Label,
		State:       string(sw.State),
		Interval:    sw.Interval.String(),
		LastCheckIn: sw.LastCheckIn,
		Deadline:    sw.Deadline(),
	})
}
