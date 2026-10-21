package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	dir := t.TempDir()
	data, meta, out := filepath.Join(dir, "e.json"), filepath.Join(dir, "m.json"), filepath.Join(dir, "public")
	os.WriteFile(data, []byte(`[{"year":2026,"name":"A","sort_name":"A, B","level":"county","position":"Twentieth Judicial District - Boulder County Court Judge"}]`), 0o644)
	os.WriteFile(meta, []byte(`{"scraped_at":"2026-10-21T04:04:14Z"}`), 0o644)
	if err := run(data, meta, out, time.Date(2026, 10, 21, 6, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("run: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(out, "ballot", "boulder", "index.html"))
	if err != nil || !strings.Contains(string(b), "Site updated") {
		t.Errorf("run did not write a stamped ballot page: %v", err)
	}
	tests := map[string][3]string{
		"missing data": {filepath.Join(dir, "nope.json"), meta, out},
		"missing meta": {data, filepath.Join(dir, "nope.json"), out},
		"bad meta":     {data, data, out}, // an array is not a meta object
	}
	for name, a := range tests {
		if err := run(a[0], a[1], a[2], time.Now()); err == nil {
			t.Errorf("run with %s succeeded, want error", name)
		}
	}
}
