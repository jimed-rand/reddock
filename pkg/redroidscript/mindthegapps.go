package redroidscript

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func installMindTheGapps(workDir, android, archKey string) error {
	if android == "8.1.0" {
		return fmt.Errorf("MindTheGapps is not published for Android 8.1; use OpenGapps (-g)")
	}
	lookup := android
	if strings.HasSuffix(android, "_64only") {
		lookup = strings.TrimSuffix(android, "_64only")
	}
	if (lookup == "10.0.0" || lookup == "11.0.0") && archKey == "x86_64" {
		return fmt.Errorf("MindTheGapps has no official x86_64 package for Android 10/11; use OpenGapps (-g) on x86_64, or Android 12+ with -mtg")
	}

	vermap, ok := mindTheGappsLinks[android]
	if !ok {
		vermap = mindTheGappsLinks[lookup]
	}
	if vermap == nil {
		return fmt.Errorf("MindTheGapps: unknown android line %q", android)
	}
	row, ok := vermap[archKey]
	if !ok {
		return fmt.Errorf("MindTheGapps: no package for architecture %s (android %s)", archKey, android)
	}
	if len(row) < 2 {
		return fmt.Errorf("MindTheGapps: bad link row")
	}
	url := row[0]
	wantMD5 := row[1]

	cache, err := DownloadCacheDir()
	if err != nil {
		return err
	}
	dlPath := filepath.Join(cache, "mindthegapps.zip")
	if err := DownloadTo(url, dlPath, wantMD5); err != nil {
		return err
	}

	extractTo := filepath.Join(workDir, ".tmp", "mindthegapps", "extract")
	_ = os.RemoveAll(extractTo)
	if err := os.MkdirAll(extractTo, 0755); err != nil {
		return err
	}
	printGreen("Extracting MindTheGapps archive…")
	if err := unzipTo(dlPath, extractTo); err != nil {
		return err
	}

	copyDir := filepath.Join(workDir, "mindthegapps")
	_ = os.RemoveAll(copyDir)
	src := filepath.Join(extractTo, "system")
	return copyTree(src, filepath.Join(copyDir, "system"))
}
