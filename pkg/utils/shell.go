package utils

import (
	"fmt"
	"os"
	"os/exec"

	"redway/pkg/container"
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
		return fmt.Errorf("The container '%s' is not running. Start it with 'redway start %s'", s.containerName, s.containerName)
	}

	pid, err := s.manager.GetPID()
	if err != nil {
		return fmt.Errorf("Failed to get container PID: %v", err)
	}

	fmt.Printf("Entering container shell (PID: %s)...\n", pid)

	cmd := exec.Command("nsenter", "-t", pid, "-a", "sh")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
