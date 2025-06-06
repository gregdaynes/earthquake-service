package main

import (
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/earthquake-service/internal/models"
	"github.com/mmcdole/gofeed"
)

func handleRoot(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger.Info("testing", "msg", "handleSomething")

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Testing"))
		},
	)
}

func handleGetData(logger *slog.Logger, config *Config, appState *State) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			currentWindowStart := time.Now().Add(time.Duration(-5) * time.Minute)
			if currentWindowStart.Before(appState.LastRun) {
				w.WriteHeader(http.StatusTooEarly)
				return
			}

			fp := gofeed.Parser{}
			feed, err := fp.ParseURL(config.AtomFeed)
			if err != nil {
				logger.Error("error fetching feed", "error", err)
				w.WriteHeader(http.StatusOK)
				appState.updateFailure()
				return
			}

			for _, item := range feed.Items {
				// this is where we want to make a new struct that satisfies the entry model
				entry := models.Entry{
					GUID:       item.GUID,
					Title:      item.Title,
					Content:    item.Content,
					Updated:    item.UpdatedParsed,
					Published:  item.PublishedParsed,
					Categories: strings.Join(item.Categories[:], ", "),
				}

				ext := item.Extensions["georss"]
				elev := ext["elev"][0].Value
				point := ext["point"][0].Value

				latlong := strings.Split(point, " ")
				latitude64, err := strconv.ParseFloat(latlong[0], 32)
				if err != nil {
					logger.Error("issue storing item", "error", err)
					appState.updateFailure()
					return
				}
				longitude64, err := strconv.ParseFloat(latlong[1], 32)
				if err != nil {
					logger.Error("issue storing item", "error", err)
					appState.updateFailure()
					return
				}
				elevation64, err := strconv.ParseInt(elev, 10, 32)
				if err != nil {
					logger.Error("issue storing item", "error", err)
					appState.updateFailure()
					return
				}

				re := regexp.MustCompile(`M([\d\.]+)`)
				match := re.FindStringSubmatch(entry.Title)
				magnitude, err := strconv.ParseFloat(match[1], 32)
				if err != nil {
					logger.Error("issue storing item", "error", err)
					appState.updateFailure()
					return
				}

				entry.Latitude = float32(latitude64)
				entry.Longitude = float32(longitude64)
				entry.Elevation = int32(elevation64)
				entry.Magnitude = float32(magnitude)

				_, err = models.Insert(entry)
				if err != nil {
					logger.Error("issue storing item", "error", err)
					appState.updateFailure()
					return
				}
			}

			w.WriteHeader(http.StatusOK)
			appState.updateSuccess()
		},
	)
}
