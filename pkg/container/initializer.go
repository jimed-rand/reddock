package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"reddock/pkg/config"
	"reddock/pkg/ui"
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

	s1 := ui.NewSpinner("Checking system requirements...")
	s1.Start()

	if err := i.checkRuntime(); err != nil {
		return fmt.Errorf("Runtime check failed: %v", err)
	}

	if err := i.checkKernelModules(); err != nil {
		return fmt.Errorf("Kernel module check failed: %v", err)
	}
	s1.Finish("System requirements met")

	if strings.HasPrefix(i.container.ImageURL, "redroid/redroid:") {
		fmt.Printf("Pulling official Redroid image %s...\n", i.container.ImageURL)
		if err := i.pullImage(); err != nil {
			return fmt.Errorf("Failed to pull image: %v", err)
		}
		fmt.Println("Image pulled successfully")
	} else {
		s2 := ui.NewSpinner("Verifying custom image availability...")
		s2.Start()
		
		if err := i.verifyImageExists(); err != nil {
			s2.Finish("Image verification failed")
			return fmt.Errorf("Image '%s' not found locally. Please build or pull it first.\n"+
				"For custom images built with 'reddock addons build', the image should already exist.\n"+
				"Error: %v", i.container.ImageURL, err)
		}
		s2.Finish("Custom image verified")
	}

	s3 := ui.NewSpinner("Setting up container environment...")
	s3.Start()

	if err := i.createDataDirectory(); err != nil {
		return fmt.Errorf("Failed to create data directory: %v", err)
	}
	s3.Finish("Environment setup complete")

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
	} else {
		cmd := exec.Command("modprobe", "binder_linux", "devices=binder,hwbinder,vndbinder")
		if err := cmd.Run(); err != nil {
			fmt.Println()
			fmt.Printf("Warning: modprobe binder_linux failed: %v\n", err)
			fmt.Println("You need to prepare the binder/binderfs first before using it.")
		}
	}

	return nil
}

func (i *Initializer) pullImage() error {
	return i.runtime.PullImage(i.container.ImageURL)
}

func (i *Initializer) verifyImageExists() error {
	cmd := i.runtime.Command("image", "inspect", i.container.ImageURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Image does not exist locally")
	}
	return nil
}

func (i *Initializer) createDataDirectory() error {
	if err := os.MkdirAll(i.container.DataPath, 0755); err != nil {
		return fmt.Errorf("Failed to create data directory: %v", err)
	}
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
