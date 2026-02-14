package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestPipApply(t *testing.T) {
	dir := t.TempDir()
	p := &Pip{path: filepath.Join(dir, "pip.conf")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	if err := p.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	data, _ := os.ReadFile(p.path)
	got := string(data)
	if !strings.Contains(got, "[global]") {
		t.Error("missing [global] section")
	}
	if !strings.Contains(got, "proxy = http://proxy:8080") {
		t.Error("missing proxy")
	}
	if !strings.Contains(got, "cert = /tmp/ca.pem") {
		t.Error("missing cert")
	}
}

func TestPipRemove(t *testing.T) {
	dir := t.TempDir()
	p := &Pip{path: filepath.Join(dir, "pip.conf")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	p.Apply(cfg)
	if err := p.Remove(); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	status, _ := p.Status(cfg)
	if status != "not configured" {
		t.Errorf("expected 'not configured', got %q", status)
	}
}

func TestPipName(t *testing.T) {
	p := &Pip{}
	if p.Name() != "pip" {
		t.Errorf("expected 'pip', got %q", p.Name())
	}
}
