package container

import (
	"fmt"
	"os"

	"reddock/pkg/config"
)

type Manager struct {
	config        *config.Config
	containerName string
	runtime       Runtime
}

func NewManager() *Manager {
	cfg, _ := config.Load()
	return &Manager{
		config:        cfg,
		containerName: "",
		runtime:       NewRuntime(),
	}
}

func NewManagerForContainer(containerName string) *Manager {
	cfg, _ := config.Load()
	return &Manager{
		config:        cfg,
		containerName: containerName,
		runtime:       NewRuntime(),
	}
}

func (m *Manager) getContainer() (*config.Container, error) {
	container := m.config.GetContainer(m.containerName)
	if container == nil {
		return nil, fmt.Errorf("Container '%s' is not found", m.containerName)
	}
	return container, nil
}

func (m *Manager) Start(verbose bool) error {
	if err := CheckRoot(); err != nil {
		return err
	}
	container, err := m.getContainer()
	if err != nil {
		return err
	}

	if !container.Initialized {
		return fmt.Errorf("The container '%s' is not initiated. Run 'reddock init %s' first", container.Name, container.Name)
	}

	if m.IsRunning() {
		fmt.Printf("The container '%s' is already running\n", container.Name)
		return nil
	}

	// Check if container exists but is stopped
	exists := m.runtime.Exists(container.Name)

	if exists {
		fmt.Printf("Starting existing container '%s'...\n", container.Name)
		startCmd := m.runtime.Command("start", container.Name)
		if verbose {
			startCmd.Stdout = os.Stdout
			startCmd.Stderr = os.Stderr
		}
		if err := startCmd.Run(); err != nil {
			return fmt.Errorf("Failed to start existing container: %v", err)
		}
	} else {
		fmt.Printf("Creating and starting new container '%s'...\n", container.Name)

		// Map GPU mode to redroid param
		gpuParam := "auto"
		if container.GPUMode != "" {
			gpuParam = container.GPUMode
		}

		args := []string{
			"run", "-itd",
			"--privileged",
			"--name", container.Name,
			"-v", fmt.Sprintf("%s:/data", container.GetDataPath()),
			"-p", fmt.Sprintf("%d:5555", container.Port),
			container.ImageURL,
			"androidboot.redroid_gpu_mode=" + gpuParam,
		}

		runCmd := m.runtime.Command(args...)
		if verbose {
			runCmd.Stdout = os.Stdout
			runCmd.Stderr = os.Stderr
		}
		if err := runCmd.Run(); err != nil {
			return fmt.Errorf("failed to run container: %v", err)
		}
	}

	fmt.Println("The Container started successfully")
	if verbose {
		fmt.Println("Showing logs (Press Ctrl+C to stop)...")
		logCmd := m.runtime.Command("logs", "-f", container.Name)
		logCmd.Stdout = os.Stdout
		logCmd.Stderr = os.Stderr
		logCmd.Run()
	}

	fmt.Println("\nNext steps:")
	fmt.Printf("  reddock status %s       # Check container status\n", container.Name)
	fmt.Printf("  reddock adb-connect %s  # Get ADB connection info\n", container.Name)

	return nil
}

func (m *Manager) Stop() error {
	if err := CheckRoot(); err != nil {
		return err
	}
	container, err := m.getContainer()
	if err != nil {
		return err
	}

	if !m.IsRunning() {
		fmt.Printf("The container '%s' is not running\n", container.Name)
		return nil
	}

	fmt.Printf("Stopping the container '%s'...\n", container.Name)
	if err := m.runtime.Stop(container.Name); err != nil {
		return fmt.Errorf("failed to stop container: %v", err)
	}

	fmt.Println("The container stopped successfully")
	return nil
}

func (m *Manager) Restart(verbose bool) error {
	if err := m.Stop(); err != nil {
		fmt.Printf("Warning: Stop failed: %v\n", err)
	}
	return m.Start(verbose)
}

func (m *Manager) IsRunning() bool {
	return m.runtime.IsRunning(m.containerName)
}

func (m *Manager) GetIP() (string, error) {
	// For Docker/Podman, we usually connect via localhost if ports are mapped
	// But if we want the internal IP:
	ip, err := m.runtime.Inspect(m.containerName, "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}")
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "localhost", nil // Return localhost as fallback for mapped ports
	}
	return ip, nil
}
