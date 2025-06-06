package main

import (
	"log/slog"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	config *Config,
	appState *State,
) {
	mux.Handle("/", handleRoot(logger))
	mux.Handle("/api/v1", handleGetData(logger, config, appState))
}
