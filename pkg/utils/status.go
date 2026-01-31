package utils

import (
	"fmt"

	"reddock/pkg/config"
	"reddock/pkg/container"
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

	fmt.Println("Reddock Status")
	fmt.Println("==============")

	fmt.Printf("\nContainer: %s\n", cont.Name)
	fmt.Printf("Image: %s\n", cont.ImageURL)
	fmt.Printf("Data Path: %s\n", cont.GetDataPath())
	fmt.Printf("GPU Mode: %s\n", cont.GPUMode)
	fmt.Printf("Initialized: %v\n", cont.Initialized)

	if !cont.Initialized {
		fmt.Printf("\nThe container is not initialized. Run 'reddock init %s' first.\n", cont.Name)
		return nil
	}

	if s.manager.IsRunning() {
		fmt.Println("\nThe container is RUNNING")

		ip, _ := s.manager.GetIP()
		fmt.Printf("\nADB Connection:\n")
		fmt.Printf("  adb connect localhost:5555  (via mapped port)\n")
		fmt.Printf("  Internal IP: %s\n", ip)

		fmt.Printf("\nDirect Shell Access:\n")
		fmt.Printf("  reddock shell %s\n", cont.Name)
	} else {
		fmt.Println("\nThe container is STOPPED")
		fmt.Printf("\nStart with: reddock start %s\n", cont.Name)
	}

	return nil
}
