package configurator

import (
	"fmt"
	"os"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
)

type Yum struct{}

func (y *Yum) Name() string { return "yum" }

func (y *Yum) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("yum") || detect.IsCommandAvailable("dnf")
}

func (y *Yum) confFile() string {
	if detect.IsCommandAvailable("dnf") {
		return "/etc/dnf/dnf.conf"
	}
	return "/etc/yum.conf"
}

func (y *Yum) Apply(cfg *config.Config) error {
	certPath := config.ExpandPath(cfg.CACert)
	confFile := y.confFile()

	// Build sed commands to add/update proxy lines in the [main] section
	cmds := []string{
		fmt.Sprintf("grep -q '^proxy=' %s && sed -i 's|^proxy=.*|proxy=%s|' %s || echo 'proxy=%s' >> %s",
			confFile, cfg.Proxy.HTTP, confFile, cfg.Proxy.HTTP, confFile),
	}
	if certPath != "" {
		cmds = append(cmds,
			fmt.Sprintf("grep -q '^sslcacert=' %s && sed -i 's|^sslcacert=.*|sslcacert=%s|' %s || echo 'sslcacert=%s' >> %s",
				confFile, certPath, confFile, certPath, confFile),
		)
	}

	return runSudoCommands(y.Name(), cmds)
}

func (y *Yum) Remove() error {
	confFile := y.confFile()
	return runSudoRemoveCommands(y.Name(), []string{
		fmt.Sprintf("sed -i '/^proxy=/d; /^sslcacert=/d' %s", confFile),
	})
}

func (y *Yum) Status(cfg *config.Config) (string, error) {
	data, err := os.ReadFile(y.confFile())
	if err != nil {
		return "not configured", nil
	}
	if strings.Contains(string(data), "proxy=") {
		return "configured", nil
	}
	return "not configured", nil
}
