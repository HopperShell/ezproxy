package configurator

import (
	"fmt"
	"os"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
)

type Apt struct{}

func (a *Apt) Name() string { return "apt" }

func (a *Apt) IsAvailable(osInfo detect.OSInfo) bool {
	return detect.IsCommandAvailable("apt") || detect.IsCommandAvailable("apt-get")
}

func (a *Apt) Apply(cfg *config.Config) error {
	content := fmt.Sprintf("Acquire::http::Proxy \"%s\";\nAcquire::https::Proxy \"%s\";\n", cfg.Proxy.HTTP, cfg.Proxy.HTTPS)
	return runSudoCommands(a.Name(), []string{
		fmt.Sprintf("printf '%s' > /etc/apt/apt.conf.d/99ezproxy", content),
	})
}

func (a *Apt) Remove() error {
	if _, err := os.Stat("/etc/apt/apt.conf.d/99ezproxy"); os.IsNotExist(err) {
		return nil
	}
	return runSudoRemoveCommands(a.Name(), []string{
		"rm -f /etc/apt/apt.conf.d/99ezproxy",
	})
}

func (a *Apt) Status(cfg *config.Config) (string, error) {
	if _, err := os.Stat("/etc/apt/apt.conf.d/99ezproxy"); err == nil {
		return "configured", nil
	}
	return "not configured", nil
}
