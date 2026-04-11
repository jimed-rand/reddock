package cmd

import "fmt"

// Injected at link time via -ldflags "-X reddock/cmd.Version=..." (see Makefile).
var Version = "dev"

func (c *Command) executeVersion() error {
	fmt.Printf("Reddock %s\n", Version)
	return nil
}
