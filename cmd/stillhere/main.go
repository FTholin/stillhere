package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/FTholin/stillhere/internal/api"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", api.Health)
	mux.HandleFunc("GET /version", api.Version)

	addr := ":" + port()
	logger.Info("listening", "addr", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
