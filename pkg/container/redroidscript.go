package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"reddock/pkg/config"
	"reddock/pkg/redroidscript"
	"reddock/pkg/ui"
)

// RedroidScriptAddonFlags mirrors the former redroid.py CLI (same semantics, same build output tag).
type RedroidScriptAddonFlags struct {
	Gapps         bool
	LiteGapps     bool
	MindTheGapps  bool
	NDK           bool
	Houdini       bool
	Magisk        bool
	Widevine      bool
	Android       string // empty: derive from container official image
	ScriptPath    string // deprecated: native builder ignores
	Instant       bool   // deprecated: native builder ignores
	TargetImage   string // final tagged image (reddock GPU layer)
	UpdateConfig  bool
}

var redroidScriptAndroidChoices = map[string]struct{}{
	"14.0.0": {}, "13.0.0": {}, "13.0.0_64only": {}, "12.0.0": {}, "12.0.0_64only": {},
	"11.0.0": {}, "11.0.0_64only": {}, "10.0.0": {}, "9.0.0": {}, "8.1.0": {},
}

func toNativeAddons(f RedroidScriptAddonFlags) redroidscript.AddonFlags {
	return redroidscript.AddonFlags{
		Gapps:        f.Gapps,
		LiteGapps:    f.LiteGapps,
		MindTheGapps: f.MindTheGapps,
		NDK:          f.NDK,
		Houdini:      f.Houdini,
		Magisk:       f.Magisk,
		Widevine:     f.Widevine,
	}
}

// RedroidScriptIntermediateImageName returns the tag the native builder produces for this flag set.
func RedroidScriptIntermediateImageName(android string, f RedroidScriptAddonFlags) string {
	s, err := redroidscript.IntermediateImageName(android, toNativeAddons(f))
	if err != nil {
		return "redroid/redroid:" + android
	}
	return s
}

func validateRedroidScriptAndroid(v string) error {
	if v == "" {
		return fmt.Errorf("Android version is empty")
	}
	if _, ok := redroidScriptAndroidChoices[v]; !ok {
		return fmt.Errorf("unsupported Android version %q for patch (use -a with one of: 8.1.0 … 14.0.0, *64only variants)", v)
	}
	return nil
}

// mapOfficial64OnlyTag maps official *-64only image tags to the Android API line used for add-on metadata.
func mapOfficial64OnlyTag(v string) string {
	switch v {
	case "13.0.0_64only":
		return "13.0.0"
	case "11.0.0_64only":
		return "11.0.0"
	default:
		return v
	}
}

// IsOfficialRedroidBaseImage reports whether the configured image is an official ReDroid
// Docker Hub image (the only base reddock will feed into patching).
func IsOfficialRedroidBaseImage(imageURL string) bool {
	s := strings.TrimSpace(imageURL)
	return strings.HasPrefix(s, "redroid/redroid:")
}

func resolveAndroidForRedroidScript(cont *config.Container, override string) (string, error) {
	if override != "" {
		o := mapOfficial64OnlyTag(strings.TrimSpace(override))
		return o, validateRedroidScriptAndroid(o)
	}
	if cont == nil {
		return "", fmt.Errorf("container not found")
	}
	if !strings.HasPrefix(cont.ImageURL, "redroid/redroid:") {
		return "", fmt.Errorf("container image is not official redroid/redroid; set Android with -a (see patch --android choices)")
	}
	v := mapOfficial64OnlyTag(config.ExtractVersionFromImage(cont.ImageURL))
	if v == "" {
		return "", fmt.Errorf("could not parse Android version from image %q; use -a", cont.ImageURL)
	}
	if err := validateRedroidScriptAndroid(v); err != nil {
		return "", err
	}
	return v, nil
}

func redroidScriptRuntimeFlag(rt Runtime) string {
	if rt != nil && rt.Name() == "podman" {
		return "podman"
	}
	return "docker"
}

