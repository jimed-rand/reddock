package cmd

import (
	"fmt"

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
	default:
		return fmt.Errorf("unknown command: %s", c.Name)
	}
}

func (c *Command) executeInit() error {
	var containerName string
	var image string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("container name is required. Usage: reddock init <container-name> [image-url]")
	}

	if len(c.Args) > 1 {
		image = c.Args[1]
	} else {
		image = config.DefaultImageURL
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
		return fmt.Errorf("container name is required. Usage: reddock start <container-name> [-v]")
	}

	mgr := container.NewManagerForContainer(containerName)
	return mgr.Start(verbose)
}

func (c *Command) executeStop() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("container name is required. Usage: reddock stop <container-name>")
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
		return fmt.Errorf("container name is required. Usage: reddock restart <container-name> [-v]")
	}

	mgr := container.NewManagerForContainer(containerName)
	return mgr.Restart(verbose)
}

func (c *Command) executeStatus() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("container name is required. Usage: reddock status <container-name>")
	}

	status := utils.NewStatusManager(containerName)
	return status.Show()
}

func (c *Command) executeShell() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("container name is required. Usage: reddock shell <container-name>")
	}

	shell := utils.NewShellManager(containerName)
	return shell.Enter()
}

func (c *Command) executeAdbConnect() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("container name is required. Usage: reddock adb-connect <container-name>")
	}

	adb := utils.NewAdbManager(containerName)
	return adb.ShowConnection()
}

func (c *Command) executeRemove() error {
	var containerName string

	if len(c.Args) > 0 {
		containerName = c.Args[0]
	} else {
		return fmt.Errorf("container name is required. Usage: reddock remove <container-name>")
	}

	mgr := container.NewManagerForContainer(containerName)
	return mgr.Remove()
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
		return fmt.Errorf("container name is required. Usage: reddock log <container-name>")
	}

	logger := utils.NewLogManager(containerName)
	return logger.Show()
}

func PrintUsage() {
	fmt.Println("Reddock - Redroid Container Manager")
	fmt.Println("\nUsage: reddock [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  init <name> [image]            Initialize container (name required)")
	fmt.Println("  start <name> [-v]              Start container (use -v for foreground/logs)")
	fmt.Println("  stop <name>                    Stop container (name required)")
	fmt.Println("  restart <name> [-v]            Restart container (use -v for foreground/logs)")
	fmt.Println("  status <name>                  Show container status (name required)")
	fmt.Println("  shell <name>                   Enter container shell (name required)")
	fmt.Println("  adb-connect <name>             Show ADB connection command (name required)")
	fmt.Println("  remove <name>                  Remove container and data (name required)")
	fmt.Println("  list                           List all Reddock containers")
	fmt.Println("  log <name>                     Show container logs (name required)")
	fmt.Println("\nExamples:")
	fmt.Println("  sudo reddock init android13")
	fmt.Println("  sudo reddock start android13 -v")
}
