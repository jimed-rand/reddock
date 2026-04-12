package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"reddock/pkg/config"
	"reddock/pkg/ui"
)

const (
	// EnvRedroidScriptPath is read when --script-path is unset and config has no redroid_script_path.
	EnvRedroidScriptPath = "REDDOCK_REDROID_SCRIPT"
	// EnvRedroidScriptInstant enables cached upstream clone when set to 1/true/yes (same as --instant).
	EnvRedroidScriptInstant = "REDDOCK_REDROID_SCRIPT_INSTANT"
)

// RedroidScriptRepoURL is used only for the optional instant/cached clone (not bundled in the reddock binary).
const RedroidScriptRepoURL = "https://github.com/ayasa520/redroid-script.git"

// RedroidScriptAddonFlags mirrors redroid.py CLI (same semantics, same build output tag).
type RedroidScriptAddonFlags struct {
	Gapps         bool
	LiteGapps     bool
	MindTheGapps  bool
	NDK           bool
	Houdini       bool
	Magisk        bool
	Widevine      bool
	Android       string // empty: derive from container official image
	ScriptPath    string // optional local clone
	Instant       bool   // use reddock cache clone (git fetch upstream) when no guest path is set
	TargetImage   string // final tagged image (reddock GPU layer)
	UpdateConfig  bool
}

var redroidScriptAndroidChoices = map[string]struct{}{
	"14.0.0": {}, "13.0.0": {}, "12.0.0": {}, "12.0.0_64only": {},
	"11.0.0": {}, "10.0.0": {}, "9.0.0": {}, "8.1.0": {},
}

// RedroidScriptIntermediateImageName returns the tag redroid.py produces (must stay in sync with upstream main()).
func RedroidScriptIntermediateImageName(android string, f RedroidScriptAddonFlags) string {
	tags := []string{android}
	if f.Gapps {
		tags = append(tags, "gapps")
	}
	if f.LiteGapps {
		tags = append(tags, "litegapps")
	}
	if f.MindTheGapps {
		tags = append(tags, "mindthegapps")
	}
	if f.NDK {
		tags = append(tags, "ndk")
	}
	if f.Houdini {
		tags = append(tags, "houdini")
	}
	if f.Magisk {
		tags = append(tags, "magisk")
	}
	if f.Widevine {
		tags = append(tags, "widevine")
	}
	return "redroid/redroid:" + strings.Join(tags, "_")
}

func validateRedroidScriptAndroid(v string) error {
	if v == "" {
		return fmt.Errorf("Android version is empty")
	}
	if _, ok := redroidScriptAndroidChoices[v]; !ok {
		return fmt.Errorf("unsupported Android version %q for redroid-script (use -a with one of: 8.1.0 … 14.0.0, 12.0.0_64only)", v)
	}
	return nil
}

// mapOfficial64OnlyTag maps official *-64only image tags to the Android API line redroid-script exposes.
// Upstream argparse does not list 11.0.0_64only / 13.0.0_64only; those images still pair with the same script lines.
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

