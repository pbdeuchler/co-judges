// Progressive enhancement only: every page works without this file.
document.documentElement.classList.add("js");

// County select jumps straight to that county's ballot.
for (const form of document.querySelectorAll("form.pick")) {
  const select = form.querySelector("select");
  const go = () => { if (select.value) location.href = `${form.dataset.root}ballot/${select.value}/index.html`; };
  select.addEventListener("change", go);
  form.addEventListener("submit", (e) => { e.preventDefault(); go(); });
}

// All-judges table: filter by cycle, court and text; sort by any column.
const table = document.getElementById("table");
const filters = document.getElementById("filters");
if (table && filters) {
  const body = table.tBodies[0];
  const rows = [...body.rows];
  const text = rows.map((r) => r.textContent.toLowerCase());
  const apply = () => {
    const f = new FormData(filters);
    const year = f.get("year"), cohort = f.get("cohort"), q = f.get("q").trim().toLowerCase();
    let shown = 0;
    rows.forEach((r, i) => {
      const ok = (!year || r.dataset.year === year) && (!cohort || r.dataset.cohort.startsWith(cohort)) && (!q || text[i].includes(q));
      r.hidden = !ok;
      shown += ok;
    });
    filters.elements.count.value = `${shown} of ${rows.length} evaluations`;
    document.querySelectorAll(".dist-row").forEach((d) => { d.hidden = !!year && d.dataset.year !== year; });
    document.getElementById("none").hidden = shown > 0;
  };
  filters.addEventListener("input", apply);
  filters.addEventListener("submit", (e) => e.preventDefault());
  apply();

  const value = (row, i, kind) => {
    const cell = row.cells[i];
    return kind === "num" ? parseFloat(cell.dataset.v ?? cell.textContent) : cell.textContent.trim();
  };
  table.tHead.querySelectorAll("th").forEach((th, i) => {
    th.querySelector("button").addEventListener("click", () => {
      const kind = th.dataset.sort;
      // Numbers other than rank read best high-to-low first.
      const first = kind === "num" && i !== 0 ? "descending" : "ascending";
      const dir = th.getAttribute("aria-sort") === first ? (first === "ascending" ? "descending" : "ascending") : first;
      table.tHead.querySelectorAll("th").forEach((h) => h.removeAttribute("aria-sort"));
      th.setAttribute("aria-sort", dir);
      const sign = dir === "ascending" ? 1 : -1;
      const sorted = [...body.rows].sort((a, b) => {
        const x = value(a, i, kind), y = value(b, i, kind);
        return sign * (kind === "num" ? x - y : x.localeCompare(y));
      });
      body.append(...sorted);
    });
  });
}
