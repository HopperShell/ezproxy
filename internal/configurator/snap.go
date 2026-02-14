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
	cmds := []string{
		fmt.Sprintf("snap set system proxy.http=%s", shellQuote(cfg.Proxy.HTTP)),
		fmt.Sprintf("snap set system proxy.https=%s", shellQuote(cfg.Proxy.HTTPS)),
	}
	if certPath != "" {
		cmds = append(cmds,
			fmt.Sprintf("snap set system store-certs.ezproxy=\"$(cat %s)\"", shellQuote(certPath)),
		)
	}
	return runSudoCommands(s.Name(), cmds)
}

func (s *Snap) Remove() error {
	return runSudoRemoveCommands(s.Name(), []string{
		"snap unset system proxy.http",
		"snap unset system proxy.https",
		"snap unset system store-certs.ezproxy",
	})
}

func (s *Snap) Status(cfg *config.Config) (string, error) {
	return "unknown (check manually)", nil
}
