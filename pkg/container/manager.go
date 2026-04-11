package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"reddock/pkg/config"
	"reddock/pkg/ui"
)

type Manager struct {
	runtime       Runtime
	config        *config.Config
	containerName string
}

func NewManagerForContainer(containerName string) *Manager {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: Failed to load config: %v\n", err)
		cfg = config.GetDefault()
	}
	return &Manager{
		runtime:       NewRuntime(),
		config:        cfg,
		containerName: containerName,
	}
}

func (m *Manager) Start(verbose bool) error {
	if err := CheckRoot(); err != nil {
		return err
	}

	container := m.config.GetContainer(m.containerName)
	if container == nil {
		return fmt.Errorf("Container '%s' not found. Run 'reddock init %s' first", m.containerName, m.containerName)
	}

	if !container.Initialized {
		return fmt.Errorf("Container '%s' is not initialized. Run 'reddock init %s' first", m.containerName, m.containerName)
	}

	if m.runtime.IsRunning(m.containerName) {
		fmt.Printf("Container '%s' is already running\n", m.containerName)
		return nil
	}

	spinner := ui.NewSpinner(fmt.Sprintf("Starting container '%s'...", m.containerName))
	spinner.Start()

	var err error
	if m.runtime.Exists(m.containerName) {
		err = m.runtime.StartExisting(m.containerName)
		if err != nil {
			spinner.Finish(fmt.Sprintf("Failed to start container '%s'", m.containerName))
			return fmt.Errorf("Failed to start existing container: %v", err)
		}
	} else {
		args := m.buildRunArgs(container)
		cmd := m.runtime.Command(args...)
		output, runErr := cmd.CombinedOutput()
		if runErr != nil {
			spinner.Finish(fmt.Sprintf("Failed to start container '%s'", m.containerName))
			return fmt.Errorf("Failed to start container: %s\n%s", runErr, string(output))
		}
	}

	spinner.Finish(fmt.Sprintf("Container '%s' started successfully", m.containerName))

	fmt.Println("\nContainer started!")
	fmt.Printf("ADB Connect: adb connect localhost:%d\n", container.HostADBPort())

	if verbose {
		fmt.Println("\nShowing container logs (Ctrl+C to detach)...")
		return m.showLogs()
	}

	return nil
}

func (m *Manager) buildRunArgs(container *config.Container) []string {
	args := []string{
		"run",
		"-d",
		"--privileged",
		"--name", m.containerName,
		"--hostname", m.containerName,
		"-v", fmt.Sprintf("%s:/data:z", container.GetDataPath()),
		"-p", fmt.Sprintf("%d:5555", container.Port),
	}

	// Add GPU mode if specified
	gpuMode := container.GPUMode
	if gpuMode == "" {
		gpuMode = "auto"
	}

	// Image
	args = append(args, container.ImageURL)

	// Boot arguments
	args = append(args, fmt.Sprintf("androidboot.redroid_gpu_mode=%s", gpuMode))

	return args
}

func (m *Manager) Stop() error {
	if err := CheckRoot(); err != nil {
		return err
	}

	if !m.runtime.Exists(m.containerName) {
		return fmt.Errorf("Container '%s' does not exist", m.containerName)
	}

	if !m.runtime.IsRunning(m.containerName) {
		// Even if not running, we continue to the removal step
	} else {
		spinner := ui.NewSpinner(fmt.Sprintf("Stopping container '%s'...", m.containerName))
		spinner.Start()

		if err := m.runtime.Stop(m.containerName); err != nil {
			spinner.Finish(fmt.Sprintf("Failed to stop container '%s'", m.containerName))
			return fmt.Errorf("failed to stop container: %v", err)
		}
		spinner.Finish(fmt.Sprintf("Container '%s' stopped successfully", m.containerName))
	}

	if err := m.runtime.Remove(m.containerName, false); err != nil {
		if forceErr := m.runtime.Remove(m.containerName, true); forceErr != nil {
			fmt.Printf("Warning: Could not remove stopped container: %v\n", forceErr)
		}
	}

	return nil
}

func (m *Manager) Restart(verbose bool) error {
	if err := m.Stop(); err != nil {
		if !strings.Contains(err.Error(), "is already stopped") {
			return err
		}
	}
	return m.Start(verbose)
}

func (m *Manager) IsRunning() bool {
	return m.runtime.IsRunning(m.containerName)
}

func (m *Manager) GetIP() (string, error) {
	format := "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}"
	ip, err := m.runtime.Inspect(m.containerName, format)
	if err != nil {
		return "", err
	}
	return ip, nil
}

func (m *Manager) GetContainer() *config.Container {
	if m.config == nil {
		return nil
	}
	return m.config.GetContainer(m.containerName)
}

func (m *Manager) showLogs() error {
	cmd := exec.Command(m.runtime.Name(), "logs", "-f", m.containerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
