package redroidscript

import (
	"os"
	"path/filepath"
)

const (
	ndkZipURL = "https://github.com/supremegamers/vendor_google_proprietary_ndk_translation-prebuilt/archive/9324a8914b649b885dad6f2bfd14a67e5d1520bf.zip"
	ndkMD5    = "c9572672d1045594448068079b34c350"
	ndkDir    = "vendor_google_proprietary_ndk_translation-prebuilt-9324a8914b649b885dad6f2bfd14a67e5d1520bf"
)

func installNDK(workDir string) error {
	cache, err := DownloadCacheDir()
	if err != nil {
		return err
	}
	dlPath := filepath.Join(cache, "libndktranslation.zip")
	if err := DownloadTo(ndkZipURL, dlPath, ndkMD5); err != nil {
		return err
	}
	extractTo := filepath.Join(workDir, ".tmp", "libndkunpack")
	_ = os.RemoveAll(extractTo)
	if err := os.MkdirAll(extractTo, 0755); err != nil {
		return err
	}
	printGreen("Extracting libndk archive…")
	if err := unzipTo(dlPath, extractTo); err != nil {
		return err
	}
	_ = runCmd("chmod", "-R", "+x", extractTo)

	copyDir := filepath.Join(workDir, "ndk")
	_ = os.RemoveAll(copyDir)
	src := filepath.Join(extractTo, ndkDir, "prebuilts")
	printGreen("Copying libndk library files …")
	if err := copyTree(src, filepath.Join(copyDir, "system")); err != nil {
		return err
	}
	initPath := filepath.Join(copyDir, "system", "etc", "init", "ndk_translation.rc")
	return os.Chmod(initPath, 0644)
}
