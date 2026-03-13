package html

import (
	"fmt"
	"net/url"
	"time"
)

const (
	metersPerKm   = 1000.0
	metersPerMile = 1609.344
	metersPerFoot = 3.28084
)

func FormatDistance(meters int) string {
	km := float64(meters) / metersPerKm
	miles := float64(meters) / metersPerMile
	return fmt.Sprintf("%.2f km (%.2f mi)", km, miles)
}

func FormatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%d:%02d:%02d", h, m, s)
}

func FormatPace(meters int, seconds int) string {
	if meters == 0 {
		return "-- /km (-- /mi)"
	}
	kmSeconds := float64(seconds) / (float64(meters) / metersPerKm)
	miSeconds := float64(seconds) / (float64(meters) / metersPerMile)

	kmMin := int(kmSeconds) / 60
	kmSec := int(kmSeconds) % 60
	miMin := int(miSeconds) / 60
	miSec := int(miSeconds) % 60

	return fmt.Sprintf("%d:%02d /km (%d:%02d /mi)", kmMin, kmSec, miMin, miSec)
}

func FormatElevation(meters float64) string {
	feet := meters * metersPerFoot
	return fmt.Sprintf("%.0f m (%.0f ft)", meters, feet)
}

func FormatDate(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("January 2, 2006")
}

func MapboxURL(polyline, token string) string {
	encoded := url.QueryEscape(polyline)
	return fmt.Sprintf(
		"https://api.mapbox.com/styles/v1/mapbox/dark-v11/static/path-5+fc4c02-0.8(%s)/auto/800x400?logo=false&access_token=%s",
		encoded, token,
	)
}
