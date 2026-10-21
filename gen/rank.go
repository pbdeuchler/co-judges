package main

import (
	"fmt"
	"math"
	"sort"
)

// Cohorts compare like with like: the Supreme Court and Court of Appeals share one appellate
// survey instrument, so they are pooled.
const (
	Appellate = "appellate"
	District  = "district"
	County    = "county"
)

func CohortOf(level string) string {
	switch level {
	case "supreme", "appeals":
		return Appellate
	case "district":
		return District
	case "county":
		return County
	}
	return ""
}

// Metrics are the three survey numbers the composite blends.
type Metrics struct {
	Overall    float64 // overall grade, 0-4
	Categories float64 // mean of category scores, 0-4
	AttyMeets  float64 // share of attorneys answering "meets performance standards", 0-1
}

// MetricsOf returns the metrics, or false when the survey lacks any of them.
func MetricsOf(s Survey) (Metrics, bool) {
	overall, ok1 := s.Overall["all"]
	atty, ok2 := s.Meets["att"]
	cats := make([]float64, 0, len(s.Categories))
	for _, g := range s.Categories {
		if v, ok := groupScore(g); ok {
			cats = append(cats, v)
		}
	}
	if !ok1 || !ok2 || len(cats) == 0 {
		return Metrics{}, false
	}
	return Metrics{overall, mean(cats), atty}, true
}

// groupScore is the all-respondent score, or the mean of the per-group scores the older
// (2016-2018) reports publish instead.
func groupScore(g map[string]float64) (float64, bool) {
	if v, ok := g["all"]; ok {
		return v, true
	}
	vs := make([]float64, 0, 2)
	for _, k := range []string{"att", "non"} {
		if v, ok := g[k]; ok {
			vs = append(vs, v)
		}
	}
	return mean(vs), len(vs) > 0
}

func mean(vs []float64) float64 {
	if len(vs) == 0 {
		return 0
	}
	s := 0.0
	for _, v := range vs {
		s += v
	}
	return s / float64(len(vs))
}

// Percentile is the mid-rank percentile of v in pool: ties count half, so equal scores share a
// percentile and the extremes are never exactly 0 or 100. An empty pool yields 0.
func Percentile(pool []float64, v float64) float64 {
	if len(pool) == 0 {
		return 0
	}
	below, equal := 0, 0
	for _, p := range pool {
		switch {
		case almostEqual(p, v):
			equal++
		case p < v:
			below++
		}
	}
	return 100 * (float64(below) + 0.5*float64(equal)) / float64(len(pool))
}

func almostEqual(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// SmallGroup is the largest group shown as raw ranks instead of percentiles: with ten or fewer
// judges one place moves a percentile by ten points or more.
const SmallGroup = 10

// Placement is an evaluation's standing within a group. Ranked is false when the survey was not
// published or lacks a metric; the other fields are then zero.
type Placement struct {
	Ranked    bool
	Metrics   Metrics
	Pct       Metrics // per-metric percentiles, 0-100
	Composite float64 // mean of the three percentiles
	Standing  float64 // percentile of Composite in the group, 0-100; the headline number
	Rank, Of  int     // competition rank (ties share) among ranked evaluations in the group
	RespPct   float64 // percentile of the response count in the group
	RespRank  int     // competition rank by response count (most responses = 1)
}

// Small reports whether the group is too small for percentiles to mean much.
func (p Placement) Small() bool { return p.Of <= SmallGroup }

// PeerGroup is who an evaluation is compared with: the same court level in the same cycle.
// Cycles are not pooled because the survey scale changed between them (2024 especially).
func PeerGroup(e Evaluation) string { return fmt.Sprintf("%d %s", e.Year, CohortOf(e.Level)) }

// CycleGroup is every evaluation in the same cycle, all court levels: the statewide comparison.
func CycleGroup(e Evaluation) string { return fmt.Sprint(e.Year) }

// Place ranks every evaluation against the others sharing its group key. The result is
// index-aligned with evs.
func Place(evs []Evaluation, group func(Evaluation) string) []Placement {
	out := make([]Placement, len(evs))
	byGroup := map[string][]int{}
	for i, e := range evs {
		if m, ok := MetricsOf(e.Survey); ok {
			out[i] = Placement{Ranked: true, Metrics: m}
			byGroup[group(e)] = append(byGroup[group(e)], i)
		}
	}
	for _, idx := range byGroup {
		pool := func(f func(int) float64) []float64 {
			vs := make([]float64, len(idx))
			for j, i := range idx {
				vs[j] = f(i)
			}
			return vs
		}
		ov, ca, at := pool(func(i int) float64 { return out[i].Metrics.Overall }),
			pool(func(i int) float64 { return out[i].Metrics.Categories }),
			pool(func(i int) float64 { return out[i].Metrics.AttyMeets })
		for _, i := range idx {
			m := out[i].Metrics
			p := Metrics{Percentile(ov, m.Overall), Percentile(ca, m.Categories), Percentile(at, m.AttyMeets)}
			out[i].Pct, out[i].Composite = p, (p.Overall+p.Categories+p.AttyMeets)/3
		}
		comps := pool(func(i int) float64 { return out[i].Composite })
		resps := pool(func(i int) float64 { return float64(evs[i].Survey.Responses) })
		rank, respRank := ranks(comps), ranks(resps)
		for j, i := range idx {
			out[i].Standing, out[i].Rank, out[i].Of = Percentile(comps, comps[j]), rank[j], len(idx)
			out[i].RespPct, out[i].RespRank = Percentile(resps, resps[j]), respRank[j]
		}
	}
	return out
}

// ranks gives each value its competition rank, highest first; equal values share a rank.
func ranks(vs []float64) []int {
	order := make([]int, len(vs))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return vs[order[a]] > vs[order[b]] })
	out := make([]int, len(vs))
	for r, i := range order {
		if r > 0 && almostEqual(vs[i], vs[order[r-1]]) {
			out[i] = out[order[r-1]]
		} else {
			out[i] = r + 1
		}
	}
	return out
}
