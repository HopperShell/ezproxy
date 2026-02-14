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

type Npm struct {
	path string
}

func (n *Npm) Name() string { return "npm" }

func (n *Npm) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("npm")
}

func (n *Npm) getPath() string {
	if n.path != "" {
		return n.path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".npmrc")
}

func (n *Npm) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	var b strings.Builder
	fmt.Fprintf(&b, "proxy=%s\n", cfg.Proxy.HTTP)
	fmt.Fprintf(&b, "https-proxy=%s\n", cfg.Proxy.HTTPS)
	if certPath != "" {
		fmt.Fprintf(&b, "cafile=%s\n", certPath)
	}
	return fileutil.UpsertMarkerBlock(n.getPath(), b.String(), "#")
}

func (n *Npm) Remove() error {
	return fileutil.RemoveMarkerBlock(n.getPath(), "#")
}

func (n *Npm) Status(cfg *config.Config) (string, error) {
	if fileutil.HasMarkerBlock(n.getPath(), "#") {
		return "configured", nil
	}
	return "not configured", nil
}
