package main

import (
	"log/slog"
	"net/http"

	"github.com/earthquake-service/internal/models"
)

func addRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	config *Config,
	appState *State,
	entries *models.EntryModel,
) {
	mux.Handle("/", handleRoot(logger))
	mux.Handle("/api/v1", handleGetData(logger, config, appState, entries))
}
