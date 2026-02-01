package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"reddock/pkg/addons"
)

func (c *Command) executeAddons() error {
	if len(c.Args) == 0 {
		return c.showAddonsHelp()
	}

	subCommand := c.Args[0]
	subArgs := c.Args[1:]

	switch subCommand {
	case "list":
		return c.executeAddonsList()
	case "install":
		return c.executeAddonsInstall(subArgs)
	case "build":
		return c.executeAddonsBuild(subArgs)
	default:
		return fmt.Errorf("unknown addons subcommand: %s", subCommand)
	}
}

func (c *Command) showAddonsHelp() error {
	fmt.Println("Addons Management")
	fmt.Println("\nUsage: reddock addons [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  list                           List available addons")
	fmt.Println("  install <addon> <version>      Install an addon")
	fmt.Println("  build <name> <version> <addons...>  Build custom image with addons")
	fmt.Println("\nAvailable Addons:")
	fmt.Println("  houdini       - Intel Houdini ARM translation (x86/x86_64 only)")
	fmt.Println("  ndk           - NDK ARM translation (x86/x86_64 only)")
	fmt.Println("  litegapps     - LiteGapps (Google Apps)")
	fmt.Println("  mindthegapps  - MindTheGapps (Google Apps)")
	fmt.Println("  opengapps     - OpenGapps (Google Apps, Android 11 only)")
	fmt.Println("\nExamples:")
	fmt.Println("  reddock addons list")
	fmt.Println("  reddock addons build android13-gapps 13.0.0 litegapps ndk")
	fmt.Println("  reddock addons build android11-full 11.0.0 opengapps houdini")
	return nil
}

func (c *Command) executeAddonsList() error {
	manager := addons.NewAddonManager()
	addonNames := manager.ListAddons()

	fmt.Println("Available Addons:")
	fmt.Println(strings.Repeat("-", 50))

	for _, name := range addonNames {
		addon, _ := manager.GetAddon(name)
		versions := addon.SupportedVersions()
		fmt.Printf("%-15s - %s\n", name, addon.Name())
		fmt.Printf("                Supported versions: %v\n", versions)
		fmt.Println()
	}

	return nil
}

func (c *Command) executeAddonsInstall(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: reddock addons install <addon> <version>")
	}

	addonName := args[0]
	version := args[1]
	arch := getHostArch()

	manager := addons.NewAddonManager()
	defer manager.Cleanup()

	return manager.InstallAddon(addonName, version, arch)
}

func (c *Command) executeAddonsBuild(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: reddock addons build <image-name> <android-version> <addon1> [addon2] ...")
	}

	imageName := args[0]
	version := args[1]
	addonNames := args[2:]
	arch := getHostArch()

	manager := addons.NewAddonManager()
	defer manager.Cleanup()

	for _, addonName := range addonNames {
		if _, err := manager.GetAddon(addonName); err != nil {
			return fmt.Errorf("invalid addon: %s", addonName)
		}
	}

	baseImage := fmt.Sprintf("redroid/redroid:%s-latest", version)
	return manager.BuildImage(baseImage, imageName, version, arch, addonNames)
}

func getHostArch() string {
	machine := runtime.GOARCH

	mapping := map[string]string{
		"386":   "x86",
		"amd64": "x86_64",
		"arm64": "arm64",
		"arm":   "arm",
	}

	if arch, ok := mapping[machine]; ok {
		return arch
	}
	return "x86_64"
}
