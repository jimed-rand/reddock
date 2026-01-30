package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"redway/pkg/config"
)

type Manager struct {
	config *config.Config
}

func NewManager() *Manager {
	cfg, _ := config.Load()
	return &Manager{config: cfg}
}

func (m *Manager) Start() error {
	if !m.config.Initialized {
		return fmt.Errorf("The container is not initialized. Run 'redway init' first")
	}

	if m.IsRunning() {
		fmt.Println("The container is already running")
		return nil
	}

	fmt.Printf("Starting the container '%s'...\n", m.config.ContainerName)

	logPath := m.config.LogFile

	cmd := exec.Command("lxc-start",
		"-l", "debug",
		"-o", logPath,
		"-n", m.config.ContainerName)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	fmt.Println("The container started successfully")
	fmt.Printf("\nLog file: %s\n", logPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  redway status       # Check container status")
	fmt.Println("  redway adb-connect  # Get ADB connection info")

	return nil
}

func (m *Manager) Stop() error {
	if !m.IsRunning() {
		fmt.Println("The container is not running")
		return nil
	}

	fmt.Printf("Stopping the container '%s'...\n", m.config.ContainerName)

	cmd := exec.Command("lxc-stop", "-k", "-n", m.config.ContainerName)

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
	if m.IsRunning() {
		fmt.Println("Stopping the container first...")
		if err := m.Stop(); err != nil {
			return err
		}
	}

	fmt.Printf("Removing the container '%s'...\n", m.config.ContainerName)

	cmd := exec.Command("lxc-destroy", "-n", m.config.ContainerName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container: %v", err)
	}

	fmt.Printf("Remove data directory? [y/N]: ")
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" {
		if err := os.RemoveAll(m.config.DataPath); err != nil {
			fmt.Printf("Warning: Could not remove data directory: %v\n", err)
		} else {
			fmt.Printf("Data directory removed: %s\n", m.config.DataPath)
		}
	}

	configPath := config.GetConfigPath()
	if err := os.Remove(configPath); err != nil {
		fmt.Printf("Warning: Could not remove config: %v\n", err)
	}

	fmt.Println("The container removed successfully")
	return nil
}

func (m *Manager) IsRunning() bool {
	cmd := exec.Command("lxc-info", "-n", m.config.ContainerName, "-s")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "RUNNING")
}

func (m *Manager) GetInfo() (string, error) {
	cmd := exec.Command("lxc-info", m.config.ContainerName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (m *Manager) GetIP() (string, error) {
	cmd := exec.Command("lxc-info", m.config.ContainerName, "-i")
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
	cmd := exec.Command("lxc-info", m.config.ContainerName, "-p")
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
