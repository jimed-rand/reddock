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
	case "addons":
		return c.executeAddons()
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
			return fmt.Errorf("Invalid selection!")
		}

		if choice == len(config.AvailableImages)+1 {
			fmt.Print("Enter custom image URL: ")
			fmt.Scanln(&image)
			if image == "" {
				return fmt.Errorf("Image URL is required!")
			}
			offerAddons = false
		} else {
			image = config.AvailableImages[choice-1].URL
			if strings.HasPrefix(image, "redroid/redroid") {
				offerAddons = true
			}
		}
	}

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

				arch := "x86_64"
				if runtime.GOARCH == "arm64" {
					arch = "arm64"
				}

				customImageName := fmt.Sprintf("reddock-custom:%s-%s", containerName, version)

				fmt.Print("\nDo you want to publish this image to Docker Hub? [y/N]: ")
				var publishImage string
				fmt.Scanln(&publishImage)

				pushToRegistry := false
				if strings.ToLower(publishImage) == "y" || strings.ToLower(publishImage) == "yes" {
					fmt.Print("Enter Docker Hub image name (format: username/repository:tag): ")
					reader := bufio.NewReader(os.Stdin)
					dockerHubName, _ := reader.ReadString('\n')
					dockerHubName = strings.TrimSpace(dockerHubName)

					if dockerHubName != "" {
						customImageName = dockerHubName
						pushToRegistry = true
					}
				}

				fmt.Printf("\nBuilding custom image '%s' with addons: %v\n", customImageName, selectedAddons)

				if err := am.BuildCustomImage(image, customImageName, version, arch, selectedAddons, pushToRegistry); err != nil {
					return fmt.Errorf("Failed to build custom image: %v", err)
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
	removeImage := false

	for _, arg := range c.Args {
		if arg == "--image" || arg == "-i" {
			removeImage = true
		} else if containerName == "" {
			containerName = arg
		}
	}

	if containerName == "" {
		return fmt.Errorf("Container name is required! Usage: reddock remove <container-name> [--image]")
	}

	remover := container.NewRemover(containerName)
	return remover.Remove(removeImage)
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
	if len(c.Args) < 1 {
		return fmt.Errorf("Usage: reddock dockerfile <subcommand> <container-name> [options]\n\n" +
			"Subcommands:\n" +
			"  show <container>              Show generated Dockerfile\n" +
			"  edit <container>              Edit Dockerfile with nano\n" +
			"  build <container> [image]     Build Docker image from Dockerfile\n" +
			"  commit <container> <image>    Commit running container to new image\n" +
			"  install <container> <addon>   Install addon to running container\n" +
			"  interactive <container>       Interactive Dockerfile workflow")
	}

	subcommand := c.Args[0]

	switch subcommand {
	case "show":
		if len(c.Args) < 2 {
			return fmt.Errorf("Container name is required! Usage: reddock dockerfile show <container-name>")
		}
		generator := container.NewDockerfileGenerator(c.Args[1])
		return generator.Show()

	case "edit":
		if len(c.Args) < 2 {
			return fmt.Errorf("Container name is required! Usage: reddock dockerfile edit <container-name>")
		}
		generator := container.NewDockerfileGenerator(c.Args[1])
		return generator.Edit()

	case "build":
		if len(c.Args) < 2 {
			return fmt.Errorf("Container name is required! Usage: reddock dockerfile build <container-name> [image-name]")
		}
		containerName := c.Args[1]
		imageName := fmt.Sprintf("reddock/%s:custom", containerName)
		if len(c.Args) > 2 {
			imageName = c.Args[2]
		}
		generator := container.NewDockerfileGenerator(containerName)
		// Save Dockerfile first
		if err := generator.SaveToFile(generator.GetDockerfilePath()); err != nil {
			return err
		}
		return generator.Build(imageName)

	case "commit":
		if len(c.Args) < 3 {
			return fmt.Errorf("Usage: reddock dockerfile commit <container-name> <new-image-name> [message]")
		}
		containerName := c.Args[1]
		imageName := c.Args[2]
		message := ""
		if len(c.Args) > 3 {
			message = strings.Join(c.Args[3:], " ")
		}
		generator := container.NewDockerfileGenerator(containerName)
		return generator.CommitContainer(imageName, message)

	case "install":
		if len(c.Args) < 3 {
			return fmt.Errorf("Usage: reddock dockerfile install <container-name> <addon-name>\n\n" +
				"This installs prepared addon files to a RUNNING container.\n" +
				"First prepare the addon with: reddock addons prepare <addon> <version>")
		}
		containerName := c.Args[1]
		addonName := c.Args[2]
		generator := container.NewDockerfileGenerator(containerName)
		// Check if container is running
		mgr := container.NewManagerForContainer(containerName)
		if !mgr.IsRunning() {
			return fmt.Errorf("Container '%s' is not running. Start it first with: sudo reddock start %s", containerName, containerName)
		}
		return generator.InstallAddonToRunningContainer("/tmp/reddock-addons", addonName)

	case "interactive", "i":
		if len(c.Args) < 2 {
			return fmt.Errorf("Container name is required! Usage: reddock dockerfile interactive <container-name>")
		}
		generator := container.NewDockerfileGenerator(c.Args[1])
		return generator.Interactive()

	default:
		// Backward compatibility: treat first arg as container name for "show"
		generator := container.NewDockerfileGenerator(subcommand)
		return generator.Show()
	}
}

