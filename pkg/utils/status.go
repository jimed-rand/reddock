package utils

import (
	"fmt"

	"redway/pkg/config"
	"redway/pkg/container"
)

type StatusManager struct {
	manager       *container.Manager
	config        *config.Config
	containerName string
}

func NewStatusManager(containerName string) *StatusManager {
	cfg, _ := config.Load()
	return &StatusManager{
		manager:       container.NewManagerForContainer(containerName),
		config:        cfg,
		containerName: containerName,
	}
}

func (s *StatusManager) Show() error {
	cont := s.config.GetContainer(s.containerName)
	if cont == nil {
		return fmt.Errorf("container '%s' not found", s.containerName)
	}

	fmt.Println("Redway Status")
	fmt.Println("=============")

	fmt.Printf("\nContainer: %s\n", cont.Name)
	fmt.Printf("Image: %s\n", cont.ImageURL)
	fmt.Printf("Data Path: %s\n", cont.DataPath)
	fmt.Printf("GPU Mode: %s\n", cont.GPUMode)
	fmt.Printf("Initialized: %v\n", cont.Initialized)

	if !cont.Initialized {
		fmt.Printf("\nThe container is not initialized. Run 'redway init %s' first.\n", cont.Name)
		return nil
	}

	fmt.Println("\nContainer Information:")

	info, err := s.manager.GetInfo()
	if err != nil {
		fmt.Printf("  Error getting info: %v\n", err)
		return nil
	}

	fmt.Print(info)

	if s.manager.IsRunning() {
		fmt.Println("\nThe container is RUNNING")

		if ip, err := s.manager.GetIP(); err == nil {
			fmt.Printf("\nADB Connection:\n")
			fmt.Printf("  adb connect %s:5555\n", ip)
		}

		if pid, err := s.manager.GetPID(); err == nil {
			fmt.Printf("\nDirect Shell Access:\n")
			fmt.Printf("  nsenter -t %s -a sh\n", pid)
		}
	} else {
		fmt.Println("\nThe container is STOPPED")
		fmt.Printf("\nStart with: redway start %s\n", cont.Name)
	}

	return nil
}
