package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"reddock/pkg/addons"
	"reddock/pkg/config"
	"reddock/pkg/container"
	"reddock/pkg/utils"
)

type Command struct {
	Name string
	Args []string
}

func NewCommand(name string, args []string) *Command {
	return &Command{
		Name: name,
		Args: args,
	}
}

func CheckRoot() error {
	return container.CheckRoot()
}

func (c *Command) Execute() error {
	switch c.Name {
	case "init":
		return c.executeInit()
	case "start":
		return c.executeStart()
	case "stop":
		return c.executeStop()
	case "restart":
		return c.executeRestart()
	case "status":
		return c.executeStatus()
	case "shell":
		return c.executeShell()
	case "adb-connect":
		return c.executeAdbConnect()
	case "remove":
		return c.executeRemove()
	case "list":
		return c.executeList()
	case "log":
		return c.executeLog()
	case "prune":
		return c.executePrune()
	case "version":
		return c.executeVersion()
	case "dockerfile":
		return c.executeDockerfile()
	default:
		return fmt.Errorf("Unknown command: %s", c.Name)
	}
}

func (c *Command) executeInit() error {
	var containerName string
	var image string
	offerAddons := false

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		fmt.Print("Enter container name: ")
		_, err := fmt.Scanln(&containerName)
		if err != nil || containerName == "" {
			return fmt.Errorf("Container name is required!")
		}
	}

	if len(c.Args) > 1 {
		image = c.Args[1]
		// If provided via CLI, check if it's official to offer addons
		if strings.HasPrefix(image, "redroid/redroid") {
			offerAddons = true
		}
	} else {
		fmt.Println("\nAvailable Redroid Images:")
		for i, img := range config.AvailableImages {
			fmt.Printf("[%d] %s (%s)\n", i+1, img.Name, img.URL)
		}
		fmt.Printf("[%d] Custom Image (Enter your own Docker image)\n", len(config.AvailableImages)+1)

		fmt.Printf("\nSelect an image [1-%d]: ", len(config.AvailableImages)+1)
		var choice int
		fmt.Scanln(&choice)

		if choice < 1 || choice > len(config.AvailableImages)+1 {
			return fmt.Errorf("invalid selection")
		}

		if choice == len(config.AvailableImages)+1 {
			fmt.Print("Enter custom image URL: ")
			fmt.Scanln(&image)
			if image == "" {
				return fmt.Errorf("image URL is required")
			}
			// Custom image: Do NOT offer addons, use as is.
			offerAddons = false
		} else {
			image = config.AvailableImages[choice-1].URL
			if strings.HasPrefix(image, "redroid/redroid") {
				offerAddons = true
			}
		}
	}

	// Addon selection (Only for official images)
	if offerAddons {
		fmt.Print("\nDo you want to install addons? [y/N]: ")
		var installAddons string
		fmt.Scanln(&installAddons)

		if strings.ToLower(installAddons) == "y" || strings.ToLower(installAddons) == "yes" {
			am := addons.NewAddonManager()
			availableAddons := am.ListAddons()

			fmt.Println("\nAvailable Addons:")
			for i, name := range availableAddons {
				fmt.Printf("[%d] %s\n", i+1, name)
			}

			fmt.Print("\nEnter addon numbers separated by comma (e.g., 1,3): ")

			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			var selectedAddons []string
			if input != "" {
				parts := strings.Split(input, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					var idx int
					fmt.Sscanf(p, "%d", &idx)
					if idx >= 1 && idx <= len(availableAddons) {
						selectedAddons = append(selectedAddons, availableAddons[idx-1])
					}
				}
			}

			if len(selectedAddons) > 0 {
				version := config.ExtractVersionFromImage(image)
				if version == "" {
					fmt.Print("Could not detect Android version. Please enter version (e.g., 11.0.0): ")
					fmt.Scanln(&version)
				}

				// Determine architecture
				arch := "x86_64"
				if runtime.GOARCH == "arm64" {
					arch = "arm64"
				}

				customImageName := fmt.Sprintf("reddock-custom:%s-%s", containerName, version)
				fmt.Printf("\nBuilding custom image '%s' with addons: %v\n", customImageName, selectedAddons)

				if err := am.BuildImage(image, customImageName, version, arch, selectedAddons); err != nil {
					return fmt.Errorf("failed to build custom image: %v", err)
				}
				image = customImageName
			}
		}
	}

	init := container.NewInitializer(containerName, image)
	return init.Initialize()
}

