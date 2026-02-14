package configurator

import (
	"fmt"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

type EnvVars struct {
	profiles []string // override for testing; nil = auto-detect
}

func (e *EnvVars) Name() string { return "env_vars" }

func (e *EnvVars) IsAvailable(_ detect.OSInfo) bool { return true }

func (e *EnvVars) getProfiles() []string {
	if e.profiles != nil {
		return e.profiles
	}
	return detect.ShellProfiles()
}

func (e *EnvVars) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)

	var b strings.Builder
	fmt.Fprintf(&b, "export HTTP_PROXY=%s\n", cfg.Proxy.HTTP)
	fmt.Fprintf(&b, "export HTTPS_PROXY=%s\n", cfg.Proxy.HTTPS)
	fmt.Fprintf(&b, "export http_proxy=%s\n", cfg.Proxy.HTTP)
	fmt.Fprintf(&b, "export https_proxy=%s\n", cfg.Proxy.HTTPS)
	fmt.Fprintf(&b, "export NO_PROXY=%s\n", cfg.Proxy.NoProxy)
	fmt.Fprintf(&b, "export no_proxy=%s\n", cfg.Proxy.NoProxy)
	if certPath != "" {
		fmt.Fprintf(&b, "export SSL_CERT_FILE=%s\n", certPath)
		fmt.Fprintf(&b, "export REQUESTS_CA_BUNDLE=%s\n", certPath)
		fmt.Fprintf(&b, "export CURL_CA_BUNDLE=%s\n", certPath)
		fmt.Fprintf(&b, "export NODE_EXTRA_CA_CERTS=%s\n", certPath)
	}
	fmt.Fprintf(&b, "export HOMEBREW_CURLRC=1\n")

	for _, profile := range e.getProfiles() {
		if err := fileutil.UpsertMarkerBlock(profile, b.String(), "#"); err != nil {
			return fmt.Errorf("updating %s: %w", profile, err)
		}
	}
	return nil
}

func (e *EnvVars) Remove() error {
	for _, profile := range e.getProfiles() {
		if err := fileutil.RemoveMarkerBlock(profile, "#"); err != nil {
			return fmt.Errorf("cleaning %s: %w", profile, err)
		}
	}
	return nil
}

func (e *EnvVars) Status(cfg *config.Config) (string, error) {
	profiles := e.getProfiles()
	if len(profiles) == 0 {
		return "not configured", nil
	}
	for _, profile := range profiles {
		if fileutil.HasMarkerBlock(profile, "#") {
			return "configured", nil
		}
	}
	return "not configured", nil
}
