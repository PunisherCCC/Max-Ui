# Max-ui

Max-Ui is a customized VPN/proxy management panel based on 3X-UI. It keeps the familiar panel workflow, adds more VPN protocols, supports switching between Xray-core and Sing-box, and ships operator-focused controls for traffic accounting, admins, resellers, bulk actions, and server management.

## What Is Included

- Xray-core panel management.
- Database-driven Xray-core and Sing-box runtimes with validated, rollback-safe switching.
- Per-inbound bandwidth multiplier for quota/accounting only.
- Multi-admin support with per-inbound access.
- Reseller accounts with metered traffic balance.
- Bulk client and inbound operations.
- Account freeze/unfreeze.
- Link export as TXT and PDF.
- Improved web UI theme and improved Ubuntu/server-side `max-ui` menu.
- Stable Vue/Ant Design web runtime with the redesigned dark admin interface.
- GitHub Actions build and release pipeline.

## Supported Protocols

- Xray protocols such as VMess, VLESS, Trojan, Shadowsocks, Hysteria, WireGuard and others supported by the panel.
- PPTP
- L2TP RAW
- L2TP/IPsec
- OpenVPN
- OpenConnect
- SSTP
- IKEv2
- WireGuard (C)
- AmneziaWG
- MTProto Proxy
- SSH gateway

## Install

The installer downloads the latest `max-ui-amd64` release asset from this repository, installs the recommended Ubuntu/Debian runtime packages, and prepares the runtime binaries the panel expects under `/opt/max-ui/bin`.

```bash
curl -Ls https://raw.githubusercontent.com/PunisherCCC/Max-Ui/refs/heads/main/deploy.sh | sudo bash
```

After installation, open the server menu with:

```bash
max-ui
```

During installation Max-ui prepares:

- `bin/xray-linux-amd64` from the latest Xray-core release.
- `bin/sing-box-linux-amd64` from this project's latest release, compiled from upstream sing-box with the V2Ray statistics API required for client traffic accounting.
- `bin/telemt` for MTProto Proxy.
- Distro packages such as OpenVPN, strongSwan, xl2tpd, ocserv, nftables and iproute2 when `apt-get` is available.

Optional packages such as `pptpd` and `accel-ppp` are installed when your Ubuntu/Debian repository provides them.

## Uninstall

```bash
sudo /opt/max-ui/max-ui-amd64 --uninstall
```

## Core Modes

Max-Ui can run either:

- **Xray-core**: the default runtime.
- **Sing-box**: a generated runtime for VLESS, VMess, Trojan, Shadowsocks, Mixed, and HTTP inbounds.

The Sing-box template supplies log, DNS, route, experimental, and outbound settings. Max-Ui replaces its inbound list with enabled database inbounds, removes disabled or depleted clients, translates TLS/Reality and supported V2Ray transports, and enables per-inbound and per-user traffic statistics. The complete generated file is validated by the installed Sing-box binary before the current core is stopped.

Core switching is transactional: Max-Ui only saves the new selection after the target starts successfully. If startup fails, it restores the previous core and returns the validation/startup error to the panel.

The generated runtime file is written to:

```text
bin/sing-box.json
```

Expected Sing-box binary path:

```text
bin/sing-box-<GOOS>-<GOARCH>
```

For the Linux amd64 installer this is created as:

```text
/opt/max-ui/bin/sing-box-linux-amd64
```

Enabled protocols that cannot be translated safely, including Max-Ui's separate system VPN services, block a switch to Sing-box with an actionable error instead of being silently omitted. Disable those inbounds or keep Xray active. Traffic multipliers and user quotas use the same accounting pipeline under both cores.

## Bandwidth Multiplier

Each inbound can enable a multiplier that affects quota/accounting only.

Example:

- Multiplier: `2`
- Actual traffic: `1 GB`
- Counted quota usage: `2 GB`

The multiplier:

- Applies from the first byte.
- Applies equally to upload and download.
- Does not change actual throughput.
- Keeps lifetime `all_time` traffic as raw real usage.
- Applies to both client counters and inbound counters.

## GitHub Builds

GitHub Actions builds and smoke-checks the Linux amd64 binary on every push to `main`.

On version tags such as `v1.8.17`, the workflow creates a GitHub Release and uploads:

```text
max-ui-amd64
sing-box-linux-amd64
```

The install command uses the latest release asset.

## Web UI

The panel keeps the stable Vue runtime required by the current Ant Design Vue components and applies the new Max-ui dark dashboard theme on top of it. This avoids blank tables, raw controls, and half-rendered pages on production installs.

The active theme is defined in:

```text
web/assets/css/tokens.css
web/assets/css/components.css
```

## Build From Source

Requirements:

- Go
- Git
- Linux build environment
- CGO toolchain for SQLite

```bash
git clone https://github.com/PunisherCCC/Max-Ui.git
cd Max-Ui
go mod download
CGO_ENABLED=1 go build -o max-ui-amd64 main.go
./max-ui-amd64 -v
```

## Verification

The GitHub build workflow verifies:

- Go modules download correctly.
- Focused traffic multiplier regression tests pass.
- The Linux amd64 binary compiles.
- The compiled binary can print its version.
- Version tags publish both the Max-Ui panel and its tested, statistics-enabled sing-box runtime.

## Notes

- The Go module path remains compatible with upstream 3X-UI imports:

```text
github.com/mhsanaei/3x-ui/v2
```

- Some bundled runtime assets are generated during full release builds. This repository keeps placeholder files so normal GitHub compilation works from a clean checkout.
- For production use, test on a clean Ubuntu or Debian server before migrating live users.

## License

This project follows the license included in [LICENSE](LICENSE).
