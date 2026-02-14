# ezproxy Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI tool that configures corporate proxy settings across all common dev tools in one command.

**Architecture:** Declarative config file (`~/.ezproxy/config.yaml`) + apply/remove pattern. Each tool has a Configurator implementation. Marker blocks for idempotent file modifications.

**Tech Stack:** Go 1.22+, `gopkg.in/yaml.v3` for config, no CLI framework (stdlib `flag` or simple arg parsing to keep it lean).

---

## Phase 1: Foundation

### Task 1: Initialize Go module and project skeleton

**Files:**
- Create: `go.mod`
- Create: `cmd/ezproxy/main.go`
- Create: `internal/config/config.go`

**Step 1: Initialize Go module**

Run: `go mod init github.com/andrew/ezproxy`

**Step 2: Create minimal main.go**

```go
// cmd/ezproxy/main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ezproxy <init|apply|remove|status>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		fmt.Println("init: not yet implemented")
	case "apply":
		fmt.Println("apply: not yet implemented")
	case "remove":
		fmt.Println("remove: not yet implemented")
	case "status":
		fmt.Println("status: not yet implemented")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
```

**Step 3: Verify it builds**

Run: `go build -o ezproxy ./cmd/ezproxy && ./ezproxy`
Expected: `Usage: ezproxy <init|apply|remove|status>`

**Step 4: Commit**

```bash
git add go.mod cmd/
git commit -m "feat: initialize Go module with CLI skeleton"
```

---

