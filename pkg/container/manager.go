package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"reddock/pkg/config"
)

type Manager struct {
	config        *config.Config
	containerName string
}

func NewManager() *Manager {
	cfg, _ := config.Load()
	return &Manager{
		config:        cfg,
		containerName: "",
	}
}

func NewManagerForContainer(containerName string) *Manager {
	cfg, _ := config.Load()
	return &Manager{
		config:        cfg,
		containerName: containerName,
	}
}

func (m *Manager) getContainer() (*config.Container, error) {
	container := m.config.GetContainer(m.containerName)
	if container == nil {
		return nil, fmt.Errorf("container '%s' not found", m.containerName)
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
		return fmt.Errorf("the container '%s' is not initialized. Run 'reddock init %s' first", container.Name, container.Name)
	}

	if m.IsRunning() {
		fmt.Printf("The container '%s' is already running\n", container.Name)
		return nil
	}

	// Check if container exists but is stopped
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name=^"+container.Name+"$", "--format", "{{.Names}}")
	output, _ := cmd.Output()
	exists := strings.TrimSpace(string(output)) == container.Name

	if exists {
		fmt.Printf("Starting existing container '%s'...\n", container.Name)
		startCmd := exec.Command("docker", "start", container.Name)
		if verbose {
			startCmd.Stdout = os.Stdout
			startCmd.Stderr = os.Stderr
		}
		if err := startCmd.Run(); err != nil {
			return fmt.Errorf("failed to start existing container: %v", err)
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

		runCmd := exec.Command("docker", args...)
		if verbose {
			runCmd.Stdout = os.Stdout
			runCmd.Stderr = os.Stderr
		}
		if err := runCmd.Run(); err != nil {
			return fmt.Errorf("failed to run container: %v", err)
		}
	}

	fmt.Println("The container started successfully")
	if verbose {
		fmt.Println("Showing logs (Press Ctrl+C to stop)...")
		logCmd := exec.Command("docker", "logs", "-f", container.Name)
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
	cmd := exec.Command("docker", "stop", container.Name)
	if err := cmd.Run(); err != nil {
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

func (m *Manager) Remove() error {
	if err := CheckRoot(); err != nil {
		return err
	}
	container, err := m.getContainer()
	if err != nil {
		return err
	}

	fmt.Printf("Removing the container '%s'...\n", container.Name)

	// Force remove the docker container
	cmd := exec.Command("docker", "rm", "-f", container.Name)
	cmd.Run()

	fmt.Printf("Remove data directory? (%s) [y/N]: ", container.GetDataPath())
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" {
		if err := os.RemoveAll(container.GetDataPath()); err != nil {
			fmt.Printf("Warning: Could not remove data directory: %v\n", err)
		} else {
			fmt.Printf("Data directory removed: %s\n", container.GetDataPath())
		}
	}

	m.config.RemoveContainer(container.Name)
	if err := config.Save(m.config); err != nil {
		fmt.Printf("Warning: Could not update config: %v\n", err)
	}

	fmt.Println("The container removed successfully")
	return nil
}

func (m *Manager) IsRunning() bool {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", m.containerName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

func (m *Manager) GetIP() (string, error) {
	// For Docker, we usually connect via localhost if ports are mapped
	// But if we want the internal IP:
	cmd := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", m.containerName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	ip := strings.TrimSpace(string(output))
	if ip == "" {
		return "localhost", nil // Return localhost as fallback for mapped ports
	}
	return ip, nil
}
