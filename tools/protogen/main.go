package main

import (
	"flag"
	"fmt"
	"os"
)

// main parses generator flags and writes protocol artifacts.
func main() {
	specPath := flag.String("spec", "", "protocol yaml spec path")
	outDir := flag.String("out", ".", "output directory")
	realm := flag.String("realm", "handshake-security", "realm filter")
	direction := flag.String("direction", "c2s", "packet direction: c2s|s2c")
	flag.Parse()
	if err := generateFile(*specPath, *outDir, *realm, *direction); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "protogen: %v\n", err)
		os.Exit(1)
	}
}