### Task 2: Config loading and saving

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write the failing test**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	yaml := `proxy:
  http: http://proxy.corp.com:8080
  https: http://proxy.corp.com:8080
  no_proxy: localhost,127.0.0.1,.corp.com

ca_cert: /path/to/cert.pem

tools:
  env_vars: true
  git: true
  pip: false
`
	os.WriteFile(configPath, []byte(yaml), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Proxy.HTTP != "http://proxy.corp.com:8080" {
		t.Errorf("got HTTP proxy %q", cfg.Proxy.HTTP)
	}
	if cfg.Proxy.NoProxy != "localhost,127.0.0.1,.corp.com" {
		t.Errorf("got NoProxy %q", cfg.Proxy.NoProxy)
	}
	if cfg.CACert != "/path/to/cert.pem" {
		t.Errorf("got CACert %q", cfg.CACert)
	}
	if !cfg.Tools["env_vars"] {
		t.Error("expected env_vars=true")
	}
	if !cfg.Tools["git"] {
		t.Error("expected git=true")
	}
	if cfg.Tools["pip"] {
		t.Error("expected pip=false")
	}
}

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		Proxy: ProxyConfig{
			HTTP:    "http://proxy:8080",
			HTTPS:   "http://proxy:8080",
			NoProxy: "localhost",
		},
		CACert: "/tmp/ca.pem",
		Tools:  map[string]bool{"git": true, "pip": false},
	}

	err := Save(configPath, cfg)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}
	if loaded.Proxy.HTTP != cfg.Proxy.HTTP {
		t.Errorf("round-trip failed: got %q", loaded.Proxy.HTTP)
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	result := ExpandPath("~/foo/bar")
	expected := filepath.Join(home, "foo/bar")
	if result != expected {
		t.Errorf("ExpandPath: got %q, want %q", result, expected)
	}

	result2 := ExpandPath("/absolute/path")
	if result2 != "/absolute/path" {
		t.Errorf("ExpandPath absolute: got %q", result2)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -v`
Expected: FAIL (types not defined)

**Step 3: Write implementation**

```go
// internal/config/config.go
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

// DefaultTools returns all tool names with default enabled=true.
func DefaultTools() map[string]bool {
	return map[string]bool{
		"env_vars":  true,
		"git":       true,
		"pip":       true,
		"npm":       true,
		"yarn":      true,
		"docker":    true,
		"curl":      true,
		"wget":      true,
		"go":        true,
		"cargo":     true,
		"conda":     true,
		"brew":      true,
		"snap":      true,
		"apt":       true,
		"yum":       true,
		"ssh":       true,
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

// ExpandPath replaces ~ with the user's home directory.
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

// CACertAbsPath returns the absolute path to the CA cert.
func (c *Config) CACertAbsPath() string {
	return ExpandPath(c.CACert)
}
```

**Step 4: Add yaml dependency and run tests**

Run: `cd /Users/andrew/Projects/ezproxy && go get gopkg.in/yaml.v3 && go test ./internal/config/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/ go.mod go.sum
git commit -m "feat: add config loading, saving, and path expansion"
```

---

### Task 3: File utility - marker block management

**Files:**
- Create: `internal/fileutil/fileutil.go`
- Create: `internal/fileutil/fileutil_test.go`

**Step 1: Write the failing tests**

```go
// internal/fileutil/fileutil_test.go
package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpsertMarkerBlock_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	content := "export FOO=bar\nexport BAZ=qux\n"
	err := UpsertMarkerBlock(path, content, "#")
	if err != nil {
		t.Fatalf("UpsertMarkerBlock failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	if !strings.Contains(got, "# >>> ezproxy >>>") {
		t.Error("missing start marker")
	}
	if !strings.Contains(got, "# <<< ezproxy <<<") {
		t.Error("missing end marker")
	}
	if !strings.Contains(got, "export FOO=bar") {
		t.Error("missing content")
	}
}

func TestUpsertMarkerBlock_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	existing := "# existing stuff\nPATH=/usr/bin\n"
	os.WriteFile(path, []byte(existing), 0644)

	content := "export PROXY=http://proxy:8080\n"
	err := UpsertMarkerBlock(path, content, "#")
	if err != nil {
		t.Fatalf("UpsertMarkerBlock failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	if !strings.HasPrefix(got, "# existing stuff\n") {
		t.Error("existing content should be preserved at start")
	}
	if !strings.Contains(got, "export PROXY=http://proxy:8080") {
		t.Error("new content missing")
	}
}

func TestUpsertMarkerBlock_Replace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	// First insert
	UpsertMarkerBlock(path, "OLD CONTENT\n", "#")
	// Replace
	UpsertMarkerBlock(path, "NEW CONTENT\n", "#")

	data, _ := os.ReadFile(path)
	got := string(data)
	if strings.Contains(got, "OLD CONTENT") {
		t.Error("old content should be replaced")
	}
	if !strings.Contains(got, "NEW CONTENT") {
		t.Error("new content missing")
	}
	// Should only have one marker pair
	if strings.Count(got, ">>> ezproxy >>>") != 1 {
		t.Error("should have exactly one start marker")
	}
}

func TestRemoveMarkerBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	before := "line1\n"
	after := "line2\n"
	os.WriteFile(path, []byte(before), 0644)

	UpsertMarkerBlock(path, "PROXY STUFF\n", "#")

	// Append something after the block
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString(after)
	f.Close()

	err := RemoveMarkerBlock(path, "#")
	if err != nil {
		t.Fatalf("RemoveMarkerBlock failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	if strings.Contains(got, "ezproxy") {
		t.Error("marker block should be removed")
	}
	if !strings.Contains(got, "line1") {
		t.Error("content before block should remain")
	}
	if !strings.Contains(got, "line2") {
		t.Error("content after block should remain")
	}
}

func TestHasMarkerBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	os.WriteFile(path, []byte("nothing here\n"), 0644)
	if HasMarkerBlock(path, "#") {
		t.Error("should not have marker block")
	}

	UpsertMarkerBlock(path, "stuff\n", "#")
	if !HasMarkerBlock(path, "#") {
		t.Error("should have marker block")
	}
}

func TestGetMarkerBlockContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	UpsertMarkerBlock(path, "FOO=bar\nBAZ=qux\n", "#")
	content, err := GetMarkerBlockContent(path, "#")
	if err != nil {
		t.Fatalf("GetMarkerBlockContent failed: %v", err)
	}
	if !strings.Contains(content, "FOO=bar") {
		t.Error("missing content")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/fileutil/ -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/fileutil/fileutil.go
package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	markerStart = ">>> ezproxy >>>"
	markerEnd   = "<<< ezproxy <<<"
)

func startMarker(comment string) string {
	return fmt.Sprintf("%s %s", comment, markerStart)
}

func endMarker(comment string) string {
	return fmt.Sprintf("%s %s", comment, markerEnd)
}

// UpsertMarkerBlock inserts or replaces the ezproxy marker block in a file.
// If the file doesn't exist, it creates it. The comment parameter is the
// comment prefix for the file type (e.g., "#" for shell, "//" for TOML).
func UpsertMarkerBlock(path string, content string, comment string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	existing := ""
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	}

	block := fmt.Sprintf("%s\n%s%s\n", startMarker(comment), content, endMarker(comment))

	start := startMarker(comment)
	end := endMarker(comment)

	startIdx := strings.Index(existing, start)
	endIdx := strings.Index(existing, end)

	var result string
	if startIdx >= 0 && endIdx >= 0 {
		// Replace existing block
		result = existing[:startIdx] + block + existing[endIdx+len(end):]
		// Clean up extra newlines
		if strings.HasSuffix(result, "\n\n\n") {
			result = strings.TrimRight(result, "\n") + "\n"
		}
	} else {
		// Append new block
		if existing != "" && !strings.HasSuffix(existing, "\n") {
			existing += "\n"
		}
		if existing != "" {
			existing += "\n"
		}
		result = existing + block
	}

	return os.WriteFile(path, []byte(result), 0644)
}

// RemoveMarkerBlock removes the ezproxy marker block from a file.
func RemoveMarkerBlock(path string, comment string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	content := string(data)
	start := startMarker(comment)
	end := endMarker(comment)

	startIdx := strings.Index(content, start)
	endIdx := strings.Index(content, end)

	if startIdx < 0 || endIdx < 0 {
		return nil // No block to remove
	}

	// Remove the block and any surrounding blank lines
	before := content[:startIdx]
	after := content[endIdx+len(end):]

	// Trim trailing newline from before and leading newline from after
	before = strings.TrimRight(before, "\n")
	after = strings.TrimLeft(after, "\n")

	result := ""
	if before != "" && after != "" {
		result = before + "\n" + after + "\n"
	} else if before != "" {
		result = before + "\n"
	} else if after != "" {
		result = after + "\n"
	}

	return os.WriteFile(path, []byte(result), 0644)
}

// HasMarkerBlock checks if the file contains an ezproxy marker block.
func HasMarkerBlock(path string, comment string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), startMarker(comment))
}

