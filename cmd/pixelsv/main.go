package main

import (
	"os"

	"pixelsv/internal/runtime/cli"
)

// main executes the CLI root command.
func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
