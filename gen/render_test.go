package main

import (
	"strings"
	"testing"
	"time"
)

func fixture() []Evaluation {
	low := ev("county", 2.5, 2.4, 0.5)
	low.Name, low.SortName, low.Position = "Low Judge", "Judge, Low", "Twentieth Judicial District - Boulder County Court Judge"
	high := ev("county", 3.9, 3.9, 1.0)
	high.Name, high.SortName, high.Position = "High Judge", "Judge, High", "Twentieth Judicial District - Boulder County Court Judge"
	mid := ev("county", 3.2, 3.2, 0.8)
	mid.Name, mid.SortName, mid.Position = "Mid Judge", "Judge, Mid", "Eighth Judicial District - Larimer County Court Judge"
	dnm := ev("district", 3.0, 3.0, 0.6)
	dnm.Name, dnm.SortName, dnm.District, dnm.Recommendation, dnm.Vote = "Flag Judge", "Judge, Flag", 20, "does_not_meet", &Vote{Meets: 3, Not: 7}
	none := Evaluation{Year: 2026, Name: "Quiet Judge", SortName: "Judge, Quiet", Level: "appeals", InsufficientResponses: true, Recommendation: "meets"}
	return []Evaluation{low, high, mid, dnm, none}
}

func TestBuild(t *testing.T) {
	files, err := Build(fixture(), testMeta)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"index.html", "judges/index.html", "method/index.html", "ballot/boulder/index.html",
		"ballot/yuma/index.html", "judges/judge-low/index.html", "style.css", "app.js", "fonts/fira-sans-800.woff2"} {
		if len(files[p]) == 0 {
			t.Errorf("Build did not emit %s", p)
		}
	}
	boulder := string(files["ballot/boulder/index.html"])
	for _, want := range []string{"High Judge", "Low Judge", "Flag Judge", "Quiet Judge", "is-low", "does not meet standards (7–3)", "Too few survey responses", "<dt>Commission vote was split</dt><dd>1</dd>"} {
		if !strings.Contains(boulder, want) {
			t.Errorf("Boulder ballot missing %q", want)
		}
	}
	if strings.Contains(boulder, "Mid Judge") {
		t.Error("Boulder ballot lists a Larimer County judge")
	}
	if strings.Index(boulder, "High Judge") > strings.Index(boulder, "Low Judge") {
		t.Error("ballot rows are not ordered by standing")
	}
	yuma := string(files["ballot/yuma/index.html"])
	if !strings.Contains(yuma, "No District Court judges") || !strings.Contains(yuma, "Quiet Judge") {
		t.Error("Yuma ballot should show the empty district state and the statewide judge")
	}
	if !strings.Contains(string(files["judges/judge-low/index.html"]), `href="../../style.css"`) {
		t.Error("judge page assets must be relative to its depth")
	}
}

func TestBuildEmpty(t *testing.T) {
	files, err := Build(nil, testMeta)
	if err != nil {
		t.Fatalf("Build(nil, testMeta) error: %v", err)
	}
	if !strings.Contains(string(files["ballot/boulder/index.html"]), "No District Court judges") {
		t.Error("an empty dataset should still render empty ballot states")
	}
}

