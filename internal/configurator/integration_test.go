package configurator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/fileutil"
)

// testConfig returns a standard test config with proxy, cert, and all tools enabled.
func testConfig(certPath string) *config.Config {
	return &config.Config{
		Proxy: config.ProxyConfig{
			HTTP:    "http://proxy.corp.com:8080",
			HTTPS:   "http://proxy.corp.com:8080",
			NoProxy: "localhost,127.0.0.1,.corp.com,10.0.0.0/8",
		},
		CACert: certPath,
		Tools:  config.DefaultTools(),
	}
}

// testConfigNoCert returns a config without a CA cert path.
func testConfigNoCert() *config.Config {
	return &config.Config{
		Proxy: config.ProxyConfig{
			HTTP:    "http://proxy.corp.com:8080",
			HTTPS:   "http://proxy.corp.com:8080",
			NoProxy: "localhost,127.0.0.1,.corp.com,10.0.0.0/8",
		},
		Tools: config.DefaultTools(),
	}
}

// setupFakeHome creates a temp dir with common dotfiles pre-created.
func setupFakeHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create a fake .bashrc so envvars has something to write to
	bashrc := filepath.Join(dir, ".bashrc")
	os.WriteFile(bashrc, []byte("# user bashrc\nexport PATH=/usr/bin\n"), 0644)

	// Create a fake CA cert file
	certDir := filepath.Join(dir, ".ezproxy")
	os.MkdirAll(certDir, 0755)
	os.WriteFile(filepath.Join(certDir, "corp-ca.pem"), []byte("-----BEGIN CERTIFICATE-----\nfake\n-----END CERTIFICATE-----\n"), 0644)

	return dir
}

// --- EnvVars ---

func TestIntegration_EnvVars_ApplyAndRemove(t *testing.T) {
	home := setupFakeHome(t)
	bashrc := filepath.Join(home, ".bashrc")
	certPath := filepath.Join(home, ".ezproxy", "corp-ca.pem")
	cfg := testConfig(certPath)

	e := &EnvVars{profiles: []string{bashrc}}

	// Apply
	if err := e.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(bashrc)
	got := string(data)

	// Verify proxy vars
	for _, want := range []string{
		"HTTP_PROXY=http://proxy.corp.com:8080",
		"HTTPS_PROXY=http://proxy.corp.com:8080",
		"http_proxy=http://proxy.corp.com:8080",
		"NO_PROXY=localhost,127.0.0.1,.corp.com,10.0.0.0/8",
		"SSL_CERT_FILE=" + certPath,
		"REQUESTS_CA_BUNDLE=" + certPath,
		"NODE_EXTRA_CA_CERTS=" + certPath,
		"HOMEBREW_CURLRC=1",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in bashrc", want)
		}
	}

	// Verify existing content preserved
	if !strings.Contains(got, "# user bashrc") {
		t.Error("existing content should be preserved")
	}

	// Status should be configured
	status, _ := e.Status(cfg)
	if status != "configured" {
		t.Errorf("expected 'configured', got %q", status)
	}

	// Remove
	if err := e.Remove(); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	data, _ = os.ReadFile(bashrc)
	got = string(data)
	if strings.Contains(got, "HTTP_PROXY") {
		t.Error("proxy vars should be removed")
	}
	if !strings.Contains(got, "# user bashrc") {
		t.Error("existing content should remain after remove")
	}

	status, _ = e.Status(cfg)
	if status != "not configured" {
		t.Errorf("expected 'not configured', got %q", status)
	}
}

// --- Pip ---

func TestIntegration_Pip_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pip.conf")
	certPath := "/tmp/corp-ca.pem"
	cfg := testConfig(certPath)

	p := &Pip{path: path}

	if err := p.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "[global]")
	assertContains(t, got, "proxy = http://proxy.corp.com:8080")
	assertContains(t, got, "cert = /tmp/corp-ca.pem")

	status, _ := p.Status(cfg)
	assertEqual(t, "configured", status)

	// Remove
	p.Remove()
	status, _ = p.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Npm ---

