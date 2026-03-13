# OpenCode Build Plan — Strava Race Site

A sequenced set of prompts for building the project with OpenCode.
Use **Plan Mode (Tab)** before every prompt to review what the agent intends
to do before it touches any files.

---

## Before You Start

1. Create a new GitHub repo (public)
2. Clone it locally and `cd` into it
3. Place `AGENTS.md` at the root
4. Open OpenCode in that directory

---

## Phase 1 — Project Scaffolding

**Goal**: Initialize the Go module, directory structure, and placeholder files.

```
Read AGENTS.md carefully. Then scaffold the project:

- Initialize a Go module named github.com/ddominguez/running-events
- Create the directory structure from AGENTS.md exactly as specified
- Create empty placeholder files:
    data/races.json  (empty JSON array: [])
    data/sync.json   ({"last_synced_at":0,"last_activity_epoch":0})
    site/assets/.gitkeep
- Create a basic README.md describing the project
- Do not write any Go code yet
```

**Check**: `go mod init` ran, all directories exist, placeholder files are valid JSON.

---

## Phase 2 — Strava Client and Token Refresh

**Goal**: Implement the Strava API client with token refresh logic.

```
Read AGENTS.md. Implement internal/strava/ with two things:

1. A Client struct that holds an access token and makes authenticated GET requests
   to the Strava v3 API base URL https://www.strava.com/api/v3

2. A RefreshToken function that:
   - POSTs to https://www.strava.com/oauth/token with grant_type=refresh_token
   - Reads STRAVA_CLIENT_ID, STRAVA_CLIENT_SECRET, STRAVA_REFRESH_TOKEN from env
   - Returns the new access_token, refresh_token, and expires_at
   - Returns a clear error if any env var is missing

Use only Go stdlib (net/http, encoding/json). No third-party HTTP libraries.
Write a unit test for RefreshToken that mocks the HTTP response.
```

**Check**: `go test ./internal/strava/...` passes.

---

## Phase 3 — JSON Store

**Goal**: Implement read/write helpers for the two data files.

```
Read AGENTS.md. Implement internal/store/ with:

1. A Race struct matching the races.json schema in AGENTS.md, including
   the summary_polyline field
2. LoadRaces(path string) ([]Race, error) — reads data/races.json
3. SaveRaces(path string, races []Race) error — writes data/races.json,
   sorted by start_date descending
4. LoadSync(path string) (SyncState, error) — reads data/sync.json
5. SaveSync(path string, state SyncState) error — writes data/sync.json

SyncState has fields: LastSyncedAt int64, LastActivityEpoch int64

All writes must be atomic: write to a temp file then os.Rename.
Write unit tests for all functions using temp directories.
```

**Check**: `go test ./internal/store/...` passes.

---

## Phase 4 — GitHub Variable Updater

**Goal**: Implement the function that writes the rotated refresh token back to GitHub Actions.

```
Read AGENTS.md. Create internal/github/ with a function UpdateActionsVariable that:

- Takes variableName string and value string
- Reads GITHUB_REPOSITORY (format: owner/repo) and GH_PAT from env
- Makes a PATCH request to:
  https://api.github.com/repos/{owner}/{repo}/actions/variables/{variableName}
  with body: {"name": variableName, "value": value}
  and header: Authorization: Bearer {GH_PAT}
- Returns an error if the request fails or returns a non-2xx status

Use only Go stdlib. Write a unit test that mocks the GitHub API response.
```

**Check**: `go test ./...` still passes.

---

## Phase 5 — Fetch Command

**Goal**: Implement the main fetch binary.

```
Read AGENTS.md. Implement cmd/fetch/main.go that:

1. Loads data/sync.json to get LastActivityEpoch
2. Calls strava.RefreshToken() to get a fresh access token
3. Calls github.UpdateActionsVariable("STRAVA_REFRESH_TOKEN", newRefreshToken)
   to persist the rotated token
4. Calls GET /athlete/activities?after={LastActivityEpoch}&per_page=100,
   paginating while len(results) == 100
5. Filters activities where workout_type == 1 and sport_type == "Run" (running races only)
6. Maps each matching activity to a store.Race, pulling summary_polyline
   from map.summary_polyline in the API response
7. Loads existing data/races.json, deduplicates by activity ID, appends new races
8. Saves updated data/races.json
9. Updates data/sync.json: LastActivityEpoch = max start_date epoch seen across
   ALL fetched activities (not just races), LastSyncedAt = time.Now().Unix()
10. Logs how many new races were found using log/slog
11. Exits 0 in all cases

Use log/slog for all logging.
```

