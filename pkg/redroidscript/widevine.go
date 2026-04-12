package redroidscript

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// widevineLinks: arch -> android -> [url, md5]
var widevineLinks = map[string]map[string][2]string{
	"x86_64": {
		"11.0.0": {"https://github.com/supremegamers/vendor_google_proprietary_widevine-prebuilt/archive/48d1076a570837be6cdce8252d5d143363e37cc1.zip", "f587b8859f9071da4bca6cea1b9bed6a"},
		"12.0.0": {"https://github.com/supremegamers/vendor_google_proprietary_widevine-prebuilt/archive/3bba8b6e9dd5ffad5b861310433f7e397e9366e8.zip", "3e147bdeeb7691db4513d93cfa6beb23"},
		"13.0.0": {"https://github.com/supremegamers/vendor_google_proprietary_widevine-prebuilt/archive/a8524d608431573ef1c9313822d271f78728f9a6.zip", "5c55df61da5c012b4e43746547ab730f"},
	},
	"arm64": {
		"11.0.0": {"https://github.com/supremegamers/vendor_google_proprietary_widevine-prebuilt/archive/a1a19361d36311bee042da8cf4ced798d2c76d98.zip", "fed6898b5cfd2a908cb134df97802554"},
	},
}

func installWidevine(workDir, android, archKey string) error {
	byArch, ok := widevineLinks[archKey]
	if !ok {
		return fmt.Errorf("widevine: no builds for arch %q (upstream only ships x86_64 11–13 and arm64 11)", archKey)
	}
	base := strings.TrimSuffix(android, "_64only")
	z, ok := byArch[base]
	if !ok {
		return fmt.Errorf("widevine: no package for android %q on %s", android, archKey)
	}
	url, want := z[0], z[1]

	cache, err := DownloadCacheDir()
	if err != nil {
		return err
	}
	dlPath := filepath.Join(cache, "widevine.zip")
	if err := DownloadTo(url, dlPath, want); err != nil {
		return err
	}
	extractTo := filepath.Join(workDir, ".tmp", "widevineunpack")
	_ = os.RemoveAll(extractTo)
	if err := os.MkdirAll(extractTo, 0755); err != nil {
		return err
	}
	printGreen("Extracting widevine archive…")
	if err := unzipTo(dlPath, extractTo); err != nil {
		return err
	}
	_ = runCmd("chmod", "-R", "+x", extractTo)

	name, err := zipCommitDirName(url)
	if err != nil {
		return err
	}
	src := filepath.Join(extractTo, "vendor_google_proprietary_widevine-prebuilt-"+name, "prebuilts")
	copyDir := filepath.Join(workDir, "widevine")
	_ = os.RemoveAll(copyDir)
	printGreen("Copying widevine library files …")
	if err := copyTree(src, filepath.Join(copyDir, "vendor")); err != nil {
		return err
	}

	if strings.Contains(archKey, "x86") && base == "11.0.0" {
		vlib := filepath.Join(copyDir, "vendor", "lib")
		vlib64 := filepath.Join(copyDir, "vendor", "lib64")
		_ = os.MkdirAll(vlib, 0755)
		_ = os.MkdirAll(vlib64, 0755)
		target := "./libprotobuf-cpp-lite-3.9.1.so"
		_ = os.Remove(filepath.Join(vlib, "libprotobuf-cpp-lite.so"))
		_ = os.Remove(filepath.Join(vlib64, "libprotobuf-cpp-lite.so"))
		if err := os.Symlink(target, filepath.Join(vlib, "libprotobuf-cpp-lite.so")); err != nil {
			return err
		}
		if err := os.Symlink(target, filepath.Join(vlib64, "libprotobuf-cpp-lite.so")); err != nil {
			return err
		}
	}

	initDir := filepath.Join(copyDir, "vendor", "etc", "init")
	es, err := os.ReadDir(initDir)
	if err != nil {
		return err
	}
	for _, e := range es {
		if strings.HasSuffix(e.Name(), ".rc") {
			_ = os.Chmod(filepath.Join(initDir, e.Name()), 0644)
		}
	}
	return nil
}
