package main

import (
	"math"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestCohortOf(t *testing.T) {
	for level, want := range map[string]string{"supreme": Appellate, "appeals": Appellate, "district": District, "county": County, "probate": "", "": ""} {
		if got := CohortOf(level); got != want {
			t.Errorf("CohortOf(%q) = %q, want %q", level, got, want)
		}
	}
}

func TestMetricsOf(t *testing.T) {
	tests := []struct {
		name string
		s    Survey
		want Metrics
		ok   bool
	}{
		{"modern report", Survey{Overall: map[string]float64{"all": 3.5}, Meets: map[string]float64{"att": .9},
			Categories: map[string]map[string]float64{"A": {"all": 3}, "B": {"all": 4}}}, Metrics{3.5, 3.5, .9}, true},
		{"split groups averaged", Survey{Overall: map[string]float64{"all": 3}, Meets: map[string]float64{"att": .5},
			Categories: map[string]map[string]float64{"A": {"att": 3, "non": 4}, "B": {"non": 2}}}, Metrics{3, 2.75, .5}, true},
		{"missing overall", Survey{Meets: map[string]float64{"att": .9}, Categories: map[string]map[string]float64{"A": {"all": 3}}}, Metrics{}, false},
		{"missing attorney share", Survey{Overall: map[string]float64{"all": 3}, Categories: map[string]map[string]float64{"A": {"all": 3}}}, Metrics{}, false},
		{"no categories", Survey{Overall: map[string]float64{"all": 3}, Meets: map[string]float64{"att": .9}}, Metrics{}, false},
		{"empty category groups", Survey{Overall: map[string]float64{"all": 3}, Meets: map[string]float64{"att": .9},
			Categories: map[string]map[string]float64{"A": {}}}, Metrics{}, false},
		{"empty survey", Survey{}, Metrics{}, false},
	}
	for _, tt := range tests {
		got, ok := MetricsOf(tt.s)
		if ok != tt.ok || !near(got.Overall, tt.want.Overall) || !near(got.Categories, tt.want.Categories) || !near(got.AttyMeets, tt.want.AttyMeets) {
			t.Errorf("%s: MetricsOf = %+v, %v; want %+v, %v", tt.name, got, ok, tt.want, tt.ok)
		}
	}
}

func TestGroupScore(t *testing.T) {
	tests := []struct {
		g    map[string]float64
		want float64
		ok   bool
	}{
		{map[string]float64{"all": 3, "att": 1}, 3, true},
		{map[string]float64{"att": 2, "non": 4}, 3, true},
		{map[string]float64{"non": 2}, 2, true},
		{map[string]float64{"peer": 2}, 0, false},
		{nil, 0, false},
	}
	for _, tt := range tests {
		if got, ok := groupScore(tt.g); ok != tt.ok || !near(got, tt.want) {
			t.Errorf("groupScore(%v) = %v, %v; want %v, %v", tt.g, got, ok, tt.want, tt.ok)
		}
	}
}

func TestPercentile(t *testing.T) {
	pool := []float64{1, 2, 2, 3}
	tests := []struct {
		pool []float64
		v    float64
		want float64
	}{
		{pool, 3, 87.5}, // 3 below + half of itself
		{pool, 2, 50},   // 1 below + half of two ties
		{pool, 1, 12.5}, // lowest is never 0
		{pool, 5, 100},  // above the pool
		{pool, 0, 0},    // below the pool
		{nil, 3, 0},     // empty pool
		{[]float64{2}, 2, 50},
	}
	for _, tt := range tests {
		if got := Percentile(tt.pool, tt.v); !near(got, tt.want) {
			t.Errorf("Percentile(%v, %v) = %v, want %v", tt.pool, tt.v, got, tt.want)
		}
	}
}

func ev(level string, overall, cat, atty float64) Evaluation {
	return Evaluation{Year: 2026, Name: "x", Level: level, Survey: Survey{
		Overall: map[string]float64{"all": overall}, Meets: map[string]float64{"att": atty},
		Categories: map[string]map[string]float64{"Fairness": {"all": cat}}}}
}

func TestPlace(t *testing.T) {
	evs := []Evaluation{
		ev("district", 3.8, 3.8, 1.0), // best district
		ev("district", 3.0, 3.0, 0.5), // worst district
		ev("district", 3.4, 3.4, 0.8), // tied middle
		ev("district", 3.4, 3.4, 0.8), // tied middle
		ev("county", 2.0, 2.0, 0.1),   // only county: ranks alone
		ev("supreme", 3.9, 3.9, 0.9),  // appellate cohort pools supreme ...
		ev("appeals", 3.5, 3.5, 0.7),  // ... with appeals
		{Year: 2026, Name: "unpublished", Level: "district"},
	}
	got := Place(evs, PeerGroup)
	if len(got) != len(evs) {
		t.Fatalf("Place returned %d placements for %d evaluations", len(got), len(evs))
	}
	tests := []struct {
		i        int
		ranked   bool
		rank, of int
		standing float64
	}{
		{0, true, 1, 4, 87.5},
		{1, true, 4, 4, 12.5},
		{2, true, 2, 4, 50},
		{3, true, 2, 4, 50}, // ties share rank and percentile
		{4, true, 1, 1, 50},
		{5, true, 1, 2, 75},
		{6, true, 2, 2, 25},
		{7, false, 0, 0, 0}, // unranked, not counted in Of
	}
	for _, tt := range tests {
		p := got[tt.i]
		if p.Ranked != tt.ranked || p.Rank != tt.rank || p.Of != tt.of || !near(p.Standing, tt.standing) {
			t.Errorf("evs[%d]: got ranked=%v rank=%d of=%d standing=%v; want %v %d %d %v", tt.i, p.Ranked, p.Rank, p.Of, p.Standing, tt.ranked, tt.rank, tt.of, tt.standing)
		}
	}
	if c := got[0].Composite; !near(c, 87.5) {
		t.Errorf("composite of best district = %v, want 87.5 (mean of three 87.5 percentiles)", c)
	}
	if evs[7].Survey.Overall != nil {
		t.Error("Place mutated its input")
	}
}

func TestPlaceEmpty(t *testing.T) {
	if got := Place(nil, PeerGroup); len(got) != 0 {
		t.Errorf("Place(nil) = %v, want empty", got)
	}
}

func TestPlaceSeparatesCycles(t *testing.T) {
	old, cur := ev("district", 3.9, 3.9, 1.0), ev("district", 3.0, 3.0, 0.5)
	old.Year = 2022
	got := Place([]Evaluation{old, cur}, PeerGroup)
	for i, p := range got {
		if p.Rank != 1 || p.Of != 1 {
			t.Errorf("evs[%d]: rank %d of %d; want 1 of 1 (different cycles are not peers)", i, p.Rank, p.Of)
		}
	}
	if PeerGroup(old) == PeerGroup(cur) {
		t.Error("PeerGroup must differ across cycles")
	}
	if PeerGroup(ev("supreme", 3, 3, 1)) != PeerGroup(ev("appeals", 3, 3, 1)) {
		t.Error("Supreme Court and Court of Appeals share a peer group within a cycle")
	}
}

func withResponses(e Evaluation, n int) Evaluation { e.Survey.Responses = n; return e }

func TestPlaceStatewideAndResponses(t *testing.T) {
	evs := []Evaluation{
		withResponses(ev("district", 3.8, 3.8, 1.0), 40),
		withResponses(ev("county", 3.0, 3.0, 0.5), 10),
		withResponses(ev("appeals", 3.4, 3.4, 0.8), 10),
	}
	peer, state := Place(evs, PeerGroup), Place(evs, CycleGroup)
	for i := range evs {
		if peer[i].Rank != 1 || peer[i].Of != 1 {
			t.Errorf("evs[%d]: peer rank %d of %d; each is alone in its court level", i, peer[i].Rank, peer[i].Of)
		}
	}
	want := []struct {
		rank, respRank int
		respPct        float64
	}{{1, 1, 83.333333}, {3, 2, 33.333333}, {2, 2, 33.333333}} // two judges tie on 10 responses
	for i, w := range want {
		if state[i].Rank != w.rank || state[i].Of != 3 || state[i].RespRank != w.respRank || !near(state[i].RespPct, w.respPct) {
			t.Errorf("evs[%d] statewide: rank %d/%d resp #%d %.4f; want %d/3 #%d %.4f", i, state[i].Rank, state[i].Of, state[i].RespRank, state[i].RespPct, w.rank, w.respRank, w.respPct)
		}
	}
}

func TestSmall(t *testing.T) {
	for of, want := range map[int]bool{1: true, SmallGroup: true, SmallGroup + 1: false, 0: true} {
		if got := (Placement{Of: of}).Small(); got != want {
			t.Errorf("Placement{Of: %d}.Small() = %v, want %v", of, got, want)
		}
	}
}

func TestRanks(t *testing.T) {
	tests := []struct {
		in   []float64
		want []int
	}{
		{[]float64{1, 3, 2}, []int{3, 1, 2}},
		{[]float64{2, 2, 1}, []int{1, 1, 3}}, // competition ranking: ties share, next skips
		{[]float64{5}, []int{1}},
		{nil, []int{}},
	}
	for _, tt := range tests {
		got := ranks(tt.in)
		if len(got) != len(tt.want) {
			t.Fatalf("ranks(%v) = %v, want %v", tt.in, got, tt.want)
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("ranks(%v) = %v, want %v", tt.in, got, tt.want)
				break
			}
		}
	}
}
