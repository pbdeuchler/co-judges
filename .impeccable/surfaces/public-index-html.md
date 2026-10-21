---
version: 1
slug: "public-index-html"
primary_target: "public/index.html"
related_targets: ["gen/templates"]
---

# Bench Record: static site (all pages)

Mode: operate. Audience: Colorado voters with a 2026 mail ballot (primary); researchers/press (depth). Job: find the judges on my county's ballot and see how each compares to peers; drill into one judge across cycles; browse the statewide ranked table; read the method.

Constraints: static HTML from Go templates; works without JS (JS only enhances sorting/filter/select navigation); every derived number traceable to the source PDF; rankings labeled as this site's, not the state's; response counts beside scores.

## Direction contract

THESIS: The ballot as a concourse of decisions: one yellow wayfinding band per decision point, each judge's composite percentile hung at monumental scale. Refuses the voter-guide card grid with score badges.

OWN-WORLD: Saturated sign yellow reserved for wayfinding only (top band, current-location markers, county select); black humanist sans (Fira Sans) in sentence case; numbers in black inset squares at monumental scale; court pictograms as letter squares (A / D / C); warm gray-white ground, white rows with hairline borders; one reserved ink (course purple) marks "this judge" in every distribution and nowhere else.

STORY: Visitor picks a county, sees every judge on that ballot ranked by percentile among same-level peers, notices low outliers instantly, opens a judge to see the three component numbers, attorney vs non-attorney halves, category scores, past cycles and the commission's verdict, then follows the link to the state PDF.

FIRST VIEWPORT: Yellow band (88px, 6px black underline) with mark, nav (My ballot / All judges / Method), county select at right. Below: h1 "Boulder County · 2026 retention ballot", subline, then ranked rows: black percentile square (56px numerals) | court pictogram | name + court | cohort distribution with this judge in purple | survey score + n. Right rail: ballot-at-a-glance counts. Low-percentile rows invert the square (white with black frame).

FORM: Terminal wayfinding signage, bolder reroll leader (dealt position 1). Raises: orienteering (one reserved ink), j-card (attorney/non-attorney hard halves in fixed order), streaming wall (focus dims cohort), type specimen (hierarchy by scale contrast alone). seed:021f84cec7b436a0

Signature interaction: hovering/focusing a judge row or a distribution mark dims the rest of the cohort and lifts that judge's purple mark; distributions are real SVG histograms of the pooled cohort composite.

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
