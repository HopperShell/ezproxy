package configurator

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/andrew/ezproxy/internal/fileutil"
)

// runSudoCommands prompts the user for confirmation, then runs each command
// via "sudo sh -c". If DryRun is enabled, it prints the commands instead.
// If AutoYes is enabled, it skips the confirmation prompt.
func runSudoCommands(toolName string, cmds []string) error {
	if len(cmds) == 0 {
		return nil
	}

	if fileutil.DryRun {
		fmt.Printf("\n  [dry-run] Would run (requires sudo):\n")
		for _, cmd := range cmds {
			fmt.Printf("    sudo sh -c '%s'\n", cmd)
		}
		return nil
	}

	fmt.Printf("\n  [%s] The following commands require sudo:\n", toolName)
	for _, cmd := range cmds {
		fmt.Printf("    sudo sh -c '%s'\n", cmd)
	}

	if !fileutil.AutoYes {
		var confirm bool
		err := huh.NewConfirm().
			Title("Run these commands now?").
			Affirmative("Yes").
			Negative("No").
			Value(&confirm).
			Run()
		if err != nil || !confirm {
			fmt.Printf("  Skipped. Run the commands above manually.\n")
			return nil
		}
	}

	for _, cmd := range cmds {
		c := exec.Command("sudo", "sh", "-c", cmd)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("command failed: sudo sh -c '%s': %w", cmd, err)
		}
	}

	return nil
}

// runSudoRemoveCommands is like runSudoCommands but best-effort (warns on errors).
func runSudoRemoveCommands(toolName string, cmds []string) error {
	if len(cmds) == 0 {
		return nil
	}

	if fileutil.DryRun {
		fmt.Printf("\n  [dry-run] Would run (requires sudo):\n")
		for _, cmd := range cmds {
			fmt.Printf("    sudo sh -c '%s'\n", cmd)
		}
		return nil
	}

	fmt.Printf("\n  [%s] The following removal commands require sudo:\n", toolName)
	for _, cmd := range cmds {
		fmt.Printf("    sudo sh -c '%s'\n", cmd)
	}

	if !fileutil.AutoYes {
		var confirm bool
		err := huh.NewConfirm().
			Title("Run these commands now?").
			Affirmative("Yes").
			Negative("No").
			Value(&confirm).
			Run()
		if err != nil || !confirm {
			fmt.Printf("  Skipped. Run the commands above manually.\n")
			return nil
		}
	}

	for _, cmd := range cmds {
		c := exec.Command("sudo", "sh", "-c", cmd)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: %s\n", err)
		}
	}

	return nil
}

// shellQuote wraps a string in single quotes for safe shell embedding.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
