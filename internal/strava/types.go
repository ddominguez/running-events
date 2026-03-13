package strava

type Activity struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	StartDate          string  `json:"start_date"`
	Distance           int     `json:"distance"`
	MovingTime         int     `json:"moving_time"`
	ElapsedTime        int     `json:"elapsed_time"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
	Type               string  `json:"type"`
	SportType          string  `json:"sport_type"`
	WorkoutType        int     `json:"workout_type"`
	Map                struct {
		SummaryPolyline string `json:"summary_polyline"`
	} `json:"map"`
}
