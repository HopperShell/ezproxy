package configurator

import (
	"fmt"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
)

type Yum struct{}

func (y *Yum) Name() string { return "yum" }

func (y *Yum) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("yum") || detect.IsCommandAvailable("dnf")
}

func (y *Yum) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	confFile := "/etc/yum.conf"
	if detect.IsCommandAvailable("dnf") {
		confFile = "/etc/dnf/dnf.conf"
	}

	fmt.Printf("\n[sudo required] To configure yum/dnf proxy, add to %s:\n", confFile)
	fmt.Printf("  proxy=%s\n", cfg.Proxy.HTTP)
	if certPath != "" {
		fmt.Printf("  sslcacert=%s\n", certPath)
	}
	fmt.Println("Also handled by system CA store (update-ca-trust).")
	return nil
}

func (y *Yum) Remove() error {
	fmt.Println("\n[manual] Remove proxy= and sslcacert= lines from /etc/yum.conf or /etc/dnf/dnf.conf")
	return nil
}

func (y *Yum) Status(cfg *config.Config) (string, error) {
	return "unknown (check manually)", nil
}
