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
	mux.Handle("GET /api/v1/update", handleUpdateEntries(logger, config, appState, entries))
	mux.Handle("GET /api/v1/", handleGetEntries(logger, entries))
	mux.Handle("GET /", handleRoot())
}
