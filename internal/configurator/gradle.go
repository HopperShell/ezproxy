package configurator

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrew/ezproxy/internal/config"
	"github.com/andrew/ezproxy/internal/detect"
	"github.com/andrew/ezproxy/internal/fileutil"
)

type Gradle struct {
	path string // override for testing
}

func (g *Gradle) Name() string { return "gradle" }

func (g *Gradle) IsAvailable(_ detect.OSInfo) bool {
	return detect.IsCommandAvailable("gradle")
}

func (g *Gradle) getPath() string {
	if g.path != "" {
		return g.path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gradle", "gradle.properties")
}

func (g *Gradle) Apply(cfg *config.Config) error {
	path := g.getPath()

	httpHost, httpPort := parseProxyURL(cfg.Proxy.HTTP)
	httpsHost, httpsPort := parseProxyURL(cfg.Proxy.HTTPS)
	nonProxy := toJavaNonProxyHosts(cfg.Proxy.NoProxy)

	lines := []string{
		fmt.Sprintf("systemProp.http.proxyHost=%s", httpHost),
		fmt.Sprintf("systemProp.http.proxyPort=%s", httpPort),
		fmt.Sprintf("systemProp.http.nonProxyHosts=%s", nonProxy),
		fmt.Sprintf("systemProp.https.proxyHost=%s", httpsHost),
		fmt.Sprintf("systemProp.https.proxyPort=%s", httpsPort),
		fmt.Sprintf("systemProp.https.nonProxyHosts=%s", nonProxy),
	}

	content := strings.Join(lines, "\n") + "\n"
	return fileutil.UpsertMarkerBlock(path, content, "#")
}

func (g *Gradle) Remove() error {
	return fileutil.RemoveMarkerBlock(g.getPath(), "#")
}

func (g *Gradle) Status(cfg *config.Config) (string, error) {
	if fileutil.HasMarkerBlock(g.getPath(), "#") {
		return "configured", nil
	}
	return "not configured", nil
}

// parseProxyURL extracts host and port from a proxy URL like http://proxy:8080
func parseProxyURL(proxyURL string) (host, port string) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return proxyURL, "8080"
	}
	host = u.Hostname()
	port = u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return host, port
}

// toJavaNonProxyHosts converts comma-separated NO_PROXY to pipe-separated Java format.
// Also converts .domain.com to *.domain.com as Java expects.
func toJavaNonProxyHosts(noProxy string) string {
	parts := strings.Split(noProxy, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Java uses *.domain.com, NO_PROXY uses .domain.com
		if strings.HasPrefix(p, ".") {
			p = "*" + p
		}
		// Strip CIDR notation - Java nonProxyHosts doesn't support it
		if strings.Contains(p, "/") {
			continue
		}
		result = append(result, p)
	}
	return strings.Join(result, "|")
}
