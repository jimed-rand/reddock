package utils

import (
	"fmt"

	"redway/pkg/config"
	"redway/pkg/container"
)

type StatusManager struct {
	manager *container.Manager
	config  *config.Config
}

func NewStatusManager() *StatusManager {
	cfg, _ := config.Load()
	return &StatusManager{
		manager: container.NewManager(),
		config:  cfg,
	}
}

func (s *StatusManager) Show() error {
	fmt.Println("Redway Status")
	fmt.Println("=============")

	fmt.Printf("\nContainer: %s\n", s.config.ContainerName)
	fmt.Printf("Image: %s\n", s.config.ImageURL)
	fmt.Printf("Data Path: %s\n", s.config.DataPath)
	fmt.Printf("GPU Mode: %s\n", s.config.GPUMode)
	fmt.Printf("Initialized: %v\n", s.config.Initialized)

	if !s.config.Initialized {
		fmt.Println("\nThe container is not initialized. Run 'redway init' first.")
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
		fmt.Println("\nStart with: redway start")
	}

	return nil
}
