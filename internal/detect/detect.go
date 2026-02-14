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

func ShellProfiles() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".profile"),
	}

	var profiles []string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			profiles = append(profiles, p)
		}
	}
	return profiles
}
