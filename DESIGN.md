---
name: Bench Record
description: Colorado judicial retention evaluations as comparable numbers, signed like a transit concourse.
colors:
  sign: "#ffcc00"
  ink: "#111111"
  ink-2: "#4f4e49"
  ground: "#f1f1ee"
  panel: "#ffffff"
  rule: "#d9d8d2"
  ctx: "#b9b7ae"
  tint: "#e6e5df"
  you: "#6a2c91"
  alert: "#a3201b"
  alert-soft: "#fbe3e1"
typography:
  numeral:
    fontFamily: "Fira Sans, system-ui, sans-serif"
    fontSize: "clamp(2.5rem, 2rem + 2vw, 3.5rem)"
    fontWeight: 800
    lineHeight: 1
    letterSpacing: "-0.02em"
    fontFeature: "\"tnum\""
  display:
    fontFamily: "Fira Sans, system-ui, sans-serif"
    fontSize: "clamp(1.75rem, 1.2rem + 2vw, 2.5rem)"
    fontWeight: 700
    lineHeight: 1.15
    letterSpacing: "-0.01em"
  title:
    fontFamily: "Fira Sans, system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 700
    lineHeight: 1.2
  headline:
    fontFamily: "Fira Sans, system-ui, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 700
    lineHeight: 1.5
  body:
    fontFamily: "Fira Sans, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.5
    fontFeature: "\"tnum\""
  small:
    fontFamily: "Fira Sans, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: "Fira Sans, system-ui, sans-serif"
    fontSize: "0.6875rem"
    fontWeight: 500
    lineHeight: 1
    letterSpacing: "0.08em"
rounded:
  none: "0px"
  dot: "50%"
spacing:
  s1: "4px"
  s2: "8px"
  s3: "12px"
  s4: "16px"
  s5: "24px"
  s6: "32px"
  s7: "48px"
  s8: "72px"
components:
  band:
    backgroundColor: "{colors.sign}"
    textColor: "{colors.ink}"
    padding: "16px 48px"
    height: "88px"
  sign-panel:
    backgroundColor: "{colors.sign}"
    textColor: "{colors.ink}"
    typography: "{typography.headline}"
    padding: "12px 16px"
  percentile-square:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.panel}"
    typography: "{typography.numeral}"
    rounded: "{rounded.none}"
    width: "112px"
    height: "88px"
  percentile-square-low:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
  ranked-row:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "12px 24px 12px 12px"
  select:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "8px 12px"
    height: "44px"
  metric-card:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    rounded: "{rounded.none}"
    padding: "16px"
  footer:
    backgroundColor: "{colors.ink}"
    textColor: "{colors.panel}"
    typography: "{typography.small}"
    padding: "24px 48px"
---

# Design System: Bench Record

## Overview

**Creative North Star: "The Concourse of Decisions"**

