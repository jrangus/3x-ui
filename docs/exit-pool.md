# Multi-country exit subscription groups

This optional integration combines 3x-ui with
[`hanhaoops/xnvgatebox`](https://github.com/hanhaoops/xnvgatebox). The companion
is installed automatically on supported hosts, but it does not alter existing
clients or routes until an administrator creates an exit-pool subscription group.

## Install the exit engine

The normal 3x-ui one-line installer installs this component automatically on
Linux x86_64/systemd hosts. For an existing installation, or to retry after a
temporary network/dependency failure, run:

```bash
x-ui exit-pool install
```

The installer pins a reviewed xnvgatebox release, its installer commit, and two
independent SHA-256 values stored in 3x-ui. It refuses installation if either
download differs from the reviewed artifacts. Systemd, TUN and network
namespaces are required. Docker deployments are not supported by this helper.

Keep at least 1.5 GB free for a new combined installation, or 1 GB for an
in-place upgrade. Typical final amd64 disk use is about 250–350 MB for 3X-UI
plus 100–250 MB for xnvgatebox and its Ubuntu/Debian dependencies. Temporary
downloads, extraction, package caches, logs and databases require additional
headroom; CentOS can require more while compiling its SSTP package.

Installation does not loop. The upstream installer takes an exclusive file
lock, and 3X-UI restarts the exit service once after installation to apply its
loopback and systemd hardening. Messages such as `restoring tunnels`, failed
candidate probes and exit replacement are runtime health checks. They can
continue for a while when dozens of slots already exist, but they do not mean
the installer is running again. Do not rerun `install` or `secure` in response
to those messages.

Use `x-ui exit-pool access` only when you need to change candidate countries,
the number of exits, probe settings, or third-party SOCKS resources. Day-to-day
node assignment is performed in the 3x-ui panel.

## Candidate-pool capacity and quality boundary

The public source commonly contains only about 95–100 rows at one instant; the
previous 96-result ceiling was therefore source supply, not a hard-coded limit.
This candidate build keeps valid rows from recent official snapshots for up to
24 hours. History is added only while the combined pool is below 512; a larger
live official snapshot is never truncated. A retained row is accepted only when
its OpenVPN profile is still a regular file under the pool directory and its
path, size and SHA-256 digest all match. Custom sources and unverified mirror
snapshots are never added to the rolling history.

Retention increases the number of candidates, not the number of trusted exits.
Every retained candidate is reset to `not_tested`; xnvgatebox must still create
the tunnel and pass its existing live connectivity, actual exit-IP, country and
policy checks before 3x-ui displays it as an available exit. Thus an offline
historical row may fail and disappear, but it cannot silently become a usable
lower-quality exit or fall back to the VPS address.

After installing this candidate build, an existing exit-pool installation can
apply the overlay and request an immediate pool refresh with:

```bash
sudo x-ui exit-pool optimize
```

Growth happens across changing official snapshots, so the first refresh can
still show roughly the same count. Source metadata in `node_pool.json` records
`current_candidates`, `retained_candidates`, and `candidate_total` separately.

## Create one subscription containing several regions

1. In xnvgatebox, create the required exits and wait until they show `up`.
2. Open **Xray configuration → Multi-country exit pool** in 3x-ui.
3. Use **Actual exit country** to filter by the country measured through the
   running tunnel. This does not rely on the server's advertised country.
4. Choose one existing local VLESS, VMess or Trojan inbound as the mother node.
5. Select several ready regional exits, enter a group name, and click **Create**.
6. Open the 3x-ui **Clients** page, search for the displayed subscription ID,
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

3x-ui fails closed if `listen_addr` is not exactly `127.0.0.1`. The embedded
console cannot change xnvgatebox network settings or invoke its upstream online
updater. A systemd drop-in also applies process hardening while retaining the
network-namespace capabilities required to create exits. The management HTTP
port must not be opened in UFW, firewalld, security groups, or provider firewall
rules.

## Operations

```bash
x-ui exit-pool status
x-ui exit-pool access
x-ui exit-pool update
x-ui exit-pool optimize
x-ui exit-pool uninstall
```

`x-ui exit-pool update` reinstalls the exact xnvgatebox build reviewed and
pinned by the installed 3x-ui version. It deliberately does not follow a mutable
"latest" release. Upgrade 3x-ui first when a newer reviewed exit-engine version
is published, then run the command above.

After installing or updating, verify the boundary with:

```bash
x-ui exit-pool secure
systemctl is-active xnvgatebox-fanout.service
sudo ss -lntp | grep xnvgatebox
```

The management listener should appear only on `127.0.0.1`, never `0.0.0.0` or
`[::]`.

Delete exit-pool subscription groups in the panel before uninstalling the exit
engine. VPN Gate exits are operated by third parties, can disappear without
notice, and should not be trusted with sensitive unencrypted traffic.

`xnvgatebox` is GPL-3.0-only and remains a separately distributed component.
The helper downloads its pinned upstream release rather than embedding the
binary in 3x-ui.
