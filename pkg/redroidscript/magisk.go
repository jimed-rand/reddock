package redroidscript

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

const (
	magiskURL = "https://github.com/ayasa520/Magisk/releases/download/v30.6/app-debug.apk"
	magiskMD5 = "77ef9f3538c0767ea45ee5c946f84bc6"
)

var magiskLibRename = regexp.MustCompile(`^lib(.*)\.so$`)

const magiskBootanimOrig = `
service bootanim /system/bin/bootanimation
    class core animation
    user graphics
    group graphics audio
    disabled
    oneshot
    ioprio rt 0
    task_profiles MaxPerformance
    
`

func magiskBootanimExtra() string {
	const s = `
on post-fs-data
    start logd
    exec u:r:su:s0 root root -- /system/etc/init/magisk/magiskpolicy --live --magisk
    exec u:r:magisk:s0 root root -- /system/etc/init/magisk/magiskpolicy --live --magisk
    exec u:r:update_engine:s0 root root -- /system/etc/init/magisk/magiskpolicy --live --magisk
    exec u:r:su:s0 root root -- /system/etc/init/magisk/magisk --auto-selinux --setup-sbin /system/etc/init/magisk /sbin
    exec u:r:su:s0 root root -- /sbin/magisk --auto-selinux --post-fs-data
on nonencrypted
    exec u:r:su:s0 root root -- /sbin/magisk --auto-selinux --service
on property:vold.decrypt=trigger_restart_framework
    exec u:r:su:s0 root root -- /sbin/magisk --auto-selinux --service
on property:sys.boot_completed=1
    mkdir /data/adb/magisk 755
    exec u:r:su:s0 root root -- /sbin/magisk --auto-selinux --boot-complete
    exec -- /system/bin/sh -c "if [ ! -e /data/data/io.github.huskydg.magisk ] ; then pm install /system/etc/init/magisk/magisk.apk ; fi"
   
on property:init.svc.zygote=restarting
    exec u:r:su:s0 root root -- /sbin/magisk --auto-selinux --zygote-restart
   
on property:init.svc.zygote=stopped
    exec u:r:su:s0 root root -- /sbin/magisk --auto-selinux --zygote-restart
    `
	return s
}

func installMagisk(workDir, archKey string) error {
	cache, err := DownloadCacheDir()
	if err != nil {
		return err
	}
	dlPath := filepath.Join(cache, "magisk.apk")
	if err := DownloadTo(magiskURL, dlPath, magiskMD5); err != nil {
		return err
	}

	extractTo := filepath.Join(workDir, ".tmp", "magisk_unpack")
	_ = os.RemoveAll(extractTo)
	if err := os.MkdirAll(extractTo, 0755); err != nil {
		return err
	}
	printGreen("Extracting Magisk apk …")
	if err := unzipTo(dlPath, extractTo); err != nil {
		return err
	}

	archMap := map[string]string{
		"x86": "x86", "x86_64": "x86_64", "arm": "armeabi-v7a", "arm64": "arm64-v8a",
	}
	abi, ok := archMap[archKey]
	if !ok {
		return fmt.Errorf("magisk: unsupported arch %q", archKey)
	}
	libDir := filepath.Join(extractTo, "lib", abi)
	if st, err := os.Stat(libDir); err != nil || !st.IsDir() {
		return fmt.Errorf("magisk: expected lib dir %s inside apk extract", libDir)
	}

	copyDir := filepath.Join(workDir, "magisk")
	_ = os.RemoveAll(copyDir)
	magiskDir := filepath.Join(copyDir, "system", "etc", "init", "magisk")
	if err := os.MkdirAll(magiskDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(copyDir, "sbin"), 0755); err != nil {
		return err
	}

	printGreen("Copying magisk libs …")
	err = filepath.WalkDir(libDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		base := filepath.Base(path)
		m := magiskLibRename.FindStringSubmatch(base)
		if m == nil {
			return nil
		}
		dst := filepath.Join(magiskDir, m[1])
		if err := copyRegular(path, dst, 0644); err != nil {
			return err
		}
		return os.Chmod(dst, 0755)
	})
	if err != nil {
		return err
	}
	if err := copyRegular(dlPath, filepath.Join(magiskDir, "magisk.apk"), 0644); err != nil {
		return err
	}

	bootanimPath := filepath.Join(copyDir, "system", "etc", "init", "bootanim.rc")
	if err := os.MkdirAll(filepath.Dir(bootanimPath), 0755); err != nil {
		return err
	}
	gzPath := bootanimPath + ".gz"
	gzf, err := os.Create(gzPath)
	if err != nil {
		return err
	}
	gw := gzip.NewWriter(gzf)
	if _, err := io.WriteString(gw, magiskBootanimOrig); err != nil {
		gw.Close()
		gzf.Close()
		return err
	}
	if err := gw.Close(); err != nil {
		gzf.Close()
		return err
	}
	if err := gzf.Close(); err != nil {
		return err
	}
	content := magiskBootanimOrig + magiskBootanimExtra()
	if err := os.WriteFile(bootanimPath, []byte(content), 0644); err != nil {
		return err
	}
	return nil
}
