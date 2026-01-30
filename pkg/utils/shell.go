package utils

import (
	"fmt"
	"os"
	"os/exec"

	"redway/pkg/container"
)

type ShellManager struct {
	manager *container.Manager
}

func NewShellManager() *ShellManager {
	return &ShellManager{
		manager: container.NewManager(),
	}
}

func (s *ShellManager) Enter() error {
	if !s.manager.IsRunning() {
		return fmt.Errorf("The container is not running. Start it with 'redway start'")
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
