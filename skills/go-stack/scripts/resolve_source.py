#!/usr/bin/env python3
"""Resolve selected Go module source without downloading dependencies."""

from __future__ import annotations

import argparse
import fnmatch
import json
import os
import re
import subprocess
import sys
from pathlib import Path

SCHEMA = "fmind.dev/dependency-source/v1"
MAX_EXCERPT_LINES = 40


class ResolutionError(RuntimeError):
    """An actionable source-resolution failure."""


def go(command: list[str], project: Path) -> str:
    environment = os.environ.copy()
    environment.update({"GOPROXY": "off", "GOSUMDB": "off", "GOTOOLCHAIN": "local"})
    result = subprocess.run(
        ["go", *command],
        cwd=project,
        env=environment,
        check=False,
        capture_output=True,
        text=True,
    )
    if result.returncode:
        detail = result.stderr.strip() or result.stdout.strip()
        raise ResolutionError(
            f"`go {' '.join(command)}` failed without network access: {detail}; sync the module cache explicitly"
        )
    return result.stdout


def go_environment(project: Path) -> dict[str, str]:
    values = json.loads(go(["env", "-json", "GOMOD", "GOMODCACHE", "GOPRIVATE"], project))
    module_file = values.get("GOMOD", "")
    if not module_file or module_file == os.devnull or not Path(module_file).is_file():
        raise ResolutionError(f"{project} is not inside a Go module; select a project containing go.mod")
    cache = Path(values.get("GOMODCACHE", ""))
    if not cache.is_dir():
        raise ResolutionError(
            f"Go module cache is absent at {cache}; download dependencies explicitly before inspection"
        )
    return values


def package_metadata(project: Path, package: str) -> dict[str, object]:
    payload = json.loads(go(["list", "-mod=readonly", "-json", package], project))
    if payload.get("Error"):
        raise ResolutionError(f"package {package!r} is unresolved: {payload['Error']}")
    module = payload.get("Module")
    if not isinstance(module, dict):
        raise ResolutionError(f"package {package!r} belongs to the standard library, not a project dependency")
    return payload


def private_module(path: str, patterns: str) -> bool:
    for pattern in patterns.split(","):
        pattern = pattern.strip()
        if pattern and (fnmatch.fnmatch(path, pattern) or fnmatch.fnmatch(path, f"{pattern}/*")):
            return True
    return False


def source_module(module: dict[str, object]) -> tuple[dict[str, object], dict[str, object] | None]:
    replacement = module.get("Replace")
    selected = replacement if isinstance(replacement, dict) else module
    if not isinstance(selected, dict) or not selected.get("Dir"):
        path = module.get("Path", "unknown")
        version = module.get("Version", "unknown")
        raise ResolutionError(f"source for {path}@{version} is missing from the configured module cache")
    return selected, replacement if isinstance(replacement, dict) else None


def definition_pattern(symbol: str) -> re.Pattern[str]:
    receiver, separator, name = symbol.rpartition(".")
    escaped = re.escape(name if separator else symbol)
    if receiver:
        escaped_receiver = re.escape(receiver)
        return re.compile(rf"^func\s+\([^)]*\*?{escaped_receiver}(?:\[[^]]+\])?\)\s+{escaped}\s*\(")
    return re.compile(rf"^(?:func\s+(?:\([^)]*\)\s+)?{escaped}\s*\(|type\s+{escaped}\b|(?:var|const)\s+{escaped}\b)")


def defining_source(package_dir: Path, symbol: str) -> tuple[Path, int, list[str]]:
    pattern = definition_pattern(symbol)
    matches: list[tuple[Path, int, list[str]]] = []
    for path in sorted(package_dir.glob("*.go")):
        if path.name.endswith("_test.go"):
            continue
        lines = path.read_text(encoding="utf-8", errors="replace").splitlines()
        matches.extend((path, index, lines) for index, line in enumerate(lines, 1) if pattern.match(line.strip()))
    if not matches:
        raise ResolutionError(f"symbol {symbol!r} was not found in {package_dir}")
    if len(matches) > 1:
        locations = ", ".join(f"{path}:{line}" for path, line, _ in matches[:5])
        raise ResolutionError(f"symbol {symbol!r} is ambiguous ({locations}); qualify a method as Receiver.Method")
    return matches[0]


def generated(lines: list[str]) -> bool:
    marker = "\n".join(lines[:10]).lower()
    return "code generated" in marker and "do not edit" in marker


def resolve(args: argparse.Namespace) -> dict[str, object]:
    project = Path(args.project).resolve()
    environment = go_environment(project)
    package = package_metadata(project, args.package)
    module = package["Module"]
    assert isinstance(module, dict)
    selected, replacement = source_module(module)
    module_dir = Path(str(selected["Dir"])).resolve()
    package_dir = Path(str(package.get("Dir", ""))).resolve()
    if not package_dir.is_dir():
        raise ResolutionError(f"package source is absent at {package_dir}; refresh the module cache explicitly")
    if not package_dir.is_relative_to(module_dir):
        raise ResolutionError(f"package source {package_dir} escapes selected module source {module_dir}")
    path, start, lines = defining_source(package_dir, args.symbol)
    if generated(lines):
        raise ResolutionError(f"{path} is generated source; inspect its generator or upstream source instead")
    end = min(len(lines), start + MAX_EXCERPT_LINES - 1)
    return {
        "schema": SCHEMA,
        "language": "go",
        "dependency": module.get("Path", args.package),
        "version": module.get("Version") or "local",
        "source_path": str(module_dir),
        "defining_file": str(path.resolve()),
        "symbol": args.symbol,
        "excerpt": {"start_line": start, "end_line": end, "text": "\n".join(lines[start - 1 : end])},
        "provenance": {
            "environment": str(environment["GOMOD"]),
            "editable": False,
            "replacement": replacement,
            "private": private_module(str(module.get("Path", "")), environment.get("GOPRIVATE", "")),
        },
    }


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("package", help="selected dependency import path")
    result.add_argument("symbol", help="symbol name or Receiver.Method")
    result.add_argument("--project", default=".")
    return result


def main() -> int:
    try:
        print(json.dumps(resolve(parser().parse_args()), indent=2))
    except (OSError, ValueError, ResolutionError) as error:
        print(f"error: {error}", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
