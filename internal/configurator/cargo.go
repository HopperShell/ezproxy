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

type Cargo struct {
	path string
}

func (c *Cargo) Name() string { return "cargo" }

func (c *Cargo) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("cargo")
}

func (c *Cargo) getPath() string {
	if c.path != "" {
		return c.path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cargo", "config.toml")
}

func (c *Cargo) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	var b strings.Builder
	fmt.Fprintf(&b, "[http]\n")
	fmt.Fprintf(&b, "proxy = \"%s\"\n", cfg.Proxy.HTTP)
	if certPath != "" {
		fmt.Fprintf(&b, "cainfo = \"%s\"\n", certPath)
	}
	return fileutil.UpsertMarkerBlock(c.getPath(), b.String(), "#")
}

func (c *Cargo) Remove() error {
	return fileutil.RemoveMarkerBlock(c.getPath(), "#")
}

func (c *Cargo) Status(cfg *config.Config) (string, error) {
	if fileutil.HasMarkerBlock(c.getPath(), "#") {
		return "configured", nil
	}
	return "not configured", nil
}
