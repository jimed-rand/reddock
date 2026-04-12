package redroidscript

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func unzipTo(srcZip, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, zf := range r.File {
		out := filepath.Join(destDir, zf.Name)
		if zf.FileInfo().IsDir() {
			if err := os.MkdirAll(out, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
			return err
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		wf, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, zf.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(wf, rc)
		rc.Close()
		cerr := wf.Close()
		if copyErr != nil {
			return copyErr
		}
		if cerr != nil {
			return cerr
		}
	}
	return nil
}

func zipCommitDirName(zipURL string) (string, error) {
	// Match Python re.findall("([a-zA-Z0-9]+)\\.zip", url) on GitHub archive URLs.
	base := filepath.Base(zipURL)
	var name string
	for _, suf := range []string{".zip", ".ZIP"} {
		if len(base) > len(suf) && base[len(base)-len(suf):] == suf {
			name = base[:len(base)-len(suf)]
			break
		}
	}
	if name == "" {
		return "", fmt.Errorf("could not parse archive name from %q", zipURL)
	}
	return name, nil
}
