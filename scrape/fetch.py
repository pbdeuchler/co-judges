"""Mirror judicialperformance.colorado.gov evaluation pages and survey PDFs into cache/.

Idempotent: files already on disk are not re-downloaded.
Usage: uv run --script scrape/fetch.py [years...]
"""
# /// script
# requires-python = ">=3.11"
# ///
import re
import sys
from datetime import datetime, timezone
import urllib.parse
import urllib.request
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

BASE = "https://judicialperformance.colorado.gov"
UA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36"
CACHE = Path(__file__).resolve().parent.parent / "cache"
YEARS = ["2016", "2018", "2020", "2022", "2024", "2026"]
INDEX_PATHS = [
    "/know-your-judges/{y}-judicial-performance-evaluations",
    "/{y}-judicial-performance-evaluations",
    "/{y}-judicial-performance-evaluations-full-list",
]


def get(url: str) -> bytes | None:
    req = urllib.request.Request(url, headers={"User-Agent": UA, "Accept": "*/*"})
    try:
        with urllib.request.urlopen(req, timeout=60) as r:
            return r.read()
    except Exception as e:  # noqa: BLE001 - report and continue; a missing page is data
        print(f"  ! {url}: {e}", file=sys.stderr)
        return None


def cached(url: str, dest: Path) -> bytes | None:
    if dest.exists() and dest.stat().st_size:
        return dest.read_bytes()
    body = get(url)
    if body:
        dest.parent.mkdir(parents=True, exist_ok=True)
        dest.write_bytes(body)
    return body


def pdf_name(href: str) -> str:
    return urllib.parse.unquote(href.rsplit("/", 1)[-1])


def judge_slugs(year: str) -> list[str]:
    pages = (cached(BASE + p.format(y=year), CACHE / year / "index" / (p.format(y=year).strip("/").replace("/", "_") + ".html"))
             for p in INDEX_PATHS)
    # Most cycles link judges by slug; 2022 links some by bare Drupal node id.
    found = (re.findall(rf"/([a-z0-9-]+-{year}-evaluation|node/\d+)\b", b.decode("utf-8", "replace")) for b in pages if b)
    return sorted({s for f in found for s in f})


def fetch_judge(year: str, slug: str) -> int:
    page = cached(f"{BASE}/{slug}", CACHE / year / "pages" / f"{slug.replace('/', '-')}.html")
    if not page:
        return 0
    hrefs = set(re.findall(r'href="([^"]+\.pdf)"', page.decode("utf-8", "replace"), re.I))
    pdfs = [h for h in hrefs if not re.search(r"Post-Evaluation|Smart%20Act|factsheet|Transcript", h, re.I)]
    for h in pdfs:
        url = h if h.startswith("http") else BASE + h
        cached(url, CACHE / year / "pdf" / slug.replace("/", "-") / pdf_name(h))
    return len(pdfs)


def main(years: list[str]) -> None:
    for y in years:
        slugs = judge_slugs(y)
        print(f"{y}: {len(slugs)} judges")
        with ThreadPoolExecutor(8) as ex:
            n = sum(ex.map(lambda s: fetch_judge(y, s), slugs))
        print(f"{y}: {n} pdfs")
    # parse.py publishes this as the site's "last scraped" time
    (CACHE / "fetched_at").write_text(datetime.now(timezone.utc).isoformat(timespec="seconds"))


if __name__ == "__main__":
    main(sys.argv[1:] or YEARS)
