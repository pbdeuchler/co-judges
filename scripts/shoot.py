"""Capture review screenshots of the built site with the locally installed Chrome.

Usage: uv run --script scripts/shoot.py
"""
# /// script
# requires-python = ">=3.11,<3.14"
# dependencies = ["playwright==1.55.0"]
# ///
from pathlib import Path

from playwright.sync_api import sync_playwright

ROOT = Path(__file__).resolve().parent.parent
SITE, OUT = ROOT / "public", ROOT / ".impeccable" / "review"


def main() -> None:
    judge = next((SITE / "judges").glob("kotlarczyk*")) / "index.html"
    shots = [
        ("ballot/boulder/index.html", 1440, "desktop.png"), ("ballot/boulder/index.html", 390, "mobile.png"),
        (str(judge.relative_to(SITE)), 1440, "judge-desktop.png"), (str(judge.relative_to(SITE)), 390, "judge-mobile.png"),
        ("index.html", 1440, "index-desktop.png"), ("judges/index.html", 1440, "table-desktop.png"),
        ("ballot/washington/index.html", 1440, "empty-desktop.png"),
    ]
    OUT.mkdir(parents=True, exist_ok=True)
    with sync_playwright() as p:
        browser = p.chromium.launch(channel="chrome")
        for path, width, name in shots:
            page = browser.new_page(viewport={"width": width, "height": 900}, device_scale_factor=1)
            page.goto((SITE / path).as_uri())
            page.wait_for_load_state("networkidle")
            overflow = page.evaluate("document.documentElement.scrollWidth - innerWidth")
            page.screenshot(path=OUT / name, full_page=True)
            print(f"{name}: horizontal overflow {overflow}px")
            page.close()
        browser.close()


if __name__ == "__main__":
    main()
