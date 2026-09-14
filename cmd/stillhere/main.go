package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/FTholin/stillhere/internal/api"
	"github.com/FTholin/stillhere/internal/store"
	"github.com/FTholin/stillhere/internal/switches"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	srv := api.New(store.NewMemory(), switches.SystemClock{}, logger)

	s := &http.Server{
		Addr:              ":" + port(),
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("listening", "addr", s.Addr)
	if err := s.ListenAndServe(); err != nil {
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
