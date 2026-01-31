package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"reddock/pkg/config"
)

type Initializer struct {
	config    *config.Config
	container *config.Container
	runtime   Runtime
}

func NewInitializer(containerName, image string) *Initializer {
	cfg, _ := config.Load()

	container := cfg.GetContainer(containerName)
	if container == nil {
		// Find next available port
		port := 5555
		for _, c := range cfg.Containers {
			if c.Port >= port {
				port = c.Port + 1
			}
		}

		container = &config.Container{
			Name:        containerName,
			ImageURL:    image,
			DataPath:    config.GetDefaultDataPath(containerName),
			LogFile:     containerName + ".log",
			GPUMode:     config.DefaultGPUMode,
			Port:        port,
			Initialized: false,
		}
		cfg.AddContainer(container)
		config.Save(cfg)
	} else {
		container.ImageURL = image
		config.Save(cfg)
	}

	return &Initializer{
		config:    cfg,
		container: container,
		runtime:   NewRuntime(),
	}
}

func (i *Initializer) Initialize() error {
	fmt.Println("Initiating the Reddock container...")
	fmt.Printf("Container: %s\n", i.container.Name)
	fmt.Printf("Image: %s\n\n", i.container.ImageURL)

	if err := CheckRoot(); err != nil {
		return err
	}

	steps := []struct {
		name string
		fn   func() error
	}{
		{"Checking Runtime installation", i.checkRuntime},
		{"Checking kernel modules", i.checkKernelModules},
		{"Pulling Redroid image", i.pullImage},
		{"Creating data directory", i.createDataDirectory},
	}

	for _, step := range steps {
		fmt.Printf("[-] %s...\n", step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s failed: %v", step.name, err)
		}
	}

	i.container.Initialized = true
	i.config.AddContainer(i.container)
	if err := config.Save(i.config); err != nil {
		return fmt.Errorf("Failed to save the config: %v", err)
	}

	fmt.Println("\nThe container has been initiated successfully!")
	fmt.Println("\nNext steps:")
	fmt.Printf("  reddock start %s        # Start the container\n", i.container.Name)
	fmt.Printf("  reddock adb-connect %s  # Get ADB connection info\n", i.container.Name)
	fmt.Printf("  reddock shell %s        # Access container shell\n", i.container.Name)

	return nil
}

func (i *Initializer) checkRuntime() error {
	if !i.runtime.IsInstalled() {
		return fmt.Errorf("%s is not found. Please install Docker or Podman", i.runtime.Name())
	}
	fmt.Printf("Using runtime: %s\n", i.runtime.Name())
	return nil
}

func (i *Initializer) checkKernelModules() error {
	binderFound := false
	binderPaths := []string{
		"/sys/module/binder_linux",
		"/sys/module/binder",
		"/dev/binderfs",
		"/dev/binder",
	}

	for _, path := range binderPaths {
		if _, err := os.Stat(path); err == nil {
			binderFound = true
			break
		}
	}

	if binderFound {
		fmt.Println("Binder support detected")
	} else {
		fmt.Println("Binder support not detected. Attempting to load module...")
		// Try to load, but don't fail if we can't (might be in container or handled by host)
		cmd := exec.Command("modprobe", "binder_linux", "devices=binder,hwbinder,vndbinder")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: modprobe binder_linux failed: %v\n", err)
			fmt.Println("You need to prepare the binder/binderfs first before using it.")
		} else {
			fmt.Println("Binder module loaded successfully")
		}
	}

	return nil
}

func (i *Initializer) pullImage() error {
	fmt.Printf("Pulling the image %s...\n", i.container.ImageURL)
	return i.runtime.PullImage(i.container.ImageURL)
}

func (i *Initializer) createDataDirectory() error {
	if err := os.MkdirAll(i.container.DataPath, 0755); err != nil {
		return fmt.Errorf("Failed to create data directory: %v", err)
	}

	fmt.Printf("Data directory: %s\n", i.container.DataPath)
	return nil
}

type Lister struct {
	config *config.Config
}

func NewLister() *Lister {
	cfg, _ := config.Load()
	return &Lister{config: cfg}
}

func (l *Lister) ListReddockContainers() error {
	containers := l.config.ListContainers()
	if len(containers) == 0 {
		fmt.Println("No Reddock containers found.")
		return nil
	}

	fmt.Printf("%-20s %-40s %-10s\n", "NAME", "IMAGE", "STATUS")
	fmt.Println(strings.Repeat("-", 70))

	runtime := NewRuntime()
	for _, c := range containers {
		status := "Initiated"
		if s, err := runtime.Inspect(c.Name, "{{.State.Status}}"); err == nil {
			status = s
		} else {
			status = "Stopped"
		}
		fmt.Printf("%-20s %-40s %-10s\n", c.Name, c.ImageURL, status)
	}

	return nil
}
