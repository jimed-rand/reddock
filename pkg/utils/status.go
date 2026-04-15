package utils

import (
	"fmt"
	"os"

	"reddock/pkg/config"
	"reddock/pkg/container"
	"reddock/pkg/sysinfo"
)

type StatusManager struct {
	manager       *container.Manager
	config        *config.Config
	containerName string
}

func NewStatusManager(containerName string) *StatusManager {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: Failed to load config: %v\n", err)
		cfg = config.GetDefault()
	}
	return &StatusManager{
		manager:       container.NewManagerForContainer(containerName),
		config:        cfg,
		containerName: containerName,
	}
}

func (s *StatusManager) Show() error {
	cont := s.config.GetContainer(s.containerName)
	if cont == nil {
		return fmt.Errorf("Container '%s' not found", s.containerName)
	}

	fmt.Println("Reddock Status")
	fmt.Println("==============")

	fmt.Printf("\nContainer: %s\n", cont.Name)
	fmt.Printf("Image: %s\n", cont.ImageURL)
	fmt.Printf("Data Path: %s\n", cont.GetDataPath())
	fmt.Printf("GPU Mode: %s\n", cont.GPUMode)
	fmt.Printf("Initiated: %v\n", cont.Initialized)

	b := sysinfo.ProbeBinderHost()
	fmt.Print("\nHost binder (kernel): ")
	switch {
	case b.HostBinderUsable():
		fmt.Println("OK — binder module and/or device nodes detected.")
	case b.BinderLinuxInstallable():
		fmt.Println("binder_linux is packaged for this kernel but not active; load the module or finish binderfs device setup.")
	default:
		fmt.Println("not detected — install binder_linux (DKMS/KMP) or configure binderfs for this kernel.")
	}

	lsm := sysinfo.ProbeHostLSM()
	fmt.Print("\nHost MAC (LSM): ")
	fmt.Println(lsm.HostLSMStatusLine())
	sysinfo.PrintHostLSMWarnings(os.Stdout, lsm)

	if !cont.Initialized {
		fmt.Printf("\nThe container is not initiated. Run 'reddock init %s' first.\n", cont.Name)
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
