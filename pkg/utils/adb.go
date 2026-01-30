package utils

import (
	"fmt"

	"redway/pkg/container"
)

type AdbManager struct {
	manager *container.Manager
}

func NewAdbManager() *AdbManager {
	return &AdbManager{
		manager: container.NewManager(),
	}
}

func (a *AdbManager) ShowConnection() error {
	if !a.manager.IsRunning() {
		return fmt.Errorf("The container is not running. Start it with 'redway start'")
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
