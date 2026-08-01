from __future__ import annotations

import json
import os
import subprocess
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).with_name("resolve_source.py")


class ResolveGoSourceTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.project = self.root / "project"
        self.dependency = self.root / "dependency"
        self.cache = self.root / "module-cache"
        self.project.mkdir()
        self.dependency.mkdir()
        self.cache.mkdir()
        (self.project / "go.mod").write_text(
            "module example.com/project\n\ngo 1.26\n\nrequire example.com/dependency v1.2.3\n\nreplace example.com/dependency => ../dependency\n",
            encoding="utf-8",
        )
        (self.dependency / "go.mod").write_text("module example.com/dependency\n\ngo 1.26\n", encoding="utf-8")
        (self.dependency / "dependency.go").write_text(
            "package dependency\n\ntype Target struct {\n\tValue int\n}\n\nfunc (Target) Method() int { return 1 }\n",
            encoding="utf-8",
        )

    def run_resolver(
        self, package: str = "example.com/dependency", symbol: str = "Target"
    ) -> subprocess.CompletedProcess[str]:
        environment = os.environ.copy()
        environment["GOMODCACHE"] = str(self.cache)
        environment["GOPRIVATE"] = "example.com/*"
        return subprocess.run(
            ["python3", str(SCRIPT), package, symbol, "--project", str(self.project)],
            env=environment,
            check=False,
            capture_output=True,
            text=True,
        )

    def test_resolves_replacement_and_selected_version(self) -> None:
        result = self.run_resolver()

        self.assertEqual(result.returncode, 0, result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["schema"], "fmind.dev/dependency-source/v1")
        self.assertEqual(payload["version"], "v1.2.3")
        self.assertEqual(payload["provenance"]["replacement"]["Dir"], str(self.dependency))
        self.assertTrue(payload["provenance"]["private"])

    def test_resolves_qualified_method(self) -> None:
        result = self.run_resolver(symbol="Target.Method")

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("func (Target) Method", json.loads(result.stdout)["excerpt"]["text"])

    def test_reports_absent_dependency_without_downloading(self) -> None:
        result = self.run_resolver(package="example.com/absent")

        self.assertEqual(result.returncode, 2)
        self.assertIn("without network access", result.stderr)


if __name__ == "__main__":
    unittest.main()
