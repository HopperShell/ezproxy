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

type Conda struct {
	path string
}

func (c *Conda) Name() string { return "conda" }

func (c *Conda) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("conda")
}

func (c *Conda) getPath() string {
	if c.path != "" {
		return c.path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".condarc")
}

func (c *Conda) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	var b strings.Builder
	fmt.Fprintf(&b, "proxy_servers:\n")
	fmt.Fprintf(&b, "  http: %s\n", cfg.Proxy.HTTP)
	fmt.Fprintf(&b, "  https: %s\n", cfg.Proxy.HTTPS)
	if certPath != "" {
		fmt.Fprintf(&b, "ssl_verify: %s\n", certPath)
	}
	return fileutil.UpsertMarkerBlock(c.getPath(), b.String(), "#")
}

func (c *Conda) Remove() error {
	return fileutil.RemoveMarkerBlock(c.getPath(), "#")
}

func (c *Conda) Status(cfg *config.Config) (string, error) {
	if fileutil.HasMarkerBlock(c.getPath(), "#") {
		return "configured", nil
	}
	return "not configured", nil
}