func resolveAndroidForRedroidScript(cont *config.Container, override string) (string, error) {
	if override != "" {
		o := mapOfficial64OnlyTag(strings.TrimSpace(override))
		return o, validateRedroidScriptAndroid(o)
	}
	if cont == nil {
		return "", fmt.Errorf("container not found")
	}
	if !strings.HasPrefix(cont.ImageURL, "redroid/redroid:") {
		return "", fmt.Errorf("container image is not official redroid/redroid; set Android with -a (see redroid-script --android choices)")
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

// findRedroidScriptRoot resolves a directory that contains redroid.py (guest clone).
// Accepts the repo root or a parent folder such as .../redroid-script when the clone lives in .../redroid-script/redroid-script/.
func findRedroidScriptRoot(userPath string) (string, error) {
	userPath = filepath.Clean(userPath)
	candidates := []string{
		userPath,
		filepath.Join(userPath, "redroid-script"),
	}
	var tried []string
	for _, dir := range candidates {
		tried = append(tried, dir)
		py := filepath.Join(dir, "redroid.py")
		if st, err := os.Stat(py); err == nil && !st.IsDir() {
			if err := validateRedroidScriptLayout(dir); err != nil {
				return "", err
			}
			return dir, nil
		}
	}
	return "", fmt.Errorf("redroid-script not found: need redroid.py under one of:\n  %s\nClone https://github.com/ayasa520/redroid-script and set --script-path, %s, or redroid_script_path in config.json",
		strings.Join(tried, "\n  "), EnvRedroidScriptPath)
}

func validateRedroidScriptLayout(dir string) error {
	for _, rel := range []string{"stuff", "tools", filepath.Join("tools", "helper.py")} {
		p := filepath.Join(dir, rel)
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("redroid-script layout invalid (missing %s): %w", rel, err)
		}
	}
	return nil
}

func wantInstantMode(f RedroidScriptAddonFlags, cfg *config.Config) bool {
	if f.Instant {
		return true
	}
	if s := strings.ToLower(strings.TrimSpace(os.Getenv(EnvRedroidScriptInstant))); s == "1" || s == "true" || s == "yes" {
		return true
	}
	if cfg != nil && cfg.RedroidScriptInstant {
		return true
	}
	return false
}

// ensureRedroidScriptCache clones or refreshes ayasa520/redroid-script under ~/.config/reddock/cache/ (legacy “instant” path).
func ensureRedroidScriptCache() (string, error) {
	dest := filepath.Join(config.GetConfigDir(), "cache", "redroid-script")
	py := filepath.Join(dest, "redroid.py")
	if _, err := os.Stat(py); err == nil {
		_ = exec.Command("git", "-C", dest, "pull", "--ff-only").Run()
		if err := validateRedroidScriptLayout(dest); err != nil {
			return "", fmt.Errorf("cached redroid-script at %s is invalid: %w", dest, err)
		}
		return dest, nil
	}
	parent := filepath.Dir(dest)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return "", fmt.Errorf("create cache dir: %w", err)
	}
	spinner := ui.NewSpinner("Cloning ayasa520/redroid-script (instant mode) …")
	spinner.Start()
	cmd := exec.Command("git", "clone", "--depth", "1", RedroidScriptRepoURL, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		spinner.Finish("Clone failed")
		return "", fmt.Errorf("git clone redroid-script: %w\n%s", err, string(out))
	}
	spinner.Finish("redroid-script ready")
	if err := validateRedroidScriptLayout(dest); err != nil {
		return "", err
	}
	return dest, nil
}

// ResolveRedroidScriptRoot prefers a guest clone (--script-path, env, config path); otherwise optional instant cache when wantInstant is true.
func ResolveRedroidScriptRoot(scriptPathFromFlag string, cfg *config.Config, wantInstant bool) (string, error) {
	if p := strings.TrimSpace(scriptPathFromFlag); p != "" {
		return findRedroidScriptRoot(p)
	}
	if p := strings.TrimSpace(os.Getenv(EnvRedroidScriptPath)); p != "" {
		return findRedroidScriptRoot(p)
	}
	if cfg != nil {
		if p := strings.TrimSpace(cfg.RedroidScriptPath); p != "" {
			return findRedroidScriptRoot(p)
		}
	}
	if wantInstant {
		if _, err := exec.LookPath("git"); err != nil {
			return "", fmt.Errorf("git not found in PATH (required for instant mode clone of redroid-script)")
		}
		repo, err := ensureRedroidScriptCache()
		if err != nil {
			return "", err
		}
		fmt.Println("Using reddock instant cache (upstream redroid-script clone). For a fixed local tree, use --script-path or redroid_script_path.")
		return repo, nil
	}
	return "", fmt.Errorf("redroid-script path not set. Either use a local clone:\n"+
		"  reddock dockerfile addons <name> --script-path /path/to/redroid-script ...\n"+
		"  export %s=/path/to/redroid-script\n"+
		"  set \"redroid_script_path\" in %s\n"+
		"Or use instant mode (reddock clones upstream to its cache):\n"+
		"  reddock dockerfile addons <name> --instant ...\n"+
		"  export %s=1\n"+
		"  set \"redroid_script_instant\": true in %s",
		EnvRedroidScriptPath, config.GetConfigPath(),
		EnvRedroidScriptInstant, config.GetConfigPath())
}

func redroidScriptRuntimeFlag(rt Runtime) string {
	if rt != nil && rt.Name() == "podman" {
		return "podman"
	}
	return "docker"
}

func pipInstallRedroidScriptDeps(repo string) error {
	req := filepath.Join(repo, "requirements.txt")
	if _, err := os.Stat(req); err != nil {
		return nil
	}
	cmd := exec.Command("python3", "-m", "pip", "install", "-q", "-r", req)
	cmd.Dir = repo
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pip install -r requirements.txt failed (need python3-pip?): %w", err)
	}
	return nil
}

