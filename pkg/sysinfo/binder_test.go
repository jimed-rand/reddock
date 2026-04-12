package sysinfo

import "testing"

func TestParseProcModulesLineNames(t *testing.T) {
	content := `binder_linux 491520 5 - Live 0x0000000000000000
ext4 999 0 - Live 0x0
binder 123 0 - Live 0x0
`
	names := ParseProcModulesLineNames(content)
	if !names["binder_linux"] {
		t.Fatal("expected binder_linux")
	}
	if !names["binder"] {
		t.Fatal("expected binder")
	}
	if !names["ext4"] {
		t.Fatal("expected ext4")
	}
}

func TestBinderHostInfoBinderLinuxInstallable(t *testing.T) {
	b := BinderHostInfo{ModinfoPathBinderLinux: "/lib/modules/1.0/updates/dkms/binder_linux.ko"}
	if !b.BinderLinuxInstallable() {
		t.Fatal("expected installable via modinfo path")
	}
	b = BinderHostInfo{KOFinderPathBinderLinux: "/lib/modules/1.0/extra/binder_linux.ko"}
	if !b.BinderLinuxInstallable() {
		t.Fatal("expected installable via ko finder")
	}
	b = BinderHostInfo{}
	if b.BinderLinuxInstallable() {
		t.Fatal("expected not installable")
	}
}
