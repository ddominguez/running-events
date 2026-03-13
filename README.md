# Running Events

A static site generator that fetches running race activities from the Strava API and publishes them as a GitHub Pages site.

## Overview

This project automatically pulls my race activities from Strava and generates a static website displaying my race history. A GitHub Actions cron job runs periodically to fetch new activities and regenerate the site.

## Commands

### fetch

```bash
go run ./cmd/fetch
```

Fetches new race activities from Strava since the last sync. Updates `data/races.json` and `data/sync.json`.

### generate

```bash
go run ./cmd/generate
```

Generates the static HTML site from race data. Outputs to the `site/` directory.

## Environment Variables

| Variable | Type | Description |
|----------|------|-------------|
| `STRAVA_CLIENT_ID` | Secret | Strava app client ID |
| `STRAVA_CLIENT_SECRET` | Secret | Strava app client secret |
| `STRAVA_REFRESH_TOKEN` | Variable | Strava refresh token (auto-updated) |
| `GH_PAT` | Secret | GitHub PAT with repo and actions:write scopes |
| `MAPBOX_TOKEN` | Secret | Mapbox public token for static map images |

## Local Development

1. Set the required environment variables
2. Run `go mod download` to fetch dependencies
3. Run `go run ./cmd/fetch` to fetch races
4. Run `go run ./cmd/generate` to generate the site
5. Serve the `site/` directory locally to preview
