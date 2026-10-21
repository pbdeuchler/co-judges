package main

import (
	"slices"
	"testing"
)

func TestDistrictCounties(t *testing.T) {
	tests := []struct {
		district, year int
		want           []string
	}{
		{20, 2026, []string{"Boulder"}},
		{18, 2026, []string{"Arapahoe"}},
		{18, 2024, []string{"Arapahoe", "Douglas", "Elbert", "Lincoln"}}, // before the 23rd split off
		{23, 2026, []string{"Douglas", "Elbert", "Lincoln"}},
		{23, 2024, nil}, // did not exist yet
		{99, 2026, nil}, // unknown district
		{0, 2026, nil},
	}
	for _, tt := range tests {
		if got := DistrictCounties(tt.district, tt.year); !slices.Equal(got, tt.want) {
			t.Errorf("DistrictCounties(%d, %d) = %v, want %v", tt.district, tt.year, got, tt.want)
		}
	}
}

func TestCounties(t *testing.T) {
	cs := Counties()
	if len(cs) != 64 {
		t.Errorf("Counties() has %d entries, want Colorado's 64", len(cs))
	}
	if !slices.IsSorted(cs) || len(slices.Compact(slices.Clone(cs))) != len(cs) {
		t.Error("Counties() must be sorted and unique")
	}
}

func TestCountyOf(t *testing.T) {
	for position, want := range map[string]string{
		"Twentieth Judicial District - Boulder County Court Judge": "Boulder",
		"Second Judicial District - Denver County Judge":           "Denver",
		"Seventh Judicial District - San Miguel County Judge":      "San Miguel",
		"Twentieth Judicial District - District Court Judge":       "",
		"Colorado Court of Appeals":                                "",
		"":                                                         "",
	} {
		if got := CountyOf(position); got != want {
			t.Errorf("CountyOf(%q) = %q, want %q", position, got, want)
		}
	}
}

func TestOnBallot(t *testing.T) {
	district := Evaluation{Year: 2026, Level: "district", District: 20}
	douglas2024 := Evaluation{Year: 2024, Level: "district", District: 18}
	county := Evaluation{Year: 2026, Level: "county", Position: "Twentieth Judicial District - Boulder County Court Judge"}
	tests := []struct {
		name   string
		e      Evaluation
		county string
		want   bool
	}{
		{"appellate everywhere", Evaluation{Level: "supreme"}, "Yuma", true},
		{"district in its county", district, "Boulder", true},
		{"district elsewhere", district, "Weld", false},
		{"18th covered Douglas in 2024", douglas2024, "Douglas", true},
		{"county court in its county", county, "Boulder", true},
		{"county court elsewhere", county, "Broomfield", false},
		{"unknown level", Evaluation{Level: "magistrate"}, "Boulder", false},
	}
	for _, tt := range tests {
		if got := OnBallot(tt.e, tt.county); got != tt.want {
			t.Errorf("%s: OnBallot = %v, want %v", tt.name, got, tt.want)
		}
	}
}
