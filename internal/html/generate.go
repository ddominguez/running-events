package html

import (
	"html/template"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/ddominguez/running-events/internal/store"
)

type RaceData struct {
	store.Race
	Year               int
	Slug               string
	FormattedDate      string
	FormattedDistance  string
	FormattedDuration  string
	FormattedPace      string
	FormattedElevation string
	HasMap             bool
}

type IndexData struct {
	HasRaces bool
	Years    []int
	Races    map[int][]RaceData
}

type DetailData struct {
	Race      RaceData
	MapboxURL string
}

func GenerateSite(races []store.Race, mapboxToken string) error {
	grouped := groupByYear(races)

	if err := os.MkdirAll("site/assets", 0755); err != nil {
		return err
	}

	indexData := buildIndexData(grouped)
	indexTmpl := template.Must(template.New("index").Parse(IndexTemplate))
	f, err := os.Create("site/index.html")
	if err != nil {
		return err
	}
	if err := indexTmpl.Execute(f, indexData); err != nil {
		f.Close()
		return err
	}
	f.Close()

	for year, yearRaces := range grouped {
		dir := "site/" + strconv.Itoa(year)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		for _, r := range yearRaces {
			slug := Slug(r.Name)
			detailData := buildDetailData(r, mapboxToken)
			detailTmpl := template.Must(template.New("detail").Parse(DetailTemplate))
			f, err := os.Create(dir + "/" + slug + ".html")
			if err != nil {
				return err
			}
			if err := detailTmpl.Execute(f, detailData); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

func groupByYear(races []store.Race) map[int][]store.Race {
	grouped := make(map[int][]store.Race)
	for _, r := range races {
		t, err := time.Parse(time.RFC3339, r.StartDate)
		if err != nil {
			continue
		}
		year := t.Year()
		grouped[year] = append(grouped[year], r)
	}

	// Races are already sorted by start_date in store.SaveRaces,
	// so this sort may be redundant. Revisit later.
	for year := range grouped {
		slices.SortFunc(grouped[year], func(a, b store.Race) int {
			if a.StartDate > b.StartDate {
				return -1
			}
			if a.StartDate < b.StartDate {
				return 1
			}
			return 0
		})
	}

	return grouped
}

func buildIndexData(grouped map[int][]store.Race) IndexData {
	years := make([]int, 0, len(grouped))
	for year := range grouped {
		years = append(years, year)
	}
	// Years come from a map, so sorting is necessary.
	slices.SortFunc(years, func(a, b int) int {
		if a > b {
			return -1
		}
		if a < b {
			return 1
		}
		return 0
	})

	racesByYear := make(map[int][]RaceData)
	for _, year := range years {
		for _, r := range grouped[year] {
			racesByYear[year] = append(racesByYear[year], toRaceData(r))
		}
	}

	return IndexData{
		HasRaces: len(years) > 0,
		Years:    years,
		Races:    racesByYear,
	}
}

func buildDetailData(r store.Race, mapboxToken string) DetailData {
	rd := toRaceData(r)
	var mapboxURL string
	if r.SummaryPolyline != "" {
		mapboxURL = MapboxURL(r.SummaryPolyline, mapboxToken)
	}
	return DetailData{
		Race:      rd,
		MapboxURL: mapboxURL,
	}
}

func toRaceData(r store.Race) RaceData {
	t, _ := time.Parse(time.RFC3339, r.StartDate)
	year := t.Year()
	return RaceData{
		Race:               r,
		Year:               year,
		Slug:               Slug(r.Name),
		FormattedDate:      FormatDate(r.StartDate),
		FormattedDistance:  FormatDistance(r.Distance),
		FormattedDuration:  FormatDuration(r.MovingTime),
		FormattedPace:      FormatPace(r.Distance, r.MovingTime),
		FormattedElevation: FormatElevation(r.TotalElevationGain),
		HasMap:             r.SummaryPolyline != "",
	}
}
