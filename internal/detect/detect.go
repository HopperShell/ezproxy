package detect

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type OSInfo struct {
	OS     string // "darwin" or "linux"
	Distro string // "debian", "ubuntu", "fedora", "rhel", "centos", "arch", "" (macOS)
}

func DetectOS() OSInfo {
	info := OSInfo{OS: runtime.GOOS}
	if runtime.GOOS == "linux" {
		info.Distro = detectLinuxDistro()
	}
	return info
}

func detectLinuxDistro() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	content := strings.ToLower(string(data))
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "id=") {
			id := strings.Trim(strings.TrimPrefix(line, "id="), "\"")
			return id
		}
	}
	return ""
}

func (o OSInfo) IsDebian() bool {
	return o.Distro == "debian" || o.Distro == "ubuntu" || o.Distro == "pop" || o.Distro == "mint"
}

func (o OSInfo) IsRHEL() bool {
	return o.Distro == "fedora" || o.Distro == "rhel" || o.Distro == "centos" || o.Distro == "rocky" || o.Distro == "alma"
}

func (o OSInfo) IsArch() bool {
	return o.Distro == "arch" || o.Distro == "manjaro" || o.Distro == "endeavouros"
}

func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// DetectShell returns the name of the user's login shell (e.g. "zsh", "bash", "fish").
// It checks $SHELL first, then falls back to looking at running processes.
func DetectShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		return filepath.Base(shell)
	}
	return ""
}

// ShellProfiles returns the profile files that should be modified for the
// user's detected shell. Only returns files for the user's actual shell,
// plus .profile for POSIX login shell compatibility.
func ShellProfiles() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	shell := DetectShell()

	var candidates []string
	switch shell {
	case "zsh":
		candidates = []string{
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".zprofile"),
		}
	case "bash":
		candidates = []string{
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".bash_profile"),
		}
		// If neither .bashrc nor .bash_profile exist, fall back to .profile
		// (.profile is read by bash login shells when .bash_profile is absent)
	case "fish":
		candidates = []string{
			filepath.Join(home, ".config", "fish", "conf.d", "ezproxy.fish"),
		}
	default:
		// Unknown shell — try common profile files that exist
		candidates = []string{
			filepath.Join(home, ".profile"),
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".bash_profile"),
			filepath.Join(home, ".zshrc"),
		}
	}

	var profiles []string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			profiles = append(profiles, p)
		}
	}

	// For bash: if no bash-specific files exist, fall back to .profile
	if shell == "bash" && len(profiles) == 0 {
		profile := filepath.Join(home, ".profile")
		if _, err := os.Stat(profile); err == nil {
			profiles = append(profiles, profile)
		}
	}

	return profiles
}

// IsFishShell returns true if the user's shell is fish.
// Fish uses different syntax (set -gx instead of export).
func IsFishShell() bool {
	return DetectShell() == "fish"
}
