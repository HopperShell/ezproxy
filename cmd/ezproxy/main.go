package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/configurator"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ezproxy <command> [args] [flags]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  init              Interactive setup wizard")
		fmt.Println("  apply             Apply proxy config to all enabled tools")
		fmt.Println("  remove            Remove proxy config from all tools")
		fmt.Println("  status            Show current config status per tool")
		fmt.Println("  manage            Interactive tool manager (toggle tools on/off)")
		fmt.Println("  enable <tool>     Enable a tool and apply its config")
		fmt.Println("  disable <tool>    Disable a tool and remove its config")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  --dry-run         Preview changes without modifying files")
		fmt.Println("  --yes, -y         Skip confirmations (for scripting)")
		os.Exit(1)
	}

	// Parse global flags from anywhere in args
	var cleaned []string
	for _, arg := range os.Args {
		switch arg {
		case "--dry-run":
			fileutil.DryRun = true
		case "--yes", "-y":
			fileutil.AutoYes = true
		default:
			cleaned = append(cleaned, arg)
		}
	}
	os.Args = cleaned

	switch os.Args[1] {
	case "init":
		cmdInit()
	case "apply":
		cmdApply()
	case "remove":
		cmdRemove()
	case "status":
		cmdStatus()
	case "manage":
		cmdManage()
	case "enable":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ezproxy enable <tool>")
			os.Exit(1)
		}
		cmdEnable(os.Args[2])
	case "disable":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: ezproxy disable <tool>")
			os.Exit(1)
		}
		cmdDisable(os.Args[2])
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
		profiles := detect.ShellProfiles()
		if len(profiles) > 0 {
			fmt.Printf("\nDone! Restart your shell or run 'source %s' to apply env vars.\n", profiles[0])
		} else {
			fmt.Println("\nDone! Restart your shell to apply env vars.")
		}
	}
}

