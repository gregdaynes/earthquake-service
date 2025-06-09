package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/earthquake-service/internal/models"
	"github.com/mmcdole/gofeed"
)

func handleRoot() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(""))
		},
	)
}

func handleGetEntries(logger *slog.Logger, entries *models.EntryModel) http.Handler {
	type Point struct {
		GUID       string  `json:"id"`
		Title      string  `json:"title"`
		Content    string  `json:"content"`
		Categories string  `json:"categoires"`
		Elevation  int32   `json:"elevation"`
		Time       string  `json:"time"`
		Latitude   float32 `json:"latitude"`
		Longitude  float32 `json:"longitude"`
		Magnitude  float32 `json:"magnitude"`
	}

	type Response struct {
		Message string  `json:"message"`
		Data    []Point `json:"data"`
		Count   int     `json:"count"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			qCoords := r.URL.Query()["coords"][0]
			if len(qCoords) == 0 {
				logger.Info("no coordinates provided")
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}

			coords := strings.Split(qCoords, ",")

			swlng, err := strconv.ParseFloat(coords[0], 32)
			if err != nil {
				logger.Error("SW Longitude is invalid", "error", err)
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}
			swlat, err := strconv.ParseFloat(coords[1], 32)
			if err != nil {
				logger.Error("SW Latitude is invalid", "error", err)
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}
			nelng, err := strconv.ParseFloat(coords[2], 32)
			if err != nil {
				logger.Error("NE Longitude is invalid", "error", err)
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}
			nelat, err := strconv.ParseFloat(coords[3], 32)
			if err != nil {
				logger.Error("NE latitude is invalid", "error", err)
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}

			// query the db for the coordinates
			results := entries.QueryWithBounds(swlat, nelat, swlng, nelng)

			var data []Point
			var count int
			for _, point := range results {
				data = append(data, Point{
					GUID:       point.GUID,
					Title:      point.Title,
					Content:    point.Content,
					Categories: point.Categories,
					Time:       string(point.Time.Format(time.RFC3339)),
					Elevation:  point.Elevation,
					Latitude:   point.Latitude,
					Longitude:  point.Longitude,
					Magnitude:  point.Magnitude,
				})
				count = count + 1
			}

			resp := Response{
				Message: "Recent points",
				Data:    data,
				Count:   count,
			}

			js, _ := json.Marshal(resp)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(js)

			logger.Info("GetEntries",
				"time_ms", time.Since(start),
				"sq_lat", swlat,
				"ne_lat", nelat,
				"sw_lng", swlng,
				"ne_lng", nelng,
				"count", count)
		},
	)
}

func handleUpdateEntries(logger *slog.Logger, config *Config, appState *State, entryModel *models.EntryModel) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
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

			var count int
			for _, item := range feed.Items {
				// this is where we want to make a new struct that satisfies the entry model
				entry := models.Entry{
					GUID:       item.GUID,
					Title:      item.Title,
					Content:    item.Content,
					Categories: strings.Join(item.Categories[:], ", "),
					Published:  item.PublishedParsed,
					Updated:    item.UpdatedParsed,
				}

				ext := item.Extensions["georss"]
				elev := ext["elev"][0].Value
				point := ext["point"][0].Value

				latlong := strings.Split(point, " ")

				// latitude
				latitude64, err := strconv.ParseFloat(latlong[0], 32)
				if err != nil {
					logger.Error("parsing latitude", "error", err)
					continue
				}
				entry.Latitude = float32(latitude64)

				// longitude
				longitude64, err := strconv.ParseFloat(latlong[1], 32)
				if err != nil {
					logger.Error("parsing longitude", "error", err)
					continue
				}
				entry.Longitude = float32(longitude64)

				// elevation
				elevation64, err := strconv.ParseInt(elev, 10, 32)
				if err != nil {
					logger.Error("parsing elevation", "error", err)
					continue
				}
				entry.Elevation = int32(elevation64)

				// magnitude
				re := regexp.MustCompile(`M([\d\.]+)`)
				match := re.FindStringSubmatch(entry.Title)
				magnitude, err := strconv.ParseFloat(match[1], 32)
				if err != nil {
					logger.Error("parsing magnitude", "error", err)
					continue
				}
				entry.Magnitude = float32(magnitude)

				tc := firstN(item.Content, 20)
				t, err := time.Parse(time.RFC3339, tc)
				if err != nil {
					logger.Error("parsing event time", "error", err)
					continue
				}
				entry.Time = &t

				_, err = entryModel.Insert(entry)
				if err != nil {
					logger.Error("issue storing item", "error", err)
					appState.updateFailure()
					return
				}
				count = count + 1
			}

			w.WriteHeader(http.StatusOK)
			appState.updateSuccess()

			logger.Info("Update Entries",
				"time_ms", time.Since(start),
				"count", count)
		},
	)
}

// time of event
func firstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}
