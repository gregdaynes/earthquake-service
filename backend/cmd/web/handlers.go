package main

import (
	"encoding/json"
	"fmt"
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

func handleGetEntries(logger *slog.Logger, config *Config, entries *models.EntryModel) http.Handler {
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
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			qCoords := r.URL.Query()["coords"][0]
			if len(qCoords) == 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotAcceptable)
				w.Write([]byte(""))
				return
			}

			coords := strings.Split(qCoords, ",")

			swlng, err := strconv.ParseFloat(coords[0], 32)
			if err != nil {
				logger.Error("")
			}
			swlat, err := strconv.ParseFloat(coords[1], 32)
			if err != nil {
				logger.Error("")
			}
			nelng, err := strconv.ParseFloat(coords[2], 32)
			if err != nil {
				logger.Error("")
			}
			nelat, err := strconv.ParseFloat(coords[3], 32)
			if err != nil {
				logger.Error("")
			}

			// query the db for the coordinates
			results := entries.QueryWithBounds(swlat, nelat, swlng, nelng)

			var data []Point

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
			}

			resp := Response{
				Message: "Recent points",
				Data:    data,
			}

			js, _ := json.Marshal(resp)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(js)
			fmt.Printf("%s took %v\n", "getEntries", time.Since(start))
		},
	)
}

func handleUpdateEntries(logger *slog.Logger, config *Config, appState *State, entryModel *models.EntryModel) http.Handler {
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
					logger.Error("issue storing item", "error", err)
					appState.updateFailure()
					return
				}

				// longitude
				longitude64, err := strconv.ParseFloat(latlong[1], 32)
				if err != nil {
					logger.Error("issue storing item", "error", err)
					appState.updateFailure()
					return
				}

				// elevation
				elevation64, err := strconv.ParseInt(elev, 10, 32)
				if err != nil {
					logger.Error("issue storing item", "error", err)
					appState.updateFailure()
					return
				}

				// magnitude
				re := regexp.MustCompile(`M([\d\.]+)`)
				match := re.FindStringSubmatch(entry.Title)
				magnitude, err := strconv.ParseFloat(match[1], 32)
				if err != nil {
					logger.Error("issue storing item", "error", err)
					appState.updateFailure()
					return
				}

				tc := firstN(item.Content, 20)
				t, err := time.Parse(time.RFC3339, tc)
				if err != nil {
					logger.Error("issue parsing time", "error", err)
					appState.updateFailure()
					return
				}

				entry.Latitude = float32(latitude64)
				entry.Longitude = float32(longitude64)
				entry.Elevation = int32(elevation64)
				entry.Magnitude = float32(magnitude)
				entry.Time = &t

				_, err = entryModel.Insert(entry)
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
