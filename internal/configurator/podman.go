package configurator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

type Podman struct{}

func (p *Podman) Name() string { return "podman" }

func (p *Podman) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("podman")
}

func (p *Podman) configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "containers", "containers.conf")
}

func (p *Podman) Apply(cfg *config.Config) error {
	path := p.configPath()

	// Podman containers.conf uses TOML. We use marker blocks to manage
	// our env entries in the [containers] section.
	content := fmt.Sprintf(`[containers]
env = [
  "http_proxy=%s",
  "https_proxy=%s",
  "no_proxy=%s",
  "HTTP_PROXY=%s",
  "HTTPS_PROXY=%s",
  "NO_PROXY=%s",
]
`, cfg.Proxy.HTTP, cfg.Proxy.HTTPS, cfg.Proxy.NoProxy,
		cfg.Proxy.HTTP, cfg.Proxy.HTTPS, cfg.Proxy.NoProxy)

	return fileutil.UpsertMarkerBlock(path, content, "#")
}

func (p *Podman) Remove() error {
	path := p.configPath()
	return fileutil.RemoveMarkerBlock(path, "#")
}

func (p *Podman) Status(cfg *config.Config) (string, error) {
	path := p.configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return "not configured", nil
	}
	if strings.Contains(string(data), cfg.Proxy.HTTP) {
		return "configured", nil
	}
	return "not configured", nil
}