// GetMarkerBlockContent returns the content inside the marker block.
func GetMarkerBlockContent(path string, comment string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content := string(data)
	start := startMarker(comment)
	end := endMarker(comment)

	startIdx := strings.Index(content, start)
	endIdx := strings.Index(content, end)

	if startIdx < 0 || endIdx < 0 {
		return "", fmt.Errorf("no marker block found in %s", path)
	}

	blockContent := content[startIdx+len(start)+1 : endIdx]
	return blockContent, nil
}
```

**Step 4: Run tests**

Run: `go test ./internal/fileutil/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/fileutil/
git commit -m "feat: add marker block file utilities for idempotent config"
```

---

### Task 4: OS and tool detection

**Files:**
- Create: `internal/detect/detect.go`
- Create: `internal/detect/detect_test.go`

**Step 1: Write the failing test**

```go
// internal/detect/detect_test.go
package detect

import (
	"runtime"
	"testing"
)

func TestDetectOS(t *testing.T) {
	info := DetectOS()
	if runtime.GOOS == "darwin" {
		if info.OS != "darwin" {
			t.Errorf("expected darwin, got %s", info.OS)
		}
	} else if runtime.GOOS == "linux" {
		if info.OS != "linux" {
			t.Errorf("expected linux, got %s", info.OS)
		}
	}
}

func TestIsCommandAvailable(t *testing.T) {
	// "ls" should always be available
	if !IsCommandAvailable("ls") {
		t.Error("ls should be available")
	}
	// made-up command should not
	if IsCommandAvailable("ezproxy_nonexistent_command_xyz") {
		t.Error("fake command should not be available")
	}
}

func TestShellProfiles(t *testing.T) {
	profiles := ShellProfiles()
	if len(profiles) == 0 {
		t.Error("should find at least one shell profile")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/detect/ -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/detect/detect.go
package detect

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type OSInfo struct {
	OS     string // "darwin" or "linux"
	Distro string // "debian", "ubuntu", "fedora", "rhel", "centos", "arch", "" (macOS)
}

func DetectOS() OSInfo {
	info := OSInfo{OS: runtime.GOOS}
	if runtime.GOOS == "linux" {
		info.Distro = detectLinuxDistro()
	}
	return info
}

func detectLinuxDistro() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	content := strings.ToLower(string(data))
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "id=") {
			id := strings.Trim(strings.TrimPrefix(line, "id="), "\"")
			return id
		}
	}
	return ""
}

