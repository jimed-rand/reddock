package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"redway/pkg/config"
)

type LXCPreparer struct {
	config *config.Config
}

type Initializer struct {
	config    *config.Config
	container *config.Container
}

func NewLXCPreparer() *LXCPreparer {
	cfg, _ := config.Load()
	return &LXCPreparer{
		config: cfg,
	}
}

func NewInitializer(containerName, image string) *Initializer {
	cfg, _ := config.Load()

	container := cfg.GetContainer(containerName)
	if container == nil {
		container = &config.Container{
			Name:        containerName,
			ImageURL:    image,
			DataPath:    config.GetDefaultDataPath(containerName),
			LogFile:     containerName + ".log",
			GPUMode:     config.DefaultGPUMode,
			Initialized: false,
		}
		cfg.AddContainer(container)
	} else {
		container.ImageURL = image
	}

	return &Initializer{
		config:    cfg,
		container: container,
	}
}

func (p *LXCPreparer) PrepareLXC() error {
	fmt.Println("Preparing LXC system...")

	steps := []struct {
		name string
		fn   func() error
	}{
		{"Checking kernel modules", p.checkKernelModules},
		{"Checking LXC tools", p.checkLXCTools},
		{"Setting up LXC networking", p.setupLXCNetworking},
		{"Preparing LXC capabilities", p.prepareLXCCapabilities},
		{"Adjusting OCI template", p.adjustOCITemplate},
		{"Checking required tools", p.checkRequiredTools},
	}

	for _, step := range steps {
		fmt.Printf("[*] %s...\n", step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s failed: %v", step.name, err)
		}
	}

	p.config.LXCReady = true
	if err := config.Save(p.config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Println("\nLXC system prepared successfully!")
	return nil
}

func (p *LXCPreparer) UnprepareLXC() error {
	fmt.Println("Cleaning up LXC system...")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if len(cfg.Containers) > 0 {
		return fmt.Errorf("cannot unprepare LXC: containers still exist. Remove all containers first with 'redway remove'")
	}

	steps := []struct {
		name string
		fn   func() error
	}{
		{"Removing NAT rules", p.removeNATRules},
		{"Removing bridge", p.removeBridge},
		{"Cleaning default config", p.cleanDefaultConfig},
	}

	for _, step := range steps {
		fmt.Printf("[*] %s...\n", step.name)
		if err := step.fn(); err != nil {
			fmt.Printf("Warning: %s failed: %v\n", step.name, err)
		}
	}

	p.config.LXCReady = false
	if err := config.Save(p.config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Println("\nLXC system cleanup completed!")
	fmt.Println("Note: Kernel modules and installed packages were not removed")
	return nil
}

func (i *Initializer) Initialize() error {
	fmt.Println("Initializing the container...")
	fmt.Printf("Container: %s\n", i.container.Name)
	fmt.Printf("Image: %s\n\n", i.container.ImageURL)

	if !i.config.LXCReady {
		return fmt.Errorf("LXC system is not prepared. Please run 'redway prepare-lxc' first as root/sudo")
	}

	if err := i.cleanupExistingContainer(); err != nil {
		return fmt.Errorf("pre-init cleanup failed: %v", err)
	}

	steps := []struct {
		name string
		fn   func() error
	}{
		{"Creating container", i.createContainer},
		{"Fixing container filesystem", i.fixContainerFilesystem},
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

	i.container.Initialized = true
	i.config.AddContainer(i.container)
	if err := config.Save(i.config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Println("\nThe container has been initialized successfully!")
	fmt.Println("\nNext steps:")
	fmt.Printf("  redway start %s        # Start the container\n", i.container.Name)
	fmt.Printf("  redway adb-connect %s  # Get ADB connection info\n", i.container.Name)
	fmt.Printf("  redway shell %s        # Access container shell\n", i.container.Name)

	return nil
}

func (p *LXCPreparer) checkKernelModules() error {
	binderFound := false
	binderPaths := []string{
		"/sys/module/binder_linux",
		"/sys/module/binder",
		"/dev/binderfs",
		"/dev/binder",
	}

	for _, path := range binderPaths {
		if _, err := os.Stat(path); err == nil {
			binderFound = true
			break
		}
	}

	if binderFound {
		fmt.Println("Binder support (binderfs/binder module) detected")
		return nil
	}

	fmt.Println("Binder support (binderfs/binder module) not detected")
	fmt.Println("You need to enable the binderfs in your kernel or install binder module.")
	fmt.Println("                                               ")
	fmt.Println("If container fails to start, try loading modules:")
	fmt.Println("                                               ")
	fmt.Println("sudo modprobe binder_linux devices=\"binder,hwbinder,vndbinder\"")
	fmt.Println("                                               ")
	fmt.Println("See README for more information.")

	return nil
}

func (p *LXCPreparer) checkLXCTools() error {
	tools := []string{"lxc-create", "lxc-start", "lxc-stop", "lxc-info"}

	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s not found. Please install lxc-utils", tool)
		}
	}

	fmt.Println("LXC tools available")
	return nil
}

func (p *LXCPreparer) setupLXCNetworking() error {
	bridgeName := config.DefaultBridgeName
	bridgeIP := config.DefaultBridgeIP

	if _, err := os.Stat(fmt.Sprintf("/sys/class/net/%s", bridgeName)); err == nil {
		fmt.Printf("LXC bridge (%s) already exists\n", bridgeName)
		return p.setupNAT()
	}

	fmt.Printf("Setting up LXC networking with bridge %s...\n", bridgeName)

	commands := [][]string{
		{"ip", "link", "add", "name", bridgeName, "type", "bridge"},
		{"ip", "addr", "add", fmt.Sprintf("%s/24", bridgeIP), "dev", bridgeName},
		{"ip", "link", "set", "dev", bridgeName, "up"},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run %v: %v", args, err)
		}
	}

	if err := p.setupNAT(); err != nil {
		fmt.Printf("Warning: NAT setup failed: %v\n", err)
	}

	fmt.Println("LXC networking configured successfully")
	return nil
}

func (p *LXCPreparer) setupNAT() error {
	bridgeSubnet := config.DefaultBridgeSubnet
	fmt.Printf("Setting up NAT for container networking (%s)...\n", bridgeSubnet)

	checkCmd := exec.Command("sh", "-c", fmt.Sprintf("iptables -t nat -C POSTROUTING -s %s ! -d %s -j MASQUERADE 2>/dev/null", bridgeSubnet, bridgeSubnet))
	if err := checkCmd.Run(); err == nil {
		fmt.Println("NAT rule already exists")
		return p.ensureIPForwarding()
	}

	natCmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", bridgeSubnet, "!", "-d", bridgeSubnet, "-j", "MASQUERADE")
	if err := natCmd.Run(); err != nil {
		return fmt.Errorf("failed to add NAT rule: %v", err)
	}

	fmt.Println("NAT rule added successfully")
	return p.ensureIPForwarding()
}

func (p *LXCPreparer) ensureIPForwarding() error {
	content, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
	if err == nil && strings.TrimSpace(string(content)) == "1" {
		fmt.Println("IP forwarding already enabled")
		return nil
	}

	fmt.Println("Enabling IP forwarding...")
	fwdCmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := fwdCmd.Run(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %v", err)
	}

	sysctlConf := "/etc/sysctl.d/99-lxc-ip-forward.conf"
	if _, err := os.Stat(sysctlConf); os.IsNotExist(err) {
		if err := os.WriteFile(sysctlConf, []byte("net.ipv4.ip_forward=1\n"), 0644); err != nil {
			fmt.Printf("Warning: Could not persist IP forwarding setting: %v\n", err)
		} else {
			fmt.Println("IP forwarding persisted to sysctl")
		}
	}

	return nil
}

func (p *LXCPreparer) prepareLXCCapabilities() error {
	fmt.Println("Preparing LXC capabilities...")

	lxcConfDir := "/etc/lxc"
	if err := os.MkdirAll(lxcConfDir, 0755); err != nil {
		return fmt.Errorf("failed to create lxc config dir: %v", err)
	}

	defaultConf := filepath.Join(lxcConfDir, "default.conf")
	content := fmt.Sprintf(`lxc.net.0.type = veth
lxc.net.0.link = %s
lxc.net.0.flags = up
lxc.net.0.hwaddr = ee:aa:ca:fe:55:01
`, config.DefaultBridgeName)

	if err := os.WriteFile(defaultConf, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write default lxc config: %v", err)
	}

	fmt.Println("Default LXC configuration updated with custom bridge")
	return nil
}

func (p *LXCPreparer) adjustOCITemplate() error {
	templatePath := "/usr/share/lxc/templates/lxc-oci"

	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		fmt.Println("Note: lxc-oci template not found, skipping adjustment")
		return nil
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %v", err)
	}

	modified := strings.ReplaceAll(string(content), "set -eu", "set -u")

	if err := os.WriteFile(templatePath, []byte(modified), 0755); err != nil {
		return fmt.Errorf("failed to write template: %v", err)
	}

	fmt.Println("OCI template adjusted")
	return nil
}

