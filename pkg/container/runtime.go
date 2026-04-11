package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Runtime interface {
	Name() string
	Command(args ...string) *exec.Cmd
	IsInstalled() bool
	PullImage(image string) error
	PushImage(image string) error
	Run(args ...string) error
	Stop(containerName string) error
	StartExisting(containerName string) error
	Remove(containerName string, force bool) error
	RemoveImage(image string) error
	Inspect(containerName string, format string) (string, error)
	Exists(containerName string) bool
	IsRunning(containerName string) bool
	PruneImages() (string, error)
	IsAuthenticated() (bool, string, error)
}

type GenericRuntime struct {
	binary string
}

func NewRuntime() Runtime {
	return &GenericRuntime{binary: "docker"}
}

// ValidateDockerEngine ensures the docker CLI is available (reddock uses the docker CLI only).
func ValidateDockerEngine() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker was not found in PATH: reddock requires the docker CLI")
	}
	return nil
}

// PrintWaydroidDockerNotice warns that Docker (with the daemon/socket active) can interfere with Waydroid.
func PrintWaydroidDockerNotice() {
	fmt.Println()
	fmt.Println("Compatibility note: With the Docker daemon running and its socket enabled, LXC-related")
	fmt.Println("workloads such as Waydroid may fail or be blocked. If you rely on Waydroid, you may need")
	fmt.Println("to stop Docker, not using reddock, or adjust your setup so only one stack owns those resources.")
}

func (r *GenericRuntime) Name() string {
	return r.binary
}

func (r *GenericRuntime) Command(args ...string) *exec.Cmd {
	return exec.Command(r.binary, args...)
}

func (r *GenericRuntime) IsInstalled() bool {
	_, err := exec.LookPath(r.binary)
	return err == nil
}

func (r *GenericRuntime) PullImage(image string) error {
	cmd := r.Command("pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *GenericRuntime) PushImage(image string) error {
	cmd := r.Command("push", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *GenericRuntime) Run(args ...string) error {
	cmd := r.Command(append([]string{"run"}, args...)...)
	return cmd.Run()
}

func (r *GenericRuntime) Stop(containerName string) error {
	return r.Command("stop", containerName).Run()
}

func (r *GenericRuntime) StartExisting(containerName string) error {
	return r.Command("start", containerName).Run()
}

func (r *GenericRuntime) Remove(containerName string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, containerName)
	return r.Command(args...).Run()
}

func (r *GenericRuntime) RemoveImage(image string) error {
	return r.Command("rmi", image).Run()
}

func (r *GenericRuntime) Inspect(containerName string, format string) (string, error) {
	cmd := r.Command("inspect", "-f", format, containerName)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *GenericRuntime) Exists(containerName string) bool {
	// Implement generic exists check using ps -a or inspect
	// Using ps -a with filter is robust
	cmd := r.Command("ps", "-a", "--filter", "name=^"+containerName+"$", "--format", "{{.Names}}")
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output)) == containerName
}

func (r *GenericRuntime) IsRunning(containerName string) bool {
	state, err := r.Inspect(containerName, "{{.State.Running}}")
	if err != nil {
		return false
	}
	return state == "true"
}

func (r *GenericRuntime) PruneImages() (string, error) {
	cmd := r.Command("image", "prune", "-f")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (r *GenericRuntime) IsAuthenticated() (bool, string, error) {
	cmd := r.Command("info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", err
	}
	outStr := string(output)
	lines := strings.Split(outStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Username:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return true, strings.TrimSpace(parts[1]), nil
			}
		}
	}
	return false, "", nil
}