// IsDebian returns true for Debian-family distros.
func (o OSInfo) IsDebian() bool {
	return o.Distro == "debian" || o.Distro == "ubuntu" || o.Distro == "pop" || o.Distro == "mint"
}

// IsRHEL returns true for Red Hat-family distros.
func (o OSInfo) IsRHEL() bool {
	return o.Distro == "fedora" || o.Distro == "rhel" || o.Distro == "centos" || o.Distro == "rocky" || o.Distro == "alma"
}

// IsArch returns true for Arch-based distros.
func (o OSInfo) IsArch() bool {
	return o.Distro == "arch" || o.Distro == "manjaro" || o.Distro == "endeavouros"
}

// IsCommandAvailable checks if a command exists in PATH.
func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// ShellProfiles returns paths to existing shell profile files.
func ShellProfiles() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".profile"),
	}

	var profiles []string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			profiles = append(profiles, p)
		}
	}
	return profiles
}
```

**Step 4: Run tests**

Run: `go test ./internal/detect/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/detect/
git commit -m "feat: add OS/distro detection and tool availability checks"
```

---

## Phase 2: Configurator Interface + First Configurators

### Task 5: Configurator interface and registry

**Files:**
- Create: `internal/configurator/configurator.go`

**Step 1: Write the interface and registry**

```go
// internal/configurator/configurator.go
package configurator

import (
	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
)

// Configurator is the interface all tool configurators implement.
type Configurator interface {
	// Name returns the tool name (matches key in config.yaml tools map).
	Name() string
	// IsAvailable returns true if the tool is installed/relevant on this system.
	IsAvailable(osInfo detect.OSInfo) bool
	// Apply writes proxy configuration for this tool.
	Apply(cfg *config.Config) error
	// Remove undoes proxy configuration for this tool.
	Remove() error
	// Status returns "configured", "not configured", or "stale".
	Status(cfg *config.Config) (string, error)
}

// All returns all registered configurators in apply order.
func All() []Configurator {
	return []Configurator{
		&SystemCA{},
		&EnvVars{},
		&Git{},
		&Pip{},
		&Npm{},
		&Yarn{},
		&Docker{},
		&Curl{},
		&Wget{},
		&Cargo{},
		&Conda{},
		&Brew{},
		&Snap{},
		&Apt{},
		&Yum{},
		&SSH{},
	}
}
```

**Step 2: Commit**

```bash
git add internal/configurator/configurator.go
git commit -m "feat: add Configurator interface and registry"
```

---

### Task 6: EnvVars configurator

**Files:**
- Create: `internal/configurator/envvars.go`
- Create: `internal/configurator/envvars_test.go`

**Step 1: Write failing test**

```go
// internal/configurator/envvars_test.go
package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestEnvVarsApply(t *testing.T) {
	dir := t.TempDir()
	bashrc := filepath.Join(dir, ".bashrc")
	os.WriteFile(bashrc, []byte("# existing\n"), 0644)

	e := &EnvVars{profiles: []string{bashrc}}
	cfg := &config.Config{
		Proxy: config.ProxyConfig{
			HTTP:    "http://proxy:8080",
			HTTPS:   "http://proxy:8080",
			NoProxy: "localhost",
		},
		CACert: "/tmp/ca.pem",
	}

	if err := e.Apply(cfg); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	data, _ := os.ReadFile(bashrc)
	got := string(data)
	if !strings.Contains(got, "HTTP_PROXY=http://proxy:8080") {
		t.Error("missing HTTP_PROXY")
	}
	if !strings.Contains(got, "http_proxy=http://proxy:8080") {
		t.Error("missing lowercase http_proxy")
	}
	if !strings.Contains(got, "NODE_EXTRA_CA_CERTS=/tmp/ca.pem") {
		t.Error("missing NODE_EXTRA_CA_CERTS")
	}
	if !strings.Contains(got, "HOMEBREW_CURLRC=1") {
		t.Error("missing HOMEBREW_CURLRC")
	}
	if !strings.Contains(got, "# existing") {
		t.Error("existing content should be preserved")
	}
}

