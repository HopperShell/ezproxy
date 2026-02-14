package configurator

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

type Maven struct{}

func (m *Maven) Name() string { return "maven" }

func (m *Maven) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("mvn")
}

func (m *Maven) settingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".m2", "settings.xml")
}

// Maven settings.xml types
type mavenSettings struct {
	XMLName xml.Name     `xml:"settings"`
	Proxies *mavenProxies `xml:"proxies,omitempty"`
	Other   []xmlNode    `xml:",any"`
}

type mavenProxies struct {
	Proxy []mavenProxy `xml:"proxy"`
}

type mavenProxy struct {
	ID            string `xml:"id"`
	Active        bool   `xml:"active"`
	Protocol      string `xml:"protocol"`
	Host          string `xml:"host"`
	Port          string `xml:"port"`
	NonProxyHosts string `xml:"nonProxyHosts,omitempty"`
}

type xmlNode struct {
	XMLName xml.Name
	Content []byte `xml:",innerxml"`
}

func (m *Maven) Apply(cfg *config.Config) error {
	path := m.settingsPath()

	httpHost, httpPort := parseProxyURL(cfg.Proxy.HTTP)
	httpsHost, httpsPort := parseProxyURL(cfg.Proxy.HTTPS)
	nonProxy := toJavaNonProxyHosts(cfg.Proxy.NoProxy)

	newProxies := &mavenProxies{
		Proxy: []mavenProxy{
			{
				ID:            "ezproxy-http",
				Active:        true,
				Protocol:      "http",
				Host:          httpHost,
				Port:          httpPort,
				NonProxyHosts: nonProxy,
			},
			{
				ID:            "ezproxy-https",
				Active:        true,
				Protocol:      "https",
				Host:          httpsHost,
				Port:          httpsPort,
				NonProxyHosts: nonProxy,
			},
		},
	}

	if fileutil.DryRun {
		data, _ := xml.MarshalIndent(newProxies, "  ", "  ")
		fmt.Printf("\n  [dry-run] Would merge into %s:\n", path)
		for _, line := range strings.Split(string(data), "\n") {
			fmt.Printf("    %s\n", line)
		}
		return nil
	}

	// Read or create settings.xml
	var settings mavenSettings
	if data, err := os.ReadFile(path); err == nil {
		xml.Unmarshal(data, &settings)
	}

	// Remove any existing ezproxy proxies, keep user's other proxies
	if settings.Proxies != nil {
		var kept []mavenProxy
		for _, p := range settings.Proxies.Proxy {
			if !strings.HasPrefix(p.ID, "ezproxy-") {
				kept = append(kept, p)
			}
		}
		newProxies.Proxy = append(kept, newProxies.Proxy...)
	}
	settings.Proxies = newProxies

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	out, err := xml.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	content := xml.Header + string(out) + "\n"
	return os.WriteFile(path, []byte(content), 0644)
}

func (m *Maven) Remove() error {
	path := m.settingsPath()

	if fileutil.DryRun {
		fmt.Printf("\n  [dry-run] Would remove ezproxy proxy entries from %s\n", path)
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var settings mavenSettings
	if err := xml.Unmarshal(data, &settings); err != nil {
		return nil
	}

	if settings.Proxies == nil {
		return nil
	}

	var kept []mavenProxy
	for _, p := range settings.Proxies.Proxy {
		if !strings.HasPrefix(p.ID, "ezproxy-") {
			kept = append(kept, p)
		}
	}

	if len(kept) == 0 {
		settings.Proxies = nil
	} else {
		settings.Proxies.Proxy = kept
	}

	out, err := xml.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	content := xml.Header + string(out) + "\n"
	return os.WriteFile(path, []byte(content), 0644)
}

func (m *Maven) Status(cfg *config.Config) (string, error) {
	data, err := os.ReadFile(m.settingsPath())
	if err != nil {
		return "not configured", nil
	}
	if strings.Contains(string(data), "ezproxy-http") {
		return "configured", nil
	}
	return "not configured", nil
}
