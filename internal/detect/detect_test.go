package detect

import (
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

func TestShellProfiles(t *testing.T) {
	profiles := ShellProfiles()
	if len(profiles) == 0 {
		t.Error("should find at least one shell profile")
	}
}
