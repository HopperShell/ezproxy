# ezproxy Design Document

## Problem

Corporate HTTP/HTTPS proxies cause configuration pain on Linux and macOS. Every tool has its own proxy config format, its own CA certificate trust mechanism, and its own env var conventions. Developers waste hours manually configuring pip, docker, curl, git, npm, and dozens of other tools every time they set up a machine or switch contexts.

## Solution

`ezproxy` is a Go CLI tool that takes a single config file declaring proxy URL, CA cert path, and NO_PROXY list, and applies the correct configuration to every supported tool in one command.

## Audience

Power-user CLI for developers. Not an org-wide deployment tool.

## CLI Interface

```
ezproxy init          # Interactive setup -> creates ~/.ezproxy/config.yaml
ezproxy apply         # Apply proxy config to all tools + install CA cert
ezproxy remove        # Undo all proxy configurations
ezproxy status        # Show what's configured, what's missing, what's stale
```

## Config File

`~/.ezproxy/config.yaml`:

```yaml
proxy:
  http: http://proxy.corp.com:8080
  https: http://proxy.corp.com:8080
  no_proxy: localhost,127.0.0.1,.corp.com,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16

ca_cert: ~/.ezproxy/corp-ca.pem

tools:
  env_vars: true
  git: true
  pip: true
  npm: true
  yarn: true
  docker: true
  curl: true
  wget: true
  go: true
  cargo: true
  conda: true
  brew: true
  snap: true
  apt: true
  yum: true
  ssh: true
  system_ca: true
```

`ezproxy init` interactively asks for proxy URL and cert file path, copies the cert into `~/.ezproxy/`, and generates the config with sensible defaults.

## Tool Configurations (Verified)

### Environment Variables

Appends a fenced marker block to `~/.bashrc`, `~/.zshrc`, and `~/.profile`:

```bash
# >>> ezproxy >>>
export HTTP_PROXY=http://proxy.corp.com:8080
export HTTPS_PROXY=http://proxy.corp.com:8080
export http_proxy=http://proxy.corp.com:8080
export https_proxy=http://proxy.corp.com:8080
export NO_PROXY=localhost,127.0.0.1,.corp.com
export no_proxy=localhost,127.0.0.1,.corp.com
export SSL_CERT_FILE=~/.ezproxy/corp-ca.pem
export REQUESTS_CA_BUNDLE=~/.ezproxy/corp-ca.pem
export CURL_CA_BUNDLE=~/.ezproxy/corp-ca.pem
export NODE_EXTRA_CA_CERTS=~/.ezproxy/corp-ca.pem
# <<< ezproxy <<<
```

Both upper and lowercase because different tools read different variants.

### git

```bash
git config --global http.proxy http://proxy.corp.com:8080
git config --global http.sslCAInfo ~/.ezproxy/corp-ca.pem
```

- There is NO `https.proxy` -- `http.proxy` covers both HTTP and HTTPS.
- Git also respects `http_proxy`/`https_proxy` env vars but config takes precedence.
- Env var override: `GIT_SSL_CAINFO`.

### pip

Config file:
- Linux: `~/.config/pip/pip.conf`
- macOS: `$HOME/Library/Application Support/pip/pip.conf`

Format (INI):
```ini
[global]
proxy = http://proxy.corp.com:8080
cert = /path/to/corp-ca.pem
```

Also respects: `PIP_CERT`, `REQUESTS_CA_BUNDLE` env vars.

### npm

Writes `~/.npmrc`:

```ini
proxy=http://proxy.corp.com:8080
https-proxy=http://proxy.corp.com:8080
cafile=/path/to/corp-ca.pem
```

Note: `https-proxy` uses `http://` protocol, not `https://`.

### yarn v1 (Classic)

Writes `~/.yarnrc`:

```
proxy "http://proxy.corp.com:8080"
https-proxy "http://proxy.corp.com:8080"
cafile "/path/to/corp-ca.pem"
```

### yarn v2+ (Berry)

Writes `~/.yarnrc.yml`:

```yaml
httpProxy: "http://proxy.corp.com:8080"
httpsProxy: "http://proxy.corp.com:8080"
caFilePath: "/path/to/corp-ca.pem"
```

### docker

**Client proxy** -- writes/updates `~/.docker/config.json`:

```json
{
  "proxies": {
    "default": {
      "httpProxy": "http://proxy.corp.com:8080",
      "httpsProxy": "http://proxy.corp.com:8080",
      "noProxy": "localhost,127.0.0.1,.corp.com"
    }
  }
}
```

**Daemon (Linux only)** -- writes `/etc/systemd/system/docker.service.d/ezproxy.conf`:

```ini
[Service]
Environment="HTTP_PROXY=http://proxy.corp.com:8080"
Environment="HTTPS_PROXY=http://proxy.corp.com:8080"
Environment="NO_PROXY=localhost,127.0.0.1,.corp.com"
```

Then prompts user to run `sudo systemctl daemon-reload && sudo systemctl restart docker`.

**macOS Docker Desktop** -- prints instructions to configure via GUI (Settings > Resources > Proxies). Docker Desktop reads macOS system proxy and keychain CA automatically after restart.

**CA certs for registries (Linux)** -- copies cert to `/etc/docker/certs.d/<registry>/ca.crt` (`.crt` extension required, NOT `.cert`).

### curl

Writes `~/.curlrc`:

```
# >>> ezproxy >>>
proxy = "http://proxy.corp.com:8080"
cacert = "/path/to/corp-ca.pem"
# <<< ezproxy <<<
```

### wget

Writes `~/.wgetrc`:

