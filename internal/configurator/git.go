package configurator

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
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

	if fileutil.DryRun {
		fmt.Println("\n  [dry-run] Would run:")
		for _, args := range cmds {
			fmt.Printf("    %s\n", strings.Join(args, " "))
		}
		return nil
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
	if fileutil.DryRun {
		fmt.Println("\n  [dry-run] Would run:")
		for _, key := range keys {
			fmt.Printf("    git config --global --unset %s\n", key)
		}
		return nil
	}
	for _, key := range keys {
		exec.Command("git", "config", "--global", "--unset", key).Run()
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
