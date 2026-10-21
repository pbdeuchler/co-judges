package main

import (
	"bytes"
	"cmp"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"math"
	"slices"
	"strings"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

// Ballot is the cycle the front door and county pages cover.
const Ballot = 2026

// Entry is an evaluation with its standing among peers (embedded) and statewide (Global).
type Entry struct {
	Evaluation
	Key string
	Placement
	Global Placement
}

// Judge is one person across every cycle they were evaluated, newest first.
type Judge struct {
	Key     string
	Entries []Entry
}

func (j Judge) Latest() Entry { return j.Entries[0] }

// Site is everything the templates read.
type Site struct {
	Entries  []Entry
	Judges   map[string]Judge
	Bins     map[string][]int                // composite histogram per group key (PeerGroup or CycleGroup)
	Pools    map[string]map[string][]float64 // metric values per group key
	Counties []string
	Years    []int
}

func NewSite(evs []Evaluation) Site {
	keys, places, global := PersonKeys(evs), Place(evs, PeerGroup), Place(evs, CycleGroup)
	s := Site{Judges: map[string]Judge{}, Bins: map[string][]int{}, Pools: map[string]map[string][]float64{}, Counties: Counties()}
	for i, e := range evs {
		en := Entry{e, keys[i], places[i], global[i]}
		s.Entries = append(s.Entries, en)
		j := s.Judges[en.Key]
		s.Judges[en.Key] = Judge{en.Key, append(j.Entries, en)}
		if !slices.Contains(s.Years, e.Year) {
			s.Years = append(s.Years, e.Year)
		}
		if en.Ranked {
			for g, comp := range map[string]float64{PeerGroup(e): en.Composite, CycleGroup(e): en.Global.Composite} {
				if s.Pools[g] == nil {
					s.Pools[g] = map[string][]float64{}
				}
				for k, v := range map[string]float64{"composite": comp, "overall": en.Metrics.Overall, "categories": en.Metrics.Categories, "atty": en.Metrics.AttyMeets, "responses": float64(e.Survey.Responses)} {
					s.Pools[g][k] = append(s.Pools[g][k], v)
				}
			}
		}
	}
	for k, j := range s.Judges {
		slices.SortFunc(j.Entries, func(a, b Entry) int { return b.Year - a.Year })
		s.Judges[k] = j
	}
	for c, p := range s.Pools {
		s.Bins[c] = bins(p["composite"], 0, 100, 20)
	}
	slices.SortFunc(s.Entries, byStanding)
	slices.Sort(s.Years)
	slices.Reverse(s.Years)
	return s
}

// byStanding orders by cycle (newest first), ranked before unranked, then the headline number.
func byStanding(a, b Entry) int {
	return cmp.Or(b.Year-a.Year, boolOrder(b.Ranked, a.Ranked), cmp.Compare(headline(b).Standing, headline(a).Standing), strings.Compare(a.SortName, b.SortName))
}

func boolOrder(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return 1
	}
	return -1
}

// BallotGroup is one court section of a county's ballot.
type BallotGroup struct {
	Title, ID string
	Entries   []Entry
}

func (s Site) BallotFor(county string, year int) []BallotGroup {
	groups := []BallotGroup{
		{"Statewide · Supreme Court and Court of Appeals", "appellate", nil},
		{"District Court", "district", nil},
		{county + " County Court", "county", nil},
	}
	for _, e := range s.Entries {
		if e.Year != year || !OnBallot(e.Evaluation, county) {
			continue
		}
		g := map[string]int{Appellate: 0, District: 1, County: 2}[CohortOf(e.Level)]
		groups[g].Entries = append(groups[g].Entries, e)
	}
	if len(groups[1].Entries) > 0 {
		groups[1].Title = ordinal(groups[1].Entries[0].District) + " Judicial District · District Court"
	}
	return groups
}

// Glance counts what a voter scanning a set of entries should notice.
type Glance struct{ Total, Low, NotMeet, Split, NoSurvey int }

