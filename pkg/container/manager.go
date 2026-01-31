package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"redway/pkg/config"
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

func (m *Manager) Start() error {
	if err := CheckRoot(); err != nil {
		return err
	}
	container, err := m.getContainer()
	if err != nil {
		return err
	}

	if !container.Initialized {
		return fmt.Errorf("The container '%s' is not initialized. Run 'redway init %s' first", container.Name, container.Name)
	}

	if m.IsRunning() {
		fmt.Printf("The container '%s' is already running\n", container.Name)
		return nil
	}

	fmt.Printf("Starting the container '%s'...\n", container.Name)

	logPath := container.LogFile

	cmd := exec.Command("lxc-start",
		"-l", "debug",
		"-o", logPath,
		"-n", container.Name)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	fmt.Println("The container started successfully")
	fmt.Printf("\nLog file: %s\n", logPath)
	fmt.Println("\nNext steps:")
	fmt.Printf("  redway status %s       # Check container status\n", container.Name)
	fmt.Printf("  redway adb-connect %s  # Get ADB connection info\n", container.Name)

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

	cmd := exec.Command("lxc-stop", "-k", "-n", container.Name)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %v", err)
	}

	fmt.Println("The container stopped successfully")
	return nil
}

func (m *Manager) Restart() error {
	fmt.Println("Restarting the container...")

	if err := m.Stop(); err != nil {
		return err
	}

	return m.Start()
}

func (m *Manager) Remove() error {
	if err := CheckRoot(); err != nil {
		return err
	}
	container, err := m.getContainer()
	if err != nil {
		return err
	}

	if m.IsRunning() {
		fmt.Println("Stopping the container first...")
		if err := m.Stop(); err != nil {
			return err
		}
	}

	fmt.Printf("Removing the container '%s'...\n", container.Name)

	cmd := exec.Command("lxc-destroy", "-n", container.Name)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: lxc-destroy failed: %v\n", err)
		fmt.Println("Attempting force cleanup...")
	}

	containerPath := container.GetContainerPath()
	if _, err := os.Stat(containerPath); err == nil {
		fmt.Printf("Force removing container directory: %s\n", containerPath)
		if err := os.RemoveAll(containerPath); err != nil {
			fmt.Printf("Warning: Could not remove container directory: %v\n", err)
		} else {
			fmt.Println("Container directory removed")
		}
	}

	// Remove log file if it exists
	if _, err := os.Stat(container.LogFile); err == nil {
		fmt.Printf("Removing log file: %s\n", container.LogFile)
		os.Remove(container.LogFile)
	}

	fmt.Printf("Remove data directory and image cache? [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" {
		// Remove data directory
		if err := os.RemoveAll(container.DataPath); err != nil {
			fmt.Printf("Warning: Could not remove data directory: %v\n", err)
		} else {
			fmt.Printf("Data directory removed: %s\n", container.DataPath)
		}

		// Remove OCI image cache if it exists
		// lxc-oci template usually caches in /var/cache/lxc/oci
		cachePath := "/var/cache/lxc/oci"
		if _, err := os.Stat(cachePath); err == nil {
			fmt.Println("Cleaning OCI image cache...")
			// Note: This removes ALL cached OCI images.
			// Currently we don't have a safe way to remove only the one used by this container
			// without potentially breaking others, but user asked to remove "embed images".
			if err := os.RemoveAll(cachePath); err != nil {
				fmt.Printf("Warning: Could not remove OCI cache: %v\n", err)
			} else {
				fmt.Println("OCI image cache cleaned")
			}
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
	container, err := m.getContainer()
	if err != nil {
		return false
	}

	cmd := exec.Command("lxc-info", "-n", container.Name, "-s")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "RUNNING")
}

func (m *Manager) GetInfo() (string, error) {
	container, err := m.getContainer()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("lxc-info", container.Name)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (m *Manager) GetIP() (string, error) {
	container, err := m.getContainer()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("lxc-info", container.Name, "-i")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "IP:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("no IP address found")
}

func (m *Manager) GetPID() (string, error) {
	container, err := m.getContainer()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("lxc-info", container.Name, "-p")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "PID:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("no PID found")
}

func CheckRoot() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this command must be run as root (superuser)")
	}
	if os.Getenv("SUDO_USER") != "" || os.Getenv("SUDO_COMMAND") != "" {
		return fmt.Errorf("running with 'sudo' is blocked. Please use 'su -' or run as root directly to ensure correct environment and paths")
	}
	return nil
}

type Lister struct{}

func NewLister() *Lister {
	return &Lister{}
}

func (l *Lister) List() error {
	fmt.Println("LXC Containers:")

	cmd := exec.Command("lxc-ls", "-f")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (l *Lister) ListRedwayContainers() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if len(cfg.Containers) == 0 {
		fmt.Println("No containers configured")
		return nil
	}

	fmt.Println("Redway Containers:")
	fmt.Println("==================")

	for _, container := range cfg.ListContainers() {
		status := "Not initialized"
		if container.Initialized {
			status = "Initialized"
		}

		mgr := NewManagerForContainer(container.Name)
		running := "Stopped"
		if mgr.IsRunning() {
			running = "Running"
		}

		fmt.Printf("\nName:        %s\n", container.Name)
		fmt.Printf("Image:       %s\n", container.ImageURL)
		fmt.Printf("Status:      %s\n", status)
		fmt.Printf("Running:     %s\n", running)
		fmt.Printf("Data Path:   %s\n", container.DataPath)
		fmt.Printf("Log File:    %s\n", container.LogFile)
	}

	return nil
}