**Check**: `go build ./cmd/fetch` compiles cleanly. Review in Plan Mode before implementing.

---

## Phase 6 — Slug Helper

**Goal**: Implement and test the slug generation function in isolation.

```
Read AGENTS.md. Add a Slug(name string) string function to internal/html/ that:

- Lowercases the input
- Replaces all spaces and punctuation with hyphens
- Collapses consecutive hyphens into one
- Trims leading and trailing hyphens

Write a table-driven unit test covering these cases at minimum:
  "New York City Marathon"     → "new-york-city-marathon"
  "Brooklyn Half-Marathon"     → "brooklyn-half-marathon"
  "Queens 10K"                 → "queens-10k"
  "Some Race (2024)"           → "some-race-2024"
  "  Weird  Spacing  "         → "weird-spacing"
```

**Check**: `go test ./internal/html/...` passes.

---

## Phase 7 — Generate Command

**Goal**: Implement the generate binary that produces the full site directory tree.

```
Read AGENTS.md. Implement cmd/generate/main.go and internal/html/ that:

1. Loads data/races.json
2. Reads MAPBOX_TOKEN from environment (fatal if missing)
3. Groups races by year from start_date, sorted year descending,
   races within each year sorted by start_date descending

4. Writes site/assets/style.css — the single shared stylesheet:
   - @import a monospaced Google Font at the top
   - CSS variables on :root for colors and typography
   - Styles for both index and detail pages
   - Dark background, light text
   - No CSS frameworks, vanilla CSS only

5. Renders site/index.html:
   - <link rel="stylesheet" href="assets/style.css"> (relative path)
   - Year section headers, only for years with at least one race
   - Each race links to ./{year}/{slug}.html (relative path)
   - Per race: name, date as "November 3, 2024", distance in km and miles
   - If no races: graceful "no races yet" message

6. For each race, creates site/{year}/ if needed, renders site/{year}/{slug}.html:
   - <link rel="stylesheet" href="../assets/style.css"> (relative path, one level up)
   - Full stats: name, date, distance (km + miles), moving time as H:MM:SS,
     pace per km and per mile, elevation gain in meters and feet
   - Mapbox <img src="{mapbox_url}"> only if summary_polyline is non-empty,
     polyline must be url.QueryEscape'd
   - Back link to ../ (relative path)

7. All templates and the CSS content are Go string literals embedded in source,
   not files read from disk at runtime
8. No external HTTP requests

Helper functions to implement and test in internal/html/:
  FormatDistance(meters float64) string       → "42.20 km (26.22 mi)"
  FormatDuration(seconds int) string          → "3:30:45"
  FormatPace(meters float64, seconds int) string → "4:59 /km (8:02 /mi)"
  FormatElevation(meters float64) string      → "245 m (804 ft)"
  FormatDate(t time.Time) string              → "November 3, 2024"
  MapboxURL(polyline, token string) string

IMPORTANT: The site is hosted at a subpath on GitHub Pages. All internal links
and asset references must use relative paths. Never use paths starting with /.
```

**Check**: `go run ./cmd/generate` produces `site/assets/style.css`, `site/index.html`,
and at least one detail page if races.json is non-empty. Open both page types in a
browser and verify layout, stats, and CSS are correct. Run `go test ./internal/html/...`.

---

## Phase 8 — GitHub Actions Workflow

**Goal**: Automate fetch + generate + commit on a schedule.