func cmdRemove() {
	cfg := loadConfig()
	osInfo := detect.DetectOS()

	if fileutil.DryRun {
		fmt.Println("DRY RUN: showing what would be removed (no files modified)")
	} else if !fileutil.AutoYes {
		var confirm bool
		err := huh.NewConfirm().
			Title("Remove all proxy configuration?").
			Description("This will undo proxy settings for all enabled tools.").
			Affirmative("Yes, remove").
			Negative("Cancel").
			Value(&confirm).
			Run()
		if err != nil || !confirm {
			fmt.Println("Cancelled.")
			return
		}
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

	fmt.Printf("Proxy:    %s\n", cfg.Proxy.HTTP)
	if cfg.Proxy.HTTPS != cfg.Proxy.HTTP {
		fmt.Printf("HTTPS:    %s\n", cfg.Proxy.HTTPS)
	}
	fmt.Printf("NO_PROXY: %s\n", cfg.Proxy.NoProxy)
	if cfg.CACert != "" {
		fmt.Printf("CA Cert:  %s\n", cfg.CACert)
	}
	fmt.Println()
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

func cmdManage() {
	cfg := loadConfig()
	osInfo := detect.DetectOS()
	allConfigurators := configurator.All()

	// Build options with current state
	var toolOptions []huh.Option[string]
	for _, c := range allConfigurators {
		enabled := cfg.Tools[c.Name()]
		label := c.Name()

		// Add status info to label
		if !c.IsAvailable(osInfo) {
			label += " (not installed)"
		} else {
			status, _ := c.Status(cfg)
			if status != "" && status != "not configured" {
				label += " [" + status + "]"
			}
		}

		toolOptions = append(toolOptions,
			huh.NewOption(label, c.Name()).Selected(enabled),
		)
	}

	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Manage tools").
				Description("Space to toggle, enter to apply changes.").
				Options(toolOptions...).
				Height(len(toolOptions)+2).
				Value(&selected),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled.\n")
		os.Exit(1)
	}

	// Build new enabled set
	newEnabled := make(map[string]bool, len(selected))
	for _, name := range selected {
		newEnabled[name] = true
	}

	// Diff against current state and apply changes
	var enabled, disabled []string
	for _, c := range allConfigurators {
		name := c.Name()
		wasEnabled := cfg.Tools[name]
		nowEnabled := newEnabled[name]

		if wasEnabled && !nowEnabled {
			disabled = append(disabled, name)
		} else if !wasEnabled && nowEnabled {
			enabled = append(enabled, name)
		}
	}

	if len(enabled) == 0 && len(disabled) == 0 {
		fmt.Println("No changes.")
		return
	}

	// Update config
	for _, c := range allConfigurators {
		cfg.Tools[c.Name()] = newEnabled[c.Name()]
	}
	if err := config.Save(configPath(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	// Apply newly enabled tools
	for _, name := range enabled {
		c := findConfigurator(name)
		if c == nil || !c.IsAvailable(osInfo) {
			fmt.Printf("  %-12s enabled (not installed, will configure when available)\n", name)
			continue
		}
		if err := c.Apply(cfg); err != nil {
			fmt.Printf("  %-12s enabled, ERROR applying: %v\n", name, err)
		} else {
			fmt.Printf("  %-12s ✓ enabled and configured\n", name)
		}
	}

	// Remove newly disabled tools
	for _, name := range disabled {
		c := findConfigurator(name)
		if c == nil {
			continue
		}
		if err := c.Remove(); err != nil {
			fmt.Printf("  %-12s disabled, ERROR removing: %v\n", name, err)
		} else {
			fmt.Printf("  %-12s ✓ disabled and removed\n", name)
		}
	}

	fmt.Printf("\n%d enabled, %d disabled.\n", len(enabled), len(disabled))
}

func findConfigurator(name string) configurator.Configurator {
	for _, c := range configurator.All() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func cmdEnable(tool string) {
	cfg := loadConfig()

	c := findConfigurator(tool)
	if c == nil {
		fmt.Fprintf(os.Stderr, "Unknown tool: %s\n", tool)
		fmt.Fprintln(os.Stderr, "Available tools:")
		for _, c := range configurator.All() {
			fmt.Fprintf(os.Stderr, "  %s\n", c.Name())
		}
		os.Exit(1)
	}

	if cfg.Tools[tool] {
		fmt.Printf("%s is already enabled.\n", tool)
		return
	}

	cfg.Tools[tool] = true
	if err := config.Save(configPath(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	osInfo := detect.DetectOS()
	if !c.IsAvailable(osInfo) {
		fmt.Printf("Enabled %s (not installed, will be configured when available).\n", tool)
		return
	}

	if err := c.Apply(cfg); err != nil {
		fmt.Printf("Enabled %s but failed to apply: %v\n", tool, err)
	} else {
		fmt.Printf("Enabled and configured %s.\n", tool)
	}
}

func cmdDisable(tool string) {
	cfg := loadConfig()

	c := findConfigurator(tool)
	if c == nil {
		fmt.Fprintf(os.Stderr, "Unknown tool: %s\n", tool)
		fmt.Fprintln(os.Stderr, "Available tools:")
		for _, c := range configurator.All() {
			fmt.Fprintf(os.Stderr, "  %s\n", c.Name())
		}
		os.Exit(1)
	}

	if enabled, exists := cfg.Tools[tool]; exists && !enabled {
		fmt.Printf("%s is already disabled.\n", tool)
		return
	}

	cfg.Tools[tool] = false
	if err := config.Save(configPath(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	if err := c.Remove(); err != nil {
		fmt.Printf("Disabled %s but failed to remove config: %v\n", tool, err)
	} else {
		fmt.Printf("Disabled and removed config for %s.\n", tool)
	}
}

func cmdInit() {
	defaultNoProxy := "localhost,127.0.0.1,.corp.com,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"

	var (
		httpProxy  string
		httpsProxy string
		noProxy    string = defaultNoProxy
		certInput  string
	)

	// Pre-fill from existing config if present
	if existing, err := config.Load(configPath()); err == nil {
		httpProxy = existing.Proxy.HTTP
		httpsProxy = existing.Proxy.HTTPS
		noProxy = existing.Proxy.NoProxy
		certInput = existing.CACert
		fmt.Println("Existing config found - values pre-filled. Edit as needed.")
		fmt.Println()
	}

	// Page 1: Proxy settings
	proxyForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("HTTP Proxy URL").
				Description("e.g. http://proxy.corp.com:8080").
				Value(&httpProxy).
				Validate(huh.ValidateNotEmpty()),

			huh.NewInput().
				Title("HTTPS Proxy URL").
				Description("Leave blank to use the same as HTTP proxy").
				Value(&httpsProxy),

			huh.NewInput().
				Title("NO_PROXY").
				Description("Comma-separated hosts/CIDRs to bypass the proxy").
				Value(&noProxy),

			huh.NewInput().
				Title("CA Certificate Path").
				Description("Path to PEM file (optional, leave blank to skip)").
				Value(&certInput),
		),
	).WithTheme(huh.ThemeCharm())

	if err := proxyForm.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled.\n")
		os.Exit(1)
	}

	if httpsProxy == "" {
		httpsProxy = httpProxy
	}

	// Copy CA cert if provided
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

	// Page 2: Tool selection via interactive checkboxes
	osInfo := detect.DetectOS()
	allConfigurators := configurator.All()

	var toolOptions []huh.Option[string]
	for _, c := range allConfigurators {
		installed := c.IsAvailable(osInfo)
		label := c.Name()
		if !installed {
			label += " (not installed)"
		}
		toolOptions = append(toolOptions,
			huh.NewOption(label, c.Name()).Selected(installed),
		)
	}

	var enabledTools []string
	toolForm := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Tools to configure").
				Description("Use arrow keys to navigate, space to toggle, enter to confirm.").
				Options(toolOptions...).
				Height(len(toolOptions)+2).
				Value(&enabledTools),
		),
	).WithTheme(huh.ThemeCharm())

	if err := toolForm.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled.\n")
		os.Exit(1)
	}

	// Build tools map from selection
	tools := config.DefaultTools()
	enabledSet := make(map[string]bool, len(enabledTools))
	for _, name := range enabledTools {
		enabledSet[name] = true
	}
	for name := range tools {
		tools[name] = enabledSet[name]
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
