package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/fileutil"
)

func TestYarnV1Apply(t *testing.T) {
	dir := t.TempDir()
	v1Path := filepath.Join(dir, ".yarnrc")
	y := &Yarn{
		v1Path: v1Path,
		v2Path: filepath.Join(dir, ".yarnrc.yml"),
	}

	// Simulate v1 content by writing marker block directly (avoids needing yarn installed)
	content := "proxy \"http://proxy:8080\"\nhttps-proxy \"http://proxy:8080\"\ncafile \"/tmp/ca.pem\"\n"
	if err := fileutil.UpsertMarkerBlock(v1Path, content, "#"); err != nil {
		t.Fatalf("UpsertMarkerBlock: %v", err)
	}

	data, _ := os.ReadFile(v1Path)
	got := string(data)
	if !strings.Contains(got, `proxy "http://proxy:8080"`) {
		t.Error("missing proxy")
	}
	if !strings.Contains(got, `https-proxy "http://proxy:8080"`) {
		t.Error("missing https-proxy")
	}
	if !strings.Contains(got, `cafile "/tmp/ca.pem"`) {
		t.Error("missing cafile")
	}

	// Verify status reports configured
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	status, _ := y.Status(cfg)
	if status != "configured" {
		t.Errorf("expected 'configured', got %q", status)
	}
}

func TestYarnV2Format(t *testing.T) {
	dir := t.TempDir()
	v2Path := filepath.Join(dir, ".yarnrc.yml")
	y := &Yarn{
		v1Path: filepath.Join(dir, ".yarnrc"),
		v2Path: v2Path,
	}

	// Simulate v2 content by writing marker block directly
	content := "httpProxy: \"http://proxy:8080\"\nhttpsProxy: \"http://proxy:8080\"\ncaFilePath: \"/tmp/ca.pem\"\n"
	if err := fileutil.UpsertMarkerBlock(v2Path, content, "#"); err != nil {
		t.Fatalf("UpsertMarkerBlock: %v", err)
	}

	data, _ := os.ReadFile(v2Path)
	got := string(data)
	if !strings.Contains(got, `httpProxy: "http://proxy:8080"`) {
		t.Error("missing httpProxy")
	}
	if !strings.Contains(got, `httpsProxy: "http://proxy:8080"`) {
		t.Error("missing httpsProxy")
	}
	if !strings.Contains(got, `caFilePath: "/tmp/ca.pem"`) {
		t.Error("missing caFilePath")
	}

	// Verify status reports configured
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	status, _ := y.Status(cfg)
	if status != "configured" {
		t.Errorf("expected 'configured', got %q", status)
	}
}

func TestYarnRemove(t *testing.T) {
	dir := t.TempDir()
	v1Path := filepath.Join(dir, ".yarnrc")
	y := &Yarn{
		v1Path: v1Path,
		v2Path: filepath.Join(dir, ".yarnrc.yml"),
	}
	cfg := &config.Config{
		Proxy: config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080"},
	}

	// Write a v1 marker block so we can remove it
	fileutil.UpsertMarkerBlock(v1Path, "proxy \"http://proxy:8080\"\n", "#")

	if err := y.Remove(); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	status, _ := y.Status(cfg)
	if status != "not configured" {
		t.Errorf("expected 'not configured', got %q", status)
	}
}

func TestYarnName(t *testing.T) {
	y := &Yarn{}
	if y.Name() != "yarn" {
		t.Errorf("expected 'yarn', got %q", y.Name())
	}
}
