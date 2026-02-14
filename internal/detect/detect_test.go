package detect

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectOS(t *testing.T) {
	info := DetectOS()
	if runtime.GOOS == "darwin" {
		if info.OS != "darwin" {
			t.Errorf("expected darwin, got %s", info.OS)
		}
	} else if runtime.GOOS == "linux" {
		if info.OS != "linux" {
			t.Errorf("expected linux, got %s", info.OS)
		}
	}
}

func TestIsCommandAvailable(t *testing.T) {
	if !IsCommandAvailable("ls") {
		t.Error("ls should be available")
	}
	if IsCommandAvailable("ezproxy_nonexistent_command_xyz") {
		t.Error("fake command should not be available")
	}
}

func TestDetectShell(t *testing.T) {
	// Save and restore SHELL
	origShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", origShell)

	os.Setenv("SHELL", "/bin/zsh")
	if got := DetectShell(); got != "zsh" {
		t.Errorf("DetectShell() = %q, want zsh", got)
	}

	os.Setenv("SHELL", "/usr/bin/bash")
	if got := DetectShell(); got != "bash" {
		t.Errorf("DetectShell() = %q, want bash", got)
	}

	os.Setenv("SHELL", "/usr/bin/fish")
	if got := DetectShell(); got != "fish" {
		t.Errorf("DetectShell() = %q, want fish", got)
	}
}

func TestShellProfiles(t *testing.T) {
	// ShellProfiles depends on $SHELL and existing files in $HOME.
	// Just verify it returns something for the current user.
	profiles := ShellProfiles()
	if len(profiles) == 0 {
		t.Error("should find at least one shell profile")
	}
}

func TestShellProfilesZsh(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("SHELL", "/bin/zsh")

	// Create .zshrc
	os.WriteFile(filepath.Join(dir, ".zshrc"), []byte(""), 0644)

	profiles := ShellProfiles()
	if len(profiles) == 0 {
		t.Fatal("expected at least one profile")
	}
	for _, p := range profiles {
		base := filepath.Base(p)
		if base != ".zshrc" && base != ".zprofile" {
			t.Errorf("unexpected profile for zsh: %s", p)
		}
	}
}

func TestShellProfilesBash(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("SHELL", "/bin/bash")

	// Create both bash files
	os.WriteFile(filepath.Join(dir, ".bashrc"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dir, ".bash_profile"), []byte(""), 0644)

	profiles := ShellProfiles()
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d: %v", len(profiles), profiles)
	}

	names := map[string]bool{}
	for _, p := range profiles {
		names[filepath.Base(p)] = true
	}
	if !names[".bashrc"] {
		t.Error("missing .bashrc")
	}
	if !names[".bash_profile"] {
		t.Error("missing .bash_profile")
	}
}

func TestShellProfilesBashFallbackToProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("SHELL", "/bin/bash")

	// No .bashrc or .bash_profile, only .profile
	os.WriteFile(filepath.Join(dir, ".profile"), []byte(""), 0644)

	profiles := ShellProfiles()
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d: %v", len(profiles), profiles)
	}
	if filepath.Base(profiles[0]) != ".profile" {
		t.Errorf("expected .profile fallback, got %s", profiles[0])
	}
}

func TestIsFishShell(t *testing.T) {
	origShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", origShell)

	os.Setenv("SHELL", "/usr/bin/fish")
	if !IsFishShell() {
		t.Error("expected IsFishShell() = true")
	}

	os.Setenv("SHELL", "/bin/zsh")
	if IsFishShell() {
		t.Error("expected IsFishShell() = false")
	}
}
