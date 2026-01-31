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
}

func NewInitializer(containerName, image string) *Initializer {
	cfg, _ := config.Load()

	container := cfg.GetContainer(containerName)
	if container == nil {
		container = &config.Container{
			Name:        containerName,
			ImageURL:    image,
			DataPath:    config.GetDefaultDataPath(containerName),
			LogFile:     containerName + ".log",
			GPUMode:     config.DefaultGPUMode,
			Port:        5555, // Default ADB port
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
	}
}

func (i *Initializer) Initialize() error {
	fmt.Println("Initializing the Reddock container...")
	fmt.Printf("Container: %s\n", i.container.Name)
	fmt.Printf("Image: %s\n\n", i.container.ImageURL)

	if err := CheckRoot(); err != nil {
		return err
	}

	steps := []struct {
		name string
		fn   func() error
	}{
		{"Checking Docker installation", i.checkDocker},
		{"Checking kernel modules", i.checkKernelModules},
		{"Pulling Redroid image", i.pullImage},
		{"Creating data directory", i.createDataDirectory},
	}

	for _, step := range steps {
		fmt.Printf("[*] %s...\n", step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s failed: %v", step.name, err)
		}
	}

	i.container.Initialized = true
	i.config.AddContainer(i.container)
	if err := config.Save(i.config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Println("\nThe container has been initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Printf("  reddock start %s        # Start the container\n", i.container.Name)
	fmt.Printf("  reddock adb-connect %s  # Get ADB connection info\n", i.container.Name)
	fmt.Printf("  reddock shell %s        # Access container shell\n", i.container.Name)

	return nil
}

func (i *Initializer) checkDocker() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found. Please install Docker")
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
		fmt.Println("Binder support detected")
	} else {
		fmt.Println("Binder support not detected. Attempting to load module...")
		cmd := exec.Command("modprobe", "binder_linux", "devices=binder,hwbinder,vndbinder")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: modprobe binder_linux failed: %v\n", err)
			fmt.Println("If you are on Ubuntu/Debian, you might need: apt install linux-modules-extra-$(uname -r)")
		} else {
			fmt.Println("Binder module loaded successfully")
		}
	}

	// Also check ashmem
	if _, err := os.Stat("/dev/ashmem"); err != nil {
		fmt.Println("Ashmem support not detected. Attempting to load module...")
		cmd := exec.Command("modprobe", "ashmem_linux")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: modprobe ashmem_linux failed: %v\n", err)
		} else {
			fmt.Println("Ashmem module loaded successfully")
		}
	} else {
		fmt.Println("Ashmem support detected")
	}

	return nil
}

func (i *Initializer) pullImage() error {
	fmt.Printf("Pulling image %s...\n", i.container.ImageURL)
	cmd := exec.Command("docker", "pull", i.container.ImageURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (i *Initializer) createDataDirectory() error {
	if err := os.MkdirAll(i.container.DataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
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

	for _, c := range containers {
		status := "Initialized"
		cmd := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", c.Name)
		if output, err := cmd.Output(); err == nil {
			status = strings.TrimSpace(string(output))
		} else {
			status = "Stopped"
		}
		fmt.Printf("%-20s %-40s %-10s\n", c.Name, c.ImageURL, status)
	}

	return nil
}