The site is terminal wayfinding signage applied to a ballot. Each decision point (a court group on the ballot, a cycle on a judge's page) is announced by a saturated yellow sign, the way an airport labels a gate. Under each sign, judges hang as white rows on a warm gray-white ground, each led by its percentile in a black inset square at monumental scale. Hierarchy comes from scale contrast inside one humanist sans, not from color or decoration.

Color is scarce and assigned, not themed. Yellow means "you are here / go here". Purple means "this judge" and nothing else. Everything else is ink, gray, and white. The comparison itself (where one judge sits in a real distribution of peers) is the visual signature: histograms of the peer group (same cycle and court level), gray bars with one purple mark.

The world is light-ground only, flat, square-cornered, and dense in the way a departures board is dense: many rows, each instantly scannable left to right (number, court, name, distribution, grade and response count).

**Key Characteristics:**
- Yellow wayfinding signs head every decision group. No court letter squares: single letters read as grades next to rankings.
- Percentile numerals in black squares at 40–56px, weight 800, tabular figures.
- One reserved ink (purple) marks this judge in every chart.
- Square corners, hairline borders, no shadows for depth.
- Fira Sans in sentence case throughout; hierarchy by size and weight alone.

## Colors

A near-achromatic ink-and-paper system carrying two reserved signal colors, each with exactly one job.

### Primary
- **Sign Yellow** (sign): Wayfinding only. The top band, court-group and cycle sign panels, the skip link, text selection, county-index hover, and the yellow rule that caps the footer. Never in data, never as decoration inside a row.

### Secondary
- **Course Purple** (you): "This judge". The judge's bin and position tick in every histogram, the judge's dot on category range plots. Nowhere else: not links, not focus, not headings.

### Tertiary
- **Verdict Red** (alert) on **Verdict Blush** (alert-soft): Only the commission's "does not meet standards" flag, always paired with a triangle icon and the words, never color alone.

### Neutral
- **Ink** (ink): All text, the percentile squares, the footer field, strong rules, row hover borders, and the focus ring. The tokens `--sign-ink`, `--block`, `--rule-2`, and `--focus` are aliases of this value; `--block-ink` aliases panel white.
- **Graphite** (ink-2): Secondary text: court and location lines, captions, axis labels, response counts.
- **Concourse Ground** (ground): Page background, a warm gray-white.
- **Panel White** (panel): Rows, cards, tables, inputs.
- **Hairline** (rule): 1px borders around rows, cards, tables and between table rows.
- **Cohort Gray** (ctx): The peer bars in histograms and the range line in category plots: context that must recede behind the purple mark.
- **Row Tint** (tint): Table row hover.

### Named Rules
**The Wayfinding Yellow Rule.** Yellow appears only on surfaces that orient: where you are, which group you are in, where to go next. If removing it would not change how someone navigates, it does not belong.

**The One Reserved Ink Rule.** Purple marks this judge and only this judge. A chart without a "this judge" has no purple.

## Typography

**Display Font:** Fira Sans (with system-ui, sans-serif), self-hosted at 400, 500, 700, 800
**Body Font:** Fira Sans
**Label/Mono Font:** none; tabular figures (`font-variant-numeric: tabular-nums`) are set on body so every number column aligns.

**Character:** One humanist sans used like signage: heavy and tight at the top, plain and legible underneath, always sentence case.

### Hierarchy
- **Numeral** (800, clamp 40–56px, line-height 1, -0.02em): The percentile inside the black square. The glance-rail counts (40px, 800), metric values (32px, 800), and survey grade (24px, 700) step down from it.
- **Display** (700, clamp 28–40px, 1.15, -0.01em): Page h1 only.
- **Headline** (700, 18px): Sign panels, rail and section headings (`--h2`). Prose h2 runs slightly larger (22px).
- **Title** (700, 20px, 1.2): Judge name in a row.
- **Body** (400, 16px, 1.5): Running text; prose capped at 60–76ch.
- **Small** (400, 14px): Court lines, table text, footer, notes. Rank and axis text runs 12–13px in graphite.
- **Label** (500, 11px, 0.08em, uppercase): Only the unit word under a percentile numeral ("percentile", "no survey"), a measurement unit bound to the number it labels.

### Named Rules
**The Scale Contrast Rule.** Hierarchy is made by size and weight jumps within Fira Sans, never by a second family or by color.

**The Sentence Case Rule.** Headings, nav, and signs are sentence case. Uppercase is limited to the unit under a numeral.

## Layout

Content sits in a 1440px max container with 48px side padding (16px under 720px) and 32px top padding. The ballot page is a two-column split: ranked rows (fluid) and a 340px sticky "at a glance" rail, 48px gap. Groups are separated by 48px; rows inside a group by 8px.

A ranked row is a four-column grid: 112px percentile square | fluid name block | 240px distribution with peer and statewide rank | 150px grade, response count and response standing, 24px gaps. At 1100px the rail moves above the rows; the row reflows into square + name + grade over a full-width distribution. At 720px the row becomes square + name, then distribution, then grade inline; nav wraps to its own line in the band.

Spacing follows a 4px-based scale (4, 8, 12, 16, 24, 32, 48, 72), with 72px reserved for page bottom padding.

## Elevation & Depth

Flat. There are no drop shadows anywhere. Depth and grouping come from the ground/panel value step (gray-white page, white rows), 1px hairlines, and heavy black bars at the boundaries (the band's 6px ink underline, the footer's 6px yellow cap). `box-shadow` is used only as an inset frame: the 4px ink frame of a low-percentile square, the 2px frame on a low score in tables, and a 2px white halo separating the purple dot from its range line.

### Named Rules
**The Flat Concourse Rule.** Surfaces never lift. State is shown by border weight (hairline to ink on hover) and by dimming context, not by elevation.

## Shapes

Square-cornered everywhere in chrome: rows, cards, inputs, buttons, squares, signs (0px). Percentile squares are true rectangles with heavy ink. The only round forms are data marks: the purple dot on category range plots (50%) and a 1px corner softening on histogram bars. Borders are either hairline (1px rule), structural (2px ink on inputs and table heads), or framing (3px on letter squares, 4px inset on low squares, 6px on band and footer edges).

## Components

### Wayfinding band
Yellow field, 88px minimum, 6px ink underline. Holds the mark (52px ink square with a yellow "J", then "Bench Record" at 26px/800), the primary nav at 18px/500, and the county picker pushed right. Nav links are undecorated; hover shows a 3px translucent ink underline, the current page a 3px solid ink underline and 700 weight.

### Sign panel
A yellow bar heading each decision group (court group on the ballot, cycle on the judge page): 18px/700 ink text, 12px 16px padding. Sits directly above its rows with 12px below.

### Percentile square
Black inset square (112 x 88px; 80 x 76px on phones) with the numeral in white at 800 and the unit label beneath. Lowest-quarter judges invert it: white with a 4px ink inset frame, so "low" reads without color. Unranked judges get a transparent square with a hairline frame and an en dash. Peer groups of 10 or fewer show the raw rank instead of a percentile: the numeral is the rank and the unit label reads "of N".

### Split-vote marker
Inline 700 ink text with a 1em plus-shaped icon and a 2px yellow underline: "Split vote 5–4". Shown in the commission line whenever the commission's vote was not unanimous; display only, never scored. "Does not meet" keeps the red flag.

### Ranked row
The whole row is one link: white panel, 1px hairline border, hover/focus border goes ink (0.2s, cubic-bezier(.16,1,.3,1)). Commission "does not meet" flags render in alert red with a triangle icon. Low rows add a bold "Weakest:" line.

### Cohort histogram (signature)
An SVG of the peer group's composite distribution (same cycle, same court level): gray bars, the judge's bin in purple, a 2px purple tick at the judge's exact value extending 4px past the baseline. 44px tall in rows, 64px in metric cards, with min/max axis labels beneath on cards. On row hover/focus the gray bars fade to 35% (0.35s) and the purple tick grows 15% from the baseline; reduced-motion removes all transitions.

### Category range plot
A dot plot per category: a 2px cohort-gray line spanning the peer min–max, an 18px ink tick at the peer median, a 14px purple dot for this judge with a 2px white halo, min and max printed beneath.

### Inputs and buttons
White, 2px ink border, square corners, 44px minimum height, 8px x 12px padding, inherited font. Buttons are the same shape at 700 weight. No hover fill; focus is the global ring.

### Focus
3px solid ink outline, 2px offset, on every focusable element.

### Cards and tables
Metric cards and tables are white panels with 1px hairline borders and 16px padding. Table heads carry a 2px ink bottom border and stick; rows divide by hairlines and tint on hover; numeric columns right-align. Attorney / non-attorney "halves" are a hard two-column split, attorneys first, divided by a hairline.

### Footer
Ink field with white 14px text, capped by a 6px yellow rule.

## Do's and Don'ts

### Do:
- **Do** head every decision group with a yellow sign panel.
- **Don't** add court letter squares (A/D/C); next to rankings they read as letter grades.
- **Do** lead every judge with the percentile in a black square, and invert to a white square with a 4px ink frame for the lowest quarter.
- **Do** show this judge in purple (#6a2c91) against cohort gray in every distribution, and keep response counts next to every score.
- **Do** keep corners square and depth flat; signal hover by switching the hairline border to ink.
- **Do** pair every color signal with a shape or words (inverted square, triangle plus "does not meet standards").

### Don't:
- **Don't** use yellow inside data, on cards, or as a highlight for emphasis.
- **Don't** use purple for links, focus, headings, or anything other than this judge.
- **Don't** add drop shadows or rounded corners to chrome.
- **Don't** introduce a second typeface or set headings in uppercase.
- **Don't** add a dark theme; the world is light ground only.
