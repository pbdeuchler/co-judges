# CLAUDE.md (compressed agent notes)

Pipeline: scrape/fetch.py (urllib, browser UA required — site 403s default UA; idempotent cache/) → scrape/parse.py (pdfplumber; pure extraction → data/evaluations.json) → gen/ Go (Place/ranking + html/template render → public/). Run py via `uv run --script`; shoot.py needs `--python 3.12` (greenlet won't build on 3.14).

Scrape gotchas:
- 2022 index links judges as /node/NNNN; slugs stored as node-NNNN.
- Old pages misalign anchor text/hrefs → retention PDF chosen by title-page year (`report_year`), appellate allows year-1. Interim reports excluded.
- `&nbsp;` everywhere: normalize whitespace; "Honorable" sometimes its own line.
- PDF layouts per cycle in LAYOUT (parse.py). Categories parsed from whitespace-flattened table with numbers allowed between name words (category_row). Use LAST occurrence of section markers (TOC/legend come first).
- Narrative <p> filtered of page chrome (Retention Year:/Reports:/Please click/*).
- Vote parser (vote()): normalizes number words/dashes/"N commissioners to M"; pair N-M capped at 12 total (years/other numbers); "voting commissioners" excluded from clause counts. Output {meets,not}|{unanimous}|{majority}|null. Votes are display-only (user decision); Vote.Split flags.
- Category value 0.0 = report placeholder for "no eligible raters" (e.g. 2020 Fairness w/ 0 non-attorney responses) → dropped as missing.

Ranking (gen/rank.go): Place(evs, groupFn). Peer = PeerGroup (cycle+cohort), statewide = CycleGroup (cycle) → Entry.Placement + Entry.Global. Within-cycle (user decision). Pools/Bins keyed by both group keys. SmallGroup=10: groups ≤10 display raw peer rank, no peer percentiles; headline(e) = Global when Small → big square + low flag + row order use statewide percentile (label "statewide"). RespPct/RespRank = response-count standing. Composite = mean of mid-rank percentiles (overall, mean category, att meets). Standing = percentile of composite (headline number; LowPct 25). Rank = competition rank by composite.
Why: 2024+ survey scale lowered trial-court medians ~0.3; method page shows per-cycle medians (medianGrade). Appellate groups 4–11/cycle → coarse percentiles.

Person keys: last-first (suffix jr/iii dropped); names colliding within a cycle get -district suffix.
Ballot: district→counties map in ballot.go (23rd split from 18th in 2025; DistrictCounties handles year). County judges matched by "X County (Court )?Judge" in position.

Design: .caveat panel (layout.html, every page, under band) = plain-language data disclaimer, ~9th-grade; median response count computed (medianResponses). No court letter squares (read as grades; user removed). CSS class .split is the page layout — split-vote marker is .vote-split. impeccable operate surface; contract in .impeccable/surfaces/public-index-html.md; DESIGN.md + .impeccable/design.json record the system. Light only (dark mode removed per finish review; contract ground). Yellow = wayfinding: top band + court/cycle group headers (.sign). Purple (--you) = "this judge" marks only; focus ring ink. Category compare = dot plot on cohort min–max (catRange); mobile shows median as text. Fira Sans self-hosted (gen/static/fonts). Numbers 2dp except monumental percentile sign (integer floor).
Tests: go test ./gen (table-driven, happy+sad).
Deploy: ./deploy.sh (go test → gen → aws s3 sync to co-judges-client w/ per-type Cache-Control → invalidate dist found by alias co-judges.philip.computer). Infra in ../terraform/prod: client_sites["co-judges"] (cloudfront.tf; `root` field = default_root_object) + project bucket co-judges-client (projects.tf). Links must use explicit index.html (no SPA fallback, S3 OAC → missing key = 403).
Stamps: fetch.py writes cache/fetched_at; parse.py → data/meta.json {scraped_at, parsed_at} (fallback newest cache mtime). gen: LoadMeta + BuiltAt=time.Now() injected via run(); Build(evs, meta) stays pure. Shown in .caveat .stamp, America/Denver (time/tzdata embedded).
