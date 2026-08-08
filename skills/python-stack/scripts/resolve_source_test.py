from __future__ import annotations

import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).with_name("resolve_source.py")
STACK = SCRIPT.parent.parent
SCRIPT_SKILL = STACK.parent / "python-script"


class ResolvePythonSourceTest(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.project = Path(self.temporary.name)
        self.site = self.project / ".venv" / "lib" / "python3.14" / "site-packages"
        self.site.mkdir(parents=True)

    def install(self, version: str = "1.2.3", suffix: str = "", name: str = "demo-pkg") -> Path:
        metadata = self.site / f"demo_pkg-{version}{suffix}.dist-info"
        metadata.mkdir()
        (metadata / "METADATA").write_text(f"Name: {name}\nVersion: {version}\n\n", encoding="utf-8")
        return metadata

    def run_resolver(
        self, *arguments: str, environment: dict[str, str] | None = None
    ) -> subprocess.CompletedProcess[str]:
        clean_environment = {
            key: value for key, value in os.environ.items() if key not in {"VIRTUAL_ENV", "UV_PROJECT_ENVIRONMENT"}
        }
        clean_environment.update(environment or {})
        return subprocess.run(
            [sys.executable, str(SCRIPT), "demo-pkg", "Target", "--project", str(self.project), *arguments],
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

    def test_distribution_name_normalization_collapses_separator_runs(self) -> None:
        self.install(name="Demo..Pkg")
        package = self.site / "demo_pkg"
        package.mkdir()
        (package / "__init__.py").write_text("class Target:\n    pass\n", encoding="utf-8")

        result = self.run_resolver()

        self.assertEqual(result.returncode, 0, result.stderr)

    def test_resolves_generated_source(self) -> None:
        self.install()
        package = self.site / "demo_pkg"
        package.mkdir()
        (package / "__init__.py").write_text(
            "# Generated file\n# DO NOT EDIT\nclass Target:\n    pass\n", encoding="utf-8"
        )

        result = self.run_resolver()

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("class Target", json.loads(result.stdout)["excerpt"]["text"])

    def test_rejects_stale_editable_source(self) -> None:
        metadata = self.install()
        missing = self.project / "missing-checkout"
        (metadata / "direct_url.json").write_text(
            json.dumps({"url": missing.as_uri(), "dir_info": {"editable": True}}), encoding="utf-8"
        )

        result = self.run_resolver()

        self.assertEqual(result.returncode, 2)
        self.assertIn("editable install source is stale", result.stderr)

    def test_rejects_ambiguous_symbol(self) -> None:
        self.install()
        package = self.site / "demo_pkg"
        package.mkdir()
        (package / "one.py").write_text("class Target:\n    pass\n", encoding="utf-8")
        (package / "two.py").write_text("def Target():\n    return None\n", encoding="utf-8")

        result = self.run_resolver()

        self.assertEqual(result.returncode, 2)
        self.assertIn("symbol 'Target' is ambiguous", result.stderr)


class PythonTemplateContractTest(unittest.TestCase):
    def read(self, relative: str) -> str:
        return (STACK / relative).read_text(encoding="utf-8")

    def test_default_test_gate_measures_coverage_and_excludes_integration(self) -> None:
        mise = self.read("references/mise.toml")
        lefthook = self.read("references/lefthook.yml")

        self.assertIn('uv run pytest -m "not integration" --cov --cov-fail-under=85', mise)
        self.assertIn('[tasks."test:integration"]', mise)
        self.assertIn("uv run pytest -m integration", mise)
        self.assertIn('run = "gitleaks dir --verbose ."', mise)
        self.assertNotIn("git rev-parse", mise)
        self.assertNotIn("check:leaks:staged", mise)
        self.assertNotIn("check:leaks:staged", lefthook)
        self.assertIn('depends = ["test"]\nrun = "uv run coverage html"', mise)
        self.assertIn("pip-audit --skip-editable", mise)
        self.assertIn("--cache-dir .cache/pip-audit", mise)

    def test_postgres_is_an_opt_in_fixture(self) -> None:
        conftest = self.read("references/conftest.py")
        smoke = self.read("references/test_smoke.py")
        integration = self.read("references/test_integration.py")

        self.assertNotIn("pytest_configure", conftest)
        self.assertNotIn("pytest_unconfigure", conftest)
        self.assertIn('scope="session"', conftest)
        self.assertNotIn("PostgresContainer", smoke)
        self.assertIn("@pytest.mark.integration", integration)

    def test_web_template_uses_safe_direct_sqlalchemy_configuration(self) -> None:
        project = self.read("references/pyproject.toml.template")
        application = self.read("references/init.py")
        environment = self.read("references/env.example")

        self.assertNotIn('"advanced-alchemy', project)
        self.assertNotIn("advanced_alchemy", application)
        self.assertIn("async_sessionmaker", application)
        self.assertIn('dependencies={"db_session": Provide(', application)
        self.assertIn('@get("/health")', application)
        self.assertIn('@get("/ready")', application)
        self.assertNotIn('allow_origins=["*"]', application)
        self.assertIn("cors_origins: list[str]", application)
        self.assertRegex(application, r"(?m)^    database_url: SecretStr$", "DATABASE_URL must be required")
        self.assertIn("DATABASE_URL=postgresql+asyncpg://", environment)

    def test_scaffold_has_a_module_entrypoint_for_the_import_package(self) -> None:
        module_entrypoint = self.read("references/main.py")

        self.assertIn("from . import main", module_entrypoint)

    def test_library_and_cli_profiles_have_honest_runtime_surfaces(self) -> None:
        project = self.read("references/pyproject.toml.template")
        library = self.read("references/init-library.py")
        cli = self.read("references/init-cli.py")
        cli_test = self.read("references/test_cli.py")
        skill = self.read("SKILL.md")

        self.assertNotIn('"typer>=', project)
        self.assertNotIn("def main", library)
        self.assertIn("import typer", cli)
        self.assertIn("Annotated[str, typer.Option", cli)
        self.assertIn("CliRunner", cli_test)
        self.assertIn("Library: remove `[project.scripts]`", skill)
        self.assertIn("CLI: add `typer`", skill)
        self.assertIn("Data, ML, and notebooks are extension profiles", skill)

    def test_agent_workflow_is_uvx_offline_first_and_normalizes_generated_typing(self) -> None:
        skill = self.read("SKILL.md")

        self.assertIn("uvx google-agents-cli install", skill)
        self.assertIn("uvx google-agents-cli playground", skill)
        self.assertIn("tests/unit", skill)
        self.assertIn("explicit approval", skill)
        self.assertIn("blanket `[tool.ty.rules]`", skill)

    def test_script_template_disables_local_rendering_and_documents_locking(self) -> None:
        template = (SCRIPT_SKILL / "references/script.py").read_text(encoding="utf-8")
        skill = (SCRIPT_SKILL / "SKILL.md").read_text(encoding="utf-8")

        self.assertIn("pretty_exceptions_show_locals=False", template)
        self.assertIn("show_locals=False", template)
        self.assertNotIn("show_locals=True", template)
        self.assertIn("uv lock --script", skill)


if __name__ == "__main__":
    unittest.main()
