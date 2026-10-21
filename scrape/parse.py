"""Parse cached evaluation pages + survey PDFs into data/evaluations.json.

Pure extraction: no ranking or derived statistics (the Go site generator owns those).
Usage: uv run --script scrape/parse.py
"""
# /// script
# requires-python = ">=3.11"
# dependencies = ["pdfplumber==0.11.4"]
# ///
import html
import json
import re
import sys
from datetime import datetime, timezone
import urllib.parse
from concurrent.futures import ProcessPoolExecutor
from pathlib import Path

import pdfplumber

ROOT = Path(__file__).resolve().parent.parent
CACHE = ROOT / "cache"
OUT = ROOT / "data" / "evaluations.json"
META = ROOT / "data" / "meta.json"
BASE = "https://judicialperformance.colorado.gov"

CATEGORIES = {
    "trial": ["Case Management", "Application and Knowledge of Law", "Communications",
              "Diligence", "Demeanor", "Fairness"],
    "appellate": ["General", "Writing", "Inter-Judge Evaluation"],
}
ORDINALS = {w: i for i, w in enumerate(
    "First Second Third Fourth Fifth Sixth Seventh Eighth Ninth Tenth Eleventh Twelfth Thirteenth "
    "Fourteenth Fifteenth Sixteenth Seventeenth Eighteenth Nineteenth Twentieth Twenty-First "
    "Twenty-Second Twenty-Third".split(), 1)}
NUM = r"(?:\d+\.\d+|\d+|N/A|-)"

# Column layouts of each cycle's trial-court summary tables, in the order values appear in the PDF
# text. Trailing columns not named (response counts, statewide averages) are dropped.
LAYOUT = {
    "2016": {"overall": ["all", "att", "non"], "meets": ["all", "att", "non"], "cat": ["att", "non"]},
    "2018": {"overall": ["all", "att", "non"], "meets": ["all", "att", "non"], "cat": ["att", "non"]},
    "2020": {"overall": ["all"], "meets": ["att", "peer_att", "non", "peer_non"], "cat": ["all"]},
    "2022": {"overall": ["all"], "meets": ["att", "peer_att", "non", "peer_non"], "cat": ["all"]},
    "2024": {"overall": ["all"], "meets": ["att", "non"], "cat": ["all"]},
    "2026": {"overall": ["all"], "meets": ["att", "non"], "cat": ["all"]},
}


def num(s: str) -> float | None:
    return None if s in ("N/A", "-") else float(s)


def label(keys: list[str], vals: list[str]) -> dict:
    return {k: num(v) for k, v in zip(keys, vals) if num(v) is not None}


# ---------- HTML ----------

def text_lines(fragment: str) -> list[str]:
    s = re.sub(r"<(script|style)[^>]*>.*?</\1>", "", fragment, flags=re.S)
    s = html.unescape(re.sub(r"<[^>]+>", "\n", s))
    return [re.sub(r"\s+", " ", l).strip() for l in s.splitlines() if l.strip()]


def court_level(position: str) -> str:
    if "Supreme" in position:
        return "supreme"
    if "Appeals" in position:
        return "appeals"
    return "county" if re.search(r"County (Court )?Judge", position) else "district"


NUMBER_WORDS = {w: str(i) for i, w in enumerate(
    "zero one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen".split())}


def vote(text: str, rec: str) -> dict | None:
    """Commission vote as {"meets": n, "not": n} commissioners; {"unanimous": True} or
    {"majority": True} when the page gives no counts. Absent/recused members are not votes."""
    t = re.sub(r"[\u2013\u2014]", "-", text.lower())
    t = re.sub(r"\b(" + "|".join(NUMBER_WORDS) + r")\b", lambda m: NUMBER_WORDS[m.group(1)], t)
    t = re.sub(r"(\d)\s+(?:commissioners?\s+)?to\s+(\d)", r"\1-\2", t)
    # "10-0", "(7-2)", "by a vote of 5-3": counts for and against the recommendation. Bounds skip years.
    for m in re.finditer(r"\b(\d{1,2})\s*-\s*(\d{1,2})\b", t):
        a, b = int(m.group(1)), int(m.group(2))
        if a + b <= 12 and a + b >= 2:  # commissions seat at most 11-12; larger pairs are other numbers
            return {"meets": a, "not": b} if rec != "does_not_meet" else {"meets": b, "not": a}
    # "Six members ... voted that ... met ..., one voted does not meet", "with one voting no".
    meets = against = 0
    for n, clause in re.findall(r"\b(\d{1,2})\s+(?:members?(?:\s+of\s+the\s+commission)?\s+|commissioners?\s+)?(?:voted|voting)\b(?!\s+commissioners?)([^,.;]*?)(?=,|;|\.|\band\b|\bwhile\b|$)", t):
        if re.search(r"\bnot\b|\bno\b", clause):
            against += int(n)
        else:
            meets += int(n)
    if meets + against:
        return {"meets": meets, "not": against}
    if m := re.search(r"\b(\d{1,2})\s+members?\s+of\b[^.]*?\bagree", t):
        return {"meets": int(m.group(1)), "not": 0} if rec != "does_not_meet" else {"meets": 0, "not": int(m.group(1))}
    if "unanimous" in t:
        return {"unanimous": True}
    return {"majority": True} if "majority vote" in t else None


