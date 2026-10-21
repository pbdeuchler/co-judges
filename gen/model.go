package main

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// Evaluation is one judge's evaluation in one retention cycle, as scraped by scrape/parse.py.
type Evaluation struct {
	Year                  int      `json:"year"`
	Slug                  string   `json:"slug"`
	URL                   string   `json:"url"`
	Name                  string   `json:"name"`
	SortName              string   `json:"sort_name"`
	Position              string   `json:"position"`
	Level                 string   `json:"level"` // supreme | appeals | district | county
	District              int      `json:"district"`
	Recommendation        string   `json:"recommendation"` // meets | does_not_meet | no_opinion
	RecommendationText    string   `json:"recommendation_text"`
	Vote                  *Vote    `json:"vote"`
	Narrative             []string `json:"narrative"`
	ReportURL             string   `json:"report_url"`
	InsufficientResponses bool     `json:"insufficient_responses"`
	Survey                Survey   `json:"survey"`
}

// Vote is the commission's vote. Meets/Not count commissioners; Unanimous or Majority is set when
// the page states the outcome without counts. A nil *Vote means the page did not say.
type Vote struct {
	Meets     int  `json:"meets"`
	Not       int  `json:"not"`
	Unanimous bool `json:"unanimous"`
	Majority  bool `json:"majority"`
}

// Tally is the vote as published, the recommendation's side first ("9–0", "7–3"), or "".
func (v *Vote) Tally(rec string) string {
	switch {
	case v == nil || v.Meets+v.Not == 0:
		return ""
	case rec == "does_not_meet":
		return fmt.Sprintf("%d–%d", v.Not, v.Meets)
	}
	return fmt.Sprintf("%d–%d", v.Meets, v.Not)
}

// Split reports a known non-unanimous vote.
func (v *Vote) Split() bool { return v != nil && (v.Majority || (v.Meets > 0 && v.Not > 0)) }

// Survey holds the numbers from the cycle's retention survey PDF. Maps are keyed by respondent
// group: "all", "att" (attorneys), "non" (non-attorneys); absent keys were not reported.
type Survey struct {
	Responses  int                           `json:"responses"`
	ReportYear int                           `json:"report_year"`
	Overall    map[string]float64            `json:"overall"`
	Meets      map[string]float64            `json:"meets"`
	Categories map[string]map[string]float64 `json:"categories"`
	History    []struct {
		Year  int     `json:"year"`
		Score float64 `json:"score"`
		N     int     `json:"n"`
	} `json:"history"`
}

// Load decodes evaluations and validates the fields the site depends on.
func Load(r io.Reader) ([]Evaluation, error) {
	var evs []Evaluation
	if err := json.NewDecoder(r).Decode(&evs); err != nil {
		return nil, fmt.Errorf("decode evaluations: %w", err)
	}
	for i, e := range evs {
		if e.Year == 0 || e.Name == "" || CohortOf(e.Level) == "" {
			return nil, fmt.Errorf("evaluation %d (%q): missing year, name or known level", i, e.Slug)
		}
	}
	return evs, nil
}

// Meta stamps the site: when the state's site was last scraped and when this build ran.
type Meta struct {
	ScrapedAt time.Time `json:"scraped_at"`
	BuiltAt   time.Time `json:"-"`
}

// LoadMeta decodes data/meta.json; a missing scrape time is an error so a stale stamp never ships.
func LoadMeta(r io.Reader) (Meta, error) {
	var m Meta
	if err := json.NewDecoder(r).Decode(&m); err != nil {
		return Meta{}, fmt.Errorf("decode meta: %w", err)
	}
	if m.ScrapedAt.IsZero() {
		return Meta{}, fmt.Errorf("meta: missing scraped_at")
	}
	return m, nil
}

var (
	suffixRe = regexp.MustCompile(`\b(jr|sr|ii|iii|iv)\b\.?`)
	nonAlpha = regexp.MustCompile(`[^a-z]+`)
)

// nameKey is "last-first" from a "Last, First Middle" sort name, ignoring generational suffixes.
func nameKey(sortName string) string {
	last, first, _ := strings.Cut(strings.ToLower(sortName), ",")
	last = strings.Trim(nonAlpha.ReplaceAllString(suffixRe.ReplaceAllString(last, ""), "-"), "-")
	firsts := strings.Fields(nonAlpha.ReplaceAllString(first, " "))
	if len(firsts) == 0 {
		return last
	}
	return last + "-" + firsts[0]
}

// PersonKeys links evaluations of the same judge across cycles. A name shared by two judges in
// one cycle is split by district for every evaluation carrying it.
func PersonKeys(evs []Evaluation) []string {
	seen := map[[2]any]int{}
	for _, e := range evs {
		seen[[2]any{e.Year, nameKey(e.SortName)}]++
	}
	ambiguous := map[string]bool{}
	for k, n := range seen {
		if n > 1 {
			ambiguous[k[1].(string)] = true
		}
	}
	keys := make([]string, len(evs))
	for i, e := range evs {
		k := nameKey(e.SortName)
		if ambiguous[k] {
			k = fmt.Sprintf("%s-%d", k, e.District)
		}
		keys[i] = k
	}
	return keys
}
