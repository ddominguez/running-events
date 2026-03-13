package html

var IndexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Running Events</title>
	<link rel="stylesheet" href="assets/style.css">
</head>
<body>
	<header>
		<h1>Running Events</h1>
	</header>
	<main>
		{{if .HasRaces}}
			{{range .Years}}
			<h2 class="year-header">{{.}}</h2>
			<ul class="race-list">
				{{range $race := index $.Races .}}
				<li class="race-item">
					<a href="./{{$race.Year}}/{{$race.Slug}}.html">
						<div class="race-name">{{$race.Name}}</div>
						<div class="race-date">{{$race.FormattedDate}}</div>
						<div class="race-distance">{{$race.FormattedDistance}}</div>
					</a>
				</li>
				{{end}}
			</ul>
			{{end}}
		{{else}}
		<p class="no-races">No races yet.</p>
		{{end}}
	</main>
</body>
</html>
`

var DetailTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{.Race.Name}}</title>
	<link rel="stylesheet" href="../assets/style.css">
</head>
<body>
	<a href="../" class="back-link">← Back</a>
	<main>
		<div class="detail-header">
			<h1 class="detail-title">{{.Race.Name}}</h1>
			<p class="detail-date">{{.Race.FormattedDate}}</p>
		</div>
		<div class="stats">
			<div class="stat">
				<span class="stat-label">Distance</span>
				<span class="stat-value">{{.Race.FormattedDistance}}</span>
			</div>
			<div class="stat">
				<span class="stat-label">Elapsed Time</span>
				<span class="stat-value">{{.Race.FormattedDuration}}</span>
			</div>
			<div class="stat">
				<span class="stat-label">Pace</span>
				<span class="stat-value">{{.Race.FormattedPace}}</span>
			</div>
			<div class="stat">
				<span class="stat-label">Elevation</span>
				<span class="stat-value">{{.Race.FormattedElevation}}</span>
			</div>
		</div>
		{{if .Race.HasMap}}
		<div class="map">
			<img src="{{.MapboxURL}}" alt="Race map">
		</div>
		{{end}}
	</main>
</body>
</html>
`
