package redroidscript

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
)

// DownloadTo fetches url into destPath. If wantMD5 is non-empty, verifies MD5 (redownload on mismatch).
func DownloadTo(url, destPath, wantMD5 string) error {
	for attempt := 0; attempt < 2; attempt++ {
		if wantMD5 != "" {
			if sum, err := fileMD5(destPath); err == nil && sum == wantMD5 {
				return nil
			}
			_ = os.Remove(destPath)
		} else if _, err := os.Stat(destPath); err == nil {
			return nil
		}

		if err := downloadOnce(url, destPath); err != nil {
			return err
		}
		if wantMD5 == "" {
			return nil
		}
		sum, err := fileMD5(destPath)
		if err != nil {
			return err
		}
		if sum == wantMD5 {
			return nil
		}
		_ = os.Remove(destPath)
		if attempt == 1 {
			return fmt.Errorf("md5 mismatch for %s: got %s want %s", destPath, sum, wantMD5)
		}
		printYellow("md5 mismatch, redownloading …")
	}
	return fmt.Errorf("download failed for %s", url)
}

func downloadOnce(url, destPath string) error {
	printGreen("Downloading " + url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http %s: %s", url, resp.Status)
	}
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}
	return f.Close()
}

func fileMD5(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := md5.Sum(b)
	return hex.EncodeToString(h[:]), nil
}
