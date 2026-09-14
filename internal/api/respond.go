package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/FTholin/stillhere/internal/switches"
)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err) // The status is gone: log only
	}
}

func writeError(w http.ResponseWriter, msg string, code int) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// writeDomainError is the customs desk: one place where domain errors
// become HTTP status code. Handlers never choose a code themselves.
func (srv *Server) writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, switches.ErrNotFound):
		writeError(w, "switch not found", http.StatusNotFound)

	case errors.Is(err, switches.ErrNotArmed):
		writeError(w, "switch is not armed", http.StatusConflict)

	case errors.Is(err, switches.ErrEmptySecret):
		writeError(w, "secret must not be empty", http.StatusBadRequest)

	case errors.Is(err, switches.ErrIntervalRange):
		writeError(w, "interval must be between 1 minute and 365 days", http.StatusBadRequest)
	default:
		srv.log.Error("unhandled error", "err", err)
		writeError(w, "internal error", http.StatusInternalServerError)
	}
}

func writeDecodeError(w http.ResponseWriter, err error) {
	var (
		syntaxErr *json.SyntaxError
		typeErr   *json.UnmarshalTypeError
		maxErr    *http.MaxBytesError
	)

	switch {
	case errors.Is(err, io.EOF):
		writeError(w, "body must not be empty", http.StatusBadRequest)
	case errors.As(err, &maxErr):
		writeError(w, "body too large", http.StatusRequestEntityTooLarge)
	case errors.As(err, &syntaxErr):
		writeError(w, fmt.Sprintf("invalid json at byte %d", syntaxErr.Offset), http.StatusBadRequest)
	case errors.As(err, &typeErr):
		writeError(w, fmt.Sprintf("field %q has the wrong type", typeErr.Field), http.StatusBadRequest)
	case strings.HasPrefix(err.Error(), "json: unknown field"):
		writeError(w, strings.TrimPrefix(err.Error(), "json:"), http.StatusBadRequest)

	default:
		writeError(w, "invalid json", http.StatusBadRequest)
	}
}
