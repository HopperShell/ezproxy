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

type Wget struct {
	path string
}

func (w *Wget) Name() string { return "wget" }

func (w *Wget) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("wget")
}

func (w *Wget) getPath() string {
	if w.path != "" {
		return w.path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".wgetrc")
}

func (w *Wget) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	var b strings.Builder
	fmt.Fprintf(&b, "http_proxy = %s\n", cfg.Proxy.HTTP)
	fmt.Fprintf(&b, "https_proxy = %s\n", cfg.Proxy.HTTPS)
	if certPath != "" {
		fmt.Fprintf(&b, "ca_certificate = %s\n", certPath)
	}
	return fileutil.UpsertMarkerBlock(w.getPath(), b.String(), "#")
}

func (w *Wget) Remove() error {
	return fileutil.RemoveMarkerBlock(w.getPath(), "#")
}

func (w *Wget) Status(cfg *config.Config) (string, error) {
	if fileutil.HasMarkerBlock(w.getPath(), "#") {
		return "configured", nil
	}
	return "not configured", nil
}