```
Read AGENTS.md. Create .github/workflows/sync.yml that:

- Triggers on:
    schedule: cron '0 8 * * 1'  (every Monday 8am UTC)
    workflow_dispatch            (manual trigger)
- Runs on: ubuntu-latest
- Steps:
  1. actions/checkout@v4 with fetch-depth: 0
  2. actions/setup-go@v5 with Go version matching go.mod
  3. go build ./... to verify compilation
  4. go run ./cmd/fetch with env:
       STRAVA_CLIENT_ID:      ${{ secrets.STRAVA_CLIENT_ID }}
       STRAVA_CLIENT_SECRET:  ${{ secrets.STRAVA_CLIENT_SECRET }}
       STRAVA_REFRESH_TOKEN:  ${{ vars.STRAVA_REFRESH_TOKEN }}
       GH_PAT:                ${{ secrets.GH_PAT }}
       GITHUB_REPOSITORY:     ${{ github.repository }}
  5. go run ./cmd/generate with env:
       MAPBOX_TOKEN:          ${{ secrets.MAPBOX_TOKEN }}
  6. git diff --quiet data/ site/ || (
       git config user.email "github-actions[bot]@users.noreply.github.com"
       git config user.name "github-actions[bot]"
       git add data/ site/
       git commit -m "chore: sync races [skip ci]"
       git push
     )

Add a comment in the YAML explaining why STRAVA_REFRESH_TOKEN is a Variable
and not a Secret.
```

**Check**: Review the YAML in Plan Mode before creating the file. Paste it into
https://rhysd.github.io/actionlint/ to validate before committing.

---

## Phase 9 — GitHub Pages Deployment Workflow

**Goal**: Deploy the site/ directory to GitHub Pages on every push to main.

```
Read AGENTS.md. Create .github/workflows/pages.yml that:

- Triggers on:
    push to main with paths: ['site/**']
    workflow_dispatch
- Sets permissions: contents: read, pages: write, id-token: write
- Uses concurrency to cancel in-progress deployments:
    group: pages
    cancel-in-progress: true
- Jobs:
    deploy:
      runs-on: ubuntu-latest
      environment:
        name: github-pages
        url: ${{ steps.deployment.outputs.page_url }}
      steps:
        1. actions/checkout@v4
        2. actions/configure-pages@v5
        3. actions/upload-pages-artifact@v3 with path: ./site
        4. actions/deploy-pages@v4 with id: deployment

Update README.md with:
- Full setup instructions: secrets, variables, and initial Strava OAuth flow
  to obtain the first refresh token
- How to trigger a manual sync via workflow_dispatch
- Note that GitHub Pages must be enabled manually in repo Settings >
  Pages > Source: GitHub Actions
```

**Check**: Commit everything and push. Confirm both workflows appear in the Actions
tab. Enable Pages in GitHub settings and trigger a manual workflow_dispatch run.

---

## Phase 10 — End-to-End Verification (Manual Checklist)

This phase is not an OpenCode prompt — do this yourself:

- [ ] Set `STRAVA_CLIENT_ID` as a repo Secret
- [ ] Set `STRAVA_CLIENT_SECRET` as a repo Secret
- [ ] Set `STRAVA_REFRESH_TOKEN` as a repo Variable (not Secret)
- [ ] Set `GH_PAT` as a repo Secret (`repo` + `actions:write` scopes)
- [ ] Set `MAPBOX_TOKEN` as a repo Secret
- [ ] Enable GitHub Pages: Settings > Pages > Source: GitHub Actions
- [ ] Trigger sync workflow manually via workflow_dispatch
- [ ] Verify `data/races.json` was updated and committed
- [ ] Verify `site/` was updated and committed (style.css, index, year subdirectories)
- [ ] Verify the Pages deployment workflow triggered and succeeded
- [ ] Visit your GitHub Pages URL and confirm:
    - [ ] Index page lists races grouped by year
    - [ ] Each race links to its detail page
    - [ ] Detail pages show correct stats
    - [ ] style.css loads correctly on both page types
    - [ ] Mapbox map renders on detail pages
    - [ ] Back link on detail pages returns to index (not a 404)

---

## Tips for Working with OpenCode

- **Always use Plan Mode (Tab) first** — review before any files are touched
- **One phase per session** — don't chain phases in a single prompt
- **If it goes off-track**: "Stop. Revert all changes. Re-read AGENTS.md and let's try a narrower scope."
- **After each phase**: run `go build ./...` and `go test ./...` yourself before moving on
- **Don't let it add dependencies** without explicit approval — stdlib-first is a hard constraint
- **Phase 6 (slug) and the helpers in Phase 7** are intentionally small and testable in isolation — don't skip them, they catch bugs before the full generate command is wired up
