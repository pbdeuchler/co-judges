# Bench Record

A static site that re-presents Colorado's judicial performance evaluations ([judicialperformance.colorado.gov](https://judicialperformance.colorado.gov/)) as comparable numbers. It shows each judge's survey scores, where they sit among peers at the same court level, and a relative rank. The front door is "find my ballot": pick a county, see every judge on its 2026 retention ballot.

Live at **https://co-judges.philip.computer**. Independent and unofficial: the rankings are this project's calculation, not the state's.

## Pipeline

```
judicialperformance.colorado.gov ──fetch.py──▶ cache/ ──parse.py──▶ data/evaluations.json ──gen (Go)──▶ public/
```

| Step | Command | Output |
|---|---|---|
| Mirror pages and PDFs (idempotent) | `uv run --script scrape/fetch.py [years…]` | `cache/<year>/{index,pages,pdf}/` |
| Extract numbers | `uv run --script scrape/parse.py` | `data/evaluations.json`, `data/meta.json` (scrape time) |
| Render site | `go run ./gen -data data/evaluations.json -out public` | `public/` (replaced) |
| Test | `go test ./gen` | |
| Review screenshots | `uv run --python 3.12 --script scripts/shoot.py` | `.impeccable/review/*.png` |
| Deploy | `./deploy.sh [--scrape] [--dry-run]` | S3 + CloudFront |

The site deploys anywhere as static files. Open `public/index.html` directly or serve the directory. JavaScript is optional: it adds county auto-navigation and table sort/filter.

## Deploy

`./deploy.sh` tests, rebuilds `public/` (stamping the build time), syncs it to `s3://co-judges-client`, and invalidates the CloudFront distribution serving `co-judges.philip.computer`. `--scrape` refreshes the data first; `--dry-run` lists the changes and skips the invalidation. Override `BUCKET` or `DOMAIN` with environment variables.

Cache headers: fonts are immutable for a year, CSS/JS/SVG for an hour, pages for five minutes, and every deploy invalidates `/*`.

Hosting is a private S3 bucket read by a CloudFront distribution (origin access control) whose alias is the site's domain. The distribution must use `index.html` as its default root object; the site's own links always name `index.html`, so no other rewriting is needed. Run `deploy.sh` with AWS credentials that can write the bucket and create invalidations.

## Data

`data/evaluations.json` holds one record per judge per cycle (2016, 2018, 2020, 2022, 2024, 2026; 711 records). Each record has the commission recommendation, vote and narrative from the evaluation page, plus survey numbers from that cycle's retention report PDF:

- `overall`: overall grade 0–4 (`all`; 2016–2018 also `att` / `non`)
- `meets`: share answering "meets performance standards" (`att`, `non`; 2016–2018 `all`; 2016 asked about retention)
- `categories`: category scores. Trial courts have six categories, appellate courts have General / Writing / Inter-Judge.
- `responses`, `report_year`, and `history` (annual grades, 2026 reports only)
- `vote` (page level): `{"meets": n, "not": n}`, or `{"unanimous": true}` / `{"majority": true}` when no counts are given; `null` if unstated

The scraper picks the retention report by its title-page year, not by link text, because older pages pair anchors with the wrong labels. Judges whose survey the state withheld are kept with `insufficient_responses: true` and an empty `survey`.

## Method

- **Peer groups:** court level (appellate = Supreme Court + Court of Appeals, district, county) within one cycle. Cycles are not pooled.
- **Composite index:** the mean of three mid-rank percentiles within the peer group: overall grade, mean category score, and % of attorneys saying "meets standards". Rank orders by composite, and ties share a rank.
- **Headline percentile:** the percentile of the composite within the peer group, consistent with rank. Below 25 is flagged as the bottom quarter. Groups of 10 or fewer (every appellate cycle) show their rank within the group as a raw rank ("3 of 7"), and the big number becomes the statewide percentile.
- **Statewide rank:** the same composite recomputed across every court level in the cycle, shown beside the peer rank.
- **Distributions:** each judge page charts the peer and statewide composite distributions with the judge marked; the All judges page charts every cycle statewide and per court level.
- **Response standing:** the percentile (or rank, for small groups) of each judge's survey response count, so thin samples are visible.
- **Commission vote:** parsed as commissioners for and against "meets standards". Displayed (split votes flagged, counted on ballots), not scored.
- **Why within-cycle:** the survey scale changed in 2024 and median trial-court grades fell from about 3.5 to 3.2. Pooling cycles would push recent judges down for a questionnaire change. The Method page prints the per-cycle medians.
- **Caveat:** appellate groups hold only 4–11 judges per cycle, so their percentiles move in large steps.

## Layout

```
scrape/fetch.py, parse.py   Python (uv inline deps)
gen/                        Go generator: model, rank, ballot (county↔district), render, templates/, static/
data/evaluations.json       scraped dataset (committed)
cache/                      raw mirror (not committed)
public/                     build output
deploy.sh                   build + publish to S3/CloudFront
PRODUCT.md, DESIGN.md       product and design-system records
```

## License

MIT. See [LICENSE](LICENSE). The underlying evaluations are public records published by the Colorado Office of Judicial Performance Evaluation.
