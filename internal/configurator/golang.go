package configurator

import (
	"fmt"
	"os"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

// Golang configures Go module proxy settings (GOPRIVATE, GONOSUMDB).
// HTTP_PROXY/HTTPS_PROXY are already handled by the env_vars configurator,
// and Go uses the system cert store, so this focuses on module-specific settings.
type Golang struct{}

func (g *Golang) Name() string { return "go" }

func (g *Golang) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("go")
}

func (g *Golang) Apply(cfg *config.Config) error {
	// Go uses the system cert store and respects HTTP_PROXY/HTTPS_PROXY
	// from the environment (handled by env_vars configurator).
	//
	// The main thing to configure for corporate environments is GOPRIVATE
	// and GONOSUMDB so private modules don't leak to the public sum DB.
	// We write these to the shell profile alongside the other env vars.

	shell := detect.DetectShell()
	profiles := detect.ShellProfiles()

	if len(profiles) == 0 {
		return fmt.Errorf("no shell profile found")
	}

	var content string
	if detect.IsFishShell() {
		content = "# Go module settings for corporate proxy\n" +
			"# Set GOPRIVATE to your internal module paths, e.g.:\n" +
			"#   set -gx GOPRIVATE \"github.com/yourcompany/*,git.internal.com/*\"\n" +
			"# set -gx GONOSUMDB $GOPRIVATE\n"
	} else {
		content = "# Go module settings for corporate proxy\n" +
			"# Set GOPRIVATE to your internal module paths, e.g.:\n" +
			"#   export GOPRIVATE=\"github.com/yourcompany/*,git.internal.com/*\"\n" +
			"# export GONOSUMDB=\"$GOPRIVATE\"\n"
	}

	if fileutil.DryRun {
		fmt.Printf("\n  [dry-run] Would add Go module comments to shell profile\n")
		fmt.Printf("  Note: Go uses system cert store and HTTP_PROXY from env_vars.\n")
		fmt.Printf("  Set GOPRIVATE for any internal Go module hosts.\n")
		return nil
	}

	// Only write to the first profile
	return fileutil.UpsertMarkerBlock(profiles[0], content, goMarkerComment(shell))
}

func (g *Golang) Remove() error {
	shell := detect.DetectShell()
	for _, profile := range detect.ShellProfiles() {
		fileutil.RemoveMarkerBlock(profile, goMarkerComment(shell))
	}
	return nil
}

func (g *Golang) Status(cfg *config.Config) (string, error) {
	// Check if GOPRIVATE is set in the environment
	if gp := os.Getenv("GOPRIVATE"); gp != "" {
		return fmt.Sprintf("GOPRIVATE=%s", gp), nil
	}

	// Check shell profiles for our marker
	shell := detect.DetectShell()
	for _, profile := range detect.ShellProfiles() {
		if fileutil.HasMarkerBlock(profile, goMarkerComment(shell)) {
			return "configured (GOPRIVATE not yet set)", nil
		}
	}

	return "not configured", nil
}

func goMarkerComment(shell string) string {
	if strings.Contains(shell, "fish") {
		return "#"
	}
	return "#"
}
