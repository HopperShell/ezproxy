package configurator

import (
	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

type Brew struct{}

func (b *Brew) Name() string { return "brew" }

func (b *Brew) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("brew")
}

func (b *Brew) Apply(cfg *config.Config) error {
	// Covered by env_vars configurator (HTTP_PROXY + HOMEBREW_CURLRC=1)
	return nil
}

func (b *Brew) Remove() error {
	// Covered by env_vars configurator
	return nil
}

func (b *Brew) Status(cfg *config.Config) (string, error) {
	// Check if any shell profile has the HOMEBREW_CURLRC marker
	for _, profile := range detect.ShellProfiles() {
		if fileutil.HasMarkerBlock(profile, "#") {
			return "configured (via env_vars)", nil
		}
	}
	return "not configured", nil
}
