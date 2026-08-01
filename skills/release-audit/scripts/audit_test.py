from __future__ import annotations

import importlib.util
import json
import unittest
from pathlib import Path

SCRIPT = Path(__file__).with_name("audit.py")
FIXTURES = Path(__file__).with_name("fixtures")
SPEC = importlib.util.spec_from_file_location("release_audit", SCRIPT)
assert SPEC and SPEC.loader
AUDIT = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(AUDIT)


class ReleaseAuditTest(unittest.TestCase):
    def test_required_release_states(self) -> None:
        fixtures = {
            "stale-ci.json",
            "moved-tag.json",
            "draft-release.json",
            "missing-assets.json",
            "valid-source-only.json",
        }
        self.assertEqual({path.name for path in FIXTURES.glob("*.json")}, fixtures)
        for name in fixtures:
            with self.subTest(name=name):
                fixture = json.loads((FIXTURES / name).read_text(encoding="utf-8"))
                AUDIT.validate_contract(fixture["contract"])
                report = AUDIT.evaluate("v1.2.3", fixture["contract"], fixture["evidence"])
                self.assertEqual(report["schema"], "fmind.dev/release-audit-report/v1")
                self.assertEqual(report["state"], fixture["expected_state"])
                self.assertIn("release audit:", AUDIT.human(report))

    def test_contract_requires_explicit_artifact_declaration(self) -> None:
        with self.assertRaisesRegex(AUDIT.AuditError, "artifacts must be a list"):
            AUDIT.validate_contract({"schema": AUDIT.CONTRACT_SCHEMA, "workflows": []})


if __name__ == "__main__":
    unittest.main()
