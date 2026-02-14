package configurator

import (
	"fmt"
	"runtime"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
)

type SystemCA struct{}

func (s *SystemCA) Name() string { return "system_ca" }

func (s *SystemCA) IsAvailable(_ detect.OSInfo) bool { return true }

func (s *SystemCA) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	if certPath == "" {
		return fmt.Errorf("no CA cert configured")
	}

	osInfo := detect.DetectOS()

	if runtime.GOOS == "darwin" {
		fmt.Println("\n[sudo required] To install CA cert into macOS system trust store, run:")
		fmt.Printf("  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain %s\n", certPath)
		fmt.Println("Note: macOS may prompt for admin password interactively.")
		return nil
	}

	if osInfo.IsDebian() {
		fmt.Println("\n[sudo required] To install CA cert (Debian/Ubuntu), run:")
		fmt.Printf("  sudo cp %s /usr/local/share/ca-certificates/ezproxy-corp-ca.crt\n", certPath)
		fmt.Println("  sudo update-ca-certificates")
		fmt.Println("Note: File MUST have .crt extension.")
	} else if osInfo.IsRHEL() {
		fmt.Println("\n[sudo required] To install CA cert (RHEL/Fedora), run:")
		fmt.Printf("  sudo cp %s /etc/pki/ca-trust/source/anchors/ezproxy-corp-ca.pem\n", certPath)
		fmt.Println("  sudo update-ca-trust extract")
	} else if osInfo.IsArch() {
		fmt.Println("\n[sudo required] To install CA cert (Arch), run:")
		fmt.Printf("  sudo trust anchor --store %s\n", certPath)
	} else {
		fmt.Println("\n[manual] Unknown Linux distro. To install CA cert, copy it to your system's CA trust directory and update the trust store.")
	}

	return nil
}

func (s *SystemCA) Remove() error {
	osInfo := detect.DetectOS()

	if runtime.GOOS == "darwin" {
		fmt.Println("\n[sudo required] To remove CA cert from macOS system trust store:")
		fmt.Println("  Open Keychain Access > System > Certificates, find the ezproxy cert and delete it.")
		return nil
	}

	if osInfo.IsDebian() {
		fmt.Println("\n[sudo required] To remove CA cert (Debian/Ubuntu), run:")
		fmt.Println("  sudo rm /usr/local/share/ca-certificates/ezproxy-corp-ca.crt")
		fmt.Println("  sudo update-ca-certificates --fresh")
	} else if osInfo.IsRHEL() {
		fmt.Println("\n[sudo required] To remove CA cert (RHEL/Fedora), run:")
		fmt.Println("  sudo rm /etc/pki/ca-trust/source/anchors/ezproxy-corp-ca.pem")
		fmt.Println("  sudo update-ca-trust extract")
	} else if osInfo.IsArch() {
		fmt.Println("\n[sudo required] To remove CA cert (Arch), run:")
		fmt.Println("  sudo trust anchor --remove ezproxy-corp-ca.pem")
	}

	return nil
}

func (s *SystemCA) Status(cfg *config.Config) (string, error) {
	// Can't easily check system CA store programmatically
	return "unknown (check manually)", nil
}
