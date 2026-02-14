package configurator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
)

func TestGitApplyAndRemove(t *testing.T) {
	if !detect.IsCommandAvailable("git") {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	gitconfig := filepath.Join(dir, ".gitconfig")
	os.WriteFile(gitconfig, []byte(""), 0644)
	t.Setenv("GIT_CONFIG_GLOBAL", gitconfig)

	g := &Git{}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}

	if err := g.Apply(cfg); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	out, _ := exec.Command("git", "config", "--global", "http.proxy").Output()
	if strings.TrimSpace(string(out)) != "http://proxy:8080" {
		t.Errorf("http.proxy = %q", strings.TrimSpace(string(out)))
	}

	out, _ = exec.Command("git", "config", "--global", "http.sslCAInfo").Output()
	if strings.TrimSpace(string(out)) != "/tmp/ca.pem" {
		t.Errorf("http.sslCAInfo = %q", strings.TrimSpace(string(out)))
	}

	if err := g.Remove(); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	out, err := exec.Command("git", "config", "--global", "http.proxy").Output()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		t.Error("http.proxy should be unset after Remove")
	}
}
