package redroidscript

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const openGappsDate = "20220503"

func openGappsAPI(redroidAndroid string) string {
	base := strings.TrimSuffix(redroidAndroid, "_64only")
	m := map[string]string{
		"8.1.0": "8.1", "9.0.0": "9.0", "10.0.0": "10.0",
		"11.0.0": "11.0", "12.0.0": "12.0", "13.0.0": "13.0", "14.0.0": "14.0",
	}
	if a, ok := m[base]; ok {
		return a
	}
	return "11.0"
}

func openGappsZipURL(archKey, redroidAndroid string) string {
	api := openGappsAPI(redroidAndroid)
	return fmt.Sprintf(
		"https://downloads.sourceforge.net/project/opengapps/%s/%s/open_gapps-%s-%s-pico-%s.zip",
		archKey, openGappsDate, archKey, api, openGappsDate,
	)
}

func installGapps(workDir, android, archKey string) error {
	cache, err := DownloadCacheDir()
	if err != nil {
		return err
	}
	dlPath := filepath.Join(cache, "open_gapps.zip")
	url := openGappsZipURL(archKey, android)
	if err := DownloadTo(url, dlPath, ""); err != nil {
		return fmt.Errorf("gapps download: %w", err)
	}

	extractTo := filepath.Join(workDir, ".tmp", "ogapps", "extract")
	_ = os.RemoveAll(extractTo)
	if err := os.MkdirAll(extractTo, 0755); err != nil {
		return err
	}
	printGreen("Extracting OpenGapps archive…")
	if err := unzipTo(dlPath, extractTo); err != nil {
		return err
	}

	copyDir := filepath.Join(workDir, "gapps")
	_ = os.RemoveAll(copyDir)
	appUnpack := filepath.Join(extractTo, "appunpack")
	if err := os.MkdirAll(appUnpack, 0755); err != nil {
		return err
	}

	coreDir := filepath.Join(extractTo, "Core")
	entries, err := os.ReadDir(coreDir)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".tar.lz") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	nonApks := map[string]bool{
		"defaultetc-common.tar.lz":         true,
		"defaultframework-common.tar.lz": true,
		"googlepixelconfig-common.tar.lz":  true,
		"vending-common.tar.lz":          true,
	}
	skip := map[string]bool{
		"setupwizarddefault-x86_64.tar.lz": true,
		"setupwizardtablet-x86_64.tar.lz":  true,
	}

	for _, lzFile := range names {
		if skip[lzFile] {
			continue
		}
		// Clear appunpack children (match Python loop).
		subs, _ := os.ReadDir(appUnpack)
		for _, s := range subs {
			_ = os.RemoveAll(filepath.Join(appUnpack, s.Name()))
		}
		lzPath := filepath.Join(coreDir, lzFile)
		if !nonApks[lzFile] {
			fmt.Println("    Processing app package :", lzPath)
			if err := runCmd("tar", "--lzip", "-xf", lzPath, "-C", appUnpack); err != nil {
				return err
			}
			appName, err := soleEntryName(filepath.Join(appUnpack))
			if err != nil {
				return err
			}
			xxDpi, err := soleEntryName(filepath.Join(appUnpack, appName))
			if err != nil {
				return err
			}
			appPriv, err := soleEntryName(filepath.Join(appUnpack, appName, "nodpi"))
			if err != nil {
				return err
			}
			appSrc := filepath.Join(appUnpack, appName, xxDpi, appPriv)
			apps, err := os.ReadDir(appSrc)
			if err != nil {
				return err
			}
			for _, app := range apps {
				if !app.IsDir() {
					continue
				}
				dst := filepath.Join(copyDir, "system", "priv-app", app.Name())
				if err := copyTree(filepath.Join(appSrc, app.Name()), dst); err != nil {
					return err
				}
			}
		} else {
			fmt.Println("    Processing extra package :", lzPath)
			if err := runCmd("tar", "--lzip", "-xf", lzPath, "-C", appUnpack); err != nil {
				return err
			}
			appName, err := soleEntryName(filepath.Join(appUnpack))
			if err != nil {
				return err
			}
			commonDir := filepath.Join(appUnpack, appName, "common")
			ccdirs, err := os.ReadDir(commonDir)
			if err != nil {
				return err
			}
			for _, ccdir := range ccdirs {
				if !ccdir.IsDir() {
					continue
				}
				src := filepath.Join(commonDir, ccdir.Name())
				dst := filepath.Join(copyDir, "system", ccdir.Name())
				if err := copyTree(src, dst); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func soleEntryName(dir string) (string, error) {
	es, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var names []string
	for _, e := range es {
		if e.Name() != "." && e.Name() != ".." {
			names = append(names, e.Name())
		}
	}
	if len(names) != 1 {
		return "", fmt.Errorf("expected exactly one entry in %s, got %v", dir, names)
	}
	return names[0], nil
}