func TestEnvVarsRemove(t *testing.T) {
	dir := t.TempDir()
	bashrc := filepath.Join(dir, ".bashrc")
	os.WriteFile(bashrc, []byte("# existing\n"), 0644)

	e := &EnvVars{profiles: []string{bashrc}}
	cfg := &config.Config{
		Proxy: config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080", NoProxy: "localhost"},
		CACert: "/tmp/ca.pem",
	}
	e.Apply(cfg)
	e.Remove()

	data, _ := os.ReadFile(bashrc)
	got := string(data)
	if strings.Contains(got, "ezproxy") {
		t.Error("marker block should be removed")
	}
	if !strings.Contains(got, "# existing") {
		t.Error("existing content should remain")
	}
}

func TestEnvVarsApplyIdempotent(t *testing.T) {
	dir := t.TempDir()
	bashrc := filepath.Join(dir, ".bashrc")
	os.WriteFile(bashrc, []byte(""), 0644)

	e := &EnvVars{profiles: []string{bashrc}}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080", NoProxy: "localhost"},
		CACert: "/tmp/ca.pem",
	}

	e.Apply(cfg)
	e.Apply(cfg) // second apply

	data, _ := os.ReadFile(bashrc)
	got := string(data)
	if strings.Count(got, ">>> ezproxy >>>") != 1 {
		t.Error("should have exactly one marker block after double apply")
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/configurator/ -run TestEnvVars -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/configurator/envvars.go
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
```

**Step 4: Run tests**

Run: `go test ./internal/configurator/ -run TestEnvVars -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/configurator/envvars.go internal/configurator/envvars_test.go
git commit -m "feat: add EnvVars configurator for shell profile proxy exports"
```

---

### Task 7: Git configurator

**Files:**
- Create: `internal/configurator/git.go`
- Create: `internal/configurator/git_test.go`

This uses `git config --global` commands. Test by setting a custom `GIT_CONFIG_GLOBAL` env var pointing to a temp file.

**Step 1: Write failing test**

```go
// internal/configurator/git_test.go
package configurator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
)

func TestGitApplyAndRemove(t *testing.T) {
	if !detect.IsCommandAvailable("git") {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	gitconfig := filepath.Join(dir, ".gitconfig")
	os.WriteFile(gitconfig, []byte(""), 0644)
	t.Setenv("GIT_CONFIG_GLOBAL", gitconfig)

	g := &Git{}
	cfg := &config.Config{
		Proxy: config.ProxyConfig{HTTP: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}

	if err := g.Apply(cfg); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Verify proxy was set
	out, _ := exec.Command("git", "config", "--global", "http.proxy").Output()
	if strings.TrimSpace(string(out)) != "http://proxy:8080" {
		t.Errorf("http.proxy = %q", strings.TrimSpace(string(out)))
	}

	// Verify CA was set
	out, _ = exec.Command("git", "config", "--global", "http.sslCAInfo").Output()
	if strings.TrimSpace(string(out)) != "/tmp/ca.pem" {
		t.Errorf("http.sslCAInfo = %q", strings.TrimSpace(string(out)))
	}

	// Remove
	if err := g.Remove(); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	out, err := exec.Command("git", "config", "--global", "http.proxy").Output()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		t.Error("http.proxy should be unset after Remove")
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/configurator/ -run TestGit -v`
Expected: FAIL

**Step 3: Write implementation**

```go
// internal/configurator/git.go
package configurator

import (
	"os/exec"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
)

type Git struct{}

func (g *Git) Name() string { return "git" }

func (g *Git) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("git")
}

func (g *Git) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)

	cmds := [][]string{
		{"git", "config", "--global", "http.proxy", cfg.Proxy.HTTP},
	}
	if certPath != "" {
		cmds = append(cmds, []string{"git", "config", "--global", "http.sslCAInfo", certPath})
	}

	for _, args := range cmds {
		if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
			return err
		}
	}
	return nil
}

func (g *Git) Remove() error {
	keys := []string{"http.proxy", "http.sslCAInfo"}
	for _, key := range keys {
		exec.Command("git", "config", "--global", "--unset", key).Run() // ignore errors (key may not exist)
	}
	return nil
}