func TestOrdinal(t *testing.T) {
	for n, want := range map[int]string{1: "1st", 2: "2nd", 3: "3rd", 4: "4th", 11: "11th", 12: "12th", 13: "13th", 21: "21st", 22: "22nd", 111: "111th", 0: "0th"} {
		if got := ordinal(n); got != want {
			t.Errorf("ordinal(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestBinOf(t *testing.T) {
	tests := []struct {
		v, lo, hi float64
		n, want   int
	}{
		{0, 0, 100, 20, 0}, {100, 0, 100, 20, 19}, {50, 0, 100, 20, 10},
		{-5, 0, 100, 20, 0}, {105, 0, 100, 20, 19}, {3, 3, 3, 10, 0}, // degenerate range
	}
	for _, tt := range tests {
		if got := binOf(tt.v, tt.lo, tt.hi, tt.n); got != tt.want {
			t.Errorf("binOf(%v, %v, %v, %d) = %d, want %d", tt.v, tt.lo, tt.hi, tt.n, got, tt.want)
		}
	}
}

func TestHistSVG(t *testing.T) {
	svg := string(histSVG([]int{1, 4, 0, 2}, 0, 100, 30, 200, 40, `a "b" & c`))
	if strings.Count(svg, `class="y"`) != 1 || strings.Count(svg, `class="c"`) != 3 || !strings.Contains(svg, `class="t"`) {
		t.Errorf("histogram should mark exactly one bin plus a tick: %s", svg)
	}
	if strings.Contains(svg, `"b"`) {
		t.Error("histogram label is not escaped")
	}
	if empty := string(histSVG([]int{0, 0}, 0, 1, 0.5, 10, 10, "")); !strings.Contains(empty, "<svg") {
		t.Error("an all-zero histogram must still render")
	}
}

func TestWeakest(t *testing.T) {
	tests := []struct {
		pct  Metrics
		want string
	}{
		{Metrics{10, 50, 90}, "survey grade 2.50 (10th percentile)"},
		{Metrics{50, 3, 90}, "category mean 2.40 (3rd percentile)"},
		{Metrics{50, 40, 21}, "attorneys “meets” 50.00% (21st percentile)"},
	}
	for _, tt := range tests {
		e := Entry{Placement: Placement{Metrics: Metrics{2.5, 2.4, 0.5}, Pct: tt.pct, Of: 50}}
		if got := weakest(e); got != tt.want {
			t.Errorf("weakest(%v) = %q, want %q", tt.pct, got, tt.want)
		}
	}
	small := Entry{Placement: Placement{Metrics: Metrics{2.5, 2.4, 0.5}, Pct: Metrics{10, 50, 90}, Of: 7}}
	if got := weakest(small); got != "survey grade 2.50" {
		t.Errorf("small-group weakest = %q, want no percentile", got)
	}
}

func TestCatRange(t *testing.T) {
	es := []Entry{
		{Evaluation: ev("district", 3, 2.1, .5)}, {Evaluation: ev("district", 3, 3.9, .5)},
		{Evaluation: ev("county", 3, 1.0, .5)}, // other cohort ignored
	}
	if lo, hi := catRange(es, "2026 district", "Fairness"); lo != 2.1 || hi != 3.9 {
		t.Errorf("catRange district = %v, %v; want 2.1, 3.9", lo, hi)
	}
	if lo, hi := catRange(es, "2026 appellate", "Fairness"); lo != 0 || hi != 4 {
		t.Errorf("catRange with no data = %v, %v; want the full 0-4 scale", lo, hi)
	}
}

func TestStanding(t *testing.T) {
	for _, tt := range []struct {
		p    Placement
		want string
	}{
		{Placement{Rank: 3, Of: 7, Standing: 60}, "rank 3 of 7"},
		{Placement{Rank: 7, Of: 65, Standing: 90.4}, "90th percentile, rank 7 of 65"},
	} {
		if got := standing(tt.p); got != tt.want {
			t.Errorf("standing(%+v) = %q, want %q", tt.p, got, tt.want)
		}
	}
}

func TestBuildSmallGroupsAndLetters(t *testing.T) {
	files, err := Build(fixture(), testMeta)
	if err != nil {
		t.Fatal(err)
	}
	boulder := string(files["ballot/boulder/index.html"])
	if !strings.Contains(boulder, "<small>of 4</small>") {
		t.Error("when even the statewide group is tiny, the square shows the statewide raw rank")
	}
	if !strings.Contains(boulder, "Oct 20, 2026, 10:04 PM MDT") || !strings.Contains(boulder, "Oct 21, 2026, 12:25 AM MDT") {
		t.Error("pages should carry the scrape and build stamps in Mountain Time")
	}
	if strings.Contains(boulder, `class="pic"`) {
		t.Error("court letter squares should be gone")
	}
	if !strings.Contains(boulder, "statewide 1 of 4") && !strings.Contains(boulder, "statewide 4 of 4") {
		t.Error("rows should show the statewide rank")
	}
	if !strings.Contains(string(files["judges/index.html"]), `class="dist-row" data-year="2026"`) {
		t.Error("all-judges page should carry the per-cycle distributions")
	}
}

var testMeta = Meta{
	ScrapedAt: time.Date(2026, 10, 21, 4, 4, 14, 0, time.UTC),
	BuiltAt:   time.Date(2026, 10, 21, 6, 25, 0, 0, time.UTC),
}

func TestHeadline(t *testing.T) {
	peer, state := Placement{Rank: 3, Of: 65, Standing: 90}, Placement{Rank: 14, Of: 108, Standing: 87}
	if got := headline(Entry{Placement: peer, Global: state}); got != peer {
		t.Errorf("large peer group: headline = %+v, want the peer placement", got)
	}
	small := Placement{Rank: 1, Of: 7, Standing: 92}
	if got := headline(Entry{Placement: small, Global: state}); got != state {
		t.Errorf("small peer group: headline = %+v, want the statewide placement", got)
	}
}
