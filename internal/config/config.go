package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ProxyConfig struct {
	HTTP    string `yaml:"http"`
	HTTPS   string `yaml:"https"`
	NoProxy string `yaml:"no_proxy"`
}

type Config struct {
	Proxy  ProxyConfig     `yaml:"proxy"`
	CACert string          `yaml:"ca_cert"`
	Tools  map[string]bool `yaml:"tools"`
}

func DefaultTools() map[string]bool {
	return map[string]bool{
		"env_vars":  true,
		"git":       true,
		"pip":       true,
		"npm":       true,
		"yarn":      true,
		"docker":    true,
		"podman":    true,
		"curl":      true,
		"wget":      true,
		"cargo":     true,
		"conda":     true,
		"go":        true,
		"gradle":    true,
		"maven":     true,
		"bundler":   true,
		"brew":      true,
		"snap":      true,
		"apt":       true,
		"yum":       true,
		"ssh":       false,
		"system_ca": true,
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func (c *Config) CACertAbsPath() string {
	return ExpandPath(c.CACert)
}