def recommendation(raw: str) -> str:
    r = raw.lower()
    if "not meet" in r or "not retain" in r:
        return "does_not_meet"
    if "meets" in r or "retain" in r:
        return "meets"
    return "no_opinion" if "opinion" in r else r


def parse_page(year: str, path: Path) -> dict | None:
    doc = path.read_text("utf-8", "replace")
    main = (re.search(r"<main.*?</main>", doc, re.S) or re.search(r".*", doc, re.S)).group(0)
    lines = text_lines(main)
    if "Retention Year:" not in " ".join(lines[:12]):
        return None
    i_name = next(i for i, l in enumerate(lines) if l.startswith("Honorable"))
    position = lines[i_name - 1]
    # Some pages put the name in its own element after "Honorable".
    name = (lines[i_name] if lines[i_name] != "Honorable" else lines[i_name + 1]).removeprefix("Honorable ").strip()
    dm = re.match(r"([A-Z][a-z]+(?:-[A-Z][a-z]+)?) Judicial District", position)
    i_rec = lines.index("Recommendation:")
    body_start = next((i for i, l in enumerate(lines) if "Commission" in l and i > i_rec), i_rec + 2)
    paras = [p for p in (re.findall(r"<p[^>]*>(.*?)</p>", main, re.S)) if p.strip()]
    narrative = [" ".join(text_lines(p)) for p in paras]
    # Drop page chrome (header block, report links, accessibility notice) that shares <p> tags.
    chrome = re.compile(r"^(Retention Year:|Reports:|Please click|Need an accessible|\*)|accessible PDF")
    narrative = [re.sub(r"\s+([.,;])", r"\1", p) for p in narrative if len(p) > 40 and not chrome.search(p)]
    hrefs = re.findall(r'href="([^"]+\.pdf)"', main, re.I)
    pdfs = {urllib.parse.unquote(h.rsplit("/", 1)[-1]): h if h.startswith("http") else BASE + h for h in hrefs}
    return {
        "year": int(year),
        "slug": path.stem,
        "url": f"{BASE}/{path.stem.replace('node-', 'node/')}",
        "name": name,
        "insufficient_responses": bool(re.search(r"insufficient(?:\s|&nbsp;)+respons", main, re.I)),
        "sort_name": lines[0 if lines[0] != "Home" else 1].split(f" {year} Evaluation")[0],
        "position": position,
        "level": court_level(position),
        "district": ORDINALS.get(dm.group(1)) if dm else None,
        "recommendation": recommendation(lines[i_rec + 1]),
        "recommendation_text": lines[i_rec + 1],
        "vote": vote(" ".join(narrative[:2]), recommendation(lines[i_rec + 1])),
        "narrative": narrative,
        "_pdfs": {CACHE / year / "pdf" / path.stem / n: u for n, u in pdfs.items()},
    }


# ---------- PDF ----------

def pdf_text(path: Path, pages: int = 9) -> str:
    with pdfplumber.open(path) as pdf:
        return "\n".join(p.extract_text() or "" for p in pdf.pages[:pages])


def category_row(table: str, name: str) -> list[str] | None:
    """Values of the last row for `name` in a whitespace-flattened table.

    pdfplumber wraps long names around their numbers ("Case 3.8 3.4 Management"), so numbers may
    sit between any two words of the name or after it.
    """
    nums = rf"((?:{NUM}\s+)+)"
    pat = rf"\s+(?:{nums})?".join(map(re.escape, name.split())) + rf"(?:\s+\(Attorneys\))?\s+(?:{nums})?"
    rows = [next((g for g in m.groups() if g), "") for m in re.finditer(pat, table + " ")]
    return rows[-1].split() if rows and rows[-1] else None


def report_year(t: str) -> tuple[int, bool] | None:
    """(survey year, is_interim) from the report's title page."""
    m = re.search(r"(20\d\d)\s+(Interim\s+)?(?:Judicial Performance Survey|Retention)", t[:3000])
    return (int(m.group(1)), bool(m.group(2))) if m else None


