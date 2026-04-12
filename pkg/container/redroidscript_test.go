package container

import (
	"os"
	"path/filepath"
	"testing"

	"reddock/pkg/config"
)

// Tag order must match ayasa520/redroid-script main() (gapps → litegapps → mindthegapps → ndk → houdini → magisk → widevine).
func TestRedroidScriptIntermediateImageName_matchesUpstreamOrder(t *testing.T) {
	got := RedroidScriptIntermediateImageName("11.0.0", RedroidScriptAddonFlags{
		Gapps: true, NDK: true, Magisk: true,
	})
	want := "redroid/redroid:11.0.0_gapps_ndk_magisk"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveRedroidScriptRoot_nestedGuestClone(t *testing.T) {
	root := t.TempDir()
	inner := filepath.Join(root, "redroid-script")
	if err := os.MkdirAll(filepath.Join(inner, "stuff"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(inner, "tools"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inner, "redroid.py"), []byte("#\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inner, "tools", "helper.py"), []byte("#\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{RedroidScriptPath: root}
	got, err := ResolveRedroidScriptRoot("", cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != inner {
		t.Fatalf("got %q want %q", got, inner)
	}
}
