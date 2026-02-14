package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestSSHApply(t *testing.T) {
	dir := t.TempDir()
	s := &SSH{path: filepath.Join(dir, "config")}
	cfg := &config.Config{
		Proxy: config.ProxyConfig{HTTP: "http://proxy.corp.com:8080"},
	}
	if err := s.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	data, _ := os.ReadFile(s.path)
	got := string(data)
	if !strings.Contains(got, "ProxyCommand nc -X connect -x proxy.corp.com:8080 %h %p") {
		t.Error("missing ProxyCommand")
	}
	if !strings.Contains(got, "Host *") {
		t.Error("missing Host *")
	}
}

func TestSSHRemove(t *testing.T) {
	dir := t.TempDir()
	s := &SSH{path: filepath.Join(dir, "config")}
	cfg := &config.Config{
		Proxy: config.ProxyConfig{HTTP: "http://proxy.corp.com:8080"},
	}
	s.Apply(cfg)
	s.Remove()
	data, _ := os.ReadFile(s.path)
	if strings.Contains(string(data), "ezproxy") {
		t.Error("should be cleaned")
	}
}