func (g *Git) Status(cfg *config.Config) (string, error) {
	out, err := exec.Command("git", "config", "--global", "http.proxy").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return "not configured", nil
	}
	current := strings.TrimSpace(string(out))
	if current == cfg.Proxy.HTTP {
		return "configured", nil
	}
	return "stale", nil
}
```

**Step 4: Run tests**

Run: `go test ./internal/configurator/ -run TestGit -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/configurator/git.go internal/configurator/git_test.go
git commit -m "feat: add Git configurator (http.proxy + sslCAInfo)"
```

---

### Task 8: Curl and Wget configurators

**Files:**
- Create: `internal/configurator/curl.go`
- Create: `internal/configurator/wget.go`
- Create: `internal/configurator/curl_test.go`
- Create: `internal/configurator/wget_test.go`

These both use marker blocks in dotfiles. Very similar pattern.

**Step 1: Write failing tests**

```go
// internal/configurator/curl_test.go
package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestCurlApply(t *testing.T) {
	dir := t.TempDir()
	c := &Curl{path: filepath.Join(dir, ".curlrc")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	if err := c.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	data, _ := os.ReadFile(c.path)
	got := string(data)
	if !strings.Contains(got, `proxy = "http://proxy:8080"`) {
		t.Error("missing proxy")
	}
	if !strings.Contains(got, `cacert = "/tmp/ca.pem"`) {
		t.Error("missing cacert")
	}
}

func TestCurlRemove(t *testing.T) {
	dir := t.TempDir()
	c := &Curl{path: filepath.Join(dir, ".curlrc")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	c.Apply(cfg)
	c.Remove()
	data, _ := os.ReadFile(c.path)
	if strings.Contains(string(data), "ezproxy") {
		t.Error("should be cleaned")
	}
}
```

```go
// internal/configurator/wget_test.go
package configurator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrew/ezproxy/internal/config"
)

func TestWgetApply(t *testing.T) {
	dir := t.TempDir()
	w := &Wget{path: filepath.Join(dir, ".wgetrc")}
	cfg := &config.Config{
		Proxy:  config.ProxyConfig{HTTP: "http://proxy:8080", HTTPS: "http://proxy:8080"},
		CACert: "/tmp/ca.pem",
	}
	if err := w.Apply(cfg); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	data, _ := os.ReadFile(w.path)
	got := string(data)
	if !strings.Contains(got, "http_proxy = http://proxy:8080") {
		t.Error("missing http_proxy")
	}
	if !strings.Contains(got, "ca_certificate = /tmp/ca.pem") {
		t.Error("missing ca_certificate")
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/configurator/ -run "TestCurl|TestWget" -v`
Expected: FAIL

**Step 3: Write implementations**

```go
// internal/configurator/curl.go
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

type Curl struct {
	path string // override for testing
}

func (c *Curl) Name() string { return "curl" }

func (c *Curl) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("curl")
}

func (c *Curl) getPath() string {
	if c.path != "" {
		return c.path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".curlrc")
}

func (c *Curl) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	var b strings.Builder
	fmt.Fprintf(&b, "proxy = \"%s\"\n", cfg.Proxy.HTTP)
	if certPath != "" {
		fmt.Fprintf(&b, "cacert = \"%s\"\n", certPath)
	}
	return fileutil.UpsertMarkerBlock(c.getPath(), b.String(), "#")
}

func (c *Curl) Remove() error {
	return fileutil.RemoveMarkerBlock(c.getPath(), "#")
}

func (c *Curl) Status(cfg *config.Config) (string, error) {
	if fileutil.HasMarkerBlock(c.getPath(), "#") {
		return "configured", nil
	}
	return "not configured", nil
}
```

```go
// internal/configurator/wget.go
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

type Wget struct {
	path string
}

func (w *Wget) Name() string { return "wget" }

func (w *Wget) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("wget")
}

func (w *Wget) getPath() string {
	if w.path != "" {
		return w.path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".wgetrc")
}

