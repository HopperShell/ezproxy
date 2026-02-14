package configurator

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

type SSH struct {
	path string
}

func (s *SSH) Name() string { return "ssh" }

func (s *SSH) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("ssh")
}

func (s *SSH) getPath() string {
	if s.path != "" {
		return s.path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ssh", "config")
}

func (s *SSH) Apply(cfg *config.Config) error {
	// Extract host:port from proxy URL
	proxyURL, err := url.Parse(cfg.Proxy.HTTP)
	if err != nil {
		return fmt.Errorf("parsing proxy URL: %w", err)
	}
	proxyHost := proxyURL.Host

	var b strings.Builder
	fmt.Fprintf(&b, "Host *\n")
	fmt.Fprintf(&b, "    ProxyCommand nc -X connect -x %s %%h %%p\n", proxyHost)

	if err := fileutil.UpsertMarkerBlock(s.getPath(), b.String(), "#"); err != nil {
		return err
	}

	// Warn about GNU netcat on Linux
	osInfo := detect.DetectOS()
	if osInfo.OS == "linux" {
		fmt.Println("\nNote: SSH proxy requires OpenBSD netcat (netcat-openbsd).")
		fmt.Println("GNU netcat does NOT support -X/-x proxy flags.")
		fmt.Println("Install: sudo apt install netcat-openbsd (Debian/Ubuntu)")
	}

	return nil
}

func (s *SSH) Remove() error {
	return fileutil.RemoveMarkerBlock(s.getPath(), "#")
}

func (s *SSH) Status(cfg *config.Config) (string, error) {
	if fileutil.HasMarkerBlock(s.getPath(), "#") {
		return "configured", nil
	}
	return "not configured", nil
}
