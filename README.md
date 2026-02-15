# ezproxy

One command to configure corporate proxy settings across all your dev tools.

If you've ever started at a new company and spent half a day configuring `git`, `npm`, `pip`, `docker`, and a dozen other tools to work behind a corporate proxy — this is for you.

## What it does

```
ezproxy init     # interactive setup wizard
ezproxy apply    # configure all your tools in one shot
```

That's it. ezproxy writes the correct proxy configuration to every tool it supports, handles CA certificates for SSL-inspecting proxies, and detects what's actually installed on your system so it only touches what's relevant.

## Install

```bash
# From source
go install github.com/HopperShell/ezproxy/cmd/ezproxy@latest

# Or build from repo
git clone https://github.com/HopperShell/ezproxy.git
cd ezproxy
make build
```

## Quick start

```bash
ezproxy init         # interactive setup
ezproxy apply        # configure everything
ezproxy status       # check what's configured
```

### Setup wizard

`ezproxy init` walks you through configuration with an interactive TUI:

```
  HTTP Proxy URL
  > http://proxy.corp.com:8080

  HTTPS Proxy URL
  > http://proxy.corp.com:8080

  NO_PROXY
  > localhost,127.0.0.1,.corp.com,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16

  CA Certificate Path
  >
```

Then select which tools to configure — installed tools are pre-selected, use space to toggle:

```
  Tools to configure
  Use arrow keys to navigate, space to toggle, enter to confirm.

  [x] env_vars
  [x] git
  [x] pip
  [x] npm
  [x] yarn
  [x] docker
  [ ] podman (not installed)
  [x] curl
  [x] wget
  [x] cargo
  [ ] conda (not installed)
  [x] go
  [ ] gradle (not installed)
  [ ] maven (not installed)
  [x] bundler
  [x] brew
  [ ] snap (not installed)
  [ ] apt (not installed)
  [ ] yum (not installed)
  [ ] ssh
  [x] system_ca
```

### Status view

```
$ ezproxy status

Proxy:    http://proxy.corp.com:8080
NO_PROXY: localhost,127.0.0.1,.corp.com,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
CA Cert:  ~/.ezproxy/corp-ca.pem

Tool           Status                       Available
────           ──────                       ─────────
system_ca      trusted by system            yes
env_vars       configured                   yes
git            configured                   yes
pip            configured                   yes
npm            configured                   yes
docker         configured                   yes
go             GOPRIVATE=github.com/corp/*  yes
gradle         skipped                      no (not installed)
ssh            disabled                     -
```

## Usage

```bash
# Interactive setup - prompts for proxy URL, cert, and tool selection
ezproxy init

# Apply proxy config to all enabled tools
ezproxy apply

# Preview what would change without modifying anything
ezproxy apply --dry-run
```

## Supported tools

| Tool | What gets configured |
|------|---------------------|
| **env_vars** | `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` in your shell profile |
| **git** | `git config --global http.proxy` + `http.sslCAInfo` |
| **pip** | `pip.conf` proxy and cert settings |
| **npm** | `.npmrc` proxy and cafile |
| **yarn** | `.yarnrc` / `.yarnrc.yml` (detects v1 vs v2+) |
| **docker** | `~/.docker/config.json` client proxy + daemon systemd override |
| **podman** | `~/.config/containers/containers.conf` proxy env |
| **curl** | `.curlrc` proxy setting |
| **wget** | `.wgetrc` proxy settings |
| **cargo** | `~/.cargo/config.toml` HTTP proxy |
| **conda** | `~/.condarc` proxy settings |
| **go** | GOPRIVATE/GONOSUMDB hints for corporate module hosts |
| **gradle** | `~/.gradle/gradle.properties` system proxy properties |
| **maven** | `~/.m2/settings.xml` proxy entries |
| **bundler** | `~/.bundle/config` SSL CA cert path |
| **brew** | Configured via env_vars (HOMEBREW_CURLRC) |
| **snap** | `snap set system proxy.*` |
| **apt** | `/etc/apt/apt.conf.d/99ezproxy` |
| **yum** | `/etc/yum.conf` or `/etc/dnf/dnf.conf` proxy lines |
| **ssh** | `~/.ssh/config` ProxyCommand (disabled by default) |
| **system_ca** | Installs CA cert into OS trust store (macOS Keychain / Linux ca-certificates) |
| **java_ca** | Imports CA cert into JVM trust store (`keytool -importcert` into `cacerts`) |

## Commands

```
ezproxy init              Interactive setup wizard
ezproxy apply             Apply proxy config to all enabled tools
ezproxy remove            Remove proxy config from all tools
ezproxy status            Show current config and tool status
ezproxy manage            Interactive tool manager (toggle tools on/off)
ezproxy enable <tool>     Enable a tool and apply its config
ezproxy disable <tool>    Disable a tool and remove its config
```

### Flags

```
--dry-run         Preview changes without modifying files
--yes, -y         Skip confirmations (for scripting/automation)
```

## Managing tools

All tools are enabled by default during `init` (except those not installed on your system). After setup, you can toggle individual tools:

```bash
# Interactive - opens a checkbox UI to toggle tools
ezproxy manage

# CLI - quick enable/disable
ezproxy disable ssh
ezproxy enable docker
```

## CA certificates

Corporate proxies that perform SSL inspection require their CA certificate to be trusted by each tool. ezproxy handles this automatically:

- **System trust store**: Checks if the cert is already trusted (common on enterprise machines where IT pushes certs via MDM). Skips install if found, prompts for `sudo` if not.
- **Per-tool certs**: Tools like pip, npm, git, and bundler that maintain their own cert stores get configured individually.

On macOS, ezproxy checks the System Keychain before attempting any install. On Linux, it uses Go's `x509.SystemCertPool()` to check the distro's CA bundle.

## Shell detection

ezproxy detects your shell via `$SHELL` and writes to the correct profile:

| Shell | Profile files |
|-------|--------------|
| zsh | `~/.zshrc`, `~/.zprofile` |
| bash | `~/.bashrc`, `~/.bash_profile` |
| fish | `~/.config/fish/conf.d/ezproxy.fish` |

## How it works

ezproxy uses **marker blocks** to manage its configuration in your dotfiles:

```
# >>> ezproxy >>>
export HTTP_PROXY=http://proxy.corp.com:8080
export HTTPS_PROXY=http://proxy.corp.com:8080
export NO_PROXY=localhost,127.0.0.1
# <<< ezproxy <<<
```

This means:
- `apply` is idempotent — run it as many times as you want
- `remove` cleanly strips only ezproxy's additions
- Your own config above/below the markers is never touched

## Config file

Stored at `~/.ezproxy/config.yaml`:

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
  docker: true
  ssh: false
  # ... etc
```

## Cross-platform

- **macOS** (Intel + Apple Silicon)
- **Linux** (Debian/Ubuntu, RHEL/Fedora/CentOS, Arch)

Build for all platforms:

```bash
make release
```

## License

MIT
