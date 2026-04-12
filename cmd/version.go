package cmd

import "fmt"

// Release is the public distribution tag (e.g. v2.4.2) when built for a GitHub Release.
// Injected with -ldflags "-X reddock/cmd.Release=..."; empty for local snapshot builds.
var Release = ""

// Snapshot is the internal build identity (short commit + optional -dirty).
// Injected with -ldflags "-X reddock/cmd.Snapshot=..." (see Makefile).
var Snapshot = "release"

func (c *Command) executeVersion() error {
	if Release != "" {
		fmt.Printf("Release:  %s\n", Release)
	} else {
		fmt.Printf("Release:  (not a tagged distribution)\n")
	}
	fmt.Printf("Snapshot: %s\n", Snapshot)
	return nil
}

// BannerLabel is a single line for help/usage headers: release tag if set, else snapshot.
func BannerLabel() string {
	if Release != "" {
		return Release
	}
	return Snapshot
}
