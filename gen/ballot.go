package main

import (
	"regexp"
	"slices"
	"strings"
)

// districtCounties maps Colorado judicial districts to their counties. The 23rd District
// (Douglas, Elbert, Lincoln) split from the 18th in January 2025.
var districtCounties = map[int][]string{
	1: {"Gilpin", "Jefferson"}, 2: {"Denver"}, 3: {"Huerfano", "Las Animas"}, 4: {"El Paso", "Teller"},
	5: {"Clear Creek", "Eagle", "Lake", "Summit"}, 6: {"Archuleta", "La Plata", "San Juan"},
	7: {"Delta", "Gunnison", "Hinsdale", "Montrose", "Ouray", "San Miguel"}, 8: {"Jackson", "Larimer"},
	9: {"Garfield", "Pitkin", "Rio Blanco"}, 10: {"Pueblo"}, 11: {"Chaffee", "Custer", "Fremont", "Park"},
	12: {"Alamosa", "Conejos", "Costilla", "Mineral", "Rio Grande", "Saguache"},
	13: {"Kit Carson", "Logan", "Morgan", "Phillips", "Sedgwick", "Washington", "Yuma"},
	14: {"Grand", "Moffat", "Routt"}, 15: {"Baca", "Cheyenne", "Kiowa", "Prowers"}, 16: {"Bent", "Crowley", "Otero"},
	17: {"Adams", "Broomfield"}, 18: {"Arapahoe"}, 19: {"Weld"}, 20: {"Boulder"}, 21: {"Mesa"},
	22: {"Dolores", "Montezuma"}, 23: {"Douglas", "Elbert", "Lincoln"},
}

// DistrictCounties returns the counties a district served in a given year.
func DistrictCounties(district, year int) []string {
	if year < 2025 {
		switch district {
		case 18:
			return []string{"Arapahoe", "Douglas", "Elbert", "Lincoln"}
		case 23:
			return nil
		}
	}
	return districtCounties[district]
}

// Counties lists all 64 counties alphabetically.
func Counties() []string {
	var all []string
	for _, cs := range districtCounties {
		all = append(all, cs...)
	}
	slices.Sort(all)
	return all
}

var countyCourtRe = regexp.MustCompile(`([A-Z][a-z]+(?: [A-Z][a-z]+)?) County (?:Court )?Judge`)

// CountyOf returns the county a county-court judge sits in, or "" for other positions.
func CountyOf(position string) string {
	m := countyCourtRe.FindStringSubmatch(position)
	if m == nil {
		return ""
	}
	return strings.TrimPrefix(m[1], "- ")
}

// OnBallot reports whether a county's voters decide this evaluation's retention: appellate judges
// statewide, district judges in their district, county judges in their county.
func OnBallot(e Evaluation, county string) bool {
	switch CohortOf(e.Level) {
	case Appellate:
		return true
	case District:
		return slices.Contains(DistrictCounties(e.District, e.Year), county)
	case County:
		return CountyOf(e.Position) == county
	}
	return false
}
