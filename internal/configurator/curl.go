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

type Curl struct {
	path string // override for testing
}

func (c *Curl) Name() string { return "curl" }

func (c *Curl) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("curl")
}

func (c *Curl) getPath() string {
	if c.path != "" {
		return c.path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".curlrc")
}

func (c *Curl) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	var b strings.Builder
	fmt.Fprintf(&b, "proxy = \"%s\"\n", cfg.Proxy.HTTP)
	if certPath != "" {
		fmt.Fprintf(&b, "cacert = \"%s\"\n", certPath)
	}
	return fileutil.UpsertMarkerBlock(c.getPath(), b.String(), "#")
}

func (c *Curl) Remove() error {
	return fileutil.RemoveMarkerBlock(c.getPath(), "#")
}

func (c *Curl) Status(cfg *config.Config) (string, error) {
	if fileutil.HasMarkerBlock(c.getPath(), "#") {
		return "configured", nil
	}
	return "not configured", nil
}
