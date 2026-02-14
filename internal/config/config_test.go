package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	yaml := `proxy:
  http: http://proxy.corp.com:8080
  https: http://proxy.corp.com:8080
  no_proxy: localhost,127.0.0.1,.corp.com

ca_cert: /path/to/cert.pem

tools:
  env_vars: true
  git: true
  pip: false
`
	os.WriteFile(configPath, []byte(yaml), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Proxy.HTTP != "http://proxy.corp.com:8080" {
		t.Errorf("got HTTP proxy %q", cfg.Proxy.HTTP)
	}
	if cfg.Proxy.NoProxy != "localhost,127.0.0.1,.corp.com" {
		t.Errorf("got NoProxy %q", cfg.Proxy.NoProxy)
	}
	if cfg.CACert != "/path/to/cert.pem" {
		t.Errorf("got CACert %q", cfg.CACert)
	}
	if !cfg.Tools["env_vars"] {
		t.Error("expected env_vars=true")
	}
	if !cfg.Tools["git"] {
		t.Error("expected git=true")
	}
	if cfg.Tools["pip"] {
		t.Error("expected pip=false")
	}
}

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		Proxy: ProxyConfig{
			HTTP:    "http://proxy:8080",
			HTTPS:   "http://proxy:8080",
			NoProxy: "localhost",
		},
		CACert: "/tmp/ca.pem",
		Tools:  map[string]bool{"git": true, "pip": false},
	}

	err := Save(configPath, cfg)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}
	if loaded.Proxy.HTTP != cfg.Proxy.HTTP {
		t.Errorf("round-trip failed: got %q", loaded.Proxy.HTTP)
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	result := ExpandPath("~/foo/bar")
	expected := filepath.Join(home, "foo/bar")
	if result != expected {
		t.Errorf("ExpandPath: got %q, want %q", result, expected)
	}

	result2 := ExpandPath("/absolute/path")
	if result2 != "/absolute/path" {
		t.Errorf("ExpandPath absolute: got %q", result2)
	}
}
