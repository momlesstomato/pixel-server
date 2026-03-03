package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type result struct {
	Path  string
	Count int
}

func main() {
	root := flag.String("root", ".", "workspace root")
	maxFiles := flag.Int("max", 12, "max non-test Go files per package")
	allowRaw := flag.String("allow", "pkg/protocol", "comma-separated allowlist of package paths")
	flag.Parse()

	allow := make(map[string]bool)
	for _, item := range strings.Split(*allowRaw, ",") {
		path := strings.TrimSpace(item)
		if path != "" {
			allow[path] = true
		}
	}

	checks := []string{"pkg", "services", "tools"}
	var violations []result
	for _, top := range checks {
		topPath := filepath.Join(*root, top)
		entries, err := os.ReadDir(topPath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			rel := filepath.ToSlash(filepath.Join(top, entry.Name()))
			if allow[rel] {
				continue
			}
			count, err := countGoFiles(filepath.Join(topPath, entry.Name()))
			if err != nil {
				fmt.Fprintf(os.Stderr, "packageguard: scan failed for %s: %v\n", rel, err)
				os.Exit(1)
			}
			if count > *maxFiles {
				violations = append(violations, result{Path: rel, Count: count})
			}
		}
	}

	if len(violations) == 0 {
		fmt.Printf("packageguard: ok (max=%d)\n", *maxFiles)
		return
	}

	sort.Slice(violations, func(i, j int) bool { return violations[i].Path < violations[j].Path })
	fmt.Printf("packageguard: %d package(s) exceed max=%d non-test .go files\n", len(violations), *maxFiles)
	for _, v := range violations {
		fmt.Printf("  - %s (%d files)\n", v.Path, v.Count)
	}
	os.Exit(1)
}

func countGoFiles(path string) (int, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		count++
	}
	return count, nil
}
