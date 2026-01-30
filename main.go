package main

import (
	"fmt"
	"os"

	"gossh/internal/app"
)

func main() {
	if err := app.RunWithArgs(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