func TestIntegration_Npm_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".npmrc")
	certPath := "/tmp/corp-ca.pem"
	cfg := testConfig(certPath)

	n := &Npm{path: path}

	if err := n.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "proxy=http://proxy.corp.com:8080")
	assertContains(t, got, "https-proxy=http://proxy.corp.com:8080")
	assertContains(t, got, "cafile=/tmp/corp-ca.pem")

	status, _ := n.Status(cfg)
	assertEqual(t, "configured", status)

	n.Remove()
	status, _ = n.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Curl ---

func TestIntegration_Curl_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".curlrc")
	certPath := "/tmp/corp-ca.pem"
	cfg := testConfig(certPath)

	c := &Curl{path: path}

	if err := c.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, `proxy = "http://proxy.corp.com:8080"`)
	assertContains(t, got, `cacert = "/tmp/corp-ca.pem"`)

	status, _ := c.Status(cfg)
	assertEqual(t, "configured", status)

	c.Remove()
	status, _ = c.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Wget ---

func TestIntegration_Wget_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".wgetrc")
	certPath := "/tmp/corp-ca.pem"
	cfg := testConfig(certPath)

	w := &Wget{path: path}

	if err := w.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "http_proxy = http://proxy.corp.com:8080")
	assertContains(t, got, "https_proxy = http://proxy.corp.com:8080")
	assertContains(t, got, "ca_certificate = /tmp/corp-ca.pem")

	status, _ := w.Status(cfg)
	assertEqual(t, "configured", status)

	w.Remove()
	status, _ = w.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Cargo ---

func TestIntegration_Cargo_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	certPath := "/tmp/corp-ca.pem"
	cfg := testConfig(certPath)

	c := &Cargo{path: path}

	if err := c.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "[http]")
	assertContains(t, got, `proxy = "http://proxy.corp.com:8080"`)
	assertContains(t, got, `cainfo = "/tmp/corp-ca.pem"`)

	status, _ := c.Status(cfg)
	assertEqual(t, "configured", status)

	c.Remove()
	status, _ = c.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Conda ---

func TestIntegration_Conda_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".condarc")
	certPath := "/tmp/corp-ca.pem"
	cfg := testConfig(certPath)

	c := &Conda{path: path}

	if err := c.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "proxy_servers:")
	assertContains(t, got, "http: http://proxy.corp.com:8080")
	assertContains(t, got, "https: http://proxy.corp.com:8080")
	assertContains(t, got, "ssl_verify: /tmp/corp-ca.pem")

	status, _ := c.Status(cfg)
	assertEqual(t, "configured", status)

	c.Remove()
	status, _ = c.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Yarn (v1 path) ---

func TestIntegration_Yarn_V1_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	v1Path := filepath.Join(dir, ".yarnrc")
	v2Path := filepath.Join(dir, ".yarnrc.yml")
	certPath := "/tmp/corp-ca.pem"
	cfg := testConfig(certPath)

	// Use v1 paths — isV2OrLater will return false since yarn isn't installed
	y := &Yarn{v1Path: v1Path, v2Path: v2Path}

	if err := y.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(v1Path)
	got := string(data)
	assertContains(t, got, `proxy "http://proxy.corp.com:8080"`)
	assertContains(t, got, `https-proxy "http://proxy.corp.com:8080"`)
	assertContains(t, got, `cafile "/tmp/corp-ca.pem"`)

	status, _ := y.Status(cfg)
	assertEqual(t, "configured", status)

	y.Remove()
	status, _ = y.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Docker ---

func TestIntegration_Docker_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".docker", "config.json")
	cfg := testConfigNoCert()

	d := &Docker{configPath: path}

	if err := d.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	var dockerConfig map[string]interface{}
	if err := json.Unmarshal(data, &dockerConfig); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	proxies, ok := dockerConfig["proxies"].(map[string]interface{})
	if !ok {
		t.Fatal("missing proxies key")
	}
	def, ok := proxies["default"].(map[string]interface{})
	if !ok {
		t.Fatal("missing proxies.default key")
	}
	assertEqual(t, "http://proxy.corp.com:8080", def["httpProxy"].(string))
	assertEqual(t, "http://proxy.corp.com:8080", def["httpsProxy"].(string))

	status, _ := d.Status(cfg)
	assertEqual(t, "configured", status)

	d.Remove()
	status, _ = d.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Docker preserves existing keys ---