func (c *Command) executeStart() error {
	var containerName string
	verbose := false

	for _, arg := range c.Args {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		} else if containerName == "" {
			containerName = arg
		}
	}

	if containerName == "" {
		return fmt.Errorf("Container name is required! Usage: reddock start <container-name> [-v]")
	}

	mgr := container.NewManagerForContainer(containerName)
	return mgr.Start(verbose)
}

func (c *Command) executeStop() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("Container name is required! Usage: reddock stop <container-name>")
	}

	mgr := container.NewManagerForContainer(containerName)
	return mgr.Stop()
}

func (c *Command) executeRestart() error {
	var containerName string
	verbose := false

	for _, arg := range c.Args {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		} else if containerName == "" {
			containerName = arg
		}
	}

	if containerName == "" {
		return fmt.Errorf("Container name is required! Usage: reddock restart <container-name> [-v]")
	}

	mgr := container.NewManagerForContainer(containerName)
	return mgr.Restart(verbose)
}

func (c *Command) executeStatus() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("Container name is required! Usage: reddock status <container-name>")
	}

	status := utils.NewStatusManager(containerName)
	return status.Show()
}

func (c *Command) executeShell() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("Container name is required! Usage: reddock shell <container-name>")
	}

	shell := utils.NewShellManager(containerName)
	return shell.Enter()
}

func (c *Command) executeAdbConnect() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("Container name is required! Usage: reddock adb-connect <container-name>")
	}

	adb := utils.NewAdbManager(containerName)
	return adb.ShowConnection()
}

func (c *Command) executeRemove() error {
	var containerName string
	removeAll := false

	for _, arg := range c.Args {
		if arg == "--all" || arg == "-a" {
			removeAll = true
		} else if containerName == "" {
			containerName = arg
		}
	}

	if containerName == "" {
		return fmt.Errorf("Container name is required! Usage: reddock remove <container-name> [--all]")
	}

	remover := container.NewRemover(containerName)
	return remover.Remove(removeAll)
}

func (c *Command) executeList() error {
	lister := container.NewLister()
	return lister.ListReddockContainers()
}

func (c *Command) executeLog() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("Container name is required! Usage: reddock log <container-name>")
	}

	logger := utils.NewLogManager(containerName)
	return logger.Show()
}

func (c *Command) executePrune() error {
	pruner := container.NewPruner()
	return pruner.Prune()
}

func (c *Command) executeDockerfile() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("Container name is required! Usage: reddock dockerfile <container-name>")
	}

	generator := container.NewDockerfileGenerator(containerName)
	return generator.Show()
}

func PrintUsage() {
	fmt.Printf("Reddock v%s\n", Version)
	fmt.Println("\nUsage: reddock [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  init [<name>] [<image>]        Initialize container (interactive if name/image omitted)")
	fmt.Println("  start <name> [-v]              Start container (use -v for foreground/logs)")
	fmt.Println("  stop <name>                    Stop container (name required)")
	fmt.Println("  restart <name> [-v]            Restart container (use -v for foreground/logs)")
	fmt.Println("  status <name>                  Show container status (name required)")
	fmt.Println("  shell <name>                   Enter container shell (name required)")
	fmt.Println("  adb-connect <name>             Show ADB connection command (name required)")
	fmt.Println("  remove <name> [--all]          Remove container, data, and optionally image (--all)")
	fmt.Println("  list                           List all Reddock containers")
	fmt.Println("  log <name>                     Show container logs (name required)")
	fmt.Println("  prune                          Remove unused images")
	fmt.Println("  dockerfile <name>              Show the Dockerfile for a container")
	fmt.Println("  version                        Show version information")
	fmt.Println("\nExamples:")
	fmt.Println("  sudo reddock init android13")
	fmt.Println("  sudo reddock start android13 -v")
	fmt.Println("  sudo reddock remove android13 --all")
}
