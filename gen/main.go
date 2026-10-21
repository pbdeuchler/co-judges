// Command gen renders the Bench Record static site from data/evaluations.json.
//
//	go run ./gen -data data/evaluations.json -out public
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "time/tzdata" // stamps render in Mountain Time on any build host
)

func main() {
	in := flag.String("data", "data/evaluations.json", "scraped evaluations JSON")
	metaPath := flag.String("meta", "data/meta.json", "scrape metadata JSON")
	outDir := flag.String("out", "public", "output directory (replaced)")
	flag.Parse()
	if err := run(*in, *metaPath, *outDir, time.Now()); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func run(in, metaPath, outDir string, now time.Time) error {
	f, err := os.Open(in)
	if err != nil {
		return err
	}
	defer f.Close()
	evs, err := Load(f)
	if err != nil {
		return err
	}
	mf, err := os.Open(metaPath)
	if err != nil {
		return err
	}
	defer mf.Close()
	meta, err := LoadMeta(mf)
	if err != nil {
		return err
	}
	meta.BuiltAt = now
	files, err := Build(evs, meta)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(outDir); err != nil {
		return err
	}
	for p, b := range files {
		dst := filepath.Join(outDir, p)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("%d evaluations -> %d files in %s\n", len(evs), len(files), outDir)
	return nil
}