func TestIntegration_Docker_PreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".docker", "config.json")
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(`{"credsStore": "desktop", "auths": {"ghcr.io": {}}}`), 0644)

	cfg := testConfigNoCert()
	d := &Docker{configPath: path}

	d.Apply(cfg)

	data, _ := os.ReadFile(path)
	var dockerConfig map[string]interface{}
	json.Unmarshal(data, &dockerConfig)

	if _, ok := dockerConfig["credsStore"]; !ok {
		t.Error("existing credsStore key should be preserved")
	}
	if _, ok := dockerConfig["auths"]; !ok {
		t.Error("existing auths key should be preserved")
	}
	if _, ok := dockerConfig["proxies"]; !ok {
		t.Error("proxies key should be added")
	}
}

// --- SSH ---

func TestIntegration_SSH_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".ssh", "config")
	cfg := testConfigNoCert()

	s := &SSH{path: path}

	if err := s.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "Host *")
	assertContains(t, got, "ProxyCommand nc -X connect -x proxy.corp.com:8080")

	status, _ := s.Status(cfg)
	assertEqual(t, "configured", status)

	s.Remove()
	status, _ = s.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Gradle ---

func TestIntegration_Gradle_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gradle.properties")
	cfg := testConfigNoCert()

	g := &Gradle{path: path}

	if err := g.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "systemProp.http.proxyHost=proxy.corp.com")
	assertContains(t, got, "systemProp.http.proxyPort=8080")
	assertContains(t, got, "systemProp.https.proxyHost=proxy.corp.com")
	assertContains(t, got, "systemProp.https.proxyPort=8080")
	assertContains(t, got, "systemProp.http.nonProxyHosts=localhost|127.0.0.1|*.corp.com")

	status, _ := g.Status(cfg)
	assertEqual(t, "configured", status)

	g.Remove()
	status, _ = g.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Maven ---

func TestIntegration_Maven_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.xml")
	cfg := testConfigNoCert()

	m := &Maven{path: path}

	if err := m.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "ezproxy-http")
	assertContains(t, got, "ezproxy-https")
	assertContains(t, got, "<host>proxy.corp.com</host>")
	assertContains(t, got, "<port>8080</port>")
	assertContains(t, got, "<protocol>http</protocol>")
	assertContains(t, got, "<protocol>https</protocol>")

	status, _ := m.Status(cfg)
	assertEqual(t, "configured", status)

	m.Remove()
	status, _ = m.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Maven preserves user proxies ---

func TestIntegration_Maven_PreservesUserProxies(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.xml")
	cfg := testConfigNoCert()

	// Write existing settings with a user proxy
	existing := `<?xml version="1.0" encoding="UTF-8"?>
<settings>
  <proxies>
    <proxy>
      <id>my-custom-proxy</id>
      <active>true</active>
      <protocol>http</protocol>
      <host>myproxy.local</host>
      <port>3128</port>
    </proxy>
  </proxies>
</settings>`
	os.WriteFile(path, []byte(existing), 0644)

	m := &Maven{path: path}
	m.Apply(cfg)

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "my-custom-proxy")
	assertContains(t, got, "ezproxy-http")
}

// --- Podman ---

func TestIntegration_Podman_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "containers.conf")
	cfg := testConfigNoCert()

	p := &Podman{path: path}

	if err := p.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "[containers]")
	assertContains(t, got, `"http_proxy=http://proxy.corp.com:8080"`)
	assertContains(t, got, `"HTTP_PROXY=http://proxy.corp.com:8080"`)
	assertContains(t, got, `"NO_PROXY=localhost,127.0.0.1,.corp.com,10.0.0.0/8"`)

	status, _ := p.Status(cfg)
	assertEqual(t, "configured", status)

	p.Remove()
	status, _ = p.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Bundler ---

