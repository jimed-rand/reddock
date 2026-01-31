package utils

import (
	"fmt"
	"os"
	"os/exec"

	"reddock/pkg/config"
	"reddock/pkg/container"
)

type LogManager struct {
	config        *config.Config
	containerName string
}

func NewLogManager(containerName string) *LogManager {
	cfg, _ := config.Load()
	return &LogManager{
		config:        cfg,
		containerName: containerName,
	}
}

func (l *LogManager) Show() error {
	if err := container.CheckRoot(); err != nil {
		return err
	}
	cont := l.config.GetContainer(l.containerName)
	if cont == nil {
		return fmt.Errorf("container '%s' not found", l.containerName)
	}

	fmt.Printf("Showing logs for container: %s\n", l.containerName)
	fmt.Println("Press Ctrl+C to exit")

	cmd := exec.Command("docker", "logs", "-f", l.containerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