func PrintUsage() {
	fmt.Printf("Reddock %s\n", Version)
	fmt.Println("\nUsage: reddock [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  init [<n>] [<image>]        		Initialize container (interactive if name/image omitted)")
	fmt.Println("  start <n> [-v]              		Start container (use -v for foreground/logs)")
	fmt.Println("  stop <n>                    		Stop container (name required)")
	fmt.Println("  restart <n> [-v]            		Restart container (use -v for foreground/logs)")
	fmt.Println("  status <n>                  		Show container status (name required)")
	fmt.Println("  shell <n>                   		Enter container shell (name required)")
	fmt.Println("  adb-connect <n>             		Show ADB connection command (name required)")
	fmt.Println("  remove <n> [--image]        		Remove container and data (--image to also remove image)")
	fmt.Println("  list                           	List all Reddock containers")
	fmt.Println("  log <n>                     		Show container logs (name required)")
	fmt.Println("  prune                          	Remove unused images")
	fmt.Println("  dockerfile <cmd> <n> ...       	Dockerfile management (see below)")
	fmt.Println("  addons <cmd> ...               	Addon management (see below)")
	fmt.Println("  version                        	Show version information")
	fmt.Println("\nDockerfile Subcommands:")
	fmt.Println("  dockerfile show <n>              	Show generated Dockerfile (name required)")
	fmt.Println("  dockerfile edit <n>              	Edit Dockerfile with nano (name required)")
	fmt.Println("  dockerfile build <n> [image]     	Build Docker image from Dockerfile (name required)")
	fmt.Println("  dockerfile commit <n> <image>    	Commit running container to new image (name and image required)")
	fmt.Println("  dockerfile install <n> <addon>   	Install addon to running container (name and addon required)")
	fmt.Println("  dockerfile interactive <n>       	Interactive Dockerfile workflow (name required)")
	fmt.Println("\nAddons Subcommands:")
	fmt.Println("  addons list                           	List available addons")
	fmt.Println("  addons prepare <addon> <v>            	Prepare addon for runtime install")
	fmt.Println("  addons build <n> <v> <addons...>      	Build custom image with addons")
	fmt.Println("\nExamples:")
	fmt.Println("  sudo reddock init android13")
	fmt.Println("  sudo reddock start android13 -v")
	fmt.Println("  sudo reddock remove android13")
	fmt.Println("  sudo reddock remove android13 --image  # Also remove Docker image")
	fmt.Println("  sudo reddock addons build custom-android13 13.0.0 litegapps ndk")
	fmt.Println("\nDockerfile subcommand example:")
	fmt.Println("  sudo reddock dockerfile edit android13           		# Edit with nano/vim")
	fmt.Println("  sudo reddock dockerfile build android13 myimage  		# Build the image")
	fmt.Println("  sudo reddock dockerfile install android13 houdini 		# Install to running container")
	fmt.Println("  sudo reddock dockerfile commit android13 myimage  		# Save the container state")
	fmt.Println("\nAddon subcommand example:")
	fmt.Println("  sudo reddock addons list                         		# List available addons")
	fmt.Println("  sudo reddock addons prepare houdini 11.0.0       		# Prepare addon for runtime")
	fmt.Println("  sudo reddock addons build custom-img 11.0.0 ndk  		# Build image with addons")
}
