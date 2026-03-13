package main

import (
	"log/slog"
	"os"

	"github.com/ddominguez/running-events/internal/html"
	"github.com/ddominguez/running-events/internal/store"
)

func main() {
	token := os.Getenv("MAPBOX_TOKEN")
	if token == "" {
		slog.Error("MAPBOX_TOKEN environment variable is required")
		os.Exit(1)
	}

	races, err := store.LoadRaces("data/races.json")
	if err != nil {
		slog.Error("loading races", "error", err)
		os.Exit(1)
	}

	if err := html.GenerateSite(races, token); err != nil {
		slog.Error("generating site", "error", err)
		os.Exit(1)
	}

	slog.Info("site generated successfully")
}
