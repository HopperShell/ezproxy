package configurator

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

type SystemCA struct{}

func (s *SystemCA) Name() string { return "system_ca" }

func (s *SystemCA) IsAvailable(_ detect.OSInfo) bool { return true }

func (s *SystemCA) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	if certPath == "" {
		return fmt.Errorf("no CA cert configured")
	}

	if _, err := os.Stat(certPath); err != nil {
		return fmt.Errorf("cert file not found: %s", certPath)
	}

	if !fileutil.DryRun && isCertSystemTrusted(certPath) {
		fmt.Printf("  ✓ CA cert is already trusted by the system (likely managed by IT)\n")
		return nil
	}

	osInfo := detect.DetectOS()

	if runtime.GOOS == "darwin" {
		return s.applyDarwin(certPath)
	}

	if fileutil.DryRun {
		fmt.Printf("\n  [dry-run] Would check if CA cert is already in system trust store\n")
		fmt.Printf("  [dry-run] If not found, would install via sudo\n")
		return nil
	}

	if osInfo.IsDebian() {
		return runSudoCommands(s.Name(), []string{
			fmt.Sprintf("cp %s /usr/local/share/ca-certificates/ezproxy-corp-ca.crt", shellQuote(certPath)),
			"update-ca-certificates",
		})
	}

	if osInfo.IsRHEL() {
		return runSudoCommands(s.Name(), []string{
			fmt.Sprintf("cp %s /etc/pki/ca-trust/source/anchors/ezproxy-corp-ca.pem", shellQuote(certPath)),
			"update-ca-trust extract",
		})
	}

	if osInfo.IsArch() {
		return runSudoCommands(s.Name(), []string{
			fmt.Sprintf("trust anchor --store %s", shellQuote(certPath)),
		})
	}

	fmt.Println("\n  Unknown Linux distro. Copy cert to your system's CA trust directory and update the trust store manually.")
	return nil
}

func (s *SystemCA) applyDarwin(certPath string) error {
	if fileutil.DryRun {
		fmt.Printf("\n  [dry-run] Would check if CA cert is already in macOS System Keychain\n")
		fmt.Printf("  [dry-run] If not found, would run: sudo security add-trusted-cert ...\n")
		return nil
	}

	return runSudoCommands(s.Name(), []string{
		fmt.Sprintf("security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain %s", shellQuote(certPath)),
	})
}

// isCertSystemTrusted checks whether the given PEM cert is already trusted
// by the operating system. On macOS it checks the Keychain, on Linux it uses
// Go's SystemCertPool which reads the distro's CA bundle.
func isCertSystemTrusted(certPath string) bool {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return false
	}
	block, _ := pem.Decode(certData)
	if block == nil {
		return false
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}

	if runtime.GOOS == "darwin" {
		return isDarwinCertTrusted(certPath, cert)
	}

	// Linux: check against the system cert pool
	pool, err := x509.SystemCertPool()
	if err != nil {
		return false
	}
	// Verify the cert against the system pool. Since this is a CA cert,
	// we check if it's present as a root in the pool by trying to verify
	// a chain with it as both leaf and root.
	opts := x509.VerifyOptions{
		Roots: pool,
	}
	// For CA certs, check if the system pool already contains this cert
	// by seeing if it can verify itself (self-signed CA).
	if _, err := cert.Verify(opts); err == nil {
		return true
	}
	// Fallback: check if any cert in the pool matches our cert's subject
	for _, poolCert := range pool.Subjects() {
		if string(poolCert) == string(cert.RawSubject) {
			return true
		}
	}
	return false
}

// isDarwinCertTrusted checks macOS Keychain for the cert.
func isDarwinCertTrusted(certPath string, cert *x509.Certificate) bool {
	// security verify-cert returns 0 if the cert is trusted
	if err := exec.Command("security", "verify-cert", "-c", certPath, "-L").Run(); err == nil {
		return true
	}
	// Fallback: search the System Keychain by CN
	out, err := exec.Command("security", "find-certificate",
		"-c", cert.Subject.CommonName,
		"-Z", "/Library/Keychains/System.keychain").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), cert.Subject.CommonName)
}

func (s *SystemCA) Remove() error {
	osInfo := detect.DetectOS()

	if runtime.GOOS == "darwin" {
		fmt.Println("\n  To remove CA cert from macOS: open Keychain Access > System > Certificates, find the cert and delete it.")
		return nil
	}

	if osInfo.IsDebian() {
		return runSudoRemoveCommands(s.Name(), []string{
			"rm -f /usr/local/share/ca-certificates/ezproxy-corp-ca.crt",
			"update-ca-certificates --fresh",
		})
	}

	if osInfo.IsRHEL() {
		return runSudoRemoveCommands(s.Name(), []string{
			"rm -f /etc/pki/ca-trust/source/anchors/ezproxy-corp-ca.pem",
			"update-ca-trust extract",
		})
	}

	if osInfo.IsArch() {
		return runSudoRemoveCommands(s.Name(), []string{
			"trust anchor --remove ezproxy-corp-ca.pem",
		})
	}

	return nil
}

func (s *SystemCA) Status(cfg *config.Config) (string, error) {
	certPath := config.ExpandPath(cfg.CACert)
	if certPath == "" {
		return "no cert configured", nil
	}
	if isCertSystemTrusted(certPath) {
		return "trusted by system", nil
	}
	return "not trusted by system", nil
}
