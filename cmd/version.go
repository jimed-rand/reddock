package cmd

import "fmt"

var Version = "2.22.5"

func (c *Command) executeVersion() error {
	fmt.Printf("Reddock %s\n", Version)
	return nil
}
