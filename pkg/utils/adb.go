package utils

import (
	"fmt"

	"redway/pkg/container"
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
		return fmt.Errorf("The container '%s' is not running. Start it with 'redway start %s'", a.containerName, a.containerName)
	}

	ip, err := a.manager.GetIP()
	if err != nil {
		return fmt.Errorf("Failed to get container IP: %v", err)
	}

	fmt.Println("ADB Connection Information:")
	fmt.Println("===========================")
	fmt.Printf("\nConnect to Android container:\n")
	fmt.Printf("  adb connect %s:5555\n", ip)
	fmt.Printf("\nAfter connecting, you can:\n")
	fmt.Printf("  adb shell              # Access Android shell\n")
	fmt.Printf("  adb install app.apk    # Install APK\n")
	fmt.Printf("  adb logcat             # View logs\n")
	fmt.Printf("  adb devices            # List connected devices\n")

	return nil
}
