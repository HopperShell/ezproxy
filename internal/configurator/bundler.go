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

// Bundler configures Ruby Bundler's SSL CA cert path.
// Bundler uses HTTP_PROXY/HTTPS_PROXY from the environment (handled by env_vars),
// but needs BUNDLE_SSL_CA_CERT for corporate proxy CA certs.
type Bundler struct{}

func (b *Bundler) Name() string { return "bundler" }

func (b *Bundler) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("bundle")
}

func (b *Bundler) configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".bundle", "config")
}

func (b *Bundler) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)

	if certPath == "" {
		// Bundler uses HTTP_PROXY from env (handled by env_vars).
		// Without a CA cert, there's nothing Bundler-specific to configure.
		if fileutil.DryRun {
			fmt.Printf("\n  [dry-run] No CA cert configured; Bundler uses HTTP_PROXY from env_vars.\n")
		}
		return nil
	}

	path := b.configPath()

	if fileutil.DryRun {
		fmt.Printf("\n  [dry-run] Would set in %s:\n", path)
		fmt.Printf("    BUNDLE_SSL_CA_CERT: \"%s\"\n", certPath)
		return nil
	}

	// Read existing config
	existing := make(map[string]string)
	if data, err := os.ReadFile(path); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || line == "---" {
				continue
			}
			if idx := strings.Index(line, ": "); idx > 0 {
				key := line[:idx]
				val := strings.Trim(line[idx+2:], "\"")
				existing[key] = val
			}
		}
	}

	existing["BUNDLE_SSL_CA_CERT"] = certPath

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	var buf strings.Builder
	buf.WriteString("---\n")
	for k, v := range existing {
		buf.WriteString(fmt.Sprintf("%s: \"%s\"\n", k, v))
	}

	return os.WriteFile(path, []byte(buf.String()), 0644)
}

func (b *Bundler) Remove() error {
	path := b.configPath()

	if fileutil.DryRun {
		fmt.Printf("\n  [dry-run] Would remove BUNDLE_SSL_CA_CERT from %s\n", path)
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var buf strings.Builder
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "BUNDLE_SSL_CA_CERT") {
			continue
		}
		buf.WriteString(line + "\n")
	}

	result := strings.TrimRight(buf.String(), "\n") + "\n"
	return os.WriteFile(path, []byte(result), 0644)
}

func (b *Bundler) Status(cfg *config.Config) (string, error) {
	data, err := os.ReadFile(b.configPath())
	if err != nil {
		return "not configured", nil
	}
	if strings.Contains(string(data), "BUNDLE_SSL_CA_CERT") {
		return "configured", nil
	}
	return "not configured", nil
}
