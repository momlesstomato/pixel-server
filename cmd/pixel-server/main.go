package main

import (
	"io"
	"os"

	"github.com/momlesstomato/pixel-server/core/cli"
)

// main executes the root CLI command.
func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes CLI logic and returns a process exit code.
func run(arguments []string, output io.Writer, errorOutput io.Writer) int {
	command := cli.NewRootCommand(cli.Dependencies{})
	command.SetArgs(arguments)
	command.SetOut(output)
	command.SetErr(errorOutput)
	if err := command.Execute(); err != nil {
		return 1
	}
	return 0
}
