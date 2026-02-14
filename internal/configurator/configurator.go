package configurator

import (
	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
)

// Configurator is the interface all tool configurators implement.
type Configurator interface {
	// Name returns the tool name (matches key in config.yaml tools map).
	Name() string
	// IsAvailable returns true if the tool is installed/relevant on this system.
	IsAvailable(osInfo detect.OSInfo) bool
	// Apply writes proxy configuration for this tool.
	Apply(cfg *config.Config) error
	// Remove undoes proxy configuration for this tool.
	Remove() error
	// Status returns "configured", "not configured", or "stale".
	Status(cfg *config.Config) (string, error)
}

// All returns all registered configurators in apply order.
func All() []Configurator {
	return []Configurator{
		&SystemCA{},
		&EnvVars{},
		&Git{},
		&Pip{},
		&Npm{},
		&Yarn{},
		&Docker{},
		&Curl{},
		&Wget{},
		&Cargo{},
		&Conda{},
		&Brew{},
		&Snap{},
		&Apt{},
		&Yum{},
		&SSH{},
	}
}
