package container

import (
	"os/exec"
	"strings"
)

type Runtime interface {
	Name() string
	Command(args ...string) *exec.Cmd
	IsInstalled() bool
	PullImage(image string) error
	Run(args ...string) error
	Stop(containerName string) error
	Remove(containerName string, force bool) error
	RemoveImage(image string) error
	Inspect(containerName string, format string) (string, error)
	Exists(containerName string) bool
	IsRunning(containerName string) bool
}

type GenericRuntime struct {
	binary string
}

func NewRuntime() Runtime {
	// Prefer podman if available, otherwise docker
	if _, err := exec.LookPath("podman"); err == nil {
		return &GenericRuntime{binary: "podman"}
	}
	return &GenericRuntime{binary: "docker"}
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
	return cmd.Run()
}

func (r *GenericRuntime) Run(args ...string) error {
	cmd := r.Command(append([]string{"run"}, args...)...)
	return cmd.Run()
}

func (r *GenericRuntime) Stop(containerName string) error {
	return r.Command("stop", containerName).Run()
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
