package addons

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

type HoudiniAddon struct {
	*BaseAddon
	dlLinks   map[string]map[string][]string
	extractTo string
}

func NewHoudiniAddon() *HoudiniAddon {
	versions := []string{"8.1.0", "9.0.0", "10.0.0", "11.0.0", "12.0.0", "13.0.0", "14.0.0", "15.0.0", "16.0.0"}

	dlLinks := map[string]map[string][]string{
		"8.1.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/46682f423b8497db3f96222f2669d770eff764c3.zip"},
			"md5": {"cd4dd2891aa18e7699d33dcc3fe3ffd4"},
		},
		"9.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/46682f423b8497db3f96222f2669d770eff764c3.zip"},
			"md5": {"cd4dd2891aa18e7699d33dcc3fe3ffd4"},
		},
		"10.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"11.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"12.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"13.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"14.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"15.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
		"16.0.0": {
			"url": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip"},
			"md5": {"cb7ffac26d47ec7c89df43818e126b47"},
		},
	}

	baseAddon := NewBaseAddon("Houdini", AddonTypeHoudini, versions)
	return &HoudiniAddon{
		BaseAddon: baseAddon,
		dlLinks:   dlLinks,
		extractTo: "/tmp/houdiniunpack",
	}
}

func (h *HoudiniAddon) Download(version, arch string, onStatus func(string)) error {
	if !h.IsSupported(version) {
		return fmt.Errorf("houdini not supported for Android %s", version)
	}

	if arch != "x86" && arch != "x86_64" {
		return fmt.Errorf("houdini only supports x86/x86_64 architecture")
	}

	url := h.dlLinks[version]["url"][0]
	filename := filepath.Join(h.downloadDir, "libhoudini.zip")

	if err := ensureDir(h.downloadDir); err != nil {
		return err
	}

	onStatus(fmt.Sprintf("Downloading Houdini for Android %s...", version))
	return downloadFile(url, filename)
}

func (h *HoudiniAddon) Extract(version, arch string, onStatus func(string)) error {
	filename := filepath.Join(h.downloadDir, "libhoudini.zip")

	if err := ensureDir(h.extractTo); err != nil {
		return err
	}

	onStatus("Extracting Houdini archive...")
	return extractZip(filename, h.extractTo)
}

func (h *HoudiniAddon) Copy(version, arch, outputDir string, onStatus func(string)) error {
	copyDir := filepath.Join(outputDir, "houdini")

	if err := os.RemoveAll(copyDir); err != nil {
		return err
	}

	url := h.dlLinks[version]["url"][0]
	re := regexp.MustCompile(`([a-zA-Z0-9]+)\.zip`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return fmt.Errorf("failed to extract name from URL")
	}
	name := matches[1]

	srcDir := filepath.Join(h.extractTo, "vendor_intel_proprietary_houdini-"+name, "prebuilts")
	dstDir := filepath.Join(copyDir, "system")

	onStatus("Copying Houdini library files...")
	if err := copyDir2(srcDir, dstDir); err != nil {
		return err
	}

	initRC := h.getInitRC()
	initPath := filepath.Join(copyDir, "system", "etc", "init", "houdini.rc")
	if err := ensureDir(filepath.Dir(initPath)); err != nil {
		return err
	}

	if err := os.WriteFile(initPath, []byte(initRC), 0644); err != nil {
		return err
	}

	// Houdini Hack logic
	if version != "8.1.0" {
		hackURL := "https://github.com/rote66/redroid_libhoudini_hack/archive/a2194c5e294cbbfdfe87e51eb9eddb4c3621d8c3.zip"
		hackFilename := filepath.Join(h.downloadDir, "libhoudini_hack.zip")
		hackExtractTo := filepath.Join(h.downloadDir, "hack_extract")

		onStatus("Downloading Houdini Hack...")
		if err := downloadFile(hackURL, hackFilename); err != nil {
			return err
		}

		onStatus("Extracting Houdini Hack...")
		if err := extractZip(hackFilename, hackExtractTo); err != nil {
			return err
		}

		hackName := "a2194c5e294cbbfdfe87e51eb9eddb4c3621d8c3"
		hackSrcDir := filepath.Join(hackExtractTo, "redroid_libhoudini_hack-"+hackName, version)
		hackDstDir := filepath.Join(copyDir, "system")

		onStatus("Copying Houdini Hack files...")
		if err := copyDir2(hackSrcDir, hackDstDir); err != nil {
			return err
		}

		if version != "9.0.0" {
			initPath := filepath.Join(copyDir, "system", "etc", "init", "hw", "init.rc")
			if err := os.Chmod(initPath, 0644); err != nil {
				// Don't fail if file doesn't exist, but maybe warn?
				// Just ignore error for now as per python script logic which assumes it exists if copied
			}
		}
	}

	return nil
}

func (h *HoudiniAddon) Install(version, arch, outputDir string, onStatus func(string)) error {
	if err := h.Download(version, arch, onStatus); err != nil {
		return err
	}
	if err := h.Extract(version, arch, onStatus); err != nil {
		return err
	}
	return h.Copy(version, arch, outputDir, onStatus)
}

func (h *HoudiniAddon) DockerfileInstructions() string {
	return fmt.Sprintf("COPY %s /\n", "houdini")
}

func (h *HoudiniAddon) getInitRC() string {
	return `
on early-init
    mount binfmt_misc binfmt_misc /proc/sys/fs/binfmt_misc

on property:ro.enable.native.bridge.exec=1
    copy /system/etc/binfmt_misc/arm_exe /proc/sys/fs/binfmt_misc/register
    copy /system/etc/binfmt_misc/arm_dyn /proc/sys/fs/binfmt_misc/register

on property:ro.enable.native.bridge.exec64=1
    copy /system/etc/binfmt_misc/arm64_exe /proc/sys/fs/binfmt_misc/register
    copy /system/etc/binfmt_misc/arm64_dyn /proc/sys/fs/binfmt_misc/register

on property:sys.boot_completed=1
    exec -- /system/bin/sh -c "echo ':arm_exe:M::\\\\x7f\\\\x45\\\\x4c\\\\x46\\\\x01\\\\x01\\\\x01\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x02\\\\x00\\\\x28::/system/bin/houdini:P' >> /proc/sys/fs/binfmt_misc/register"
    exec -- /system/bin/sh -c "echo ':arm_dyn:M::\\\\x7f\\\\x45\\\\x4c\\\\x46\\\\x01\\\\x01\\\\x01\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x03\\\\x00\\\\x28::/system/bin/houdini:P' >> /proc/sys/fs/binfmt_misc/register"
    exec -- /system/bin/sh -c "echo ':arm64_exe:M::\\\\x7f\\\\x45\\\\x4c\\\\x46\\\\x02\\\\x01\\\\x01\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x02\\\\x00\\\\xb7::/system/bin/houdini64:P' >> /proc/sys/fs/binfmt_misc/register"
    exec -- /system/bin/sh -c "echo ':arm64_dyn:M::\\\\x7f\\\\x45\\\\x4c\\\\x46\\\\x02\\\\x01\\\\x01\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x00\\\\x03\\\\x00\\\\xb7::/system/bin/houdini64:P' >> /proc/sys/fs/binfmt_misc/register"
`
}

func extractZip(src, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func copyDir2(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