func (p *LXCPreparer) checkRequiredTools() error {
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

func (p *LXCPreparer) removeNATRules() error {
	bridgeSubnet := config.DefaultBridgeSubnet

	checkCmd := exec.Command("sh", "-c", fmt.Sprintf("iptables -t nat -C POSTROUTING -s %s ! -d %s -j MASQUERADE 2>/dev/null", bridgeSubnet, bridgeSubnet))
	if err := checkCmd.Run(); err != nil {
		fmt.Println("NAT rule does not exist")
		return nil
	}

	delCmd := exec.Command("iptables", "-t", "nat", "-D", "POSTROUTING", "-s", bridgeSubnet, "!", "-d", bridgeSubnet, "-j", "MASQUERADE")
	if err := delCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove NAT rule: %v", err)
	}

	fmt.Println("NAT rule removed")
	return nil
}

func (p *LXCPreparer) removeBridge() error {
	bridgeName := config.DefaultBridgeName

	if _, err := os.Stat(fmt.Sprintf("/sys/class/net/%s", bridgeName)); os.IsNotExist(err) {
		fmt.Printf("Bridge %s does not exist\n", bridgeName)
		return nil
	}

	commands := [][]string{
		{"ip", "link", "set", "dev", bridgeName, "down"},
		{"ip", "link", "delete", bridgeName},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run %v: %v", args, err)
		}
	}

	fmt.Printf("Bridge %s removed\n", bridgeName)
	return nil
}

