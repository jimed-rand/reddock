package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"redway/pkg/config"
)

type Initializer struct {
	config *config.Config
	image  string
}

func NewInitializer(image string) *Initializer {
	cfg, _ := config.Load()
	cfg.ImageURL = image

	return &Initializer{
		config: cfg,
		image:  image,
	}
}

func (i *Initializer) Initialize() error {
	fmt.Println("Initializing the container...")
	fmt.Printf("Image: %s\n", i.image)
	fmt.Printf("Container: %s\n\n", i.config.ContainerName)

	steps := []struct {
		name string
		fn   func() error
	}{
		{"Checking kernel modules", i.checkKernelModules},
		{"Checking LXC tools", i.checkLXCTools},
		{"Checking LXC networking", i.checkLXCNetworking},
		{"Adjusting OCI template", i.adjustOCITemplate},
		{"Checking required tools", i.checkRequiredTools},
		{"Creating container", i.createContainer},
		{"Creating data directory", i.createDataDirectory},
		{"Adjusting container config", i.adjustContainerConfig},
		{"Applying networking workaround", i.applyNetworkingWorkaround},
	}

	for _, step := range steps {
		fmt.Printf("[*] %s...\n", step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s failed: %v", step.name, err)
		}
	}

	i.config.Initialized = true
	if err := config.Save(i.config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Println("\nThe container has been initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Printf("  redway start        # Start the container\n")
	fmt.Printf("  redway adb-connect  # Get ADB connection info\n")
	fmt.Printf("  redway shell        # Access container shell\n")

	return nil
}

func (i *Initializer) checkKernelModules() error {
	// Check for binder support (module or binderfs)
	binderFound := false
	binderPaths := []string{
		"/sys/module/binder_linux", // Module loaded
		"/sys/module/binder",       // Some distros use this name
		"/dev/binderfs",            // binderfs mounted
		"/dev/binder",              // Direct binder device
	}

	for _, path := range binderPaths {
		if _, err := os.Stat(path); err == nil {
			binderFound = true
			break
		}
	}

	// Check for ashmem support (module or memfd fallback)
	// Ashmem is no longer required for modern redroid versions (uses memfd)

	if binderFound {
		fmt.Println("Binder support (binderfs/binder module) detected")
		return nil
	}

	// Advisory warning only - don't fail, let user proceed
	fmt.Println("Binder support (binderfs/binder module) not detected")
	fmt.Println("This may be normal if your kernel has built-in support.")
	fmt.Println("If container fails to start, try loading modules:")
	fmt.Println("  sudo modprobe binder_linux devices=\"binder,hwbinder,vndbinder\"")
	fmt.Println("See README for distro-specific instructions.")

	return nil // Continue anyway - let container startup reveal actual issues
}

func (i *Initializer) checkLXCTools() error {
	tools := []string{"lxc-create", "lxc-start", "lxc-stop", "lxc-info"}

	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s not found. Please install lxc-utils", tool)
		}
	}

	fmt.Println("LXC tools available")
	return nil
}

func (i *Initializer) checkLXCNetworking() error {
	// Check if lxc-net service is active (systemd)
	cmd := exec.Command("systemctl", "is-active", "--quiet", "lxc-net")
	if err := cmd.Run(); err == nil {
		fmt.Println("    LXC networking (lxc-net) is active")
		return nil
	}

	// Fallback: check if lxcbr0 bridge exists
	if _, err := os.Stat("/sys/class/net/lxcbr0"); err == nil {
		fmt.Println("    LXC bridge (lxcbr0) is available")
		return nil
	}

	// Warning only, don't fail - some setups use different networking
	fmt.Println("    LXC networking isn't detected. Container may need manual network setup.")
	fmt.Println("      Try: sudo systemctl start lxc-net")
	return nil
}

func (i *Initializer) adjustOCITemplate() error {
	templatePath := "/usr/share/lxc/templates/lxc-oci"

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		fmt.Println("    Note: lxc-oci templates not found, skipping adjustment")
		return nil
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("Failed to read template: %v", err)
	}

	modified := strings.ReplaceAll(string(content), "set -eu", "set -u")

	if err := os.WriteFile(templatePath, []byte(modified), 0755); err != nil {
		return fmt.Errorf("Failed to write template: %v", err)
	}

	fmt.Println("OCI template adjusted")
	return nil
}

func (i *Initializer) checkRequiredTools() error {
	tools := []string{"skopeo", "umoci", "jq"}
	missing := []string{}

	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required tools: %s. Please install them", strings.Join(missing, ", "))
	}

	fmt.Println("Required tools installed")
	return nil
}

func (i *Initializer) createContainer() error {
	containerPath := i.config.GetContainerPath()

	if _, err := os.Stat(containerPath); err == nil {
		fmt.Println("Container already exists")
		return nil
	}

	fmt.Printf("Creating LXC container from %s...\n", i.image)

	cmd := exec.Command("lxc-create",
		"-n", i.config.ContainerName,
		"-t", "oci",
		"--",
		"-u", i.image)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	fmt.Println("Container created")
	return nil
}

func (i *Initializer) createDataDirectory() error {
	if err := os.MkdirAll(i.config.DataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	fmt.Printf("Data directory: %s\n", i.config.DataPath)
	return nil
}

func (i *Initializer) adjustContainerConfig() error {
	configPath := i.config.GetConfigFilePath()

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	for _, line := range lines {
		if !strings.Contains(line, "lxc.include") {
			newLines = append(newLines, line)
		}
	}

	additionalConfig := fmt.Sprintf(`
### Redway Configuration
lxc.init.cmd = /init androidboot.hardware=redroid androidboot.redroid_gpu_mode=%s
lxc.apparmor.profile = unconfined
lxc.autodev = 1
lxc.autodev.tmpfs.size = 25000000
lxc.mount.entry = %s data none bind 0 0
`, i.config.GPUMode, i.config.DataPath)

	newLines = append(newLines, additionalConfig)

	finalContent := strings.Join(newLines, "\n")

	if err := os.WriteFile(configPath, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	fmt.Println("Container config adjusted")
	return nil
}

func (i *Initializer) applyNetworkingWorkaround() error {
	workaroundPath := filepath.Join(i.config.GetRootfsPath(), "vendor", "bin", "ipconfigstore")

	if _, err := os.Stat(workaroundPath); err == nil {
		if err := os.Remove(workaroundPath); err != nil {
			fmt.Printf("Warning: Could not remove ipconfigstore: %v\n", err)
		} else {
			fmt.Println("Networking workaround applied")
		}
	} else {
		fmt.Println("Networking workaround not needed")
	}

	return nil
}