// BuildImageWithRedroidScript runs upstream redroid.py unmodified, then applies a thin Dockerfile layer
// with the same GPU CMD as reddock's DockerfileGenerator (official ReDroid + user config).
func BuildImageWithRedroidScript(containerName string, f RedroidScriptAddonFlags) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	cont := cfg.GetContainer(containerName)
	if cont == nil {
		return fmt.Errorf("container %q not found", containerName)
	}

	if !anyAddonSelected(f) {
		return fmt.Errorf("select at least one addon: -g -lg -mtg -n -i -m -w (same as redroid-script)")
	}

	android, err := resolveAndroidForRedroidScript(cont, f.Android)
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("python3"); err != nil {
		return fmt.Errorf("python3 not found in PATH (required to run redroid-script)")
	}
	if _, err := exec.LookPath("lzip"); err != nil {
		fmt.Println("Warning: lzip not found; redroid-script README lists it as a dependency. Install lzip if the build fails.")
	}

	repo, err := ResolveRedroidScriptRoot(f.ScriptPath, cfg, wantInstantMode(f, cfg))
	if err != nil {
		return err
	}
	fmt.Printf("Using redroid-script at: %s\n", repo)
	if err := pipInstallRedroidScriptDeps(repo); err != nil {
		return err
	}

	rt := NewRuntime()
	runtimeFlag := redroidScriptRuntimeFlag(rt)

	args := []string{filepath.Join(repo, "redroid.py"), "-a", android, "-c", runtimeFlag}
	if f.Gapps {
		args = append(args, "-g")
	}
	if f.LiteGapps {
		args = append(args, "-lg")
	}
	if f.MindTheGapps {
		args = append(args, "-mtg")
	}
	if f.NDK {
		args = append(args, "-n")
	}
	if f.Houdini {
		args = append(args, "-i")
	}
	if f.Magisk {
		args = append(args, "-m")
	}
	if f.Widevine {
		args = append(args, "-w")
	}

	intermediate := RedroidScriptIntermediateImageName(android, f)

	fmt.Println()
	fmt.Println("Running redroid-script (downloads addon blobs and builds intermediate image) …")
	fmt.Printf("  Intermediate image will be: %s\n", intermediate)
	cmd := exec.Command("python3", args...)
	cmd.Dir = repo
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("redroid-script failed: %w", err)
	}

	if err := rt.Command("image", "inspect", intermediate).Run(); err != nil {
		return fmt.Errorf("intermediate image %s not found after build; check redroid-script output", intermediate)
	}

	target := strings.TrimSpace(f.TargetImage)
	if target == "" {
		target = config.SuggestCustomImageName(containerName, "redroid-script")
	}
	if err := config.ValidateImageName(target); err != nil {
		return err
	}

	gpu := config.DefaultGPUMode
	if cont.GPUMode != "" {
		gpu = cont.GPUMode
	}

	layerDir, err := os.MkdirTemp("", "reddock-redroid-gpu-layer-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(layerDir)

	df := fmt.Sprintf(`# Thin layer: ReDroid GPU boot args (reddock) on top of redroid-script image
# Base: %s
FROM %s
CMD ["androidboot.redroid_gpu_mode=%s"]
`, intermediate, intermediate, gpu)
	dockerfilePath := filepath.Join(layerDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(df), 0644); err != nil {
		return err
	}

	spinner := ui.NewSpinner(fmt.Sprintf("Building final image %s …", target))
	spinner.Start()

	buildArgs := []string{"build", "-t", target, layerDir}
	if rt.Name() == "docker" {
		buildArgs = []string{"buildx", "build", "--load", "-t", target, layerDir}
	}
	bcmd := rt.Command(buildArgs...)
	out, err := bcmd.CombinedOutput()
	if err != nil {
		spinner.Finish("Build failed")
		fmt.Println(string(out))
		return fmt.Errorf("docker build GPU layer: %w", err)
	}
	spinner.Finish(fmt.Sprintf("Built %s", target))

	fmt.Println()
	fmt.Printf("Done. Intermediate: %s\n", intermediate)
	fmt.Printf("Final image (use in reddock): %s\n", target)
	fmt.Println("Set this image on your container, then init/start as usual, e.g.:")
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

// ParseRedroidScriptCLIArgs parses flags after "reddock dockerfile addons <name>".
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
				return "", f, fmt.Errorf("--script-path requires a directory (root of your redroid-script clone)")
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
