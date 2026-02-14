package configurator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

type Docker struct {
	configPath string // override for testing
}

func (d *Docker) Name() string { return "docker" }

func (d *Docker) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("docker")
}

func (d *Docker) getConfigPath() string {
	if d.configPath != "" {
		return d.configPath
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".docker", "config.json")
}

func (d *Docker) Apply(cfg *config.Config) error {
	// Client proxy config
	if err := d.applyClientConfig(cfg); err != nil {
		return fmt.Errorf("docker client config: %w", err)
	}

	// Daemon config (Linux only)
	if runtime.GOOS == "linux" {
		if err := d.applyDaemonConfig(cfg); err != nil {
			fmt.Printf("  docker daemon: %v\n", err)
		}
	} else if runtime.GOOS == "darwin" {
		fmt.Println("\n[Docker Desktop - macOS]")
		fmt.Println("Configure proxy via: Docker Desktop > Settings > Resources > Proxies")
		fmt.Printf("  HTTP Proxy:  %s\n", cfg.Proxy.HTTP)
		fmt.Printf("  HTTPS Proxy: %s\n", cfg.Proxy.HTTPS)
		fmt.Printf("  No Proxy:    %s\n", cfg.Proxy.NoProxy)
		fmt.Println("Docker Desktop reads macOS system CA certs automatically after restart.")
	}

	return nil
}

func (d *Docker) applyClientConfig(cfg *config.Config) error {
	path := d.getConfigPath()

	proxies := map[string]interface{}{
		"default": map[string]interface{}{
			"httpProxy":  cfg.Proxy.HTTP,
			"httpsProxy": cfg.Proxy.HTTPS,
			"noProxy":    cfg.Proxy.NoProxy,
		},
	}

	if fileutil.DryRun {
		data, _ := json.MarshalIndent(map[string]interface{}{"proxies": proxies}, "", "  ")
		fmt.Printf("\n  [dry-run] Would merge into %s:\n", path)
		for _, line := range strings.Split(string(data), "\n") {
			fmt.Printf("    %s\n", line)
		}
		return nil
	}

	// Read existing config
	var dockerConfig map[string]interface{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &dockerConfig); err != nil {
			dockerConfig = make(map[string]interface{})
		}
	} else {
		dockerConfig = make(map[string]interface{})
	}

	dockerConfig["proxies"] = proxies

	// Write back
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(dockerConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func (d *Docker) applyDaemonConfig(cfg *config.Config) error {
	content := fmt.Sprintf("[Service]\nEnvironment=\"HTTP_PROXY=%s\"\nEnvironment=\"HTTPS_PROXY=%s\"\nEnvironment=\"NO_PROXY=%s\"\n",
		cfg.Proxy.HTTP, cfg.Proxy.HTTPS, cfg.Proxy.NoProxy)
	return runSudoCommands("docker daemon", []string{
		"mkdir -p /etc/systemd/system/docker.service.d",
		fmt.Sprintf("printf '%s' > /etc/systemd/system/docker.service.d/ezproxy.conf", content),
		"systemctl daemon-reload && systemctl restart docker",
	})
}

func (d *Docker) Remove() error {
	path := d.getConfigPath()

	if fileutil.DryRun {
		fmt.Printf("\n  [dry-run] Would remove \"proxies\" key from %s\n", path)
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var dockerConfig map[string]interface{}
	if err := json.Unmarshal(data, &dockerConfig); err != nil {
		return nil
	}

	delete(dockerConfig, "proxies")

	out, err := json.MarshalIndent(dockerConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0644)
}

func (d *Docker) Status(cfg *config.Config) (string, error) {
	path := d.getConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return "not configured", nil
	}

	var dockerConfig map[string]interface{}
	if err := json.Unmarshal(data, &dockerConfig); err != nil {
		return "not configured", nil
	}

	if _, ok := dockerConfig["proxies"]; ok {
		return "configured", nil
	}
	return "not configured", nil
}
