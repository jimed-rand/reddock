package redroidscript

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil || rel == "." {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyRegular(path, target, d.Type())
	})
}

func copyRegular(srcPath, dstPath string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}
	r, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer r.Close()
	w, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = io.Copy(w, r)
	return err
}

func chmodRecursive(root string, perm os.FileMode) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		return os.Chmod(path, perm)
	})
}
