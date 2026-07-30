# Max-ui

Max-ui is a customized VPN/proxy management panel based on 3X-UI. It keeps the familiar Xray-core panel workflow, adds more VPN protocols, includes an optional Sing-box runtime mode, and ships operator-focused controls for traffic accounting, admins, resellers, bulk actions, and server management.

## What Is Included

- Xray-core panel management.
- Optional Sing-box core runner from a validated JSON template.
- Per-inbound bandwidth multiplier for quota/accounting only.
- Multi-admin support with per-inbound access.
- Reseller accounts with metered traffic balance.
- Bulk client and inbound operations.
- Account freeze/unfreeze.
- Link export as TXT and PDF.
- Improved web UI theme and improved Ubuntu/server-side `max-ui` menu.
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
- `bin/sing-box-linux-amd64` from the latest Sing-box release.
- `bin/telemt` for MTProto Proxy.
- Distro packages such as OpenVPN, strongSwan, xl2tpd, ocserv, nftables and iproute2 when `apt-get` is available.

Optional packages such as `pptpd` and `accel-ppp` are installed when your Ubuntu/Debian repository provides them.

## Uninstall

```bash
sudo /opt/max-ui/max-ui-amd64 --uninstall
```

## Core Modes

Max-ui can run either:

- **Xray-core**: the default and fully integrated core.
- **Sing-box**: an alternate runtime started from the Sing-box template in panel settings.

Sing-box mode validates the JSON template before start and writes it to:

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

Current scope: Sing-box is a guarded alternate runtime. Automatic conversion of every Xray inbound into Sing-box JSON is not included yet.

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

On version tags such as `v1.8.8`, the workflow creates a GitHub Release and uploads:

```text
max-ui-amd64
```

The install command uses the latest release asset.

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
- A release asset is published for version tags.

## Notes

- The Go module path remains compatible with upstream 3X-UI imports:

```text
github.com/mhsanaei/3x-ui/v2
```

- Some bundled runtime assets are generated during full release builds. This repository keeps placeholder files so normal GitHub compilation works from a clean checkout.
- For production use, test on a clean Ubuntu or Debian server before migrating live users.

## License

This project follows the license included in [LICENSE](LICENSE).
