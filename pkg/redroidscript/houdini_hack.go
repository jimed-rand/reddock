package redroidscript

import (
	"os"
	"path/filepath"
)

const (
	houdiniHackURL = "https://github.com/rote66/redroid_libhoudini_hack/archive/a2194c5e294cbbfdfe87e51eb9eddb4c3621d8c3.zip"
	houdiniHackMD5 = "8f71a58f3e54eca879a2f7de64dbed58"
)

// installHoudiniHack merges hack blobs into workDir/houdini (run after installHoudini).
func installHoudiniHack(workDir, androidLine string) error {
	cache, err := DownloadCacheDir()
	if err != nil {
		return err
	}
	dlPath := filepath.Join(cache, "libhoudini_hack.zip")
	if err := DownloadTo(houdiniHackURL, dlPath, houdiniHackMD5); err != nil {
		return err
	}
	extractTo := filepath.Join(workDir, ".tmp", "houdinihackunpack")
	_ = os.RemoveAll(extractTo)
	if err := os.MkdirAll(extractTo, 0755); err != nil {
		return err
	}
	printGreen("Extracting libhoudini_hack archive…")
	if err := unzipTo(dlPath, extractTo); err != nil {
		return err
	}
	_ = runCmd("chmod", "-R", "+x", extractTo)

	name, err := zipCommitDirName(houdiniHackURL)
	if err != nil {
		return err
	}
	src := filepath.Join(extractTo, "redroid_libhoudini_hack-"+name, androidLine)
	copyDir := filepath.Join(workDir, "houdini")
	printGreen("Copying libhoudini hack files …")
	if err := copyTree(src, filepath.Join(copyDir, "system")); err != nil {
		return err
	}
	if androidLine != "9.0.0" {
		initPath := filepath.Join(copyDir, "system", "etc", "init", "hw", "init.rc")
		return os.Chmod(initPath, 0644)
	}
	return nil
}