def choose_report(year: int, appellate: bool, texts: dict) -> Path | None:
    """The cycle's retention report. Appellate surveys are sometimes fielded the prior year."""
    ok = [year, year - 1] if appellate else [year]
    years = {p: report_year(t) for p, t in texts.items()}
    ranked = sorted((ok.index(ry[0]), str(p)) for p, ry in years.items() if ry and not ry[1] and ry[0] in ok)
    return Path(ranked[0][1]) if ranked else None


def section(t: str, start: str, end: str) -> str:
    # Last occurrence: earlier ones are the table of contents or a chart legend.
    i = t.rfind(start)
    j = t.find(end, i + 1) if i >= 0 else -1
    return t[i:j if j > 0 else None] if i >= 0 else ""


def parse_pdf(year: str, appellate: bool, t: str) -> dict:
    lay = LAYOUT[year] | ({"cat": ["all"]} if appellate else {})
    overall = re.search(r"^Overall (?:Average )?Grade:?\s+(.+)$", t, re.M)
    meets_block = section(t, "Yes, meets", "No, does not") or section(t, "% Recommending", "Retention\n")
    cat_page = section(t, "Overall Category Scores", "Summary of Responses") or section(
        t, "Individual Category Scores", "Detailed Report") or section(t, "Individual Category Scores", "Summary of Responses")
    table = " ".join(cat_page.split())
    rows = {c: category_row(table, c) for c in CATEGORIES["appellate" if appellate else "trial"]}
    # Reports print 0.0 for a category nobody in the eligible group rated (e.g. 2020 Fairness,
    # asked only of non-attorneys, for a judge with no non-attorney responses): that is missing data.
    cats = {c: g for c, r in rows.items() if r and (g := {k: v for k, v in label(lay["cat"], r).items() if v > 0})}
    resp = re.search(r"reflects these (\d+) responses", t)
    history = re.findall(r"^(20\d\d) (\d\.\d) (\d+)$", section(t, "Previous Years", "Performance Scores"), re.M)
    return {
        "responses": int(resp.group(1)) if resp else None,
        "overall": label(lay["overall"], re.findall(NUM, overall.group(1))) if overall else {},
        "meets": {k: v / 100 for k, v in label(lay["meets"], re.findall(r"(\d+)%", meets_block)).items()},
        "categories": cats,
        "history": [{"year": int(y), "score": float(s), "n": int(n)} for y, s, n in history],
    }


def parse_one(arg: tuple[str, Path]) -> dict | None:
    year, page = arg
    try:
        rec = parse_page(year, page)
    except Exception as e:  # noqa: BLE001 - one malformed page must not sink the run
        print(f"  ! {page}: {e!r}", file=sys.stderr)
        return None
    if rec is None:
        return None
    pdfs = rec.pop("_pdfs")
    texts = {}
    for p in pdfs:
        try:
            texts[p] = pdf_text(p) if p.exists() else ""
        except Exception as e:  # noqa: BLE001 - a corrupt PDF is reported as a gap, not a crash
            print(f"  ! {p}: {e}", file=sys.stderr)
    appellate = rec["level"] in ("supreme", "appeals")
    chosen = choose_report(int(year), appellate, texts)
    survey = parse_pdf(year, appellate, texts[chosen]) | {"report_year": report_year(texts[chosen])[0]} if chosen else {}
    return rec | {"report_url": pdfs.get(chosen), "survey": survey}


def main() -> None:
    jobs = [(p.parent.parent.name, p) for p in sorted(CACHE.glob("20*/pages/*.html"))]
    with ProcessPoolExecutor() as ex:
        recs = [r for r in ex.map(parse_one, jobs, chunksize=4) if r]
    OUT.parent.mkdir(exist_ok=True)
    OUT.write_text(json.dumps(recs, indent=1, ensure_ascii=False))
    print(f"{len(recs)} evaluations -> {OUT.relative_to(ROOT)}")
    META.write_text(json.dumps({"scraped_at": scraped_at(), "parsed_at": now()}, indent=1) + "\n")


def now() -> str:
    return datetime.now(timezone.utc).isoformat(timespec="seconds")


def scraped_at() -> str:
    """When fetch.py last ran; caches from before it recorded that fall back to the newest file."""
    stamp = CACHE / "fetched_at"
    if stamp.exists():
        return stamp.read_text().strip()
    newest = max(p.stat().st_mtime for p in CACHE.glob("20*/**/*") if p.is_file())
    return datetime.fromtimestamp(newest, timezone.utc).isoformat(timespec="seconds")


if __name__ == "__main__":
    main()