func (p *LXCPreparer) cleanDefaultConfig() error {
	defaultConf := "/etc/lxc/default.conf"

	if _, err := os.Stat(defaultConf); os.IsNotExist(err) {
		fmt.Println("Default LXC config does not exist")
		return nil
	}

	if err := os.Remove(defaultConf); err != nil {
		return fmt.Errorf("failed to remove default config: %v", err)
	}

	fmt.Println("Default LXC config removed")
	return nil
}

func (i *Initializer) cleanupExistingContainer() error {
	containerPath := i.container.GetContainerPath()

	if _, err := os.Stat(containerPath); os.IsNotExist(err) {
		return nil
	}

	fmt.Printf("Existing container directory found at %s\n", containerPath)
	
	cmd := exec.Command("lxc-info", "-n", i.container.Name, "-s")
	if output, err := cmd.Output(); err == nil {
		if strings.Contains(string(output), "RUNNING") {
			fmt.Println("Stopping running container...")
			stopCmd := exec.Command("lxc-stop", "-k", "-n", i.container.Name)
			stopCmd.Run()
		}
	}

	fmt.Println("Cleaning up existing container...")
	destroyCmd := exec.Command("lxc-destroy", "-n", i.container.Name)
	destroyCmd.Run()

	if _, err := os.Stat(containerPath); err == nil {
		fmt.Printf("Force removing leftover directory: %s\n", containerPath)
		if err := os.RemoveAll(containerPath); err != nil {
			return fmt.Errorf("failed to remove existing container directory: %v", err)
		}
		fmt.Println("Existing container cleaned up")
	}

	return nil
}

