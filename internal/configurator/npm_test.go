package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestNpmApply(t *testing.T) {
	dir := t.TempDir()
	n := &Npm{path: filepath.Join(dir, ".npmrc")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	if err := n.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	data, _ := os.ReadFile(n.path)
	got := string(data)
	if !strings.Contains(got, "proxy=http://proxy:8080") {
		t.Error("missing proxy")
	}
	if !strings.Contains(got, "https-proxy=http://proxy:8080") {
		t.Error("missing https-proxy")
	}
	if !strings.Contains(got, "cafile=/tmp/ca.pem") {
		t.Error("missing cafile")
	}
}

func TestNpmRemove(t *testing.T) {
	dir := t.TempDir()
	n := &Npm{path: filepath.Join(dir, ".npmrc")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	n.Apply(cfg)
	if err := n.Remove(); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	status, _ := n.Status(cfg)
	if status != "not configured" {
		t.Errorf("expected 'not configured', got %q", status)
	}
}

func TestNpmName(t *testing.T) {
	n := &Npm{}
	if n.Name() != "npm" {
		t.Errorf("expected 'npm', got %q", n.Name())
	}
}
