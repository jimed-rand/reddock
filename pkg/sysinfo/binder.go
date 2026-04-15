package sysinfo

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var errStopBinderWalk = errors.New("stop binder ko walk")

// BinderHostInfo summarizes Android binder / binderfs support on the running kernel.
// It is meant for hosts where binder_linux comes from in-tree builds, DKMS, or distro KMP
// (e.g. openSUSE), and where devices may appear as legacy /dev/binder* or under binderfs.
type BinderHostInfo struct {
	KernelRelease string

	ProcModuleBinderLinux bool
	ProcModuleBinder      bool // distinct "binder" line in /proc/modules (some trees)

	SysModuleBinderLinux bool
	SysModuleBinder      bool

	// ModinfoPathBinderLinux is non-empty when modinfo can resolve the module (installed in module tree).
	ModinfoPathBinderLinux string
	ModinfoPathBinder      string
	// KOFinderPathBinderLinux is set when a binder_linux.ko was found under /lib/modules/<release>/ (DKMS/KMP/extra).
	KOFinderPathBinderLinux string

	LegacyBinderCharDevs bool
	BinderFSBinderDevs   bool
	BinderFSInProcFS     bool // "binder" fs listed in /proc/filesystems (binderfs support)
}

// ProbeBinderHost collects binder-related signals from the host (best-effort, no root required for reads).
func ProbeBinderHost() BinderHostInfo {
	var info BinderHostInfo
	info.KernelRelease = strings.TrimSpace(readFileFirstLine("/proc/sys/kernel/osrelease"))
	if info.KernelRelease == "" {
		info.KernelRelease = strings.TrimSpace(unameRelease())
	}

	info.ProcModuleBinderLinux = procModuleLoaded("binder_linux")
	info.ProcModuleBinder = procModuleLoaded("binder")

	info.SysModuleBinderLinux = dirExists("/sys/module/binder_linux")
	info.SysModuleBinder = dirExists("/sys/module/binder")

	if p, ok := modinfoFilename("binder_linux"); ok {
		info.ModinfoPathBinderLinux = p
	}
	if p, ok := modinfoFilename("binder"); ok {
		info.ModinfoPathBinder = p
	}
	if info.ModinfoPathBinderLinux == "" && info.KernelRelease != "" {
		if p := findBinderLinuxKO(info.KernelRelease); p != "" {
			info.KOFinderPathBinderLinux = p
		}
	}

	info.LegacyBinderCharDevs = legacyBinderDevicesPresent()
	info.BinderFSBinderDevs = binderFSBinderDevicesPresent()
	info.BinderFSInProcFS = binderFSListedInProcFilesystems()

	return info
}

// BinderLinuxInstallable reports whether binder_linux appears to be packaged for this kernel
// (modinfo or a .ko under /lib/modules/<release>/), even if the module is not loaded yet.
func (b BinderHostInfo) BinderLinuxInstallable() bool {
	return b.ModinfoPathBinderLinux != "" || b.KOFinderPathBinderLinux != ""
}

// HostBinderUsable is true when reddock sees binder endpoints that redroid-style stacks use.
func (b BinderHostInfo) HostBinderUsable() bool {
	if b.LegacyBinderCharDevs || b.BinderFSBinderDevs {
		return true
	}
	// Loaded driver without visible nodes is still a strong signal (udev may use different paths).
	if b.ProcModuleBinderLinux || b.SysModuleBinderLinux {
		return true
	}
	if b.ProcModuleBinder || b.SysModuleBinder {
		return true
	}
	return false
}

// Summary returns a short, multi-line description for warnings or logs.
func (b BinderHostInfo) Summary() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("kernel: %s", orDash(b.KernelRelease)))
	lines = append(lines, fmt.Sprintf("loaded: binder_linux=%v binder=%v | sysfs: binder_linux=%v binder=%v",
		b.ProcModuleBinderLinux, b.ProcModuleBinder, b.SysModuleBinderLinux, b.SysModuleBinder))
	if b.ModinfoPathBinderLinux != "" {
		lines = append(lines, "modinfo binder_linux: "+b.ModinfoPathBinderLinux)
	} else if b.KOFinderPathBinderLinux != "" {
		lines = append(lines, "binder_linux.ko: "+b.KOFinderPathBinderLinux)
	} else {
		lines = append(lines, "modinfo binder_linux: (not found)")
	}
	if b.ModinfoPathBinder != "" {
		lines = append(lines, "modinfo binder: "+b.ModinfoPathBinder)
	}
	lines = append(lines, fmt.Sprintf("devices: legacy=%v binderfs_layout=%v | binderfs in /proc/filesystems: %v",
		b.LegacyBinderCharDevs, b.BinderFSBinderDevs, b.BinderFSInProcFS))
	return strings.Join(lines, "\n")
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func readFileFirstLine(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	line, _, _ := strings.Cut(string(data), "\n")
	return strings.TrimSpace(line)
}

func unameRelease() string {
	out, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func procModuleLoaded(name string) bool {
	f, err := os.Open("/proc/modules")
	if err != nil {
		return false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) > 0 && fields[0] == name {
			return true
		}
	}
	return false
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func isCharDev(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

func legacyBinderDevicesPresent() bool {
	// Classic android binder_linux "devices=" nodes
	for _, p := range []string{"/dev/binder", "/dev/hwbinder", "/dev/vndbinder"} {
		if isCharDev(p) {
			return true
		}
	}
	return false
}

func binderFSBinderDevicesPresent() bool {
	// Typical binderfs layout after mount + binder_ctl (Waydroid, manual setups)
	base := "/dev/binderfs"
	for _, p := range []string{
		filepath.Join(base, "binder"),
		filepath.Join(base, "hwbinder"),
		filepath.Join(base, "vndbinder"),
	} {
		if isCharDev(p) {
			return true
		}
	}
	return false
}

func binderFSListedInProcFilesystems() bool {
	data, err := os.ReadFile("/proc/filesystems")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		// Lines look like "nodev binder" or "binder"
		name := fields[len(fields)-1]
		if name == "binder" {
			return true
		}
	}
	return false
}

func modinfoFilename(module string) (string, bool) {
	cmd := exec.Command("modinfo", "-n", module)
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	path := strings.TrimSpace(string(out))
	if path == "" || strings.Contains(path, "ERROR") {
		return "", false
	}
	if _, err := os.Stat(path); err != nil {
		return path, true // modinfo said something; file may be delayed visibility
	}
	return path, true
}

func findBinderLinuxKO(release string) string {
	roots := []string{
		filepath.Join("/lib/modules", release, "kernel", "drivers", "android"),
		filepath.Join("/lib/modules", release, "updates", "dkms"),
		filepath.Join("/lib/modules", release, "updates"),
		filepath.Join("/lib/modules", release, "extra"),
		filepath.Join("/lib/modules", release, "weak-updates"),
	}
	for _, root := range roots {
		if p := walkFindBinderLinuxKO(root); p != "" {
			return p
		}
	}
	return ""
}

func walkFindBinderLinuxKO(root string) string {
	var found string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Base(path), "binder_linux.ko") {
			found = path
			return errStopBinderWalk
		}
		return nil
	})
	if err != nil && !errors.Is(err, errStopBinderWalk) {
		return ""
	}
	return found
}
