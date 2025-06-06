package main

import (
	"log/slog"
	"net/http"
)

func handleRoot(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// use thing to handle request
			logger.Info("testing", "msg", "handleSomething")

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Testing"))
		},
	)
}
