package main

import (
	"log/slog"
	"os"
)

func NewLogger(config *Config) *slog.Logger {
	slogHandlerOptions := slog.HandlerOptions{}
	slogHandlerOptions.Level = slog.LevelInfo

	if config.Debug {
		slogHandlerOptions.AddSource = true
		slogHandlerOptions.Level = slog.LevelDebug
	}

	// Use the slog.New() function to initialize a new structured logger, which
	// writes to the standard out stream and uses the default settings.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slogHandlerOptions))

	return logger
}
