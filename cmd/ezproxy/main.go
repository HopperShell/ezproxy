package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/configurator"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ezproxy <init|apply|remove|status> [--dry-run]")
		os.Exit(1)
	}

	// Check for --dry-run flag anywhere in args
	for i, arg := range os.Args {
		if arg == "--dry-run" {
			fileutil.DryRun = true
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			break
		}
	}

	switch os.Args[1] {
	case "init":
		cmdInit()
	case "apply":
		cmdApply()
	case "remove":
		cmdRemove()
	case "status":
		cmdStatus()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot determine home directory: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, ".ezproxy", "config.yaml")
}

func loadConfig() *config.Config {
	cfg, err := config.Load(configPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'ezproxy init' to create a config file.\n")
		os.Exit(1)
	}
	return cfg
}

func cmdApply() {
	cfg := loadConfig()
	osInfo := detect.DetectOS()

	if fileutil.DryRun {
		fmt.Println("DRY RUN: showing what would be configured (no files modified)")
	} else {
		fmt.Println("Applying proxy configuration...")
	}

	for _, c := range configurator.All() {
		enabled, exists := cfg.Tools[c.Name()]
		if exists && !enabled {
			fmt.Printf("  %-12s skipped (disabled)\n", c.Name())
			continue
		}
		if !c.IsAvailable(osInfo) {
			fmt.Printf("  %-12s skipped (not installed)\n", c.Name())
			continue
		}
		if err := c.Apply(cfg); err != nil {
			fmt.Printf("  %-12s ERROR: %v\n", c.Name(), err)
		} else {
			fmt.Printf("  %-12s ✓ configured\n", c.Name())
		}
	}

	if !fileutil.DryRun {
		fmt.Println("\nDone! Restart your shell or run 'source ~/.bashrc' (or ~/.zshrc) to apply env vars.")
	}
}

func cmdRemove() {
	cfg := loadConfig()
	osInfo := detect.DetectOS()

	if fileutil.DryRun {
		fmt.Println("DRY RUN: showing what would be removed (no files modified)")
	} else {
		fmt.Println("Removing proxy configuration...")
	}

	for _, c := range configurator.All() {
		enabled, exists := cfg.Tools[c.Name()]
		if exists && !enabled {
			continue
		}
		if !c.IsAvailable(osInfo) {
			continue
		}
		if err := c.Remove(); err != nil {
			fmt.Printf("  %-12s ERROR: %v\n", c.Name(), err)
		} else {
			fmt.Printf("  %-12s ✓ removed\n", c.Name())
		}
	}

	if !fileutil.DryRun {
		fmt.Println("\nDone! Restart your shell to apply changes.")
	}
}

func cmdStatus() {
	cfg := loadConfig()
	osInfo := detect.DetectOS()

	fmt.Printf("%-14s %-28s %s\n", "Tool", "Status", "Available")
	fmt.Printf("%-14s %-28s %s\n", "────", "──────", "─────────")

	for _, c := range configurator.All() {
		enabled, exists := cfg.Tools[c.Name()]
		if exists && !enabled {
			fmt.Printf("%-14s %-28s %s\n", c.Name(), "disabled", "-")
			continue
		}

		available := c.IsAvailable(osInfo)
		if !available {
			fmt.Printf("%-14s %-28s %s\n", c.Name(), "skipped", "no (not installed)")
			continue
		}

		status, err := c.Status(cfg)
		if err != nil {
			status = fmt.Sprintf("error: %v", err)
		}
		fmt.Printf("%-14s %-28s %s\n", c.Name(), status, "yes")
	}
}

func cmdInit() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("ezproxy setup")
	fmt.Println("=============")

	// HTTP proxy
	fmt.Print("HTTP proxy URL: ")
	scanner.Scan()
	httpProxy := strings.TrimSpace(scanner.Text())
	if httpProxy == "" {
		fmt.Fprintln(os.Stderr, "HTTP proxy URL is required.")
		os.Exit(1)
	}

	// HTTPS proxy
	fmt.Printf("HTTPS proxy URL [%s]: ", httpProxy)
	scanner.Scan()
	httpsProxy := strings.TrimSpace(scanner.Text())
	if httpsProxy == "" {
		httpsProxy = httpProxy
	}

	// NO_PROXY
	defaultNoProxy := "localhost,127.0.0.1,.corp.com,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
	fmt.Printf("NO_PROXY [%s]: ", defaultNoProxy)
	scanner.Scan()
	noProxy := strings.TrimSpace(scanner.Text())
	if noProxy == "" {
		noProxy = defaultNoProxy
	}

	// CA cert
	fmt.Print("Path to CA certificate PEM file (optional, press Enter to skip): ")
	scanner.Scan()
	certInput := strings.TrimSpace(scanner.Text())

	home, _ := os.UserHomeDir()
	ezproxyDir := filepath.Join(home, ".ezproxy")
	if err := os.MkdirAll(ezproxyDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", ezproxyDir, err)
		os.Exit(1)
	}

	caCertConfig := ""
	if certInput != "" {
		certInput = config.ExpandPath(certInput)
		destCert := filepath.Join(ezproxyDir, "corp-ca.pem")

		src, err := os.Open(certInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading cert: %v\n", err)
			os.Exit(1)
		}
		defer src.Close()

		dst, err := os.Create(destCert)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating cert copy: %v\n", err)
			os.Exit(1)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			fmt.Fprintf(os.Stderr, "Error copying cert: %v\n", err)
			os.Exit(1)
		}

		caCertConfig = "~/.ezproxy/corp-ca.pem"
		fmt.Printf("  Copied cert to %s\n", destCert)
	}

	// Tool selection
	tools := config.DefaultTools()
	osInfo := detect.DetectOS()
	allConfigurators := configurator.All()

	fmt.Println("\nTools (all enabled by default):")
	for _, c := range allConfigurators {
		installed := c.IsAvailable(osInfo)
		status := "✓"
		note := ""
		if !installed {
			note = " (not installed)"
		}
		fmt.Printf("  [%s] %-12s%s\n", status, c.Name(), note)
	}

	fmt.Print("\nDisable any tools? (comma-separated names, or Enter to keep all): ")
	scanner.Scan()
	disableInput := strings.TrimSpace(scanner.Text())
	if disableInput != "" {
		for _, name := range strings.Split(disableInput, ",") {
			name = strings.TrimSpace(name)
			if _, exists := tools[name]; exists {
				tools[name] = false
			} else {
				fmt.Fprintf(os.Stderr, "  Warning: unknown tool %q, skipping\n", name)
			}
		}
	}

	cfg := &config.Config{
		Proxy: config.ProxyConfig{
			HTTP:    httpProxy,
			HTTPS:   httpsProxy,
			NoProxy: noProxy,
		},
		CACert: caCertConfig,
		Tools:  tools,
	}

	cfgPath := configPath()
	if err := config.Save(cfgPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	// Show final summary
	fmt.Printf("\nConfig saved to %s\n\n", cfgPath)
	enabledCount := 0
	disabledCount := 0
	for _, v := range tools {
		if v {
			enabledCount++
		} else {
			disabledCount++
		}
	}
	fmt.Printf("  %d tools enabled", enabledCount)
	if disabledCount > 0 {
		fmt.Printf(", %d disabled", disabledCount)
	}
	fmt.Println()
	fmt.Println("\nRun 'ezproxy apply' to configure all tools.")
	fmt.Println("Run 'ezproxy apply --dry-run' to preview changes first.")
}