func (w *Wget) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	var b strings.Builder
	fmt.Fprintf(&b, "http_proxy = %s\n", cfg.Proxy.HTTP)
	fmt.Fprintf(&b, "https_proxy = %s\n", cfg.Proxy.HTTPS)
	if certPath != "" {
		fmt.Fprintf(&b, "ca_certificate = %s\n", certPath)
	}
	return fileutil.UpsertMarkerBlock(w.getPath(), b.String(), "#")
}

func (w *Wget) Remove() error {
	return fileutil.RemoveMarkerBlock(w.getPath(), "#")
}

func (w *Wget) Status(cfg *config.Config) (string, error) {
	if fileutil.HasMarkerBlock(w.getPath(), "#") {
		return "configured", nil
	}
	return "not configured", nil
}
```

**Step 4: Run tests**

Run: `go test ./internal/configurator/ -run "TestCurl|TestWget" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/configurator/curl.go internal/configurator/wget.go internal/configurator/curl_test.go internal/configurator/wget_test.go
git commit -m "feat: add Curl and Wget configurators"
```

---

### Task 9: Pip, Npm, Yarn, Cargo, Conda configurators

These all write to specific config files. Follow the exact same pattern as curl/wget — each has a struct with an overridable path, Apply writes the correct format, Remove cleans it up.

**Files:**
- Create: `internal/configurator/pip.go`
- Create: `internal/configurator/npm.go`
- Create: `internal/configurator/yarn.go`
- Create: `internal/configurator/cargo.go`
- Create: `internal/configurator/conda.go`
- Create: tests for each

Follow the same test/implement/test/commit cycle. Each configurator writes the exact format from the design doc. Key notes:

- **pip**: INI format, path differs per OS (use `runtime.GOOS`). Uses marker comments with `#`.
- **npm**: Key=value `.npmrc`, marker comments with `#`.
- **yarn**: Detect yarn version (`yarn --version`). If v1, write `.yarnrc`. If v2+, write `.yarnrc.yml` in YAML format. Marker comments with `#`.
- **cargo**: TOML `[http]` section in `~/.cargo/config.toml`. Marker comments with `#`.
- **conda**: YAML in `~/.condarc`. Marker comments with `#`.

**Commit after all five are implemented and tested:**

```bash
git add internal/configurator/pip.go internal/configurator/npm.go internal/configurator/yarn.go internal/configurator/cargo.go internal/configurator/conda.go
git add internal/configurator/*_test.go
git commit -m "feat: add Pip, Npm, Yarn, Cargo, Conda configurators"
```

---

### Task 10: Docker configurator

**Files:**
- Create: `internal/configurator/docker.go`
- Create: `internal/configurator/docker_test.go`

This is more complex — reads/writes JSON (`~/.docker/config.json`), merging `proxies` key into existing config. On Linux, also writes systemd override (prints sudo instructions). On macOS, prints Docker Desktop GUI instructions.

Key implementation notes:
- Read existing `config.json`, unmarshal to `map[string]interface{}`, set `proxies.default`, marshal back. This preserves other Docker config.
- For Remove: delete the `proxies` key, write back.
- Daemon config on Linux: write file content, print instructions to run `sudo systemctl daemon-reload && sudo systemctl restart docker`.

**Commit:**

```bash
git add internal/configurator/docker.go internal/configurator/docker_test.go
git commit -m "feat: add Docker configurator (client + daemon + Desktop instructions)"
```

---

### Task 11: System CA, Snap, Apt, Yum, SSH configurators

**Files:**
- Create: `internal/configurator/systemca.go`
- Create: `internal/configurator/snap.go`
- Create: `internal/configurator/apt.go`
- Create: `internal/configurator/yum.go`
- Create: `internal/configurator/ssh.go`
- Create: `internal/configurator/brew.go`
- Create: tests for each

Key notes:
- **systemca**: Detect OS/distro, run the correct command. Print sudo instructions rather than executing sudo directly (let the user confirm).
- **snap**: Run `snap set system` commands. Print sudo instructions.
- **apt**: Write `/etc/apt/apt.conf.d/99ezproxy`. Print sudo instructions.
- **yum**: Write proxy to yum.conf/dnf.conf. Print sudo instructions.
- **ssh**: Append ProxyCommand block to `~/.ssh/config` with markers. Warn if GNU netcat on Linux.
- **brew**: No-op (covered by env_vars), but Status checks if `HOMEBREW_CURLRC` is set.

