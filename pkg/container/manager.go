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

	// #region agent log
	agentDebugLog("manager.go:Start", "container config loaded", "H5", map[string]any{
		"containerName": m.containerName,
		"port":          container.Port,
		"imageURL":      container.ImageURL,
		"dataPath":      container.GetDataPath(),
	})
	// #endregion

	if m.runtime.IsRunning(m.containerName) {
		fmt.Printf("Container '%s' is already running\n", m.containerName)
		return nil
	}

	spinner := ui.NewSpinner(fmt.Sprintf("Starting container '%s'...", m.containerName))
	spinner.Start()

	// #region agent log
	versOut, versErr := m.runtime.Command("version", "--format", "{{.Client.Version}}").CombinedOutput()
	// #endregion

	exists := m.runtime.Exists(m.containerName)
	// #region agent log
	agentDebugLog("manager.go:Start", "docker probe + exists", "H3", map[string]any{
		"runtimeBinary": m.runtime.Name(),
		"versionErr":    fmt.Sprintf("%v", versErr),
		"versionOut":    strings.TrimSpace(string(versOut)),
		"exists":        exists,
	})
	// #endregion

	var err error
	if exists {
		err = m.runtime.StartExisting(m.containerName)
		if err != nil {
			spinner.Finish(fmt.Sprintf("Failed to start container '%s'", m.containerName))
			return fmt.Errorf("Failed to start existing container: %v", err)
		}
	} else {
		args := m.buildRunArgs(container)
		// #region agent log
		agentDebugLog("manager.go:Start", "docker run args", "H4", map[string]any{"args": args})
		// #endregion
		cmd := m.runtime.Command(args...)
		output, runErr := cmd.CombinedOutput()
		if runErr != nil {
			spinner.Finish(fmt.Sprintf("Failed to start container '%s'", m.containerName))
			return fmt.Errorf("Failed to start container: %s\n%s", runErr, string(output))
		}
	}

	// #region agent log
	st, stErr := m.runtime.Inspect(m.containerName, "{{.State.Status}}")
	exitStr, _ := m.runtime.Inspect(m.containerName, "{{.State.ExitCode}}")
	runRaw, runErr := m.runtime.Inspect(m.containerName, "{{.State.Running}}")
	runningNow := m.runtime.IsRunning(m.containerName)
	agentDebugLog("manager.go:Start", "post-start state", "H1", map[string]any{
		"usedStartExisting":       exists,
		"IsRunning":               runningNow,
		"StateStatus":             st,
		"StateRunningTemplateOut": strings.TrimSpace(runRaw),
		"inspectStatusErr":        fmt.Sprintf("%v", stErr),
		"inspectRunningErr":       fmt.Sprintf("%v", runErr),
		"ExitCodeStr":             exitStr,
	})
	// #endregion

	if !runningNow {
		logOut, logErr := m.runtime.Command("logs", "--tail", "60", m.containerName).CombinedOutput()
		spinner.Finish(fmt.Sprintf("Container '%s' did not stay running", m.containerName))
		logBlock := string(logOut)
		if logErr != nil {
			logBlock = fmt.Sprintf("(docker logs failed: %v)\n%s", logErr, logBlock)
		}
		return fmt.Errorf(
			"container is not running (docker state: %q, exit code: %s). "+
				"Check binder (binder_linux /dev/binder* or binderfs /dev/binderfs/*), ashmem/memfd, and image compatibility; see redroid-doc troubleshooting. "+
				"If the host uses SELinux (enforcing) or AppArmor, run `reddock status %s` for remediation hints.\n\nLast container logs:\n%s",
			strings.TrimSpace(st), strings.TrimSpace(exitStr), m.containerName, strings.TrimSpace(logBlock),
		)
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

	// Boot arguments (use_memfd helps on kernels without ashmem, e.g. many openSUSE/5.18+ setups)
	args = append(args, fmt.Sprintf("androidboot.redroid_gpu_mode=%s", gpuMode))
	args = append(args, "androidboot.use_memfd=true")

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

// FormatStoppedDiagnostics returns docker state and recent logs when the instance is not running.
func (m *Manager) FormatStoppedDiagnostics() string {
	if !m.runtime.Exists(m.containerName) {
		return "No Docker container with this name exists. If you removed it manually, run `reddock remove " + m.containerName + "` and `reddock init` again, or check `docker ps -a`."
	}
	st, errSt := m.runtime.Inspect(m.containerName, "{{.State.Status}}")
	exit, _ := m.runtime.Inspect(m.containerName, "{{.State.ExitCode}}")
	logs, logErr := m.runtime.Command("logs", "--tail", "45", m.containerName).CombinedOutput()
	var b strings.Builder
	fmt.Fprintf(&b, "Docker state: status=%q exit_code=%q (inspect err: %v)\n",
		strings.TrimSpace(st), strings.TrimSpace(exit), errSt)
	if logErr != nil {
		fmt.Fprintf(&b, "docker logs error: %v\n", logErr)
	}
	b.WriteString(strings.TrimSpace(string(logs)))
	fmt.Fprintf(&b, "\n\nTip: If you upgraded reddock, run `sudo reddock stop %s` then `sudo reddock start %s` once to recreate the Docker container with current boot flags; your host data directory is kept.\n",
		m.containerName, m.containerName)
	return b.String()
}

func (m *Manager) showLogs() error {
	cmd := exec.Command(m.runtime.Name(), "logs", "-f", m.containerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
