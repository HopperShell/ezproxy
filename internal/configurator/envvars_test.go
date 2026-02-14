package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestEnvVarsApply(t *testing.T) {
	dir := t.TempDir()
	bashrc := filepath.Join(dir, ".bashrc")
	os.WriteFile(bashrc, []byte("# existing\n"), 0644)

	e := &EnvVars{profiles: []string{bashrc}}
	cfg := &config.Config{
		Proxy: config.ProxyConfig{
			HTTP:    "http://proxy:8080",
			HTTPS:   "http://proxy:8080",
			NoProxy: "localhost",
		},
		CACert: "/tmp/ca.pem",
	}

	if err := e.Apply(cfg); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	data, _ := os.ReadFile(bashrc)
	got := string(data)
	if !strings.Contains(got, "HTTP_PROXY=http://proxy:8080") {
		t.Error("missing HTTP_PROXY")
	}
	if !strings.Contains(got, "http_proxy=http://proxy:8080") {
		t.Error("missing lowercase http_proxy")
	}
	if !strings.Contains(got, "NODE_EXTRA_CA_CERTS=/tmp/ca.pem") {
		t.Error("missing NODE_EXTRA_CA_CERTS")
	}
	if !strings.Contains(got, "HOMEBREW_CURLRC=1") {
		t.Error("missing HOMEBREW_CURLRC")
	}
	if !strings.Contains(got, "# existing") {
		t.Error("existing content should be preserved")
	}
}

func TestEnvVarsRemove(t *testing.T) {
	dir := t.TempDir()
	bashrc := filepath.Join(dir, ".bashrc")
	os.WriteFile(bashrc, []byte("# existing\n"), 0644)

	e := &EnvVars{profiles: []string{bashrc}}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080", NoProxy: "localhost"},
		CACert: "/tmp/ca.pem",
	}
	e.Apply(cfg)
	e.Remove()

	data, _ := os.ReadFile(bashrc)
	got := string(data)
	if strings.Contains(got, "ezproxy") {
		t.Error("marker block should be removed")
	}
	if !strings.Contains(got, "# existing") {
		t.Error("existing content should remain")
	}
}

func TestEnvVarsApplyIdempotent(t *testing.T) {
	dir := t.TempDir()
	bashrc := filepath.Join(dir, ".bashrc")
	os.WriteFile(bashrc, []byte(""), 0644)

	e := &EnvVars{profiles: []string{bashrc}}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080", NoProxy: "localhost"},
		CACert: "/tmp/ca.pem",
	}

	e.Apply(cfg)
	e.Apply(cfg)

	data, _ := os.ReadFile(bashrc)
	got := string(data)
	if strings.Count(got, ">>> ezproxy >>>") != 1 {
		t.Error("should have exactly one marker block after double apply")
	}
}