func TestIntegration_Bundler_ApplyAndRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	certPath := "/tmp/corp-ca.pem"
	cfg := testConfig(certPath)

	b := &Bundler{path: path}

	if err := b.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "BUNDLE_SSL_CA_CERT")
	assertContains(t, got, certPath)

	status, _ := b.Status(cfg)
	assertEqual(t, "configured", status)

	b.Remove()
	status, _ = b.Status(cfg)
	assertEqual(t, "not configured", status)
}

// --- Bundler preserves existing keys ---

func TestIntegration_Bundler_PreservesExistingKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	os.WriteFile(path, []byte("---\nBUNDLE_GEMFILE: \"Gemfile.custom\"\n"), 0644)

	certPath := "/tmp/corp-ca.pem"
	cfg := testConfig(certPath)
	b := &Bundler{path: path}

	b.Apply(cfg)

	data, _ := os.ReadFile(path)
	got := string(data)
	assertContains(t, got, "BUNDLE_GEMFILE")
	assertContains(t, got, "BUNDLE_SSL_CA_CERT")
}

// --- Bundler no-op without cert ---

func TestIntegration_Bundler_NoCert(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	cfg := testConfigNoCert()

	b := &Bundler{path: path}
	b.Apply(cfg)

	// File should not be created
	if _, err := os.Stat(path); err == nil {
		t.Error("bundler config should not be created without a cert")
	}
}

// --- Idempotency: apply twice, only one marker block ---

func TestIntegration_Idempotency(t *testing.T) {
	tests := []struct {
		name string
		fn   func(dir string) (Configurator, string)
	}{
		{"pip", func(dir string) (Configurator, string) {
			p := filepath.Join(dir, "pip.conf")
			return &Pip{path: p}, p
		}},
		{"npm", func(dir string) (Configurator, string) {
			p := filepath.Join(dir, ".npmrc")
			return &Npm{path: p}, p
		}},
		{"curl", func(dir string) (Configurator, string) {
			p := filepath.Join(dir, ".curlrc")
			return &Curl{path: p}, p
		}},
		{"wget", func(dir string) (Configurator, string) {
			p := filepath.Join(dir, ".wgetrc")
			return &Wget{path: p}, p
		}},
		{"cargo", func(dir string) (Configurator, string) {
			p := filepath.Join(dir, "config.toml")
			return &Cargo{path: p}, p
		}},
		{"conda", func(dir string) (Configurator, string) {
			p := filepath.Join(dir, ".condarc")
			return &Conda{path: p}, p
		}},
		{"gradle", func(dir string) (Configurator, string) {
			p := filepath.Join(dir, "gradle.properties")
			return &Gradle{path: p}, p
		}},
		{"ssh", func(dir string) (Configurator, string) {
			p := filepath.Join(dir, "ssh_config")
			return &SSH{path: p}, p
		}},
	}

	cfg := testConfig("/tmp/corp-ca.pem")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			c, path := tt.fn(dir)

			// Apply twice
			c.Apply(cfg)
			c.Apply(cfg)

			data, _ := os.ReadFile(path)
			got := string(data)
			count := strings.Count(got, ">>> ezproxy >>>")
			if count != 1 {
				t.Errorf("expected 1 marker block, got %d", count)
			}
		})
	}
}

// --- Config update: re-apply with different proxy URL ---

