from __future__ import annotations

import json
import os
import subprocess
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).with_name("resolve_source.py")


class ResolvePythonSourceTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.project = Path(self.temporary.name)
        self.site = self.project / ".venv" / "lib" / "python3.14" / "site-packages"
        self.site.mkdir(parents=True)

    def install(self, version: str = "1.2.3", suffix: str = "") -> Path:
        metadata = self.site / f"demo_pkg-{version}{suffix}.dist-info"
        metadata.mkdir()
        (metadata / "METADATA").write_text(f"Name: demo-pkg\nVersion: {version}\n\n", encoding="utf-8")
        return metadata

    def run_resolver(
        self, *arguments: str, environment: dict[str, str] | None = None
    ) -> subprocess.CompletedProcess[str]:
        clean_environment = {
            key: value for key, value in os.environ.items() if key not in {"VIRTUAL_ENV", "UV_PROJECT_ENVIRONMENT"}
        }
        clean_environment.update(environment or {})
        return subprocess.run(
            ["python3", str(SCRIPT), "demo-pkg", "Target", "--project", str(self.project), *arguments],
            env=clean_environment,
            check=False,
            capture_output=True,
            text=True,
        )

    def test_resolves_uv_environment_without_importing_package(self) -> None:
        self.install()
        package = self.site / "demo_pkg"
        package.mkdir()
        (package / "__init__.py").write_text("class Target:\n    value = 1\n", encoding="utf-8")

        result = self.run_resolver()

        self.assertEqual(result.returncode, 0, result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["schema"], "fmind.dev/dependency-source/v1")
        self.assertEqual(payload["version"], "1.2.3")
        self.assertEqual(payload["excerpt"]["text"], "class Target:\n    value = 1")
        self.assertFalse(payload["provenance"]["editable"])

    def test_resolves_editable_install(self) -> None:
        metadata = self.install()
        checkout = self.project / "checkout"
        source = checkout / "src" / "demo_pkg"
        source.mkdir(parents=True)
        (source / "core.py").write_text("def Target():\n    return 1\n", encoding="utf-8")
        (metadata / "direct_url.json").write_text(
            json.dumps({"url": checkout.as_uri(), "dir_info": {"editable": True}}), encoding="utf-8"
        )

        result = self.run_resolver()

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue(json.loads(result.stdout)["provenance"]["editable"])

    def test_rejects_ambiguous_environments(self) -> None:
        self.install()
        other = self.project / "other-venv"
        other.mkdir()

        result = self.run_resolver(environment={"VIRTUAL_ENV": str(other)})

        self.assertEqual(result.returncode, 2)
        self.assertIn("ambiguous Python environments", result.stderr)

    def test_rejects_multiple_versions(self) -> None:
        self.install()
        self.install("2.0.0", "-duplicate")

        result = self.run_resolver()

        self.assertEqual(result.returncode, 2)
        self.assertIn("multiple installed versions", result.stderr)

    def test_reports_absent_dependency(self) -> None:
        result = self.run_resolver()

        self.assertEqual(result.returncode, 2)
        self.assertIn("is absent", result.stderr)


if __name__ == "__main__":
    unittest.main()