func (i *Initializer) createContainer() error {
	containerPath := i.container.GetContainerPath()

	if _, err := os.Stat(containerPath); err == nil {
		fmt.Println("Container already exists")
		return nil
	}

	fmt.Printf("Creating LXC container from %s...\n", i.container.ImageURL)

	cmd := exec.Command("lxc-create",
		"-n", i.container.Name,
		"-t", "oci",
		"--",
		"-u", i.container.ImageURL)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	fmt.Println("Container created")
	return nil
}

func (i *Initializer) fixContainerFilesystem() error {
	rootfs := i.container.GetRootfsPath()

	etcDir := filepath.Join(rootfs, "etc")
	if stat, err := os.Stat(etcDir); err == nil && stat.IsDir() {
		fmt.Println("/etc directory already exists")
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(etcDir, 0755); err != nil {
			return fmt.Errorf("failed to create /etc directory: %v", err)
		}
		fmt.Println("Created /etc directory")
	} else {
		return fmt.Errorf("failed to stat /etc directory: %v", err)
	}

	hostnamePath := filepath.Join(etcDir, "hostname")
	if _, err := os.Stat(hostnamePath); os.IsNotExist(err) {
		if err := os.WriteFile(hostnamePath, []byte(i.container.Name+"\n"), 0644); err != nil {
			return fmt.Errorf("failed to create /etc/hostname: %v", err)
		}
		fmt.Println("Created /etc/hostname")
	} else {
		fmt.Println("/etc/hostname already exists")
	}

	hostsPath := filepath.Join(etcDir, "hosts")
	if _, err := os.Stat(hostsPath); os.IsNotExist(err) {
		hostsContent := fmt.Sprintf("127.0.0.1 localhost\n127.0.1.1 %s\n::1 localhost ip6-localhost ip6-loopback\n", i.container.Name)
		if err := os.WriteFile(hostsPath, []byte(hostsContent), 0644); err != nil {
			return fmt.Errorf("failed to create /etc/hosts: %v", err)
		}
		fmt.Println("Created /etc/hosts")
	} else {
		fmt.Println("/etc/hosts already exists")
	}

	fmt.Println("Container filesystem fixed")
	return nil
}

func (i *Initializer) createDataDirectory() error {
	if err := os.MkdirAll(i.container.DataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	fmt.Printf("Data directory: %s\n", i.container.DataPath)
	return nil
}

func (i *Initializer) adjustContainerConfig() error {
	configPath := i.container.GetConfigFilePath()

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
lxc.net.0.type = veth
lxc.net.0.link = %s
lxc.net.0.flags = up
lxc.init.cmd = /init androidboot.hardware=redroid androidboot.redroid_gpu_mode=%s
lxc.apparmor.profile = unconfined
lxc.autodev = 1
lxc.autodev.tmpfs.size = 25000000
lxc.mount.entry = %s data none bind 0 0
`, config.DefaultBridgeName, i.container.GPUMode, i.container.DataPath)

	newLines = append(newLines, additionalConfig)

	finalContent := strings.Join(newLines, "\n")

	if err := os.WriteFile(configPath, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	fmt.Println("Container config adjusted")
	return nil
}

func (i *Initializer) applyNetworkingWorkaround() error {
	workaroundPath := filepath.Join(i.container.GetRootfsPath(), "vendor", "bin", "ipconfigstore")

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
