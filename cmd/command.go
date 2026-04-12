package cmd

import (
	"fmt"
	"strings"

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
	if c.Name != "version" {
		if err := container.ValidateDockerEngine(); err != nil {
			return err
		}
	}
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
	case "patch", "redroid-script":
		return c.executePatch()
	default:
		return fmt.Errorf("Unknown command: %s", c.Name)
	}
}

func (c *Command) executeInit() error {
	var containerName string
	var image string

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
	} else {
		fmt.Println("\nAvailable Redroid Images:")
		var filteredImages []config.RedroidImage
		isARM := utils.IsARM()

		for _, img := range config.AvailableImages {
			isOfficial := strings.HasPrefix(img.URL, "redroid/redroid")

			if isOfficial {
				// Official images work on both (multi-arch), except 64only which is x86_64 only
				if isARM && img.Is64Only {
					continue
				}
				filteredImages = append(filteredImages, img)
				continue
			}

			// Community images: Hide 64only images on ARM hosts
			if isARM && img.Is64Only {
				continue
			}
			// Community images: Hide ARM-only images on x86_64 hosts
			if !isARM && img.IsARMOnly {
				continue
			}
			filteredImages = append(filteredImages, img)
		}

		for i, img := range filteredImages {
			fmt.Printf("[%d] %s (%s)\n", i+1, img.Name, img.URL)
		}
		fmt.Printf("[%d] Custom Image (Enter your own Docker image)\n", len(filteredImages)+1)

		fmt.Printf("\nSelect an image [1-%d]: ", len(filteredImages)+1)
		var choice int
		fmt.Scanln(&choice)

		if choice < 1 || choice > len(filteredImages)+1 {
			return fmt.Errorf("Invalid selection!")
		}

		if choice == len(filteredImages)+1 {
			fmt.Print("Enter custom image URL: ")
			fmt.Scanln(&image)
			if image == "" {
				return fmt.Errorf("Image URL is required!")
			}
		} else {
			image = filteredImages[choice-1].URL
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

func (c *Command) executePatch() error {
	if len(c.Args) < 1 {
		return fmt.Errorf("Usage: reddock patch <container-name> [flags]\n\n" +
			"Builds a new Docker image from the official ReDroid base using reddock’s native Go port of\n" +
			"ayasa520/redroid-script (OpenGapps 8.1–14, MindTheGapps 9–15 where packages exist, NDK/Houdini on 10–13, Widevine).\n" +
			"Requires the container's image_url to be redroid/redroid:… (official images only).\n" +
			"No Python clone: work happens in a temp directory; downloads go to ~/.cache/redroid/downloads (or XDG_CACHE_HOME).\n" +
			"Android version defaults from the container image tag when it is official.\n\n" +
			"Flags (same meaning as the original redroid.py):\n" +
			"  -a, --android VERSION   Android line (8.1.0 … 14.0.0, *64only); default from image\n" +
			"  -g, --gapps             OpenGapps (pico; 8.1–14)\n" +
			"  -lg, --litegapps        LiteGapps\n" +
			"  -mtg, --mindthegapps    MindTheGapps (8.1 not available; 10/11 not on x86_64 — use -g)\n" +
			"  -n, --ndk               libndk translation (x86/x86_64; Android 10–13)\n" +
			"  -i, --houdini           Houdini + Houdini_Hack except on 8.1 (hack skipped there)\n" +
			"  -m, --magisk            Magisk\n" +
			"  -w, --widevine          Widevine L3 (only published arch/API combos)\n" +
			"  -t, --target-image      Final image name (default: reddock-custom:<name>-redroid-script)\n" +
			"  --script-path DIR       Ignored (kept for compatibility with old scripts)\n" +
			"  --instant               Ignored (kept for compatibility)\n" +
			"  --update-config         Set container image_url to the final image in config.json\n\n" +
			"Examples:\n" +
			"  sudo reddock patch mybox -g -n -t reddock/mybox:full\n" +
			"  sudo reddock patch mybox -mtg -w -t reddock/mybox:mtg-wv")
	}
	name, flags, err := container.ParseRedroidScriptCLIArgs(c.Args)
	if err != nil {
		return err
	}
	return container.BuildImageWithRedroidScript(name, flags)
}

func PrintUsage() {
	fmt.Printf("Reddock %s\n", BannerLabel())
	fmt.Println("\nRequires the Docker CLI (docker) on PATH. If you use Waydroid, note that a running Docker")
	fmt.Println("daemon can block LXC features Waydroid needs; see messages after init.")
	fmt.Println("\nUsage: reddock [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  init [<n>] [<image>]        		Initialize container (interactive if name/image omitted)")
	fmt.Println("  start <n> [-v]              		Start container (use -v for foreground/logs)")
	fmt.Println("  stop <n>                    		Stop container (name required)")
	fmt.Println("  restart <n> [-v]            		Restart container (use -v for foreground/logs)")
	fmt.Println("  status <n>                  		Show container status (name required)")
	fmt.Println("  shell <n>                   		Enter container shell (name required)")
	fmt.Println("  adb-connect <n>             		Show ADB connection command (name required)")
	fmt.Println("  remove <n> [--image]        		Remove container/data (--image to also remove image)")
	fmt.Println("  list                           	List all Reddock-managed containers")
	fmt.Println("  log <n>                     		Show container logs (name required)")
	fmt.Println("  prune                          	Remove unused images")
	fmt.Println("  patch <n> [flags]            		Official-image addon build (redroid-script + reddock matrix; see patch usage)")
	fmt.Println("  version                        	Show version information")
	fmt.Println("\nExamples:")
	fmt.Println("  sudo reddock init android13")
	fmt.Println("  sudo reddock start android13 -v")
	fmt.Println("  sudo reddock remove android13")
	fmt.Println("  sudo reddock remove android13 --image  # Also remove Docker image")
	fmt.Println("  sudo reddock patch android13 --instant -g -n -t my/gapps-ndk")
}