For tools requiring sudo, the pattern is: generate the file content, write to a temp file, print the sudo command the user needs to run. Example output:
```
[sudo required] To configure apt proxy, run:
  sudo cp /tmp/ezproxy-apt-12345 /etc/apt/apt.conf.d/99ezproxy
```

**Commit:**

```bash
git add internal/configurator/systemca.go internal/configurator/snap.go internal/configurator/apt.go internal/configurator/yum.go internal/configurator/ssh.go internal/configurator/brew.go
git add internal/configurator/*_test.go
git commit -m "feat: add SystemCA, Snap, Apt, Yum, SSH, Brew configurators"
```

---

## Phase 3: CLI Commands

### Task 12: Wire up `apply` and `remove` commands

**Files:**
- Modify: `cmd/ezproxy/main.go`

**Step 1: Implement apply command**

Wire up main.go to:
1. Load config from `~/.ezproxy/config.yaml`
2. Call `detect.DetectOS()`
3. Loop through `configurator.All()`
4. For each: check `cfg.Tools[c.Name()]`, check `c.IsAvailable(osInfo)`, call `c.Apply(cfg)`
5. Print results table

**Step 2: Implement remove command**

Same loop, call `c.Remove()` instead.

**Step 3: Test manually**

Run: `go build -o ezproxy ./cmd/ezproxy && ./ezproxy apply`
Expected: loads config, applies configurators, prints results

**Step 4: Commit**

```bash
git add cmd/ezproxy/main.go
git commit -m "feat: wire up apply and remove CLI commands"
```

---

### Task 13: Wire up `status` command

**Files:**
- Modify: `cmd/ezproxy/main.go`

Print a table like:
```
Tool        Status          Available
────        ──────          ─────────
env_vars    configured      yes
git         configured      yes
pip         not configured  yes
docker      configured      yes
snap        skipped         no (not installed)
```

**Commit:**

```bash
git add cmd/ezproxy/main.go
git commit -m "feat: add status command with tool configuration table"
```

---

### Task 14: Wire up `init` command

**Files:**
- Modify: `cmd/ezproxy/main.go`

Interactive prompts using `fmt.Scan`:
1. Ask for HTTP proxy URL (with default suggestion)
2. Ask for HTTPS proxy URL (default: same as HTTP)
3. Ask for NO_PROXY (with sensible default)
4. Ask for path to CA cert PEM file
5. Copy cert to `~/.ezproxy/corp-ca.pem`
6. Generate `~/.ezproxy/config.yaml` with all tools enabled
7. Print "Run `ezproxy apply` to configure all tools"

**Commit:**

```bash
git add cmd/ezproxy/main.go
git commit -m "feat: add interactive init command"
```

---

## Phase 4: Polish

### Task 15: Run all tests, fix any issues

Run: `go test ./... -v`
Fix any failures.

**Commit:**

```bash
git commit -am "fix: resolve test issues"
```

---

### Task 16: Build and manual integration test

**Step 1: Build**

Run: `go build -o ezproxy ./cmd/ezproxy`

**Step 2: Test init**

Run: `./ezproxy init` and walk through the prompts.

**Step 3: Test apply**

Run: `./ezproxy apply` and verify configs were written.

**Step 4: Test status**

Run: `./ezproxy status` and verify table output.

**Step 5: Test remove**

Run: `./ezproxy remove` and verify configs were cleaned.

**Step 6: Commit any fixes**

```bash
git commit -am "fix: integration test fixes"
```

---

### Task 17: Add cross-compile Makefile

**Files:**
- Create: `Makefile`

```makefile
BINARY=ezproxy
VERSION?=0.1.0

.PHONY: build test clean release

build:
	go build -o $(BINARY) ./cmd/ezproxy

test:
	go test ./... -v

clean:
	rm -f $(BINARY) $(BINARY)-*

release:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY)-darwin-amd64 ./cmd/ezproxy
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY)-darwin-arm64 ./cmd/ezproxy
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 ./cmd/ezproxy
	GOOS=linux GOARCH=arm64 go build -o $(BINARY)-linux-arm64 ./cmd/ezproxy
```

**Commit:**

```bash
git add Makefile
git commit -m "feat: add Makefile with build, test, and cross-compile targets"
```
