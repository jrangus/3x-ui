"""3x-ui reviewed Pool Builder overlay for xnvgatebox v0.9.5."""
from __future__ import annotations

from pathlib import Path
import json
import time
from urllib.parse import urlsplit

from .common import ServiceError, ToolError, atomic_write, digest, get_bytes, utc_now, write_json
from .mirror_source import fetch_mirror
from .pool_history import merge_recent_official_history
from .source_fetch import SourceFetcher
from .vpngate import parse_csv


def build_pool(settings: dict, output: Path, csv_file=None, transport=get_bytes) -> dict:
    errors = []
    source_info = {"kind": "local_csv"}
    if csv_file:
        data = Path(csv_file).read_bytes()
        source = "local_csv"
        if len(data) > 12_000_000:
            raise ToolError("csv_too_large")
        nodes, profiles, rejected = parse_csv(data, settings)
    else:
        fetcher = SourceFetcher(transport)
        mirror = settings.get("mirror_fallback", {"enabled": True, "max_age_hours": 24})
        source_timeout = settings.get("pool_source_timeout_seconds", 10)
        primary_deadline = time.monotonic() + (30 if mirror["enabled"] else 60)
        for source in settings["source_urls"]:
            try:
                data = fetcher.fetch(source, min(source_timeout, max(0, primary_deadline - time.monotonic())))
                parsed = urlsplit(source)
                nodes, profiles, rejected = parse_csv(
                    data, settings, safe_profiles=parsed.hostname == "download.vpngate.jp"
                )
                if not nodes:
                    raise ServiceError("empty_source_pool")
                source_info = {
                    "kind": "official"
                    if parsed.hostname in ("www.vpngate.net", "vpngate.net", "download.vpngate.jp")
                    and parsed.path.rstrip("/") == "/api/iphone"
                    else "configured"
                }
                if source_info["kind"] == "official" and parsed.hostname == "download.vpngate.jp":
                    source_info["kind"] = "official_backup"
                break
            except (ServiceError, ToolError) as exc:
                errors.append(str(exc))
        else:
            if not mirror["enabled"]:
                raise ServiceError("pool_source_unavailable")
            previous = {}
            try:
                previous = json.loads((output / "node_pool.json").read_text()).get("source", {})
                if not isinstance(previous, dict):
                    previous = {}
            except (OSError, ValueError, AttributeError):
                pass
            data, source_info = fetch_mirror(fetcher, source_timeout, mirror["max_age_hours"], previous)
            source = source_info.pop("name")
            try:
                nodes, profiles, rejected = parse_csv(data, settings, safe_profiles=True)
            except ToolError:
                raise ServiceError("mirror_csv_invalid") from None
    if not nodes:
        raise ToolError("no_candidates_matching_filters")
    for sha, profile in profiles.items():
        atomic_write(output / "configs" / (sha + ".ovpn"), profile)

    current_count = len(nodes)
    # Local CSV and custom sources are deliberately excluded from rolling
    # retention; only prior official/integrity-verified mirror rows qualify.
    if source_info.get("kind") in {"official", "official_backup", "mirror"}:
        nodes, retained_count = merge_recent_official_history(settings, output, nodes)
    else:
        retained_count = 0
    generated_at = utc_now()
    pool = {
        "schema_version": 1,
        "generated_at": generated_at,
        "source": {
            "name": source,
            "sha256": digest(data),
            "rejected_rows": rejected,
            "errors": errors,
            "retrieval": "local_builder",
            "retrieved_at": generated_at,
            "current_candidates": current_count,
            "retained_candidates": retained_count,
            "candidate_total": len(nodes),
            **source_info,
        },
        "nodes": nodes,
    }
    save_pool(pool, output)
    return pool


def save_pool(pool: dict, output: Path):
    write_json(output / "node_pool.json", pool)
    atomic_write(
        output / "node_pool.json.sha256",
        (digest((output / "node_pool.json").read_bytes()) + "\n").encode(),
    )
