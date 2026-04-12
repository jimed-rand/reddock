package redroidscript

import (
	"fmt"
	"os"
	"path/filepath"
)

const houdiniInitRC = `
on early-init
    mount binfmt_misc binfmt_misc /proc/sys/fs/binfmt_misc

on property:ro.enable.native.bridge.exec=1
    copy /system/etc/binfmt_misc/arm_exe /proc/sys/fs/binfmt_misc/register
    copy /system/etc/binfmt_misc/arm_dyn /proc/sys/fs/binfmt_misc/register

on property:ro.enable.native.bridge.exec64=1
    copy /system/etc/binfmt_misc/arm64_exe /proc/sys/fs/binfmt_misc/register
    copy /system/etc/binfmt_misc/arm64_dyn /proc/sys/fs/binfmt_misc/register

on property:sys.boot_completed=1
    exec -- /system/bin/sh -c "echo ':arm_exe:M::\\x7f\\x45\\x4c\\x46\\x01\\x01\\x01\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x02\\x00\\x28::/system/bin/houdini:P' >> /proc/sys/fs/binfmt_misc/register"
    exec -- /system/bin/sh -c "echo ':arm_dyn:M::\\x7f\\x45\\x4c\\x46\\x01\\x01\\x01\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x03\\x00\\x28::/system/bin/houdini:P' >> /proc/sys/fs/binfmt_misc/register"
    exec -- /system/bin/sh -c "echo ':arm64_exe:M::\\x7f\\x45\\x4c\\x46\\x02\\x01\\x01\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x02\\x00\\xb7::/system/bin/houdini64:P' >> /proc/sys/fs/binfmt_misc/register"
    exec -- /system/bin/sh -c "echo ':arm64_dyn:M::\\x7f\\x45\\x4c\\x46\\x02\\x01\\x01\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x03\\x00\\xb7::/system/bin/houdini64:P' >> /proc/sys/fs/binfmt_misc/register"
`

var houdiniZips = map[string][2]string{
	"8.1.0":   {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/46682f423b8497db3f96222f2669d770eff764c3.zip", "cd4dd2891aa18e7699d33dcc3fe3ffd4"},
	"9.0.0":   {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/46682f423b8497db3f96222f2669d770eff764c3.zip", "cd4dd2891aa18e7699d33dcc3fe3ffd4"},
	"10.0.0":  {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip", "cb7ffac26d47ec7c89df43818e126b47"},
	"11.0.0":  {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip", "cb7ffac26d47ec7c89df43818e126b47"},
	"12.0.0":  {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip", "cb7ffac26d47ec7c89df43818e126b47"},
	"12.0.0_64only": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip", "cb7ffac26d47ec7c89df43818e126b47"},
	"13.0.0":  {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip", "cb7ffac26d47ec7c89df43818e126b47"},
	"13.0.0_64only": {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip", "cb7ffac26d47ec7c89df43818e126b47"},
	"14.0.0":  {"https://github.com/rote66/vendor_intel_proprietary_houdini/archive/debc3dc91cf12b5c5b8a1c546a5b0b7bf7f838a8.zip", "cb7ffac26d47ec7c89df43818e126b47"},
}

func installHoudini(workDir, android string) error {
	z, ok := houdiniZips[android]
	if !ok {
		return fmt.Errorf("no libhoudini build for Android %s", android)
	}
	url, want := z[0], z[1]
	cache, err := DownloadCacheDir()
	if err != nil {
		return err
	}
	dlPath := filepath.Join(cache, "libhoudini.zip")
	if err := DownloadTo(url, dlPath, want); err != nil {
		return err
	}
	extractTo := filepath.Join(workDir, ".tmp", "houdiniunpack")
	_ = os.RemoveAll(extractTo)
	if err := os.MkdirAll(extractTo, 0755); err != nil {
		return err
	}
	printGreen("Extracting libhoudini archive…")
	if err := unzipTo(dlPath, extractTo); err != nil {
		return err
	}
	_ = runCmd("chmod", "-R", "+x", extractTo)

	name, err := zipCommitDirName(url)
	if err != nil {
		return err
	}
	src := filepath.Join(extractTo, "vendor_intel_proprietary_houdini-"+name, "prebuilts")
	copyDir := filepath.Join(workDir, "houdini")
	_ = os.RemoveAll(copyDir)
	printGreen("Copying libhoudini library files …")
	if err := copyTree(src, filepath.Join(copyDir, "system")); err != nil {
		return err
	}
	initPath := filepath.Join(copyDir, "system", "etc", "init", "houdini.rc")
	if err := os.MkdirAll(filepath.Dir(initPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(initPath, []byte(houdiniInitRC), 0644); err != nil {
		return err
	}
	return nil
}
