# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Static site. Python (uv + pdfplumber) scrapes and parses source pages/PDFs into JSON; a hand-written Go binary renders static HTML/CSS from that JSON. No frontend framework. (User: "If you need programmatic generation just use a hand made Go binary".)

## Users

Primary: Colorado voters filling in the judicial retention section of their ballot, wanting a fast, numbers-first read on each judge name they see. Secondary: researchers and press who want distributions, outliers, history, and methodology.

## Product Purpose

Re-present the Colorado Office of Judicial Performance Evaluation (judicialperformance.colorado.gov) reports as comparable numbers: each judge's survey scores, where they sit relative to peers on the same court level, and a relative ranking. Covers the 2026 cycle plus prior retention cycles. Success: a voter can find their judge and understand how they compare in under a minute; a researcher can trace every number back to the source report.

## Positioning

The official site publishes one PDF per judge with no cross-judge comparison. This site's mechanism is the comparison itself: percentiles and ranks computed across all judges in a cycle and court level, with the official commission recommendation shown alongside.

## Capabilities and Constraints

- Source data: per-judge evaluation pages (commission recommendation, vote, narrative) and Market Decisions survey PDFs (0-4 scale scores: overall, six categories, attorney vs non-attorney, % "meets standards", response counts, prior-year scores).
- Rankings are derived by this site, not the state; must be labeled as such, with methodology shown.
- Small response counts (some judges have ~10 responses) must be visible next to scores.
- Survey formats differ by cycle; cross-cycle comparisons only where the scale is the same.
- Independent, unofficial project; must not present itself as the state office.

## Evidence on Hand

Scraped data only (data/). No endorsements, testimonials, or claims beyond the source reports.

## Product Principles

1. Numbers first, always traceable to the source PDF.
2. Compare like with like (same cycle, same court level).
3. Show uncertainty: response counts and close-together scores are not over-ranked.
4. Neutral voice; the commission's recommendation is reported, not editorialized.

## Accessibility & Inclusion

Public civic information: WCAG 2.2 AA, works on phones, readable without JS, color never the sole carrier of meaning.