func TestIntegration_ConfigUpdate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".npmrc")

	cfg1 := &config.Config{
		Proxy: config.ProxyConfig{
			HTTP:  "http://old-proxy:3128",
			HTTPS: "http://old-proxy:3128",
		},
	}
	cfg2 := &config.Config{
		Proxy: config.ProxyConfig{
			HTTP:  "http://new-proxy:8080",
			HTTPS: "http://new-proxy:8080",
		},
	}

	n := &Npm{path: path}

	n.Apply(cfg1)
	data, _ := os.ReadFile(path)
	assertContains(t, string(data), "old-proxy:3128")

	n.Apply(cfg2)
	data, _ = os.ReadFile(path)
	got := string(data)
	assertNotContains(t, got, "old-proxy:3128")
	assertContains(t, got, "new-proxy:8080")
	if strings.Count(got, ">>> ezproxy >>>") != 1 {
		t.Error("should still have exactly one marker block")
	}
}

// --- DryRun: files should not be modified ---

func TestIntegration_DryRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".npmrc")
	cfg := testConfigNoCert()

	// Enable dry run
	fileutil.DryRun = true
	defer func() { fileutil.DryRun = false }()

	n := &Npm{path: path}
	n.Apply(cfg)

	// File should not exist
	if _, err := os.Stat(path); err == nil {
		t.Error("file should not be created in dry-run mode")
	}
}

// --- Full lifecycle: apply all file-based tools, verify, then remove all ---

func TestIntegration_FullLifecycle(t *testing.T) {
	dir := t.TempDir()
	certPath := "/tmp/corp-ca.pem"
	cfg := testConfig(certPath)

	bashrc := filepath.Join(dir, ".bashrc")
	os.WriteFile(bashrc, []byte("# existing\n"), 0644)

	// Create all file-based configurators with injectable paths
	configurators := map[string]Configurator{
		"env_vars": &EnvVars{profiles: []string{bashrc}},
		"pip":      &Pip{path: filepath.Join(dir, "pip.conf")},
		"npm":      &Npm{path: filepath.Join(dir, ".npmrc")},
		"curl":     &Curl{path: filepath.Join(dir, ".curlrc")},
		"wget":     &Wget{path: filepath.Join(dir, ".wgetrc")},
		"cargo":    &Cargo{path: filepath.Join(dir, "config.toml")},
		"conda":    &Conda{path: filepath.Join(dir, ".condarc")},
		"yarn":     &Yarn{v1Path: filepath.Join(dir, ".yarnrc"), v2Path: filepath.Join(dir, ".yarnrc.yml")},
		"docker":   &Docker{configPath: filepath.Join(dir, "docker-config.json")},
		"ssh":      &SSH{path: filepath.Join(dir, "ssh_config")},
		"gradle":   &Gradle{path: filepath.Join(dir, "gradle.properties")},
		"maven":    &Maven{path: filepath.Join(dir, "settings.xml")},
		"podman":   &Podman{path: filepath.Join(dir, "containers.conf")},
		"bundler":  &Bundler{path: filepath.Join(dir, "bundle-config")},
	}

	// Phase 1: Apply all
	for name, c := range configurators {
		if err := c.Apply(cfg); err != nil {
			t.Errorf("Apply %s: %v", name, err)
		}
	}

	// Phase 2: Verify all configured
	for name, c := range configurators {
		status, _ := c.Status(cfg)
		if status != "configured" {
			t.Errorf("%s: expected 'configured', got %q", name, status)
		}
	}

	// Phase 3: Remove all
	for name, c := range configurators {
		if err := c.Remove(); err != nil {
			t.Errorf("Remove %s: %v", name, err)
		}
	}

	// Phase 4: Verify all not configured
	for name, c := range configurators {
		status, _ := c.Status(cfg)
		if status != "not configured" {
			t.Errorf("%s: expected 'not configured' after remove, got %q", name, status)
		}
	}

	// Phase 5: Verify existing content preserved
	data, _ := os.ReadFile(bashrc)
	if !strings.Contains(string(data), "# existing") {
		t.Error("existing bashrc content should be preserved after full lifecycle")
	}
}

// --- Helpers ---

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("expected to contain %q, got:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, notWant string) {
	t.Helper()
	if strings.Contains(got, notWant) {
		t.Errorf("expected NOT to contain %q, got:\n%s", notWant, got)
	}
}

func assertEqual(t *testing.T, want, got string) {
	t.Helper()
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
