# Strava Race Site — Agent Context

## Project Overview

A static site generator that fetches running race activities from the Strava API and publishes them as a GitHub Pages site. There is no database. All state lives in committed JSON files. A GitHub Actions cron job runs periodically to fetch new activities and regenerate the site.

## Tech Stack

- **Language**: Go
- **Hosting**: GitHub Pages (project site at `https://{username}.github.io/{repo-name}/`)
- **CI/CD**: GitHub Actions
- **External APIs**: Strava v3 REST API, Mapbox Static Images API
- **State**: JSON files committed to the repo

## Repository Structure

```
/
├── cmd/
│   ├── fetch/        # fetches new race activities from Strava
│   └── generate/     # generates HTML and CSS from race data
├── internal/
│   ├── strava/       # Strava API client and token refresh logic
│   ├── github/       # GitHub Actions variable updater
│   ├── store/        # read/write JSON data files
│   └── html/         # HTML template rendering and slug logic
├── data/
│   ├── races.json    # all saved race activities
│   └── sync.json     # last sync metadata (epoch timestamp)
├── site/             # generated output (committed, served by GitHub Pages)
│   ├── assets/
│   │   └── style.css
│   ├── index.html
│   └── {year}/
│       └── {slug}.html
├── .github/
│   └── workflows/
│       ├── sync.yml  # cron job workflow
│       └── pages.yml # GitHub Pages deployment
├── AGENTS.md         # this file
└── README.md
```

## Data Files

### data/races.json
Array of race activity objects. Each object is a subset of the Strava activity response:
```json
[
  {
    "id": 12345678,
    "name": "New York City Marathon",
    "start_date": "2024-11-03T10:00:00Z",
    "distance": 42195,
    "moving_time": 12600,
    "elapsed_time": 12700,
    "total_elevation_gain": 245.3,
    "type": "Run",
    "workout_type": 1,
    "summary_polyline": "encoded_polyline_string_here"
  }
]
```

Note: `summary_polyline` comes from `map.summary_polyline` in the Strava API response.

### data/sync.json
Tracks the last successful sync so the fetch command uses the `after` query param:
```json
{
  "last_synced_at": 1713175200,
  "last_activity_epoch": 1713175200
}
```

## Strava API

- Base URL: `https://www.strava.com/api/v3`
- Auth: OAuth2 bearer token
- Access tokens expire after 6 hours — always refresh before fetching
- Endpoint: `GET /athlete/activities?after={epoch}&per_page=100`
- A race activity has `workout_type == 1`
- The route polyline is at `map.summary_polyline` in each activity response

### Token Refresh Flow
```
POST https://www.strava.com/oauth/token
  grant_type=refresh_token
  client_id={STRAVA_CLIENT_ID}
  client_secret={STRAVA_CLIENT_SECRET}
  refresh_token={STRAVA_REFRESH_TOKEN}

Response:
  access_token  — use this for API calls
  refresh_token — save this, it rotates on each refresh
  expires_at    — unix epoch
```

## Mapbox Static Images API

Used to render a map of each race route on the detail page. The image URL is embedded
directly in the generated HTML — the visitor's browser fetches it from Mapbox at page
load time. The token is scoped in the Mapbox dashboard to only allow requests from the
GitHub Pages domain, so embedding it in HTML is acceptable.

### URL format
```
https://api.mapbox.com/styles/v1/mapbox/dark-v11/static/path-5+fc4c02-0.8({url_encoded_polyline})/auto/800x400?access_token={MAPBOX_TOKEN}
```

- `path-5+fc4c02-0.8` — 5px wide path, color #fc4c02 (Strava orange), 80% opacity
- `auto` — Mapbox automatically fits the map bounds to the polyline
- `800x400` — image dimensions in pixels
- The polyline string must be URL-encoded using Go's `net/url.QueryEscape`

### When MAPBOX_TOKEN is needed
Only at generate time — it is baked into the `<img src>` URL in the HTML output.
It is never fetched server-side or stored in the repo. Only needed in the GitHub
Actions environment and locally when running `generate`.

## Environment Variables / GitHub Actions Secrets and Variables

