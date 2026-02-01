package addons

import (
	"fmt"
	"os"
	"path/filepath"
)

type NDKAddon struct {
	*BaseAddon
	dlLink    string
	md5       string
	extractTo string
}

func NewNDKAddon() *NDKAddon {
	versions := []string{"8.1.0", "9.0.0", "10.0.0", "11.0.0", "12.0.0", "12.0.0_64only", "13.0.0", "14.0.0", "15.0.0", "16.0.0"}

	baseAddon := NewBaseAddon("NDK Translation", AddonTypeNDK, versions)
	return &NDKAddon{
		BaseAddon: baseAddon,
		dlLink:    "https://github.com/supremegamers/vendor_google_proprietary_ndk_translation-prebuilt/archive/9324a8914b649b885dad6f2bfd14a67e5d1520bf.zip",
		md5:       "c9572672d1045594448068079b34c350",
		extractTo: "/tmp/libndkunpack",
	}
}

func (n *NDKAddon) Download(version, arch string, onStatus func(string)) error {
	if !n.IsSupported(version) {
		return fmt.Errorf("NDK not supported for Android %s", version)
	}

	if arch != "x86" && arch != "x86_64" {
		return fmt.Errorf("NDK only supports x86/x86_64 architecture")
	}

	filename := filepath.Join(n.downloadDir, "libndktranslation.zip")

	if err := ensureDir(n.downloadDir); err != nil {
		return err
	}

	onStatus(fmt.Sprintf("Downloading NDK Translation for Android %s...", version))
	return downloadFile(n.dlLink, filename)
}

func (n *NDKAddon) Extract(version, arch string, onStatus func(string)) error {
	filename := filepath.Join(n.downloadDir, "libndktranslation.zip")

	if err := ensureDir(n.extractTo); err != nil {
		return err
	}

	onStatus("Extracting NDK archive...")
	return extractZip(filename, n.extractTo)
}

func (n *NDKAddon) Copy(version, arch, outputDir string, onStatus func(string)) error {
	copyDir := filepath.Join(outputDir, "ndk")

	if err := os.RemoveAll(copyDir); err != nil {
		return err
	}

	srcDir := filepath.Join(n.extractTo, "vendor_google_proprietary_ndk_translation-prebuilt-9324a8914b649b885dad6f2bfd14a67e5d1520bf", "prebuilts")
	dstDir := filepath.Join(copyDir, "system")

	onStatus("Copying NDK library files...")
	if err := copyDir2(srcDir, dstDir); err != nil {
		return err
	}

	initPath := filepath.Join(copyDir, "system", "etc", "init", "ndk_translation.rc")
	if err := os.Chmod(initPath, 0644); err != nil {
		return err
	}

	return nil
}

func (n *NDKAddon) Install(version, arch, outputDir string, onStatus func(string)) error {
	if err := n.Download(version, arch, onStatus); err != nil {
		return err
	}
	if err := n.Extract(version, arch, onStatus); err != nil {
		return err
	}
	return n.Copy(version, arch, outputDir, onStatus)
}