func GlanceOf(es []Entry) Glance {
	var g Glance
	for _, e := range es {
		g.Total++
		if e.Ranked && headline(e).Standing < LowPct {
			g.Low++
		}
		if e.Recommendation == "does_not_meet" {
			g.NotMeet++
		}
		if e.Vote.Split() {
			g.Split++
		}
		if !e.Ranked {
			g.NoSurvey++
		}
	}
	return g
}

// LowPct marks the bottom quarter of a cohort.
const LowPct = 25

func bins(vs []float64, lo, hi float64, n int) []int {
	out := make([]int, n)
	for _, v := range vs {
		out[binOf(v, lo, hi, n)]++
	}
	return out
}

func binOf(v, lo, hi float64, n int) int {
	if hi <= lo {
		return 0
	}
	return max(0, min(n-1, int(float64(n)*(v-lo)/(hi-lo))))
}

// histSVG draws a cohort histogram with one purple mark at v. Bars fill the width; the marked
// bin and a tick at the exact value carry the reserved "this judge" ink.
func histSVG(counts []int, lo, hi, v float64, w, h int, label string) template.HTML {
	peak := slices.Max(counts)
	bw := float64(w) / float64(len(counts))
	mark := -1 // NaN v draws the distribution with no judge marked
	if !math.IsNaN(v) {
		mark = binOf(v, lo, hi, len(counts))
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg class="hist" viewBox="0 0 %d %d" preserveAspectRatio="none" role="img" aria-label="%s">`, w, h+4, template.HTMLEscapeString(label))
	for i, c := range counts {
		bh := 0.0
		if peak > 0 {
			bh = math.Max(1, float64(h)*float64(c)/float64(peak))
		}
		cls := "c"
		if i == mark {
			cls = "y"
		}
		fmt.Fprintf(&b, `<rect class="%s" x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="1"/>`, cls, float64(i)*bw+1, float64(h)-bh, bw-2, bh)
	}
	if mark >= 0 {
		x := float64(w) * math.Max(0, math.Min(1, (v-lo)/math.Max(hi-lo, 1e-9)))
		fmt.Fprintf(&b, `<rect class="t" x="%.2f" y="0" width="2" height="%d"/>`, math.Min(x, float64(w)-2), h+4)
	}
	b.WriteString("</svg>")
	return template.HTML(b.String())
}

func ordinal(n int) string {
	suf := "th"
	if n%100 < 11 || n%100 > 13 {
		suf = map[int]string{1: "st", 2: "nd", 3: "rd"}[n%10]
		if suf == "" {
			suf = "th"
		}
	}
	return fmt.Sprintf("%d%s", n, suf)
}

func slug(s string) string {
	return strings.Trim(nonAlpha.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

var courtNames = map[string]string{"supreme": "Supreme Court", "appeals": "Court of Appeals", "district": "District Court", "county": "County Court"}

var cohortNames = map[string]string{Appellate: "appellate", District: "district court", County: "county court"}

func (s Site) funcs() template.FuncMap {
	return template.FuncMap{
		"f2":       func(v float64) string { return fmt.Sprintf("%.2f", v) },
		"share":    func(v float64) string { return fmt.Sprintf("%.2f%%", 100*v) },
		"sign":     func(v float64) string { return fmt.Sprintf("%.0f", math.Floor(v)) },
		"ordinal":  ordinal,
		"slug":     slug,
		"court":    func(level string) string { return courtNames[level] },
		"cohort":   func(level string) string { return cohortNames[CohortOf(level)] },
		"standing": standing,
		// medianResponses is the median survey response count among a cycle's published reports.
		"medianResponses": func(year int) int {
			var ns []int
			for _, e := range s.Entries {
				if e.Year == year && e.Survey.Responses > 0 {
					ns = append(ns, e.Survey.Responses)
				}
			}
			if len(ns) == 0 {
				return 0
			}
			slices.Sort(ns)
			return ns[(len(ns)-1)/2]
		},
		"float": func(n int) float64 { return float64(n) },
		"int":   func(v float64) int { return int(v) },
		"resp": func(p Placement) string {
			if p.Small() {
				return fmt.Sprintf("%s most of %d", ordinal(p.RespRank), p.Of)
			}
			return ordinal(int(p.RespPct)) + " percentile"
		},
		"groupHist": func(e Entry, statewide bool) template.HTML {
			g, p, who := PeerGroup(e.Evaluation), e.Placement, fmt.Sprintf("%d %s", e.Year, cohortNames[CohortOf(e.Level)])
			if statewide {
				g, p, who = CycleGroup(e.Evaluation), e.Global, fmt.Sprintf("%d statewide", e.Year)
			}
			return histSVG(s.Bins[g], 0, 100, p.Composite, 320, 64, fmt.Sprintf("Composite index distribution of %d %s judges; this judge ranks %d", p.Of, who, p.Rank))
		},
		"cycleDists": func(year int) []Dist {
			ds := []Dist{{Title: "All courts, statewide", Group: fmt.Sprint(year)}}
			for _, c := range []string{Appellate, District, County} {
				ds = append(ds, Dist{Title: courtGroupNames[c], Group: fmt.Sprintf("%d %s", year, c)})
			}
			for i, d := range ds {
				pool := s.Pools[d.Group]["composite"]
				if len(pool) == 0 {
					continue
				}
				ds[i].N = len(pool)
				ds[i].SVG = histSVG(s.Bins[d.Group], 0, 100, math.NaN(), 320, 64,
					fmt.Sprintf("Composite index distribution of %d %d %s evaluations", len(pool), year, d.Title))
			}
			return ds
		},
		"where": func(e Entry) string {
			switch CohortOf(e.Level) {
			case County:
				return CountyOf(e.Position) + " County"
			case District:
				return ordinal(e.District) + " Judicial District"
			}
			return "Statewide"
		},
		"low":      func(e Entry) bool { return e.Ranked && headline(e).Standing < LowPct },
		"headline": headline,
		"stamp":    func(t time.Time) string { return t.In(mountain).Format("Jan 2, 2006, 3:04 PM MST") },
		"glance":   GlanceOf,
		"ballotYear": func(es []Entry, year int) []Entry {
			return slices.DeleteFunc(slices.Clone(es), func(e Entry) bool { return e.Year != year })
		},
		"flatten": func(gs []BallotGroup) []Entry {
			var es []Entry
			for _, g := range gs {
				es = append(es, g.Entries...)
			}
			return es
		},
		"rowHist": func(e Entry) template.HTML {
			return histSVG(s.Bins[PeerGroup(e.Evaluation)], 0, 100, e.Composite, 200, 40,
				fmt.Sprintf("Composite index %.2f, %s of %d %s evaluations", e.Composite, ordinal(e.Rank), e.Of, fmt.Sprint(e.Year)+" "+cohortNames[CohortOf(e.Level)]))
		},
		"metricHist": func(e Entry, metric string, v float64) template.HTML {
			pool := s.Pools[PeerGroup(e.Evaluation)][metric]
			lo, hi := slices.Min(pool), slices.Max(pool)
			return histSVG(bins(pool, lo, hi, 24), lo, hi, v, 320, 64, fmt.Sprintf("%s distribution across %d %s evaluations", metric, len(pool), fmt.Sprint(e.Year)+" "+cohortNames[CohortOf(e.Level)]))
		},
		"poolMin": func(e Entry, metric string) float64 { return slices.Min(s.Pools[PeerGroup(e.Evaluation)][metric]) },
		"poolMax": func(e Entry, metric string) float64 { return slices.Max(s.Pools[PeerGroup(e.Evaluation)][metric]) },
		"catMedian": func(e Entry, cat string) float64 {
			var vs []float64
			for _, o := range s.Entries {
				if PeerGroup(o.Evaluation) == PeerGroup(e.Evaluation) {
					if v, ok := groupScore(o.Survey.Categories[cat]); ok && o.Survey.Categories[cat] != nil {
						vs = append(vs, v)
					}
				}
			}
			slices.Sort(vs)
			if len(vs) == 0 {
				return 0
			}
			return vs[len(vs)/2]
		},
		"catScore": func(g map[string]float64) float64 { v, _ := groupScore(g); return v },
		// medianGrade is the median overall grade of a cohort's ranked evaluations in one cycle.
		"medianGrade": func(cohort string, year int) string {
			var vs []float64
			for _, e := range s.Entries {
				if e.Ranked && e.Year == year && CohortOf(e.Level) == cohort {
					vs = append(vs, e.Metrics.Overall)
				}
			}
			if len(vs) == 0 {
				return "–"
			}
			slices.Sort(vs)
			return fmt.Sprintf("%.2f", vs[len(vs)/2])
		},
		"catNames": func(e Entry) []string {
			order := CategoryOrder
			var out []string
			for _, c := range order {
				if _, ok := e.Survey.Categories[c]; ok {
					out = append(out, c)
				}
			}
			return out
		},
		"catLo": func(e Entry, cat string) float64 {
			lo, _ := catRange(s.Entries, PeerGroup(e.Evaluation), cat)
			return lo
		},
		"catHi": func(e Entry, cat string) float64 {
			_, hi := catRange(s.Entries, PeerGroup(e.Evaluation), cat)
			return hi
		},
		"pos": func(lo, hi, v float64) string {
			return fmt.Sprintf("%.2f%%", 100*math.Max(0, math.Min(1, (v-lo)/math.Max(hi-lo, 1e-9))))
		},
		"weakest": weakest,
		// short is the one-line court label for the dense table.
		"short": func(e Entry) string {
			switch CohortOf(e.Level) {
			case County:
				return "County · " + CountyOf(e.Position)
			case District:
				return "District · " + ordinal(e.District)
			}
			return courtNames[e.Level]
		},
		"has":   func(m map[string]float64, k string) bool { _, ok := m[k]; return ok },
		"inc":   func(i int) int { return i + 1 },
		"row":   func(root string, e Entry) map[string]any { return map[string]any{"Root": root, "E": e} },
		"years": func() []int { return s.Years },
		"list":  func(xs ...string) []string { return xs },
	}
}

// catRange is the min and max of one category's score across a peer group.
func catRange(es []Entry, group, cat string) (float64, float64) {
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, o := range es {
		if PeerGroup(o.Evaluation) != group || o.Survey.Categories[cat] == nil {
			continue
		}
		if v, ok := groupScore(o.Survey.Categories[cat]); ok {
			lo, hi = math.Min(lo, v), math.Max(hi, v)
		}
	}
	if lo > hi {
		return 0, 4
	}
	return lo, hi
}

// weakest names the composite component with the lowest percentile, with its raw value.
func weakest(e Entry) string {
	parts := []struct {
		pct  float64
		text string
	}{
		{e.Pct.Overall, fmt.Sprintf("survey grade %.2f", e.Metrics.Overall)},
		{e.Pct.Categories, fmt.Sprintf("category mean %.2f", e.Metrics.Categories)},
		{e.Pct.AttyMeets, fmt.Sprintf("attorneys “meets” %.2f%%", 100*e.Metrics.AttyMeets)},
	}
	w := slices.MinFunc(parts, func(a, b struct {
		pct  float64
		text string
	}) int {
		return cmp.Compare(a.pct, b.pct)
	})
	if e.Small() {
		return w.text
	}
	return fmt.Sprintf("%s (%s percentile)", w.text, ordinal(int(w.pct)))
}

// Dist is one cohort's composite distribution for the distributions panel.
type Dist struct {
	Title, Group string
	N            int
	SVG          template.HTML
}

var courtGroupNames = map[string]string{Appellate: "Supreme Court and Court of Appeals", District: "District courts", County: "County courts"}

// mountain is Colorado's time zone, for the scrape and build stamps.
var mountain = func() *time.Location {
	loc, err := time.LoadLocation("America/Denver")
	if err != nil {
		panic(err) // unreachable: main.go embeds time/tzdata
	}
	return loc
}()

// headline is the standing shown in a judge's big square: the peer group's, or statewide when
// the peer group is too small for a percentile (the appellate courts).
func headline(e Entry) Placement {
	if e.Small() {
		return e.Global
	}
	return e.Placement
}

// standing phrases a placement: raw rank for small groups, percentile otherwise.
func standing(p Placement) string {
	if p.Small() {
		return fmt.Sprintf("rank %d of %d", p.Rank, p.Of)
	}
	return fmt.Sprintf("%s percentile, rank %d of %d", ordinal(int(p.Standing)), p.Rank, p.Of)
}

// CategoryOrder is the order the state's reports list categories in.
var CategoryOrder = []string{"Case Management", "Application and Knowledge of Law", "Communications",
	"Diligence", "Demeanor", "Fairness", "General", "Writing", "Inter-Judge Evaluation"}

// page is the data every template receives.
type page struct {
	Site
	Root, Nav, Title, Description string
	County                        string
	Groups                        []BallotGroup
	Judge                         Judge
	Rows                          []Entry
	Year                          int
	Meta                          Meta
}

// Build renders the whole static site as path -> file bytes.
func Build(evs []Evaluation, meta Meta) (map[string][]byte, error) {
	s := NewSite(evs)
	tpl, err := template.New("").Funcs(s.funcs()).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	out := map[string][]byte{}
	render := func(path, name string, p page) error {
		p.Site, p.Year, p.Meta = s, Ballot, meta
		p.Root = strings.Repeat("../", strings.Count(path, "/"))
		var b bytes.Buffer
		if err := tpl.ExecuteTemplate(&b, name, p); err != nil {
			return fmt.Errorf("render %s: %w", path, err)
		}
		out[path] = b.Bytes()
		return nil
	}
	var statewide []Entry
	for _, e := range s.Entries {
		if e.Year == Ballot && CohortOf(e.Level) == Appellate {
			statewide = append(statewide, e)
		}
	}
	pages := []struct {
		path, tpl string
		p         page
	}{
		{"index.html", "index", page{Nav: "ballot", Title: "Find the judges on your ballot", Description: "Colorado judicial retention evaluations, compared and ranked.", Rows: statewide}},
		{"judges/index.html", "judges", page{Nav: "judges", Title: "All judges", Description: "Every Colorado judicial performance evaluation since 2016, ranked against peers.", Rows: s.Entries}},
		{"method/index.html", "method", page{Nav: "method", Title: "How the numbers work", Description: "Sources, the composite index, cohorts and caveats."}},
	}
	for _, c := range s.Counties {
		pages = append(pages, struct {
			path, tpl string
			p         page
		}{"ballot/" + slug(c) + "/index.html", "ballot", page{Nav: "ballot", County: c, Title: c + " County · " + fmt.Sprint(Ballot) + " retention ballot",
			Description: "Every judge on the " + c + " County " + fmt.Sprint(Ballot) + " retention ballot, ranked against peers.", Groups: s.BallotFor(c, Ballot)}})
	}
	for k, j := range s.Judges {
		pages = append(pages, struct {
			path, tpl string
			p         page
		}{"judges/" + k + "/index.html", "judge", page{Nav: "judges", Title: j.Latest().Name, Description: "Judicial performance evaluations of " + j.Latest().Name + ", compared with peers.", Judge: j}})
	}
	for _, pg := range pages {
		if err := render(pg.path, pg.tpl, pg.p); err != nil {
			return nil, err
		}
	}
	err = fs.WalkDir(staticFS, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := staticFS.ReadFile(p)
		out[strings.TrimPrefix(p, "static/")] = b
		return err
	})
	return out, err
}