| Name | Type | Description |
|---|---|---|
| `STRAVA_CLIENT_ID` | Secret | Strava app client ID |
| `STRAVA_CLIENT_SECRET` | Secret | Strava app client secret |
| `STRAVA_REFRESH_TOKEN` | Variable | Current refresh token — stored as a Variable (not Secret) so it can be updated programmatically via the GitHub API after each token rotation |
| `GH_PAT` | Secret | Personal access token with `repo` and `actions:write` scopes, used to update the STRAVA_REFRESH_TOKEN variable |
| `MAPBOX_TOKEN` | Secret | Mapbox public token scoped to GitHub Pages domain, embedded in generated HTML |

## Commands

### fetch
```
go run ./cmd/fetch
```
- Reads `data/sync.json` for the last activity epoch
- Refreshes Strava access token using env vars
- Updates `STRAVA_REFRESH_TOKEN` GitHub Actions Variable via GitHub API with the new refresh token
- Calls `GET /athlete/activities?after={epoch}` paginating while len(results) == 100
- Filters for `workout_type == 1` (races only)
- Saves `map.summary_polyline` from each activity as `summary_polyline` in races.json
- Appends new races to `data/races.json`
- Updates `data/sync.json` with the latest epoch seen across all fetched activities
- Exits 0 whether or not new races were found

### generate
```
go run ./cmd/generate
```
- Reads `data/races.json`
- Reads `MAPBOX_TOKEN` from environment
- Generates a slug for each race from the activity name: lowercase, spaces and
  punctuation replaced with hyphens, consecutive hyphens collapsed
- Writes `site/assets/style.css` — the single shared stylesheet for all pages
- Renders `site/index.html`
- For each race, renders `site/{year}/{slug}.html`
- Does not make any external HTTP requests

## Generated Site Structure

```
site/
├── assets/
│   └── style.css
├── index.html
├── 2025/
│   ├── new-york-city-marathon.html
│   └── some-other-half-marathon.html
└── 2024/
    ├── brooklyn-half-marathon.html
    └── queens-10k.html
```

## HTML Pages and CSS

### Linking to style.css
The site is hosted at a subpath (`/{repo-name}/`) on GitHub Pages so all internal
links must be relative, never absolute (absolute paths starting with `/` would
resolve to the personal site root, not this project).

- From `site/index.html`: `<link rel="stylesheet" href="assets/style.css">`
- From `site/{year}/{slug}.html`: `<link rel="stylesheet" href="../assets/style.css">`

Same rule applies to the back link on detail pages — use `../` not `/`.

### site/assets/style.css
- Single shared stylesheet for all pages, written by the generate command
- Loads a monospaced Google Font via `@import` at the top of the file
- No CSS frameworks — vanilla CSS only
- CSS variables for colors and typography defined on `:root`

### site/index.html
- Links to `assets/style.css`
- Lists all races grouped by year, year descending, races descending within each year
- Year section headers only rendered for years that have at least one race
- Each race is a link to `./{year}/{slug}.html`
- Shows per race: name, date formatted as "November 3, 2024", distance in km and miles
- If no races exist: render a graceful "no races yet" message

### site/{year}/{slug}.html
- Links to `../assets/style.css`
- Full race stats: name, date, distance (km + miles), moving time as H:MM:SS,
  pace per km and per mile, elevation gain in meters and feet
- Mapbox static image: `<img src="{mapbox_url}">` — only rendered if `summary_polyline`
  is non-empty
- Back link to `../` using relative path

### Design Aesthetic
- Dark background, light text
- Monospaced font for all stats (via Google Fonts, loaded in style.css)
- No JavaScript
- Clean, editorial, runner-focused — think a race results board. Stark. Confident.
- Consistent visual language across index and detail pages

## Key Constraints

- No database of any kind
- No JavaScript in generated HTML
- The `generate` command must be fully reproducible — same input always produces same output
- All secrets and variables injected via environment variables, never hardcoded
- The `site/` directory is committed to the repo and served directly by GitHub Pages
- All internal links and asset references must use relative paths (never absolute)
- Keep dependencies minimal — prefer stdlib where possible
- Templates and the CSS content are embedded as Go string literals, not files on disk
