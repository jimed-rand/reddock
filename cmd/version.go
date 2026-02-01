package cmd

import "fmt"

var Version = "2.22.3"

func (c *Command) executeVersion() error {
	fmt.Printf("Reddock version %s\n", Version)
	return nil
}
