package configurator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestDockerApplyClient(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Create existing config with other keys
	existing := map[string]interface{}{
		"credsStore": "desktop",
		"auths":      map[string]interface{}{},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(configPath, data, 0644)

	d := &Docker{configPath: configPath}
	cfg := &config.Config{
		Proxy: config.ProxyConfig{
			HTTP:    "http://proxy:8080",
			HTTPS:   "http://proxy:8080",
			NoProxy: "localhost,127.0.0.1",
		},
	}

	if err := d.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Read back
	data, _ = os.ReadFile(configPath)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Check proxies were added
	proxies, ok := result["proxies"].(map[string]interface{})
	if !ok {
		t.Fatal("proxies key missing")
	}
	def, ok := proxies["default"].(map[string]interface{})
	if !ok {
		t.Fatal("proxies.default missing")
	}
	if def["httpProxy"] != "http://proxy:8080" {
		t.Errorf("httpProxy = %v", def["httpProxy"])
	}

	// Check existing keys preserved
	if result["credsStore"] != "desktop" {
		t.Error("credsStore should be preserved")
	}
}

func TestDockerRemove(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	d := &Docker{configPath: configPath}
	cfg := &config.Config{
		Proxy: config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080", NoProxy: "localhost"},
	}
	d.Apply(cfg)
	d.Remove()

	data, _ := os.ReadFile(configPath)
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	if _, ok := result["proxies"]; ok {
		t.Error("proxies should be removed")
	}
}
