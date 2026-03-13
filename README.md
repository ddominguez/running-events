# Running Events

A static site generator that fetches running race activities from the Strava API and publishes them as a GitHub Pages site.

## Prerequisites

- Go installed
- Strava account
- GitHub account with a repository

## GitHub Setup

### 1. Enable GitHub Pages

Go to your repository **Settings → Pages** and set:
- **Source**: GitHub Actions

### 2. Add Repository Secrets

Go to **Settings → Secrets and variables → Actions** and add:

| Secret | Description |
|--------|-------------|
| `STRAVA_CLIENT_ID` | From your Strava API application |
| `STRAVA_CLIENT_SECRET` | From your Strava API application |
| `GH_PAT` | Personal access token with `repo` and `actions:write` scopes |
| `MAPBOX_TOKEN` | Public token from Mapbox (scoped to your GitHub Pages domain) |

### 3. Add Repository Variable

| Variable | Description |
|----------|-------------|
| `STRAVA_REFRESH_TOKEN` | Initial refresh token (see Strava OAuth setup below) |

## Strava OAuth Setup

### 1. Create a Strava API Application

Go to https://www.strava.com/settings/api and create an application.

### 2. Authorize the Application

Visit this URL (replace `YOUR_CLIENT_ID`):

```
https://www.strava.com/oauth/authorize?client_id=YOUR_CLIENT_ID&response_type=code&redirect_uri=http://localhost&approval_prompt=force&scope=read
```

Copy the authorization code from the URL after you redirect to `http://localhost?code=...`.

### 3. Exchange Code for Tokens

Run this curl command (replace the placeholders):

```bash
curl -X POST https://www.strava.com/oauth/token \
  -d client_id=YOUR_CLIENT_ID \
  -d client_secret=YOUR_CLIENT_SECRET \
  -d code=AUTHORIZATION_CODE \
  -d grant_type=authorization_code
```

The response includes an `access_token` and `refresh_token`. Copy the `refresh_token` and set it as the `STRAVA_REFRESH_TOKEN` variable in your GitHub repository.

## Initial Setup

1. Clone the repository
2. Run `go mod download` to fetch dependencies
3. Run `go run ./cmd/generate` (requires `MAPBOX_TOKEN` env var)
4. Push the changes — this triggers the GitHub Pages deployment

## Usage

### Automatic Sync

The sync workflow runs automatically every Monday at 8am UTC.

### Manual Sync

Go to **Actions → Sync Races → Run workflow**.

### Manual Deploy

Go to **Actions → Deploy to GitHub Pages → Run workflow**.

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

## Environment Variables (Local Development)

| Variable | Type | Description |
|----------|------|-------------|
| `STRAVA_CLIENT_ID` | Secret | Strava app client ID |
| `STRAVA_CLIENT_SECRET` | Secret | Strava app client secret |
| `STRAVA_REFRESH_TOKEN` | Variable | Strava refresh token |
| `GH_PAT` | Secret | GitHub PAT |
| `MAPBOX_TOKEN` | Secret | Mapbox public token |
