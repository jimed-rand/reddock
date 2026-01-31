package main

import (
	"fmt"
	"os"

	"reddock/cmd"
)

// Main application entrypoint

func main() {
	if len(os.Args) < 2 {
		cmd.PrintUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	c := cmd.NewCommand(command, args)
	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
