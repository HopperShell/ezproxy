package configurator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

const javaCAAlias = "ezproxy-corp-ca"

type JavaCA struct{}

func (j *JavaCA) Name() string { return "java_ca" }

func (j *JavaCA) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("keytool")
}

func (j *JavaCA) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	if certPath == "" {
		return nil
	}

	if _, err := os.Stat(certPath); err != nil {
		return fmt.Errorf("cert file not found: %s", certPath)
	}

	cacertsPath := findJavaCacerts()
	if cacertsPath == "" {
		fmt.Println("  Could not locate JVM cacerts keystore. Set JAVA_HOME and retry.")
		return nil
	}

	// Check if already imported
	if !fileutil.DryRun && isJavaCertInstalled(cacertsPath) {
		fmt.Printf("  ✓ CA cert already in JVM trust store (%s)\n", cacertsPath)
		return nil
	}

	cmd := fmt.Sprintf(
		"keytool -importcert -alias %s -file %s -keystore %s -storepass changeit -noprompt",
		javaCAAlias, shellQuote(certPath), shellQuote(cacertsPath),
	)

	if fileutil.DryRun {
		fmt.Printf("\n  [dry-run] Would run (requires sudo):\n")
		fmt.Printf("    sudo sh -c '%s'\n", cmd)
		return nil
	}

	return runSudoCommands(j.Name(), []string{cmd})
}

func (j *JavaCA) Remove() error {
	cacertsPath := findJavaCacerts()
	if cacertsPath == "" {
		return nil
	}

	if !isJavaCertInstalled(cacertsPath) {
		return nil
	}

	cmd := fmt.Sprintf(
		"keytool -delete -alias %s -keystore %s -storepass changeit -noprompt",
		javaCAAlias, shellQuote(cacertsPath),
	)

	if fileutil.DryRun {
		fmt.Printf("\n  [dry-run] Would run (requires sudo):\n")
		fmt.Printf("    sudo sh -c '%s'\n", cmd)
		return nil
	}

	return runSudoRemoveCommands(j.Name(), []string{cmd})
}

func (j *JavaCA) Status(cfg *config.Config) (string, error) {
	certPath := config.ExpandPath(cfg.CACert)
	if certPath == "" {
		return "no cert configured", nil
	}

	cacertsPath := findJavaCacerts()
	if cacertsPath == "" {
		return "JVM cacerts not found", nil
	}

	if isJavaCertInstalled(cacertsPath) {
		return "imported into JVM", nil
	}
	return "not imported", nil
}

// isJavaCertInstalled checks if the ezproxy alias exists in the JVM keystore.
func isJavaCertInstalled(cacertsPath string) bool {
	out, err := exec.Command("keytool", "-list",
		"-alias", javaCAAlias,
		"-keystore", cacertsPath,
		"-storepass", "changeit").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), javaCAAlias)
}

// findJavaCacerts locates the JVM cacerts file.
func findJavaCacerts() string {
	// Check JAVA_HOME first
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		p := filepath.Join(javaHome, "lib", "security", "cacerts")
		if _, err := os.Stat(p); err == nil {
			return p
		}
		// Older JDK layout
		p = filepath.Join(javaHome, "jre", "lib", "security", "cacerts")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Try common locations
	candidates := []string{
		// macOS (Homebrew)
		"/opt/homebrew/opt/openjdk/libexec/openjdk.jdk/Contents/Home/lib/security/cacerts",
		"/usr/local/opt/openjdk/libexec/openjdk.jdk/Contents/Home/lib/security/cacerts",
		// macOS system Java
		"/Library/Java/JavaVirtualMachines",
		// Linux
		"/etc/pki/java/cacerts",
		"/etc/ssl/certs/java/cacerts",
	}

	for _, c := range candidates {
		if c == "/Library/Java/JavaVirtualMachines" {
			// Scan for installed JDKs on macOS
			if p := findMacOSJDKCacerts(c); p != "" {
				return p
			}
			continue
		}
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	// Last resort: ask java where it lives
	if out, err := exec.Command("java", "-XshowSettings:property", "-version").CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "java.home") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					jHome := strings.TrimSpace(parts[1])
					p := filepath.Join(jHome, "lib", "security", "cacerts")
					if _, err := os.Stat(p); err == nil {
						return p
					}
				}
			}
		}
	}

	return ""
}

// findMacOSJDKCacerts scans /Library/Java/JavaVirtualMachines for installed JDKs.
func findMacOSJDKCacerts(baseDir string) string {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			p := filepath.Join(baseDir, e.Name(), "Contents", "Home", "lib", "security", "cacerts")
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}
