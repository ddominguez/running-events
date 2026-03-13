package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

type Race struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	StartDate          string  `json:"start_date"`
	Distance           int     `json:"distance"`
	MovingTime         int     `json:"moving_time"`
	ElapsedTime        int     `json:"elapsed_time"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
	Type               string  `json:"type"`
	WorkoutType        int     `json:"workout_type"`
	SummaryPolyline    string  `json:"summary_polyline"`
}

type SyncState struct {
	LastSyncedAt      int64 `json:"last_synced_at"`
	LastActivityEpoch int64 `json:"last_activity_epoch"`
}

func LoadRaces(path string) ([]Race, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Race{}, nil
		}
		return nil, fmt.Errorf("reading races file: %w", err)
	}

	if len(data) == 0 {
		return []Race{}, nil
	}

	var races []Race
	if err := json.Unmarshal(data, &races); err != nil {
		return nil, fmt.Errorf("decoding races: %w", err)
	}

	return races, nil
}

func SaveRaces(path string, races []Race) error {
	slices.SortFunc(races, func(a, b Race) int {
		if a.StartDate > b.StartDate {
			return -1
		}
		if a.StartDate < b.StartDate {
			return 1
		}
		return 0
	})

	data, err := json.MarshalIndent(races, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding races: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(path), "races-*.json")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	_, err = tmpFile.Write(data)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmpFile.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

func LoadSync(path string) (SyncState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return SyncState{}, nil
		}
		return SyncState{}, fmt.Errorf("reading sync file: %w", err)
	}

	var state SyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return SyncState{}, fmt.Errorf("decoding sync: %w", err)
	}

	return state, nil
}

func SaveSync(path string, state SyncState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding sync: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(path), "sync-*.json")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	_, err = tmpFile.Write(data)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmpFile.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}
