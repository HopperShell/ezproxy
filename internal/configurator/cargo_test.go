package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestCargoApply(t *testing.T) {
	dir := t.TempDir()
	c := &Cargo{path: filepath.Join(dir, "config.toml")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	if err := c.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	data, _ := os.ReadFile(c.path)
	got := string(data)
	if !strings.Contains(got, "[http]") {
		t.Error("missing [http] section")
	}
	if !strings.Contains(got, `proxy = "http://proxy:8080"`) {
		t.Error("missing proxy")
	}
	if !strings.Contains(got, `cainfo = "/tmp/ca.pem"`) {
		t.Error("missing cainfo")
	}
}

func TestCargoRemove(t *testing.T) {
	dir := t.TempDir()
	c := &Cargo{path: filepath.Join(dir, "config.toml")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	c.Apply(cfg)
	if err := c.Remove(); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	status, _ := c.Status(cfg)
	if status != "not configured" {
		t.Errorf("expected 'not configured', got %q", status)
	}
}

func TestCargoName(t *testing.T) {
	c := &Cargo{}
	if c.Name() != "cargo" {
		t.Errorf("expected 'cargo', got %q", c.Name())
	}
}
