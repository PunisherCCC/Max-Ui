# Max-ui

Max-ui is an extended VPN/proxy management panel based on 3X-UI. It keeps the familiar Xray-core workflow and adds more VPN protocols, reseller/admin features, optional Sing-box core support, and quota accounting controls for operators who need more than a basic inbound manager.

> This repository is a customized fork intended to be published as **Max-ui**.

## Highlights

- Xray-core management with the existing inbound/outbound UI.
- Optional Sing-box core mode from a configurable JSON template.
- Per-inbound configurable bandwidth multiplier for quota accounting.
- Multi-admin access control with per-inbound ownership.
- Reseller accounts with metered traffic balance.
- Bulk operations for clients and inbounds.
- Export account links as TXT and PDF.
- Account freeze/unfreeze support.
- Built-in and bundled support for additional VPN protocols.

## Supported Protocols

- VMess, VLESS, Trojan, Shadowsocks, Hysteria, WireGuard and other Xray-supported protocols.
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

## Core Selection

Max-ui can start either:

- **Xray-core**: the fully integrated default core.
- **Sing-box**: an alternate core started from the Sing-box template in panel settings.

Sing-box support is designed as a guarded alternate runtime. Max-ui validates the Sing-box JSON template before starting the process. Automatic conversion from every Xray inbound to Sing-box JSON is not included yet.

Expected Sing-box binary location:

```text
bin/sing-box-<GOOS>-<GOARCH>
```

Generated Sing-box config location:

```text
bin/sing-box.json
```

## Bandwidth Multiplier

Each inbound can enable a bandwidth multiplier for accounting only.

Example:

- Multiplier: `2`
- Actual traffic: `1 GB`
- Counted quota usage: `2 GB`

The multiplier:

- Applies from the first byte.
- Applies equally to upload and download.
- Affects quota/accounting counters only.
- Does not throttle or increase actual network throughput.
- Keeps lifetime `all_time` traffic as raw real usage.

## Install

```bash
curl -Ls https://raw.githubusercontent.com/PunisherCCC/Max-Ui/refs/heads/main/deploy.sh | sudo bash
```

After installation, open the management menu with:

```bash
max-ui
```

## Uninstall

```bash
sudo /opt/max-ui/max-ui-amd64 --uninstall
```

## Build From Source

Requirements:

- Go
- Git
- A Linux build environment for production binaries

Clone and build:

```bash
git clone https://github.com/PunisherCCC/Max-Ui.git
cd max-ui
go mod download
go build -o max-ui
```

## Development Notes

The Go module path currently remains compatible with the upstream 3X-UI package path:

```text
github.com/mhsanaei/3x-ui/v2
```

This keeps imports stable while the project is being renamed and customized.

## Recent Custom Changes

- Project branding changed from vpn-ui to Max-ui.
- Added Sing-box process support beside Xray-core.
- Added panel setting to switch active proxy core.
- Added Sing-box template configuration in Xray/core settings.
- Updated traffic multiplier behavior to be per-inbound accounting from byte zero.
- Removed the old multiplier threshold field from the active inbound form.

## Tested Systems

This project is intended for modern Linux servers. The original panel targets Debian, Ubuntu, Fedora, AlmaLinux, Rocky Linux, CentOS Stream, and Arch Linux variants, but VPN daemon behavior can vary by kernel and distribution.

For production use, test on a clean server before migrating live users.

## License

This project follows the license included in [LICENSE](LICENSE).
