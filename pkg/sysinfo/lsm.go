package sysinfo

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// HostLSMInfo summarizes mandatory access controls relevant to Docker / privileged containers.
// Detection is best-effort from sysfs and optionally getenforce; it does not require root for reads.
type HostLSMInfo struct {
	SELinuxPresent bool
	SELinuxMode    string // enforcing, permissive, disabled, unknown, "" if absent

	AppArmorModulePresent bool
	AppArmorKernelEnabled bool // /sys/module/apparmor/parameters/enabled == Y
}

// ProbeHostLSM collects SELinux and AppArmor signals from the host.
func ProbeHostLSM() HostLSMInfo {
	var info HostLSMInfo
	info.SELinuxPresent, info.SELinuxMode = probeSELinux()
	info.AppArmorModulePresent, info.AppArmorKernelEnabled = probeAppArmor()
	return info
}

func probeSELinux() (present bool, mode string) {
	if _, err := os.Stat("/sys/fs/selinux"); err != nil {
		return false, ""
	}
	present = true
	b, err := os.ReadFile("/sys/fs/selinux/enforce")
	if err != nil {
		mode = modeFromGetenforce()
		if mode != "" {
			return true, mode
		}
		return true, "unknown"
	}
	switch strings.TrimSpace(string(b)) {
	case "1":
		return true, "enforcing"
	case "0":
		return true, "permissive"
	default:
		return true, "unknown"
	}
}

func modeFromGetenforce() string {
	paths := []string{"getenforce", "/usr/sbin/getenforce", "/sbin/getenforce"}
	for _, p := range paths {
		cmd := exec.Command(p)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		line := strings.TrimSpace(string(out))
		switch strings.ToLower(line) {
		case "enforcing":
			return "enforcing"
		case "permissive":
			return "permissive"
		case "disabled":
			return "disabled"
		default:
			return ""
		}
	}
	return ""
}

func probeAppArmor() (modulePresent, kernelEnabled bool) {
	if _, err := os.Stat("/sys/module/apparmor"); err != nil {
		return false, false
	}
	modulePresent = true
	b, err := os.ReadFile("/sys/module/apparmor/parameters/enabled")
	if err != nil {
		return true, false
	}
	kernelEnabled = strings.TrimSpace(string(b)) == "Y"
	return true, kernelEnabled
}

// SELinuxMayBlockDocker is true when SELinux is present and in enforcing mode.
func (h HostLSMInfo) SELinuxMayBlockDocker() bool {
	return h.SELinuxPresent && strings.EqualFold(h.SELinuxMode, "enforcing")
}

// AppArmorMayAffectDocker is true when the AppArmor module reports enabled in the kernel.
func (h HostLSMInfo) AppArmorMayAffectDocker() bool {
	return h.AppArmorKernelEnabled
}

// PrintHostLSMWarnings writes remediation hints to w when SELinux or AppArmor may interfere.
func PrintHostLSMWarnings(w io.Writer, h HostLSMInfo) {
	for _, block := range HostLSMRemediationBlocks(h) {
		fmt.Fprintln(w)
		fmt.Fprintln(w, block)
	}
}

// HostLSMRemediationBlocks returns non-empty paragraphs (one string per block) for warnings.
func HostLSMRemediationBlocks(h HostLSMInfo) []string {
	var blocks []string
	if h.SELinuxMayBlockDocker() {
		var b strings.Builder
		b.WriteString("Warning: SELinux is enforcing. It can prevent Docker from mounting volumes, ")
		b.WriteString("using certain capabilities, or running privileged Android-style containers.\n\n")
		b.WriteString("  Temporary (until reboot):  sudo setenforce 0\n")
		b.WriteString("  Check:                     getenforce    (expect: Permissive)\n\n")
		b.WriteString("  Persistent (typical Fedora/RHEL/openSUSE): edit /etc/selinux/config, set ")
		b.WriteString("SELINUX=permissive, reboot. To turn enforcing back on later: sudo setenforce 1 ")
		b.WriteString("and SELINUX=enforcing in that file, then reboot.\n\n")
		b.WriteString("  Docker-only workaround: add --security-opt label=disable to docker run ")
		b.WriteString("(reddock does not add this automatically).")
		blocks = append(blocks, b.String())
	}

	if h.AppArmorMayAffectDocker() {
		var b strings.Builder
		b.WriteString("Warning: AppArmor is enabled in the kernel. The Docker daemon may still apply ")
		b.WriteString("a profile (for example docker-default) that blocks parts of a privileged ")
		b.WriteString("redroid-style container.\n\n")
		b.WriteString("  Temporary stop (until reboot or manual start):  sudo systemctl stop apparmor\n")
		b.WriteString("  Turn it back on:                                 sudo systemctl start apparmor\n\n")
		b.WriteString("  Docker-only workaround: add --security-opt apparmor=unconfined to docker run ")
		b.WriteString("(reddock does not add this automatically).")
		blocks = append(blocks, b.String())
	}
	return blocks
}

// HostLSMStatusLine is a single-line summary for reddock status.
func (h HostLSMInfo) HostLSMStatusLine() string {
	var parts []string
	if h.SELinuxPresent {
		parts = append(parts, fmt.Sprintf("SELinux: %s", orDash(strings.ToLower(h.SELinuxMode))))
	} else {
		parts = append(parts, "SELinux: not active (no /sys/fs/selinux)")
	}
	switch {
	case h.AppArmorKernelEnabled:
		parts = append(parts, "AppArmor: enabled (kernel Y)")
	case h.AppArmorModulePresent:
		parts = append(parts, "AppArmor: module present, kernel flag not Y")
	default:
		parts = append(parts, "AppArmor: module not loaded")
	}
	return strings.Join(parts, " | ")
}
