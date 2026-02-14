package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestCondaApply(t *testing.T) {
	dir := t.TempDir()
	c := &Conda{path: filepath.Join(dir, ".condarc")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	if err := c.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	data, _ := os.ReadFile(c.path)
	got := string(data)
	if !strings.Contains(got, "proxy_servers:") {
		t.Error("missing proxy_servers")
	}
	if !strings.Contains(got, "  http: http://proxy:8080") {
		t.Error("missing http proxy")
	}
	if !strings.Contains(got, "  https: http://proxy:8080") {
		t.Error("missing https proxy")
	}
	if !strings.Contains(got, "ssl_verify: /tmp/ca.pem") {
		t.Error("missing ssl_verify")
	}
}

func TestCondaRemove(t *testing.T) {
	dir := t.TempDir()
	c := &Conda{path: filepath.Join(dir, ".condarc")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080"},
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

func TestCondaName(t *testing.T) {
	c := &Conda{}
	if c.Name() != "conda" {
		t.Errorf("expected 'conda', got %q", c.Name())
	}
}
