package main

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/earthquake-service/internal/models"
)

type State struct {
	mu            sync.Mutex
	LastRun       time.Time
	LastCompleted time.Time
	LastFailed    time.Time
}

func (s *State) updateSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastCompleted = time.Now()
	s.LastRun = time.Now()
}

func (s *State) updateFailure() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastFailed = time.Now()
	s.LastRun = time.Now()
}

func NewServer(
	ctx context.Context,
	logger *slog.Logger,
	config *Config,
	db *DB,
) http.Handler {
	mux := http.NewServeMux()

	appState := &State{}

	entries := &models.EntryModel{DB: db.Connection}

	addRoutes(
		mux,
		logger,
		config,
		appState,
		entries,
	)

	var handler http.Handler = mux
	// handler = logging.NewLoggingMiddleware(logger, handler)
	// handler = logging.NewGoogleTraceIDMiddleware(logger, handler)
	// handler = checkAuthHeaders(handler)
	return handler
}
