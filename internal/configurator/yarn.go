package configurator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

type Yarn struct {
	v1Path string // override for testing
	v2Path string // override for testing
}

func (y *Yarn) Name() string { return "yarn" }

func (y *Yarn) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("yarn")
}

func (y *Yarn) isV2OrLater() bool {
	out, err := exec.Command("yarn", "--version").Output()
	if err != nil {
		return false
	}
	version := strings.TrimSpace(string(out))
	return len(version) > 0 && version[0] >= '2'
}

func (y *Yarn) getV1Path() string {
	if y.v1Path != "" {
		return y.v1Path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".yarnrc")
}

func (y *Yarn) getV2Path() string {
	if y.v2Path != "" {
		return y.v2Path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".yarnrc.yml")
}

func (y *Yarn) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)

	if y.isV2OrLater() {
		var b strings.Builder
		fmt.Fprintf(&b, "httpProxy: \"%s\"\n", cfg.Proxy.HTTP)
		fmt.Fprintf(&b, "httpsProxy: \"%s\"\n", cfg.Proxy.HTTPS)
		if certPath != "" {
			fmt.Fprintf(&b, "caFilePath: \"%s\"\n", certPath)
		}
		return fileutil.UpsertMarkerBlock(y.getV2Path(), b.String(), "#")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "proxy \"%s\"\n", cfg.Proxy.HTTP)
	fmt.Fprintf(&b, "https-proxy \"%s\"\n", cfg.Proxy.HTTPS)
	if certPath != "" {
		fmt.Fprintf(&b, "cafile \"%s\"\n", certPath)
	}
	return fileutil.UpsertMarkerBlock(y.getV1Path(), b.String(), "#")
}

func (y *Yarn) Remove() error {
	fileutil.RemoveMarkerBlock(y.getV1Path(), "#")
	fileutil.RemoveMarkerBlock(y.getV2Path(), "#")
	return nil
}

func (y *Yarn) Status(cfg *config.Config) (string, error) {
	if fileutil.HasMarkerBlock(y.getV1Path(), "#") || fileutil.HasMarkerBlock(y.getV2Path(), "#") {
		return "configured", nil
	}
	return "not configured", nil
}
