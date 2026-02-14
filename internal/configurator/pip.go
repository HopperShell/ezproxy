package configurator

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

type Pip struct {
	path string
}

func (p *Pip) Name() string { return "pip" }

func (p *Pip) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("pip") || detect.IsCommandAvailable("pip3")
}

func (p *Pip) getPath() string {
	if p.path != "" {
		return p.path
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "pip", "pip.conf")
	}
	return filepath.Join(home, ".config", "pip", "pip.conf")
}

func (p *Pip) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	var b strings.Builder
	fmt.Fprintf(&b, "[global]\n")
	fmt.Fprintf(&b, "proxy = %s\n", cfg.Proxy.HTTP)
	if certPath != "" {
		fmt.Fprintf(&b, "cert = %s\n", certPath)
	}
	return fileutil.UpsertMarkerBlock(p.getPath(), b.String(), "#")
}

func (p *Pip) Remove() error {
	return fileutil.RemoveMarkerBlock(p.getPath(), "#")
}

func (p *Pip) Status(cfg *config.Config) (string, error) {
	if fileutil.HasMarkerBlock(p.getPath(), "#") {
		return "configured", nil
	}
	return "not configured", nil
}
