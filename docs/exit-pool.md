# Multi-country exit pool

The multi-country exit pool is an optional integration with
[`hanhaoops/xnvgatebox`](https://github.com/hanhaoops/xnvgatebox). It is disabled
by default and does not change existing inbounds, outbounds, routing rules, or
firewall rules until the administrator installs it and explicitly creates a
route binding.

## Install

Update 3x-ui to the latest development release, then run as root:

```bash
x-ui exit-pool install
```

The command downloads the pinned `xnvgatebox` release installer over HTTPS.
The upstream installer verifies its release package and checksum before
installation. Linux x86_64, systemd, TUN and network namespace support are
required. Containerized 3x-ui installations are not supported by this helper.

Open the exit-pool console:

```bash
x-ui exit-pool access
```

Use that console to refresh VPN Gate candidates, select a country, choose the
number of exits and probe concurrency, and manage VPS-direct or third-party
SOCKS5 exits. Wait until an exit reports ready before binding it in 3x-ui.

## Bind an inbound

Open **Xray configuration → Multi-country exit pool**. Ready exits registered
by the sidecar appear with a `fanout-` tag. Select one inbound and one exit,
create the binding, then use the page-level **Save** button.

Only that inbound is affected. The generated rule is restricted to TCP because
the upstream public exit path does not currently guarantee UDP, QUIC, IPv6,
Hysteria2 or TUIC forwarding. The sidecar's kill switch blocks a failed tunnel
instead of silently falling back to the VPS public IP.

## Operations

```bash
x-ui exit-pool status
x-ui exit-pool access
x-ui exit-pool update
x-ui exit-pool uninstall
```

The panel integration is currently published through this fork's `dev-latest`
release channel, the same channel used by `x-ui update-dev`.

Before uninstalling, remove exit-pool bindings in the panel and save the Xray
configuration. The integration only manages `fanout-*` outbounds and rules; it
does not modify unrelated user entries.

## Trust and licensing

VPN Gate exits are operated by third parties. They can disappear without
notice and should not be trusted with sensitive traffic. Country labels are
candidate hints; the sidecar verifies the connected public exit before marking
an exit ready.

`xnvgatebox` is distributed under GPL-3.0-only. This repository does not embed
its binary or source archive; the optional installer obtains a pinned upstream
release. Copyright, license and third-party notices remain with the upstream
project.

