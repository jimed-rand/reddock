package addons

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"reddock/pkg/config"
)

type AddonInjector struct {
	availableAddons map[string]Addon
	workDir         string
}

func NewAddonInjector() *AddonInjector {
	// Check for required dependencies
	if err := CheckDependencies(); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	addons := map[string]Addon{
		"houdini":      NewHoudiniAddon(),
		"ndk":          NewNDKAddon(),
		"litegapps":    NewLiteGappsAddon(),
		"mindthegapps": NewMindTheGappsAddon(),
		"opengapps":    NewOpenGappsAddon(),
	}

	return &AddonInjector{
		availableAddons: addons,
		workDir:         "/tmp/reddock-addons",
	}
}

func (ai *AddonInjector) GetAddon(name string) (Addon, error) {
	addon, ok := ai.availableAddons[name]
	if !ok {
		return nil, fmt.Errorf("addon '%s' not found", name)
	}
	return addon, nil
}

func (ai *AddonInjector) ListAddons() []string {
	var names []string
	for name := range ai.availableAddons {
		names = append(names, name)
	}
	return names
}

func (ai *AddonInjector) InjectToContainer(containerName, addonName, version, arch string) error {
	addon, err := ai.GetAddon(addonName)
	if err != nil {
		return err
	}

	if !addon.IsSupported(version) {
		return fmt.Errorf("%s does not support Android %s", addon.Name(), version)
	}

	if err := ensureDir(ai.workDir); err != nil {
		return err
	}

	fmt.Printf("\n=== Injecting %s to container %s ===\n", addon.Name(), containerName)

	if _, err := ai.checkContainerExists(containerName); err != nil {
		return err
	}

	fmt.Printf("Downloading %s...\n", addon.Name())
	if err := addon.Download(version, arch); err != nil {
		return fmt.Errorf("Download failed: %v", err)
	}

	fmt.Printf("Extracting %s...\n", addon.Name())
	if err := addon.Extract(version, arch); err != nil {
		return fmt.Errorf("Extract failed: %v", err)
	}

	fmt.Printf("Preparing files for injection...\n")
	addonDir := filepath.Join(ai.workDir, addonName)
	if err := addon.Copy(version, arch, ai.workDir); err != nil {
		return fmt.Errorf("Copy failed: %v", err)
	}

	fmt.Printf("Injecting files into container...\n")
	if err := ai.copyToContainer(containerName, addonDir); err != nil {
		return fmt.Errorf("Injection failed: %v", err)
	}

	fmt.Printf("Setting permissions...\n")
	if ai.isContainerRunning(containerName) {
		if err := ai.SetPermissions(containerName, addonName); err != nil {
			fmt.Printf("Warning: Failed to set some permissions: %v\n", err)
		}
	} else {
		fmt.Printf("Container is not running. Permissions will be applied when you start the container.\n")
	}

	// Update container config
	cfg, err := config.Load()
	if err == nil {
		container := cfg.GetContainer(containerName)
		if container != nil {
			alreadyExists := false
			for _, a := range container.Addons {
				if a == addonName {
					alreadyExists = true
					break
				}
			}
			if !alreadyExists {
				container.Addons = append(container.Addons, addonName)
				config.Save(cfg)
			}
		}
	}

	fmt.Printf("\nSuccessfully injected %s into %s\n", addon.Name(), containerName)
	fmt.Printf("\nNote: You may need to restart the container for changes to take effect:\n")
	fmt.Printf("  sudo reddock restart %s\n", containerName)

	return nil
}

func (ai *AddonInjector) checkContainerExists(containerName string) (string, error) {
	runtime := DetectRuntime()
	cmd := exec.Command(runtime, "inspect", "-f", "{{.State.Status}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Container '%s' not found", containerName)
	}

	return strings.TrimSpace(string(output)), nil
}

func (ai *AddonInjector) isContainerRunning(containerName string) bool {
	status, err := ai.checkContainerExists(containerName)
	if err != nil {
		return false
	}
	return status == "running"
}

func (ai *AddonInjector) copyToContainer(containerName, addonDir string) error {
	runtime := DetectRuntime()

	if _, err := os.Stat(addonDir); os.IsNotExist(err) {
		return fmt.Errorf("Addon directory not found: %s", addonDir)
	}

	entries, err := os.ReadDir(addonDir)
	if err != nil {
		return fmt.Errorf("Failed to read addon directory: %v", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(addonDir, entry.Name())

		cmd := exec.Command(runtime, "cp", srcPath, fmt.Sprintf("%s:/", containerName))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Failed to copy %s: %v", entry.Name(), err)
		}
		fmt.Printf("  ✓ Copied %s\n", entry.Name())
	}

	return nil
}

func (ai *AddonInjector) SetPermissions(containerName, addonName string) error {
	runtime := DetectRuntime()

	var commands [][]string

	switch addonName {
	case "houdini":
		commands = [][]string{
			{"chmod", "644", "/system/etc/init/houdini.rc"},
			{"chmod", "-R", "755", "/system/bin/houdini*"},
			{"chmod", "-R", "644", "/system/lib*/libhoudini*"},
		}
	case "ndk":
		commands = [][]string{
			{"chmod", "644", "/system/etc/init/ndk_translation.rc"},
			{"chmod", "-R", "644", "/system/lib*/libndk*"},
		}
	case "litegapps", "mindthegapps", "opengapps":
		commands = [][]string{
			{"chmod", "-R", "755", "/system/priv-app"},
			{"chmod", "-R", "644", "/system/etc/permissions"},
			{"chmod", "-R", "644", "/system/framework"},
		}
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(runtime, "exec", containerName)
		cmd.Args = append(cmd.Args, cmdArgs...)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: chmod failed for %v: %v\n", cmdArgs, err)
		}
	}

	return nil
}

func (ai *AddonInjector) InjectMultiple(containerName string, addons []AddonRequest) error {
	fmt.Printf("\n=== Injecting %d addons to container %s ===\n", len(addons), containerName)

	if _, err := ai.checkContainerExists(containerName); err != nil {
		return err
	}

	for i, req := range addons {
		fmt.Printf("\n[%d/%d] Processing %s...\n", i+1, len(addons), req.Name)
		if err := ai.InjectToContainer(containerName, req.Name, req.Version, req.Arch); err != nil {
			fmt.Printf("✗ Failed to inject %s: %v\n", req.Name, err)
			fmt.Printf("  Continuing with remaining addons...\n")
			continue
		}
	}

	fmt.Printf("\n=== Injection complete ===\n")
	fmt.Printf("\nRestart the container to apply changes:\n")
	fmt.Printf("  sudo reddock restart %s\n", containerName)

	return nil
}

func (ai *AddonInjector) GetSupportedVersions(addonName string) ([]string, error) {
	addon, err := ai.GetAddon(addonName)
	if err != nil {
		return nil, err
	}
	return addon.SupportedVersions(), nil
}

func (ai *AddonInjector) Cleanup() error {
	return os.RemoveAll(ai.workDir)
}

type AddonRequest struct {
	Name    string
	Version string
	Arch    string
}
