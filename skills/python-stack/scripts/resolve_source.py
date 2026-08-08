#!/usr/bin/env python3
"""Resolve installed Python source without importing dependency code."""

from __future__ import annotations

import argparse
import ast
import json
import os
import re
import sys
from pathlib import Path
from urllib.parse import unquote, urlparse

SCHEMA = "fmind.dev/dependency-source/v1"
MAX_EXCERPT_LINES = 40


class ResolutionError(RuntimeError):
    """An actionable source-resolution failure."""


def normalized(name: str) -> str:
    # Distribution names treat every run of '-', '_' and '.' as equivalent.
    # Underscores also produce the conventional import-module default.
    return re.sub(r"[-_.]+", "_", name).lower()


def selected_environment(project: Path, explicit: str | None) -> Path:
    candidates: list[tuple[str, Path]] = []
    if explicit:
        candidates.append(("--environment", Path(explicit).expanduser()))
    else:
        configured = os.environ.get("UV_PROJECT_ENVIRONMENT")
        if configured:
            path = Path(configured).expanduser()
            candidates.append(("UV_PROJECT_ENVIRONMENT", path if path.is_absolute() else project / path))
        elif (project / ".venv").is_dir():
            candidates.append(("project .venv", project / ".venv"))

        active = os.environ.get("VIRTUAL_ENV")
        if active:
            active_path = Path(active).expanduser()
            if not candidates or active_path.resolve() != candidates[0][1].resolve():
                candidates.append(("VIRTUAL_ENV", active_path))

    existing = [(origin, path.resolve()) for origin, path in candidates if path.is_dir()]
    if not existing:
        raise ResolutionError("no selected environment found; run `uv sync` or pass --environment")
    unique = {path for _, path in existing}
    if len(unique) != 1:
        choices = ", ".join(f"{origin}={path}" for origin, path in existing)
        raise ResolutionError(f"ambiguous Python environments ({choices}); pass --environment explicitly")
    return existing[0][1]


def site_packages(environment: Path) -> Path:
    candidates = sorted(environment.glob("lib/python*/site-packages"))
    windows = environment / "Lib" / "site-packages"
    if windows.is_dir():
        candidates.append(windows)
    if len(candidates) != 1:
        found = ", ".join(str(path) for path in candidates) or "none"
        raise ResolutionError(
            f"expected one site-packages directory in {environment}, found {found}; recreate the uv environment"
        )
    return candidates[0].resolve()


def distribution_metadata(site: Path, distribution: str) -> tuple[Path, dict[str, str]]:
    matches: list[tuple[Path, dict[str, str]]] = []
    for metadata_file in site.glob("*.dist-info/METADATA"):
        fields: dict[str, str] = {}
        for line in metadata_file.read_text(encoding="utf-8", errors="replace").splitlines():
            if not line:
                break
            key, separator, value = line.partition(":")
            if separator and key in {"Name", "Version"}:
                fields[key] = value.strip()
        if normalized(fields.get("Name", "")) == normalized(distribution):
            matches.append((metadata_file.parent, fields))
    if not matches:
        raise ResolutionError(f"distribution {distribution!r} is absent from {site}; run `uv sync --frozen`")
    if len(matches) > 1:
        versions = ", ".join(item[1].get("Version", "unknown") for item in matches)
        raise ResolutionError(
            f"multiple installed versions of {distribution!r} found ({versions}); recreate the environment"
        )
    return matches[0]


def source_root(site: Path, dist_info: Path, module: str) -> tuple[Path, bool]:
    module_path = Path(*module.split("."))
    direct_url = dist_info / "direct_url.json"
    if direct_url.is_file():
        payload = json.loads(direct_url.read_text(encoding="utf-8"))
        if payload.get("dir_info", {}).get("editable"):
            parsed = urlparse(payload.get("url", ""))
            if parsed.scheme != "file":
                raise ResolutionError("editable install has a non-file source URL; reinstall it from a local checkout")
            root = Path(unquote(parsed.path)).resolve()
            if not root.is_dir():
                raise ResolutionError(f"editable install source is stale: {root}; rerun `uv sync --frozen`")
            candidates = [root / module_path, root / "src" / module_path]
            sources = [path.resolve() for path in candidates if path.is_dir()]
            sources.extend(
                path.with_suffix(".py").resolve() for path in candidates if path.with_suffix(".py").is_file()
            )
            if len(sources) != 1:
                found = ", ".join(str(path) for path in sources) or "none"
                raise ResolutionError(
                    f"editable module {module!r} must resolve to one source under {root}, found {found}; pass --module"
                )
            return sources[0], True

    package = site / module_path
    single_file = package.with_suffix(".py")
    if package.is_dir():
        return package.resolve(), False
    if single_file.is_file():
        return single_file.resolve(), False
    raise ResolutionError(f"module {module!r} has no source under {site}; pass --module or reinstall the dependency")


def symbol_nodes(tree: ast.AST, symbol: str) -> list[ast.AST]:
    return [
        node
        for node in ast.walk(tree)
        if isinstance(node, (ast.ClassDef, ast.FunctionDef, ast.AsyncFunctionDef))
        and node.name == symbol
        or isinstance(node, (ast.Assign, ast.AnnAssign))
        and any(
            isinstance(target, ast.Name) and target.id == symbol
            for target in ([node.target] if isinstance(node, ast.AnnAssign) else node.targets)
        )
    ]


def defining_source(root: Path, symbol: str) -> tuple[Path, ast.AST, list[str]]:
    files = [root] if root.is_file() else sorted(root.rglob("*.py"))
    matches: list[tuple[Path, ast.AST, list[str]]] = []
    for path in files:
        lines = path.read_text(encoding="utf-8", errors="replace").splitlines()
        try:
            tree = ast.parse("\n".join(lines), filename=str(path))
        except SyntaxError:
            continue
        matches.extend((path, node, lines) for node in symbol_nodes(tree, symbol))
    if not matches:
        raise ResolutionError(f"symbol {symbol!r} was not found in {root}")
    if len(matches) > 1:
        locations = ", ".join(str(path) for path, _, _ in matches[:5])
        raise ResolutionError(f"symbol {symbol!r} is ambiguous ({locations}); pass a narrower --module")
    return matches[0]


def resolve(args: argparse.Namespace) -> dict[str, object]:
    project = Path(args.project).resolve()
    environment = selected_environment(project, args.environment)
    site = site_packages(environment)
    dist_info, metadata = distribution_metadata(site, args.distribution)
    root, editable = source_root(site, dist_info, args.module)
    path, node, lines = defining_source(root, args.symbol)
    start = node.lineno
    natural_end = getattr(node, "end_lineno", start)
    end = min(natural_end, start + MAX_EXCERPT_LINES - 1)
    return {
        "schema": SCHEMA,
        "language": "python",
        "dependency": args.distribution,
        "version": metadata.get("Version", "unknown"),
        "source_path": str(root),
        "defining_file": str(path.resolve()),
        "symbol": args.symbol,
        "excerpt": {"start_line": start, "end_line": end, "text": "\n".join(lines[start - 1 : end])},
        "provenance": {"environment": str(environment), "editable": editable, "replacement": None},
    }


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("distribution")
    result.add_argument("symbol")
    result.add_argument("--module", help="import module; defaults to the normalized distribution name")
    result.add_argument("--project", default=".")
    result.add_argument("--environment")
    return result


def main() -> int:
    args = parser().parse_args()
    args.module = args.module or normalized(args.distribution)
    try:
        print(json.dumps(resolve(args), indent=2))
    except (OSError, ValueError, ResolutionError) as error:
        print(f"error: {error}", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
