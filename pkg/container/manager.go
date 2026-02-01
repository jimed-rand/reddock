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
		return nil, fmt.Errorf("Addon '%s' not found", name)
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

func (am *AddonManager) PrepareAddon(addonName, version, arch string) error {
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

	spinner := ui.NewSpinner(fmt.Sprintf("Preparing %s...", addon.Name()))
	spinner.Start()
	defer func() {
		if !spinner.IsDone() {
			spinner.Finish("Preparation interrupted")
		}
	}()

	err = addon.Install(version, arch, am.workDir, func(msg string) {
		spinner.SetMessage(msg)
	})
	if err != nil {
		spinner.Finish(fmt.Sprintf("Failed to prepare %s", addon.Name()))
		return err
	}
	spinner.Finish(fmt.Sprintf("Successfully prepared %s", addon.Name()))
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
		}
	}

	return dockerfile.String(), nil
}

func (am *AddonManager) BuildCustomImage(baseImage, targetImage, version, arch string, addonNames []string, pushToRegistry bool) error {
	if err := ensureDir(am.workDir); err != nil {
		return err
	}

	fmt.Println("\n=== Building custom Redroid Image ===")
	fmt.Printf("Base Image: %s\n", baseImage)
	fmt.Printf("Target Image: %s\n", targetImage)
	fmt.Printf("Addons: %v\n\n", addonNames)

	runtime := getContainerRuntime()

	fmt.Printf("Pulling base image %s...\n", baseImage)
	if err := pullImage(runtime, baseImage); err != nil {
		return fmt.Errorf("Failed to pull base image: %v", err)
	}
	fmt.Println("Base image pulled successfully\n")

	for _, addonName := range addonNames {
		if err := am.PrepareAddon(addonName, version, arch); err != nil {
			fmt.Printf("Warning: Failed to prepare %s: %v\n", addonName, err)
			fmt.Printf("Continuing without %s...\n", addonName)
		}
	}

	dockerfileContent, err := am.BuildDockerfile(baseImage, addonNames)
	if err != nil {
		return err
	}

	dockerfilePath := filepath.Join(am.workDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("Failed to write Dockerfile: %v", err)
	}

	fmt.Println("Dockerfile created:")
	fmt.Println(dockerfileContent)

	spinner := ui.NewSpinner("Building Docker image...")
	spinner.Start()

	cmd := exec.Command(runtime, "build", "-t", targetImage, am.workDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		spinner.Finish("Failed to build Docker image")
		fmt.Println(string(output))
		return fmt.Errorf("Failed to build Docker image: %v", err)
	}
	spinner.Finish(fmt.Sprintf("Successfully built %s", targetImage))

	if pushToRegistry {
		if err := am.PushToRegistry(runtime, targetImage); err != nil {
			return fmt.Errorf("Failed to push to registry: %v", err)
		}
	}

	return nil
}

func (am *AddonManager) PushToRegistry(runtime, imageName string) error {
	fmt.Println("\n=== Publishing the image to Docker Hub ===")

	if !strings.Contains(imageName, "/") {
		return fmt.Errorf("The image name must include username/repository format (e.g., username/image:tag)")
	}

	fmt.Println("You need to authenticate with Docker Hub first.")
	fmt.Print("Do you want to login now? [y/N]: ")
	var doLogin string
	fmt.Scanln(&doLogin)

	if strings.ToLower(doLogin) == "y" || strings.ToLower(doLogin) == "yes" {
		fmt.Print("Docker Hub username: ")
		var username string
		fmt.Scanln(&username)

		fmt.Println("Please enter your Docker Hub password or access token:")
		loginCmd := exec.Command(runtime, "login", "-u", username, "--password-stdin")
		loginCmd.Stdin = os.Stdin
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr

		if err := loginCmd.Run(); err != nil {
			return fmt.Errorf("login failed: %v", err)
		}
		fmt.Println("Login successful!\n")
	}

	spinner := ui.NewSpinner(fmt.Sprintf("Pushing %s to Docker Hub...", imageName))
	spinner.Start()

	pushCmd := exec.Command(runtime, "push", imageName)
	output, err := pushCmd.CombinedOutput()

	if err != nil {
		spinner.Finish("Failed to push image")
		fmt.Println(string(output))
		return fmt.Errorf("Push failed: %v", err)
	}

	spinner.Finish(fmt.Sprintf("Successfully pushed %s to Docker Hub", imageName))
	fmt.Printf("\nYour image is now available at: https://hub.docker.com/r/%s\n", imageName)

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

func getContainerRuntime() string {
	if _, err := exec.LookPath("podman"); err == nil {
		return "podman"
	}
	return "docker"
}

func pullImage(runtime, image string) error {
	cmd := exec.Command(runtime, "pull", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