```
# >>> ezproxy >>>
http_proxy = http://proxy.corp.com:8080
https_proxy = http://proxy.corp.com:8080
ca_certificate = /path/to/corp-ca.pem
# <<< ezproxy <<<
```

Note: uses underscores (not dashes).

### Go

Covered by env vars (`HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`). No additional config needed.

### cargo/Rust

Writes/updates `~/.cargo/config.toml`:

```toml
# >>> ezproxy >>>
[http]
proxy = "http://proxy.corp.com:8080"
cainfo = "/path/to/corp-ca.pem"
# <<< ezproxy <<<
```

### conda

Writes/updates `~/.condarc`:

```yaml
# >>> ezproxy >>>
proxy_servers:
  http: http://proxy.corp.com:8080
  https: http://proxy.corp.com:8080
ssl_verify: /path/to/corp-ca.pem
# <<< ezproxy <<<
```

### brew (Homebrew)

Covered by env vars. Additionally sets `HOMEBREW_CURLRC=1` in shell profile so Homebrew reads `~/.curlrc`.

### snap

Runs commands (requires sudo):

```bash
sudo snap set system proxy.http="http://proxy.corp.com:8080"
sudo snap set system proxy.https="http://proxy.corp.com:8080"
sudo snap set system store-certs.ezproxy="$(cat /path/to/corp-ca.pem)"
```

CA cert support requires snapd 2.45+.

### apt

Writes `/etc/apt/apt.conf.d/99ezproxy` (requires sudo):

```
Acquire::http::Proxy "http://proxy.corp.com:8080/";
Acquire::https::Proxy "http://proxy.corp.com:8080/";
```

Note: trailing slash in URL is important.
CA: handled by system CA store (update-ca-certificates).

### yum/dnf

Writes proxy lines in `/etc/yum.conf` or `/etc/dnf/dnf.conf` (requires sudo):

```ini
proxy=http://proxy.corp.com:8080
sslcacert=/path/to/corp-ca.pem
```

Also handled by system CA store (update-ca-trust).

### SSH

Appends to `~/.ssh/config`:

```
# >>> ezproxy >>>
Host *
    ProxyCommand nc -X connect -x proxy.corp.com:8080 %h %p
# <<< ezproxy <<<
```

- macOS: built-in `nc` is OpenBSD netcat (supports `-X`/`-x`).
- Linux: requires `netcat-openbsd` package (GNU netcat does NOT support proxy flags).
- The tool detects which nc variant is installed and warns if GNU netcat is found.

### System CA Certificate

**macOS:**
```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain <cert>
```
Note: modern macOS may prompt for admin password interactively.

**Debian/Ubuntu:**
```bash
sudo cp <cert> /usr/local/share/ca-certificates/ezproxy-corp-ca.crt
sudo update-ca-certificates
```
File MUST have `.crt` extension. One cert per file.

**RHEL/Fedora:**
```bash
sudo cp <cert> /etc/pki/ca-trust/source/anchors/ezproxy-corp-ca.pem
sudo update-ca-trust extract
```
Any extension works (.pem, .crt, .cer).

**Arch:**
```bash
sudo trust anchor --store <cert>
```
Or copy to `/etc/ca-certificates/trust-source/anchors/` + `sudo update-ca-trust`.

## Architecture

```
ezproxy/
├── cmd/
│   └── ezproxy/
│       └── main.go              # CLI entrypoint (init, apply, remove, status)
├── internal/
│   ├── config/
│   │   └── config.go            # Load/save ~/.ezproxy/config.yaml
│   ├── configurator/
│   │   ├── configurator.go      # Configurator interface
│   │   ├── envvars.go
│   │   ├── git.go
│   │   ├── pip.go
│   │   ├── npm.go
│   │   ├── yarn.go              # Both v1 and Berry
│   │   ├── docker.go
│   │   ├── curl.go
│   │   ├── wget.go
│   │   ├── cargo.go
│   │   ├── conda.go
│   │   ├── brew.go
│   │   ├── snap.go
│   │   ├── apt.go
│   │   ├── yum.go
│   │   ├── ssh.go
│   │   └── systemca.go
│   ├── detect/
│   │   └── detect.go            # Detect OS, distro, installed tools
│   └── fileutil/
│       └── fileutil.go          # Marker-block insert/replace/remove
├── go.mod
└── go.sum
```

### Configurator Interface

```go
type Configurator interface {
    Name() string
    IsInstalled() bool
    Apply(cfg *Config) error
    Remove() error
    Status() (string, error)  // "configured", "not configured", "stale"
}
```

### Key Design Decisions

1. **Marker blocks** (`# >>> ezproxy >>>`/`# <<< ezproxy <<<`) for idempotent apply and clean remove in files we append to.

2. **Detect before configure** -- `detect.go` figures out macOS vs Linux, Debian vs RHEL vs Arch, which tools are installed. Irrelevant configurators are skipped silently.

3. **Sudo handling** -- Operations needing sudo (system CA, apt, yum, snap, docker daemon) clearly tell the user what needs sudo and prompt. The tool does not silently run sudo.

4. **`ezproxy status`** -- Shows a table of every tool: configured/not configured/stale, whether config matches current `config.yaml`.

5. **Docker Desktop on macOS** -- Prints instructions for GUI config rather than trying to automate it.

6. **Language: Go** -- Single binary, no runtime deps, fast startup, easy cross-compile for Linux/macOS amd64/arm64.

## Non-Goals

- No proxy authentication (NTLM/Kerberos) support
- No auto-detection of proxy availability
- No Windows support (Linux/macOS only)
- No daemon/service mode
