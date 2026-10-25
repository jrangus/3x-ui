from datetime import datetime, timedelta, timezone
import hashlib
import json
from pathlib import Path
import tempfile
import unittest

from pool_history import merge_recent_official_history


def candidate(host: str, country: str, profile: bytes, seen: datetime) -> dict:
    sha = hashlib.sha256(profile).hexdigest()
    return {
        "node_id": "vpngate-" + sha[:20],
        "hostname": host,
        "server_ip": "198.51.100.8",
        "advertised_country": country,
        "source_score": 100,
        "discovered_at": seen.isoformat(),
        "mode_a": {"status": "passed"},
        "protocols": {
            "openvpn": {
                "status": "passed",
                "profile_ref": f"configs/{sha}.ovpn",
                "profile_sha256": sha,
                "expected_exit_ip": "203.0.113.9",
                "egress": {"country": country},
            }
        },
    }


class PoolHistoryTest(unittest.TestCase):
    def test_retains_recent_bound_profile_and_clears_health(self):
        now = datetime.now(timezone.utc)
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            (output / "configs").mkdir()
            old = candidate("old.opengw.net", "JP", b"client\n", now - timedelta(hours=2))
            sha = old["protocols"]["openvpn"]["profile_sha256"]
            (output / "configs" / f"{sha}.ovpn").write_bytes(b"client\n")
            (output / "node_pool.json").write_text(
                json.dumps({"source": {"kind": "official"}, "nodes": [old]}), encoding="utf-8"
            )
            merged, retained = merge_recent_official_history({"countries": ["JP"]}, output, [], now=now)
            self.assertEqual(retained, 1)
            self.assertEqual(merged[0]["protocols"]["openvpn"]["status"], "not_tested")
            self.assertIsNone(merged[0]["protocols"]["openvpn"]["egress"])
            self.assertNotIn("mode_a", merged[0])

    def test_rejects_expired_wrong_country_and_changed_profile(self):
        now = datetime.now(timezone.utc)
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            (output / "configs").mkdir()
            old = candidate("old.opengw.net", "US", b"client\n", now - timedelta(hours=25))
            sha = old["protocols"]["openvpn"]["profile_sha256"]
            (output / "configs" / f"{sha}.ovpn").write_bytes(b"changed\n")
            (output / "node_pool.json").write_text(
                json.dumps({"source": {"kind": "official"}, "nodes": [old]}), encoding="utf-8"
            )
            merged, retained = merge_recent_official_history({"countries": ["JP"]}, output, [], now=now)
            self.assertEqual((merged, retained), ([], 0))

    def test_does_not_retain_custom_or_unverified_mirror_sources(self):
        now = datetime.now(timezone.utc)
        for source in ({"kind": "configured"}, {"kind": "mirror", "integrity_verified": False}):
            with self.subTest(source=source), tempfile.TemporaryDirectory() as directory:
                output = Path(directory)
                (output / "configs").mkdir()
                old = candidate("old.opengw.net", "JP", b"client\n", now)
                sha = old["protocols"]["openvpn"]["profile_sha256"]
                (output / "configs" / f"{sha}.ovpn").write_bytes(b"client\n")
                (output / "node_pool.json").write_text(
                    json.dumps({"source": source, "nodes": [old]}), encoding="utf-8"
                )
                self.assertEqual(
                    merge_recent_official_history({"countries": []}, output, [], now=now), ([], 0)
                )


if __name__ == "__main__":
    unittest.main()
