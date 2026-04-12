package redroidscript

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AddonFlags selects optional layers (same ordering as ayasa520/redroid-script redroid.py).
type AddonFlags struct {
	Gapps        bool
	LiteGapps    bool
	MindTheGapps bool
	NDK          bool
	Houdini      bool
	Magisk       bool
	Widevine     bool
}

// IntermediateImageName returns the docker image tag produced for this Android line and flags
// (must match addon install decisions in Build).
func IntermediateImageName(android string, f AddonFlags) (string, error) {
	sfx, err := computeTagSuffixes(android, f)
	if err != nil {
		return "", err
	}
	tags := append([]string{android}, sfx...)
	return "redroid/redroid:" + strings.Join(tags, "_"), nil
}

func computeTagSuffixes(android string, f AddonFlags) ([]string, error) {
	arch, _, err := HostTuple()
	if err != nil {
		return nil, err
	}
	var out []string
	if f.Gapps {
		out = append(out, "gapps")
	}
	if f.LiteGapps {
		out = append(out, "litegapps")
	}
	if f.MindTheGapps {
		out = append(out, "mindthegapps")
	}
	if f.NDK && ndkAndroidOK(android) && (arch == "x86" || arch == "x86_64") {
		out = append(out, "ndk")
	}
	if f.Houdini && houdiniAndroidOK(android) && (arch == "x86" || arch == "x86_64") {
		out = append(out, "houdini")
	}
	if f.Magisk {
		out = append(out, "magisk")
	}
	if f.Widevine {
		out = append(out, "widevine")
	}
	return out, nil
}

// Build prepares addon trees under workDir, writes Dockerfile, and runs docker/podman build.
// workDir must be empty or only contain leftovers this function may replace.
func Build(workDir, android, container string, f AddonFlags) error {
	workDir = filepath.Clean(workDir)
	arch, _, err := HostTuple()
	if err != nil {
		return err
	}

	var docker strings.Builder
	docker.WriteString(fmt.Sprintf("FROM redroid/redroid:%s-latest\n", android))

	if f.Gapps {
		if err := installGapps(workDir, android, arch); err != nil {
			return fmt.Errorf("gapps: %w", err)
		}
		docker.WriteString("COPY gapps /\n")
	}
	if f.LiteGapps {
		if err := installLiteGapps(workDir, android, arch); err != nil {
			return fmt.Errorf("litegapps: %w", err)
		}
		docker.WriteString("COPY litegapps /\n")
	}
	if f.MindTheGapps {
		if err := installMindTheGapps(workDir, android, arch); err != nil {
			return fmt.Errorf("mindthegapps: %w", err)
		}
		docker.WriteString("COPY mindthegapps /\n")
	}
	if f.NDK {
		if ndkAndroidOK(android) && (arch == "x86" || arch == "x86_64") {
			if err := installNDK(workDir); err != nil {
				return fmt.Errorf("ndk: %w", err)
			}
			docker.WriteString("COPY ndk /\n")
		} else {
			printYellow("WARNING: Libndk is only applied for ReDroid 10.0.0–13.x on x86/x86_64 (reddock matrix).")
		}
	}
	if f.Houdini {
		if houdiniAndroidOK(android) && (arch == "x86" || arch == "x86_64") {
			if err := installHoudini(workDir, android); err != nil {
				return fmt.Errorf("houdini: %w", err)
			}
			if android != "8.1.0" {
				hackVer := strings.TrimSuffix(android, "_64only")
				if err := installHoudiniHack(workDir, hackVer); err != nil {
					return fmt.Errorf("houdini_hack: %w", err)
				}
			}
			docker.WriteString("COPY houdini /\n")
		} else {
			printYellow("WARNING: Houdini is not enabled for this Android line in reddock.")
		}
	}
	if f.Magisk {
		if err := installMagisk(workDir, arch); err != nil {
			return fmt.Errorf("magisk: %w", err)
		}
		docker.WriteString("COPY magisk /\n")
	}
	if f.Widevine {
		if err := installWidevine(workDir, android, arch); err != nil {
			return fmt.Errorf("widevine: %w", err)
		}
		docker.WriteString("COPY widevine /\n")
	}

	sfx, err := computeTagSuffixes(android, f)
	if err != nil {
		return err
	}
	tags := append([]string{android}, sfx...)

	df := docker.String()
	fmt.Println("\nDockerfile\n" + df)
	if err := os.WriteFile(filepath.Join(workDir, "Dockerfile"), []byte(df), 0644); err != nil {
		return err
	}

	image := "redroid/redroid:" + strings.Join(tags, "_")
	cmd := exec.Command(container, "build", "-t", image, ".")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s build: %w", container, err)
	}
	printGreen("Successfully built " + image)
	return nil
}

func ndkAndroidOK(android string) bool {
	switch android {
	case "10.0.0", "11.0.0", "12.0.0", "12.0.0_64only", "13.0.0", "13.0.0_64only":
		return true
	default:
		return false
	}
}

func houdiniAndroidOK(android string) bool {
	switch android {
	case "8.1.0", "9.0.0", "10.0.0", "11.0.0", "12.0.0", "12.0.0_64only", "13.0.0", "13.0.0_64only", "14.0.0":
		return true
	default:
		return false
	}
}
