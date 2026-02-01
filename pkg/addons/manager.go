package addons

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reddock/pkg/ui"
	"strings"
)

type AddonManager struct {
	availableAddons map[string]Addon
	workDir         string
}

func NewAddonManager() *AddonManager {
	addons := map[string]Addon{
		"houdini":      NewHoudiniAddon(),
		"ndk":          NewNDKAddon(),
		"litegapps":    NewLiteGappsAddon(),
		"mindthegapps": NewMindTheGappsAddon(),
		"opengapps":    NewOpenGappsAddon(),
	}

	return &AddonManager{
		availableAddons: addons,
		workDir:         "/tmp/reddock-addons",
	}
}

func (am *AddonManager) GetAddon(name string) (Addon, error) {
	addon, ok := am.availableAddons[name]
	if !ok {
		return nil, fmt.Errorf("addon '%s' not found", name)
	}
	return addon, nil
}

func (am *AddonManager) ListAddons() []string {
	var names []string
	for name := range am.availableAddons {
		names = append(names, name)
	}
	return names
}

func (am *AddonManager) InstallAddon(addonName, version, arch string) error {
	addon, err := am.GetAddon(addonName)
	if err != nil {
		return err
	}

	if !addon.IsSupported(version) {
		return fmt.Errorf("%s does not support Android %s", addon.Name(), version)
	}

	if err := ensureDir(am.workDir); err != nil {
		return err
	}

	spinner := ui.NewSpinner(fmt.Sprintf("Installing %s...", addon.Name()))
	spinner.Start()
	defer func() {
		// Stop spinner if it's still running (in case of error before finish)
		if !spinner.IsDone() {
			spinner.Finish("Installation interrupted")
		}
	}()

	err = addon.Install(version, arch, am.workDir, func(msg string) {
		spinner.SetMessage(msg)
	})
	if err != nil {
		spinner.Finish(fmt.Sprintf("Failed to install %s", addon.Name()))
		return err
	}
	spinner.Finish(fmt.Sprintf("Successfully installed %s", addon.Name()))
	return nil
}

func (am *AddonManager) BuildDockerfile(baseImage string, addons []string) (string, error) {
	var dockerfile strings.Builder

	dockerfile.WriteString(fmt.Sprintf("FROM %s\n", baseImage))

	for _, addonName := range addons {
		addon, err := am.GetAddon(addonName)
		if err != nil {
			return "", err
		}
		instructions := addon.DockerfileInstructions()
		if instructions != "" {
			dockerfile.WriteString(instructions)
		} else {
			// Fallback (should not happen if all addons implement instructions correctly)
			dockerfile.WriteString(fmt.Sprintf("# No instructions for %s\n", addonName))
		}
	}

	return dockerfile.String(), nil
}

func (am *AddonManager) BuildImage(baseImage, targetImage, version, arch string, addonNames []string) error {
	if err := ensureDir(am.workDir); err != nil {
		return err
	}

	fmt.Println("\n=== Building Custom Redroid Image ===")
	fmt.Printf("Base Image: %s\n", baseImage)
	fmt.Printf("Target Image: %s\n", targetImage)
	fmt.Printf("Addons: %v\n\n", addonNames)

	for _, addonName := range addonNames {
		if err := am.InstallAddon(addonName, version, arch); err != nil {
			fmt.Printf("Warning: Failed to install %s: %v\n", addonName, err)
			fmt.Printf("Continuing without %s...\n", addonName)
		}
	}

	dockerfileContent, err := am.BuildDockerfile(baseImage, addonNames)
	if err != nil {
		return err
	}

	dockerfilePath := filepath.Join(am.workDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %v", err)
	}

	fmt.Println("Dockerfile created:")
	fmt.Println(dockerfileContent)

	spinner := ui.NewSpinner("Building Docker image...")
	spinner.Start()

	cmd := exec.Command("docker", "build", "-t", targetImage, am.workDir)
	// cmd.Stdout = os.Stdout // We don't want docker output to mess up spinner
	// cmd.Stderr = os.Stderr
	// But catching specific errors might be good. For now let's just capture output
	output, err := cmd.CombinedOutput()

	if err != nil {
		spinner.Finish("Failed to build Docker image")
		fmt.Println(string(output))
		return fmt.Errorf("failed to build Docker image: %v", err)
	}
	spinner.Finish(fmt.Sprintf("Successfully built %s", targetImage))

	return nil
}

func (am *AddonManager) GetSupportedVersions(addonName string) ([]string, error) {
	addon, err := am.GetAddon(addonName)
	if err != nil {
		return nil, err
	}
	return addon.SupportedVersions(), nil
}

func (am *AddonManager) Cleanup() error {
	return os.RemoveAll(am.workDir)
}
