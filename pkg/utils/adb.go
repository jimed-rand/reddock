package utils

import (
	"fmt"
	"os/exec"

	"reddock/pkg/container"
)

type AdbManager struct {
	manager       *container.Manager
	containerName string
}

func NewAdbManager(containerName string) *AdbManager {
	return &AdbManager{
		manager:       container.NewManagerForContainer(containerName),
		containerName: containerName,
	}
}

func (a *AdbManager) ShowConnection() error {
	if err := container.CheckRoot(); err != nil {
		return err
	}
	if !a.manager.IsRunning() {
		return fmt.Errorf("The container '%s' is not running. Start it with 'reddock start %s'", a.containerName, a.containerName)
	}

	ip, _ := a.manager.GetIP()
	if ip == "" {
		ip = "localhost"
	}

	fmt.Println("\nADB Information:")
	fmt.Println("===========================")
	fmt.Printf("Default connection: localhost:5555\n")
	fmt.Printf("Internal IP: %s\n", ip)

	fmt.Printf("\nAttempting to connect via ADB...\n")
	cmd := exec.Command("adb", "connect", "localhost:5555")
	output, _ := cmd.CombinedOutput()
	fmt.Printf("ADB Output: %s", string(output))

	fmt.Printf("\nYou can now use:\n")
	fmt.Printf("  adb shell              # Access Android shell\n")
	fmt.Printf("  adb install app.apk    # Install APK\n")
	fmt.Printf("  adb logcat             # View logs\n")
	fmt.Printf("  scrcpy -s localhost:5555 # Run scrcpy\n")

	return nil
}
