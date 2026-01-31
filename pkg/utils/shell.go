package utils

import (
	"fmt"
	"os"
	"os/exec"

	"reddock/pkg/container"
)

type ShellManager struct {
	manager       *container.Manager
	containerName string
}

func NewShellManager(containerName string) *ShellManager {
	return &ShellManager{
		manager:       container.NewManagerForContainer(containerName),
		containerName: containerName,
	}
}

func (s *ShellManager) Enter() error {
	if err := container.CheckRoot(); err != nil {
		return err
	}
	if !s.manager.IsRunning() {
		return fmt.Errorf("The container '%s' is not running. Start it with 'reddock start %s'", s.containerName, s.containerName)
	}

	fmt.Printf("Entering container shell for '%s'...\n", s.containerName)

	cmd := exec.Command("docker", "exec", "-it", s.containerName, "sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
