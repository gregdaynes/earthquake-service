package main

import (
	"context"
	"log/slog"
	"net/http"
)

func NewServer(
	ctx context.Context,
	logger *slog.Logger,
	config *Config,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		logger,
		config,
	)

	var handler http.Handler = mux
	// handler = logging.NewLoggingMiddleware(logger, handler)
	// handler = logging.NewGoogleTraceIDMiddleware(logger, handler)
	// handler = checkAuthHeaders(handler)
	return handler
}