// BuildImageWithRedroidScript builds a customized ReDroid image using reddock’s native Go port
// of ayasa520/redroid-script (no Python clone, no pip). Requires docker/podman, tar, and lzip when using OpenGapps.
func BuildImageWithRedroidScript(containerName string, f RedroidScriptAddonFlags) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cont := cfg.GetContainer(containerName)
	if cont == nil {
		return fmt.Errorf("container %q not found", containerName)
	}

	if !IsOfficialRedroidBaseImage(cont.ImageURL) {
		return fmt.Errorf("reddock patch only supports containers whose image_url is an official ReDroid image (redroid/redroid:…); %q uses %q", containerName, cont.ImageURL)
	}

	if !anyAddonSelected(f) {
		return fmt.Errorf("select at least one addon: -g -lg -mtg -n -i -m -w")
	}

	android, err := resolveAndroidForRedroidScript(cont, f.Android)
	if err != nil {
		return err
	}

	if f.ScriptPath != "" || f.Instant {
		fmt.Println("Note: --script-path / --instant are ignored; add-on build runs inside reddock (native).")
	}
	if cfg != nil && strings.TrimSpace(cfg.RedroidScriptPath) != "" {
		fmt.Println("Note: config redroid_script_path is ignored; add-on build runs inside reddock (native).")
	}

	if _, err := exec.LookPath("tar"); err != nil {
		return fmt.Errorf("tar not found in PATH (required for add-on build)")
	}
	if f.Gapps {
		if _, err := exec.LookPath("lzip"); err != nil {
			fmt.Println("Warning: lzip not found; OpenGapps extraction needs tar with --lzip. Install lzip if -g fails.")
		}
	}

	rt := NewRuntime()
	runtimeFlag := redroidScriptRuntimeFlag(rt)

	intermediate, err := redroidscript.IntermediateImageName(android, toNativeAddons(f))
	if err != nil {
		return err
	}

	workDir, err := os.MkdirTemp("", "reddock-patch-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(workDir) }()

	fmt.Println()
	fmt.Println("Running native redroid-script build (downloads blobs, docker build) …")
	fmt.Printf("  Intermediate image will be: %s\n", intermediate)

	if err := redroidscript.Build(workDir, android, runtimeFlag, toNativeAddons(f)); err != nil {
		return fmt.Errorf("patch build: %w", err)
	}

	if err := rt.Command("image", "inspect", intermediate).Run(); err != nil {
		return fmt.Errorf("intermediate image %s not found after build; check build output above", intermediate)
	}

	target := strings.TrimSpace(f.TargetImage)
	if target == "" {
		target = config.SuggestCustomImageName(containerName, "redroid-script")
	}
	if err := config.ValidateImageName(target); err != nil {
		return err
	}

	spinner := ui.NewSpinner(fmt.Sprintf("Tagging patched image as %s …", target))
	spinner.Start()
	tcmd := rt.Command("tag", intermediate, target)
	out, err := tcmd.CombinedOutput()
	if err != nil {
		spinner.Finish("docker tag failed")
		if len(out) > 0 {
			fmt.Println(string(out))
		}
		return fmt.Errorf("docker tag %s -> %s: %w", intermediate, target, err)
	}
	spinner.Finish(fmt.Sprintf("Tagged %s", target))

	fmt.Println()
	fmt.Printf("Done. Built image id: %s\n", intermediate)
	fmt.Printf("Tagged name (use in reddock): %s\n", target)
	fmt.Println("GPU mode is not baked into the image; `reddock start` passes androidboot.redroid_gpu_mode as usual.")
	fmt.Println("Point your container at the new image, e.g.:")
	fmt.Printf("  reddock init %s %s   # or edit ~/.config/reddock/config.json image_url\n", containerName, target)

	if f.UpdateConfig {
		cont.ImageURL = target
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saved build but failed to update config: %w", err)
		}
		fmt.Printf("Updated container %q to use image %s\n", containerName, target)
	}

	return nil
}

func anyAddonSelected(f RedroidScriptAddonFlags) bool {
	return f.Gapps || f.LiteGapps || f.MindTheGapps || f.NDK || f.Houdini || f.Magisk || f.Widevine
}

// ParseRedroidScriptCLIArgs parses flags after "reddock patch <name>".
func ParseRedroidScriptCLIArgs(args []string) (containerName string, f RedroidScriptAddonFlags, err error) {
	if len(args) < 1 {
		return "", f, fmt.Errorf("container name is required")
	}
	containerName = args[0]
	i := 1
	for i < len(args) {
		a := args[i]
		switch a {
		case "-a", "--android":
			if i+1 >= len(args) {
				return "", f, fmt.Errorf("%s requires a version (e.g. 11.0.0)", a)
			}
			f.Android = args[i+1]
			i += 2
		case "-g", "--gapps":
			f.Gapps = true
			i++
		case "-lg", "--litegapps":
			f.LiteGapps = true
			i++
		case "-mtg", "--mindthegapps":
			f.MindTheGapps = true
			i++
		case "-n", "--ndk":
			f.NDK = true
			i++
		case "-i", "--houdini":
			f.Houdini = true
			i++
		case "-m", "--magisk":
			f.Magisk = true
			i++
		case "-w", "--widevine":
			f.Widevine = true
			i++
		case "-t", "--target-image":
			if i+1 >= len(args) {
				return "", f, fmt.Errorf("%s requires an image name (e.g. reddock/my:gapps)", a)
			}
			f.TargetImage = args[i+1]
			i += 2
		case "--script-path":
			if i+1 >= len(args) {
				return "", f, fmt.Errorf("--script-path requires a directory")
			}
			f.ScriptPath = args[i+1]
			i += 2
		case "--update-config":
			f.UpdateConfig = true
			i++
		case "--instant":
			f.Instant = true
			i++
		default:
			return "", f, fmt.Errorf("unknown argument: %q", a)
		}
	}
	return containerName, f, nil
}
