"""Bounded, data-only retention for recently seen official VPN Gate candidates.

This module never marks a candidate healthy.  xnvgatebox still establishes an
OpenVPN tunnel and independently verifies its exit IP, country and policy before
3x-ui can expose the exit.  Retention only prevents a valid official profile
from disappearing immediately when two public CSV snapshots differ.
"""
from __future__ import annotations

from copy import deepcopy
from datetime import datetime, timezone
from pathlib import Path
import hashlib
import json
import os
import re
import stat


MAX_POOL_BYTES = 16 << 20
MAX_PROFILE_BYTES = 180_000
DEFAULT_HISTORY_HOURS = 24
DEFAULT_TOTAL_LIMIT = 512
_SHA256 = re.compile(r"[a-f0-9]{64}")
_HOST = re.compile(r"[a-z0-9](?:[a-z0-9.-]*[a-z0-9])?")
_TRUSTED_SOURCE_KINDS = {"official", "official_backup", "mirror"}


def _timestamp(value: object) -> datetime | None:
    if not isinstance(value, str):
        return None
    try:
        parsed = datetime.fromisoformat(value.replace("Z", "+00:00"))
    except ValueError:
        return None
    if parsed.tzinfo is None:
        return None
    return parsed.astimezone(timezone.utc)


def _regular_file(path: Path, maximum: int) -> bytes | None:
    try:
        info = path.lstat()
        if path.is_symlink() or not path.is_file() or info.st_size < 1 or info.st_size > maximum:
            return None
        flags = os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0)
        descriptor = os.open(path, flags)
        try:
            opened = os.fstat(descriptor)
            if not stat.S_ISREG(opened.st_mode) or opened.st_size < 1 or opened.st_size > maximum:
                return None
            chunks = []
            remaining = maximum + 1
            while remaining:
                chunk = os.read(descriptor, remaining)
                if not chunk:
                    break
                chunks.append(chunk)
                remaining -= len(chunk)
            data = b"".join(chunks)
        finally:
            os.close(descriptor)
        return data if 0 < len(data) <= maximum else None
    except OSError:
        return None


def _load_previous(output: Path) -> dict | None:
    raw = _regular_file(output / "node_pool.json", MAX_POOL_BYTES)
    if raw is None:
        return None
    try:
        value = json.loads(raw)
    except (UnicodeError, ValueError, TypeError):
        return None
    return value if isinstance(value, dict) else None


def _profile_is_bound(output: Path, node: dict) -> bool:
    try:
        openvpn = node["protocols"]["openvpn"]
        sha = openvpn["profile_sha256"]
        profile_ref = openvpn["profile_ref"]
    except (KeyError, TypeError):
        return False
    if not isinstance(sha, str) or not _SHA256.fullmatch(sha):
        return False
    if profile_ref != f"configs/{sha}.ovpn":
        return False
    raw = _regular_file(output / "configs" / f"{sha}.ovpn", MAX_PROFILE_BYTES)
    return raw is not None and hashlib.sha256(raw).hexdigest() == sha


def merge_recent_official_history(
    settings: dict,
    output: Path,
    current_nodes: list[dict],
    *,
    now: datetime | None = None,
    history_hours: int = DEFAULT_HISTORY_HOURS,
    total_limit: int = DEFAULT_TOTAL_LIMIT,
) -> tuple[list[dict], int]:
    """Append bounded recent candidates from the previous trusted snapshot.

    Returned historical rows remain ``not_tested`` candidates.  This is
    intentional: the runtime's existing full-chain checks remain the only path
    to a selectable exit.
    """
    previous = _load_previous(output)
    if previous is None:
        return list(current_nodes), 0
    source = previous.get("source")
    if not isinstance(source, dict) or source.get("kind") not in _TRUSTED_SOURCE_KINDS:
        return list(current_nodes), 0
    if source.get("kind") == "mirror" and source.get("integrity_verified") is not True:
        return list(current_nodes), 0

    now = (now or datetime.now(timezone.utc)).astimezone(timezone.utc)
    allowed = {
        value.upper()
        for value in settings.get("countries", [])
        if isinstance(value, str) and re.fullmatch(r"[A-Za-z]{2}", value)
    }
    nodes = list(current_nodes)
    seen_hosts = {
        str(node.get("hostname", "")).lower()
        for node in nodes
        if isinstance(node, dict)
    }
    candidates: list[tuple[datetime, int, dict]] = []
    for node in previous.get("nodes", []):
        if not isinstance(node, dict):
            continue
        hostname = str(node.get("hostname", "")).lower()
        country = str(node.get("advertised_country", "")).upper()
        seen_at = _timestamp(node.get("discovered_at"))
        if (
            not _HOST.fullmatch(hostname)
            or hostname in seen_hosts
            or not re.fullmatch(r"[A-Z]{2}", country)
            or (allowed and country not in allowed)
            or seen_at is None
        ):
            continue
        age = (now - seen_at).total_seconds()
        if age < -300 or age > history_hours * 3600 or not _profile_is_bound(output, node):
            continue
        retained = deepcopy(node)
        # Never carry a prior health claim into a new refresh.  The Go runtime
        # will perform the same live checks required for a current CSV row.
        for protocol in retained.get("protocols", {}).values():
            if isinstance(protocol, dict):
                protocol["status"] = "not_tested"
                protocol["expected_exit_ip"] = None
                protocol["egress"] = None
        retained.pop("mode_a", None)
        score = retained.get("source_score")
        score = score if isinstance(score, int) else 0
        candidates.append((seen_at, score, retained))

    candidates.sort(key=lambda item: (item[0], item[1]), reverse=True)
    room = max(0, min(total_limit, DEFAULT_TOTAL_LIMIT) - len(nodes))
    retained_nodes = [item[2] for item in candidates[:room]]
    return nodes + retained_nodes, len(retained_nodes)
