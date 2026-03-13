package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ddominguez/running-events/internal/github"
	"github.com/ddominguez/running-events/internal/store"
	"github.com/ddominguez/running-events/internal/strava"
)

func main() {
	defer os.Exit(0)

	syncState, err := store.LoadSync("data/sync.json")
	if err != nil {
		slog.Error("loading sync state", "error", err)
		return
	}

	accessToken, newRefreshToken, _, err := strava.RefreshToken()
	if err != nil {
		slog.Error("refreshing token", "error", err)
		return
	}
	slog.Info("token refreshed")

	if err := github.UpdateActionsVariable("STRAVA_REFRESH_TOKEN", newRefreshToken); err != nil {
		slog.Error("updating GitHub variable", "error", err)
		return
	}
	slog.Info("GitHub variable updated")

	client := strava.NewClient(accessToken)
	var allActivities []strava.Activity
	after := syncState.LastActivityEpoch

	for {
		endpoint := fmt.Sprintf("/athlete/activities?after=%d&per_page=100", after)
		resp, err := client.Get(endpoint)
		if err != nil {
			slog.Error("fetching activities", "error", err)
			return
		}

		var activities []strava.Activity
		if err := json.Unmarshal(resp, &activities); err != nil {
			slog.Error("decoding activities", "error", err)
			return
		}

		allActivities = append(allActivities, activities...)
		if len(activities) < 100 {
			break
		}

		lastTime, _ := time.Parse(time.RFC3339, activities[len(activities)-1].StartDate)
		after = lastTime.Unix()
	}

	var newRaces []store.Race
	var maxEpoch int64

	for _, a := range allActivities {
		t, err := time.Parse(time.RFC3339, a.StartDate)
		if err != nil {
			continue
		}
		epoch := t.Unix()
		if epoch > maxEpoch {
			maxEpoch = epoch
		}

		if a.WorkoutType == 1 && a.SportType == "Run" {
			newRaces = append(newRaces, store.Race{
				ID:                 a.ID,
				Name:               a.Name,
				StartDate:          a.StartDate,
				Distance:           a.Distance,
				MovingTime:         a.MovingTime,
				ElapsedTime:        a.ElapsedTime,
				TotalElevationGain: a.TotalElevationGain,
				Type:               a.Type,
				WorkoutType:        a.WorkoutType,
				SummaryPolyline:    a.Map.SummaryPolyline,
			})
		}
	}

	existing, err := store.LoadRaces("data/races.json")
	if err != nil {
		slog.Error("loading races", "error", err)
		return
	}

	existingIDs := make(map[int64]bool)
	for _, r := range existing {
		existingIDs[r.ID] = true
	}

	newCount := 0
	for _, r := range newRaces {
		if !existingIDs[r.ID] {
			existing = append(existing, r)
			newCount++
		}
	}

	if err := store.SaveRaces("data/races.json", existing); err != nil {
		slog.Error("saving races", "error", err)
		return
	}

	if err := store.SaveSync("data/sync.json", store.SyncState{
		LastSyncedAt:      time.Now().Unix(),
		LastActivityEpoch: maxEpoch,
	}); err != nil {
		slog.Error("saving sync state", "error", err)
		return
	}

	slog.Info("fetch complete", "new_races", newCount)
}
