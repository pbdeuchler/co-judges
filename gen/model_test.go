package main

import (
	"slices"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name, in string
		n        int
		wantErr  string
	}{
		{"valid", `[{"year":2026,"name":"A","level":"county","survey":{"overall":{"all":3.1}}}]`, 1, ""},
		{"empty list", `[]`, 0, ""},
		{"malformed json", `[{`, 0, "decode"},
		{"missing year", `[{"name":"A","level":"county"}]`, 0, "missing"},
		{"missing name", `[{"year":2026,"level":"county"}]`, 0, "missing"},
		{"unknown level", `[{"year":2026,"name":"A","level":"magistrate"}]`, 0, "missing"},
	}
	for _, tt := range tests {
		evs, err := Load(strings.NewReader(tt.in))
		switch {
		case tt.wantErr == "" && (err != nil || len(evs) != tt.n):
			t.Errorf("%s: Load = %d evs, %v; want %d, nil", tt.name, len(evs), err, tt.n)
		case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr) || evs != nil):
			t.Errorf("%s: Load error = %v, want one containing %q and no evaluations", tt.name, err, tt.wantErr)
		}
	}
}

func TestNameKey(t *testing.T) {
	for in, want := range map[string]string{
		"Alexander, William David":  "alexander-william",
		"Hood III, William W.":      "hood-william",
		"Samour Jr., Carlos":        "samour-carlos",
		`Vielma, MaryAnn "Mariana"`: "vielma-maryann",
		"Shamsid-Deen, Isaam":       "shamsid-deen-isaam",
		"Cher":                      "cher",
	} {
		if got := nameKey(in); got != want {
			t.Errorf("nameKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPersonKeys(t *testing.T) {
	evs := []Evaluation{
		{Year: 2016, SortName: "Murphy, Patrick", District: 17},
		{Year: 2016, SortName: "Murphy, Patrick William", District: 11},
		{Year: 2022, SortName: "Murphy, Patrick William", District: 11},
		{Year: 2020, SortName: "Hood, William"},
		{Year: 2026, SortName: "Hood III, William W."},
	}
	want := []string{"murphy-patrick-17", "murphy-patrick-11", "murphy-patrick-11", "hood-william", "hood-william"}
	if got := PersonKeys(evs); !slices.Equal(got, want) {
		t.Errorf("PersonKeys = %v, want %v", got, want)
	}
	if got := PersonKeys(nil); len(got) != 0 {
		t.Errorf("PersonKeys(nil) = %v, want empty", got)
	}
}

func TestVote(t *testing.T) {
	tests := []struct {
		name  string
		v     *Vote
		rec   string
		tally string
		split bool
	}{
		{"unanimous meets", &Vote{Meets: 10}, "meets", "10–0", false},
		{"split meets", &Vote{Meets: 5, Not: 4}, "meets", "5–4", true},
		{"does not meet, recommendation side first", &Vote{Meets: 3, Not: 7}, "does_not_meet", "7–3", true},
		{"unanimous without counts", &Vote{Unanimous: true}, "meets", "", false},
		{"majority without counts", &Vote{Majority: true}, "meets", "", true},
		{"unanimous against", &Vote{Not: 8}, "does_not_meet", "8–0", false},
		{"not stated", nil, "meets", "", false},
	}
	for _, tt := range tests {
		if got := tt.v.Tally(tt.rec); got != tt.tally {
			t.Errorf("%s: Tally = %q, want %q", tt.name, got, tt.tally)
		}
		if got := tt.v.Split(); got != tt.split {
			t.Errorf("%s: Split = %v, want %v", tt.name, got, tt.split)
		}
	}
}

func TestLoadMeta(t *testing.T) {
	m, err := LoadMeta(strings.NewReader(`{"scraped_at":"2026-10-21T04:04:14+00:00","parsed_at":"x"}`))
	if err != nil || m.ScrapedAt.Hour() != 4 || !m.BuiltAt.IsZero() {
		t.Errorf("LoadMeta valid = %+v, %v", m, err)
	}
	for name, in := range map[string]string{"malformed": `{`, "missing scraped_at": `{}`, "bad time": `{"scraped_at":"yesterday"}`} {
		if _, err := LoadMeta(strings.NewReader(in)); err == nil {
			t.Errorf("LoadMeta(%s) succeeded, want error", name)
		}
	}
}
