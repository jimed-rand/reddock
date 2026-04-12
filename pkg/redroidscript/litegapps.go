package redroidscript

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func liteGappsAPIKey(android string) string {
	v := strings.ReplaceAll(strings.ReplaceAll(android, "__64only", ""), "_64only", "")
	v = strings.TrimSuffix(v, "_")
	m := map[string]string{
		"15.0.0": "35", "14.0.0": "34", "13.0.0": "33", "12.0.0": "31",
		"11.0.0": "30", "10.0.0": "29", "9.0.0": "28", "8.1.0": "27",
	}
	return m[v]
}

func pickLitegapps(version, arch string) (url, md5 string, err error) {
	candidates := []string{version}
	if strings.HasSuffix(version, "_64only") {
		b := strings.TrimSuffix(version, "_64only")
		candidates = append(candidates, b+"__64only", b)
	}
	for _, vc := range candidates {
		m, ok := liteGappsLinks[vc]
		if !ok {
			continue
		}
		row, ok2 := m[arch]
		if !ok2 {
			continue
		}
		if len(row) < 2 {
			return "", "", fmt.Errorf("litegapps: bad row for %s %s", vc, arch)
		}
		return row[0], row[1], nil
	}
	return "", "", fmt.Errorf("no LiteGapps package for android %q arch %q", version, arch)
}

func installLiteGapps(workDir, android, archKey string) error {
	url, wantMD5, err := pickLitegapps(android, archKey)
	if err != nil {
		return err
	}
	cache, err := DownloadCacheDir()
	if err != nil {
		return err
	}
	dlPath := filepath.Join(cache, "litegapps.zip")
	if err := DownloadTo(url, dlPath, wantMD5); err != nil {
		return err
	}

	extractTo := filepath.Join(workDir, ".tmp", "litegapps", "extract")
	_ = os.RemoveAll(extractTo)
	if err := os.MkdirAll(extractTo, 0755); err != nil {
		return err
	}
	printGreen("Extracting LiteGapps archive…")
	if err := unzipTo(dlPath, extractTo); err != nil {
		return err
	}

	appUnpack := filepath.Join(extractTo, "appunpack")
	_ = os.RemoveAll(appUnpack)
	if err := os.MkdirAll(appUnpack, 0755); err != nil {
		return err
	}
	tarball := filepath.Join(extractTo, "files", "files.tar.xz")
	if err := runCmd("tar", "-xf", tarball, "-C", appUnpack); err != nil {
		return err
	}

	api := liteGappsAPIKey(android)
	if api == "" {
		return fmt.Errorf("litegapps: unknown API mapping for android %q", android)
	}
	srcSystem := filepath.Join(appUnpack, archKey, api, "system")
	copyDir := filepath.Join(workDir, "litegapps")
	_ = os.RemoveAll(copyDir)
	if err := copyTree(srcSystem, filepath.Join(copyDir, "system")); err != nil {
		return err
	}
	return nil
}
