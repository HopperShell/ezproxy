package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestWgetApply(t *testing.T) {
	dir := t.TempDir()
	w := &Wget{path: filepath.Join(dir, ".wgetrc")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	if err := w.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	data, _ := os.ReadFile(w.path)
	got := string(data)
	if !strings.Contains(got, "http_proxy = http://proxy:8080") {
		t.Error("missing http_proxy")
	}
	if !strings.Contains(got, "ca_certificate = /tmp/ca.pem") {
		t.Error("missing ca_certificate")
	}
}
