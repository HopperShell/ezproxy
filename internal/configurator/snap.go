package configurator

import (
	"fmt"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
)

type Snap struct{}

func (s *Snap) Name() string { return "snap" }

func (s *Snap) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("snap")
}

func (s *Snap) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	fmt.Println("\n[sudo required] To configure snap proxy, run:")
	fmt.Printf("  sudo snap set system proxy.http=\"%s\"\n", cfg.Proxy.HTTP)
	fmt.Printf("  sudo snap set system proxy.https=\"%s\"\n", cfg.Proxy.HTTPS)
	if certPath != "" {
		fmt.Printf("  sudo snap set system store-certs.ezproxy=\"$(cat %s)\"\n", certPath)
		fmt.Println("Note: CA cert support requires snapd 2.45+.")
	}
	return nil
}

func (s *Snap) Remove() error {
	fmt.Println("\n[sudo required] To remove snap proxy config, run:")
	fmt.Println("  sudo snap unset system proxy.http")
	fmt.Println("  sudo snap unset system proxy.https")
	fmt.Println("  sudo snap unset system store-certs.ezproxy")
	return nil
}

func (s *Snap) Status(cfg *config.Config) (string, error) {
	return "unknown (check manually)", nil
}
