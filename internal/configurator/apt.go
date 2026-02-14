package configurator

import (
	"fmt"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
)

type Apt struct{}

func (a *Apt) Name() string { return "apt" }

func (a *Apt) IsAvailable(osInfo detect.OSInfo) bool {
	return detect.IsCommandAvailable("apt") || detect.IsCommandAvailable("apt-get")
}

func (a *Apt) Apply(cfg *config.Config) error {
	fmt.Println("\n[sudo required] To configure apt proxy, run:")
	fmt.Printf("  echo 'Acquire::http::Proxy \"%s/\";' | sudo tee /etc/apt/apt.conf.d/99ezproxy\n", cfg.Proxy.HTTP)
	fmt.Printf("  echo 'Acquire::https::Proxy \"%s/\";' | sudo tee -a /etc/apt/apt.conf.d/99ezproxy\n", cfg.Proxy.HTTPS)
	fmt.Println("Note: CA handled by system CA store (update-ca-certificates).")
	return nil
}

func (a *Apt) Remove() error {
	fmt.Println("\n[sudo required] To remove apt proxy config, run:")
	fmt.Println("  sudo rm -f /etc/apt/apt.conf.d/99ezproxy")
	return nil
}

func (a *Apt) Status(cfg *config.Config) (string, error) {
	return "unknown (check manually)", nil
}
