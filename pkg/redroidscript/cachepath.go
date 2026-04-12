package redroidscript

import (
	"fmt"
	"os"
	"path/filepath"
)

// DownloadCacheDir mirrors tools.helper.get_download_dir() from redroid-script.
func DownloadCacheDir() (string, error) {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		p := filepath.Join(xdg, "redroid", "downloads")
		return p, os.MkdirAll(p, 0755)
	}
	user := os.Getenv("SUDO_USER")
	if user == "" {
		user = os.Getenv("USER")
	}
	if user == "" {
		return "", fmt.Errorf("USER and SUDO_USER are unset; cannot resolve download cache dir")
	}
	p := filepath.Join("/home", user, ".cache", "redroid", "downloads")
	return p, os.MkdirAll(p, 0755)
}
