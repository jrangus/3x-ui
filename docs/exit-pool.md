# Multi-country exit subscription groups

This optional integration combines 3x-ui with
[`hanhaoops/xnvgatebox`](https://github.com/hanhaoops/xnvgatebox). It is disabled
by default. Existing clients and routes remain unchanged until an administrator
creates an exit-pool subscription group.

## Install the exit engine

Run as root on a Linux x86_64 VPS:

```bash
x-ui exit-pool install
```

The installer pins a reviewed xnvgatebox release and verifies the downloaded
package. Systemd, TUN and network namespaces are required. Docker deployments
are not supported by this helper.

Use `x-ui exit-pool access` only when you need to change candidate countries,
the number of exits, probe settings, or third-party SOCKS resources. Day-to-day
node assignment is performed in the 3x-ui panel.

## Create one subscription containing several regions

1. In xnvgatebox, select the required countries and wait until their exits show
   `up`.
2. Open **Xray configuration → Multi-country exit pool** in 3x-ui.
3. Choose one existing local VLESS, VMess or Trojan inbound as the mother node.
4. Select several ready regional exits, enter a group name, and click **Create**.
5. Open the 3x-ui **Clients** page, search for the displayed subscription ID,
   and copy its subscription URL.

3x-ui creates one dedicated client identity per selected region inside the same
inbound. Those identities share one `subId`, so one subscription contains
separate US, JP, SG, or other regional entries while they all use the mother
node's existing public address and port. Xray routes each identity to its own
local xnvgatebox SOCKS exit. No extra inbound or public port is created.

Use **Edit regions** to add or remove countries from an existing group. Deleting
a group removes only its generated clients, routing rules and unused managed
outbounds.

## Network behavior

The public VPN Gate path is TCP-only. Managed rules capture both TCP and UDP for
their generated identities: TCP is forwarded through the selected exit, while
unsupported UDP fails at that exit instead of silently falling back to the VPS
public IP. Hysteria2, TUIC, QUIC and IPv6 are therefore not supported by these
groups.

The integration reads xnvgatebox only through its authenticated loopback API.
SOCKS credentials are never returned to the browser.

## Operations

```bash
x-ui exit-pool status
x-ui exit-pool access
x-ui exit-pool update
x-ui exit-pool uninstall
```

Delete exit-pool subscription groups in the panel before uninstalling the exit
engine. VPN Gate exits are operated by third parties, can disappear without
notice, and should not be trusted with sensitive unencrypted traffic.

`xnvgatebox` is GPL-3.0-only and remains a separately distributed component.
The helper downloads its pinned upstream release rather than embedding the
binary in 3x-ui.

