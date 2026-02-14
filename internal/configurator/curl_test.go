package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestCurlApply(t *testing.T) {
	dir := t.TempDir()
	c := &Curl{path: filepath.Join(dir, ".curlrc")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	if err := c.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	data, _ := os.ReadFile(c.path)
	got := string(data)
	if !strings.Contains(got, `proxy = "http://proxy:8080"`) {
		t.Error("missing proxy")
	}
	if !strings.Contains(got, `cacert = "/tmp/ca.pem"`) {
		t.Error("missing cacert")
	}
}

func TestCurlRemove(t *testing.T) {
	dir := t.TempDir()
	c := &Curl{path: filepath.Join(dir, ".curlrc")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	c.Apply(cfg)
	c.Remove()
	data, _ := os.ReadFile(c.path)
	if strings.Contains(string(data), "ezproxy") {
		t.Error("should be cleaned")
	}
}
