package main

import (
	"fmt"
	"os"

	"gossh/internal/app"
)

// Version is set at build time via -ldflags
var version = "dev"

func main() {
	// Set version in app package
	app.SetVersion(version)

	if err := app.RunWithArgs(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
