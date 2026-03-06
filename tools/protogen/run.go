package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// generateFile renders one generated realm-direction protocol file.
func generateFile(specPath string, outDir string, realm string, direction string) error {
	if specPath == "" {
		return fmt.Errorf("spec path is required")
	}
	if direction != "c2s" && direction != "s2c" {
		return fmt.Errorf("invalid direction: %s", direction)
	}
	raw, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	spec, err := decodeSpec(raw)
	if err != nil {
		return err
	}
	packets := selectPackets(spec, realm, direction)
	sort.Slice(packets, func(i int, j int) bool {
		return packets[i].ID < packets[j].ID
	})
	rendered, err := renderPackets(packets, direction)
	if err != nil {
		return err
	}
	formatted, err := format.Source(rendered)
	if err != nil {
		return err
	}
	fileName := generatedFileName(realm, direction)
	outputPath := filepath.Join(outDir, fileName)
	return os.WriteFile(outputPath, formatted, 0o644)
}

// generatedFileName builds the generator output file name.
func generatedFileName(realm string, direction string) string {
	prefix := strings.ReplaceAll(realm, "-", "_")
	return prefix + "_" + direction + "_gen.go"
}
