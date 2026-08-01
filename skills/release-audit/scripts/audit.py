#!/usr/bin/env python3
"""Audit exact-head GitHub release evidence without changing release state."""

from __future__ import annotations

import argparse
import hashlib
import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Any

CONTRACT_SCHEMA = "fmind.dev/release-audit-contract/v1"
REPORT_SCHEMA = "fmind.dev/release-audit-report/v1"
SHA_PATTERN = re.compile(r"^[0-9a-f]{40}$")


class AuditError(RuntimeError):
    """A bounded evidence-collection failure."""


def command(arguments: list[str], cwd: Path, *, allow_failure: bool = False) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(arguments, cwd=cwd, check=False, capture_output=True, text=True, timeout=60)
    if result.returncode and not allow_failure:
        detail = result.stderr.strip() or result.stdout.strip()
        raise AuditError(f"`{' '.join(arguments)}` failed: {detail}")
    return result


def output(arguments: list[str], cwd: Path) -> str:
    return command(arguments, cwd).stdout.strip()


def validate_contract(contract: dict[str, Any]) -> None:
    if contract.get("schema") != CONTRACT_SCHEMA:
        raise AuditError(f"contract schema must be {CONTRACT_SCHEMA!r}")
    if not isinstance(contract.get("workflows"), list) or not all(
        isinstance(item, str) for item in contract["workflows"]
    ):
        raise AuditError("contract workflows must be a list of names or workflow paths")
    if not isinstance(contract.get("artifacts"), list):
        raise AuditError("contract artifacts must be a list; use [] to declare a source-only release")
    supported = {"asset", "checksum", "sbom", "signature", "attestation"}
    for artifact in contract["artifacts"]:
        if not isinstance(artifact, dict) or not isinstance(artifact.get("name"), str):
            raise AuditError("each artifact requires a string name")
        if artifact.get("kind") not in supported:
            raise AuditError(f"artifact {artifact['name']!r} has an unsupported kind")
        if artifact["kind"] == "checksum" and not isinstance(artifact.get("covers"), list):
            raise AuditError(f"checksum {artifact['name']!r} requires a covers list")
        if artifact["kind"] in {"signature", "attestation"} and not isinstance(artifact.get("subject"), str):
            raise AuditError(f"{artifact['kind']} {artifact['name']!r} requires a subject")
        if artifact["kind"] == "signature" and not all(
            isinstance(artifact.get(field), str) for field in ("certificate_identity", "certificate_oidc_issuer")
        ):
            raise AuditError(f"signature {artifact['name']!r} requires certificate identity and OIDC issuer")


def remote_tag_sha(project: Path, tag: str) -> str | None:
    result = command(
        ["git", "ls-remote", "--tags", "origin", f"refs/tags/{tag}", f"refs/tags/{tag}^{{}}"],
        project,
        allow_failure=True,
    )
    if result.returncode:
        return None
    refs = dict(line.split("\t", 1)[::-1] for line in result.stdout.splitlines() if "\t" in line)
    return refs.get(f"refs/tags/{tag}^{{}}") or refs.get(f"refs/tags/{tag}")


def resolve_release_target(project: Path, repository: str, target: str) -> str | None:
    if SHA_PATTERN.fullmatch(target):
        return target
    result = command(["gh", "api", f"repos/{repository}/commits/{target}", "--jq", ".sha"], project, allow_failure=True)
    return result.stdout.strip() if result.returncode == 0 else None


def verify_download(
    project: Path, repository: str, tag: str, destination: Path, artifact: dict[str, Any]
) -> tuple[bool, str]:
    name = artifact["name"]
    path = destination / name
    if not path.is_file():
        return False, "declared asset is missing"
    verified = command(
        ["gh", "release", "verify-asset", tag, str(path), "--repo", repository], project, allow_failure=True
    )
    if verified.returncode:
        return False, "asset digest does not match GitHub release metadata"

    kind = artifact["kind"]
    if kind == "checksum":
        entries: dict[str, str] = {}
        for line in path.read_text(encoding="utf-8", errors="replace").splitlines():
            digest, separator, filename = line.partition("  ")
            if separator and re.fullmatch(r"[0-9a-fA-F]{64}", digest):
                entries[filename.lstrip("*")] = digest.lower()
        for covered in artifact["covers"]:
            covered_path = destination / covered
            if not covered_path.is_file():
                return False, f"checksum subject {covered!r} was not declared before its checksum"
            actual = hashlib.sha256(covered_path.read_bytes()).hexdigest()
            if entries.get(covered) != actual:
                return False, f"checksum mismatch for {covered!r}"
    elif kind == "sbom":
        try:
            sbom = json.loads(path.read_text(encoding="utf-8"))
        except (OSError, ValueError):
            return False, "SBOM is not valid JSON"
        if not isinstance(sbom, dict) or not ({"spdxVersion", "bomFormat"} & sbom.keys()):
            return False, "SBOM is neither SPDX JSON nor CycloneDX JSON"
    elif kind == "signature":
        subject = destination / artifact["subject"]
        if not subject.is_file():
            return False, f"signature subject {artifact['subject']!r} is missing"
        verified = command(
            [
                "cosign",
                "verify-blob",
                "--signature",
                str(path),
                "--certificate-identity",
                artifact["certificate_identity"],
                "--certificate-oidc-issuer",
                artifact["certificate_oidc_issuer"],
                str(subject),
            ],
            project,
            allow_failure=True,
        )
        if verified.returncode:
            return False, "Sigstore signature verification failed"
    elif kind == "attestation":
        subject = destination / artifact["subject"]
        if not subject.is_file():
            return False, f"attestation subject {artifact['subject']!r} is missing"
        verified = command(
            ["gh", "attestation", "verify", str(subject), "--repo", repository], project, allow_failure=True
        )
        if verified.returncode:
            return False, "GitHub attestation verification failed"
    return True, "validated"


def collect(project: Path, tag: str, contract: dict[str, Any]) -> dict[str, Any]:
    repository = output(["gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner"], project)
    head = output(["git", "rev-parse", "HEAD"], project)
    upstream_result = command(["git", "rev-parse", "@{upstream}"], project, allow_failure=True)
    tag_result = command(["git", "rev-list", "-n", "1", tag], project, allow_failure=True)
    release_result = command(["gh", "api", f"repos/{repository}/releases/tags/{tag}"], project, allow_failure=True)
    release = json.loads(release_result.stdout) if release_result.returncode == 0 else None
    release_target = resolve_release_target(project, repository, release["target_commitish"]) if release else None
    exact_runs = json.loads(
        output(["gh", "api", f"repos/{repository}/actions/runs?head_sha={head}&per_page=100"], project)
    ).get("workflow_runs", [])
    recent_runs = json.loads(output(["gh", "api", f"repos/{repository}/actions/runs?per_page=100"], project)).get(
        "workflow_runs", []
    )
    runs = list({run["id"]: run for run in [*recent_runs, *exact_runs]}.values())

    artifact_checks: dict[str, dict[str, Any]] = {}
    with tempfile.TemporaryDirectory(prefix="release-audit-") as temporary:
        destination = Path(temporary)
        for artifact in contract["artifacts"]:
            command(
                [
                    "gh",
                    "release",
                    "download",
                    tag,
                    "--repo",
                    repository,
                    "--pattern",
                    artifact["name"],
                    "--dir",
                    str(destination),
                ],
                project,
                allow_failure=True,
            )
        for artifact in contract["artifacts"]:
            valid, detail = verify_download(project, repository, tag, destination, artifact)
            artifact_checks[artifact["name"]] = {"valid": valid, "detail": detail}

    return {
        "head_sha": head,
        "upstream_sha": upstream_result.stdout.strip() if upstream_result.returncode == 0 else None,
        "tag_sha": tag_result.stdout.strip() if tag_result.returncode == 0 else None,
        "remote_tag_sha": remote_tag_sha(project, tag),
        "release": {
            "exists": release is not None,
            "draft": release.get("draft") if release else None,
            "published_at": release.get("published_at") if release else None,
            "target_sha": release_target,
            "assets": [asset["name"] for asset in release.get("assets", [])] if release else [],
        },
        "workflow_runs": [
            {
                "name": run.get("name"),
                "path": str(run.get("path", "")).split("@", 1)[0],
                "head_sha": run.get("head_sha"),
                "status": run.get("status"),
                "conclusion": run.get("conclusion"),
            }
            for run in runs
        ],
        "artifact_checks": artifact_checks,
    }


def check(name: str, passed: bool, detail: str) -> dict[str, Any]:
    return {"name": name, "passed": passed, "detail": detail}


def evaluate(tag: str, contract: dict[str, Any], evidence: dict[str, Any]) -> dict[str, Any]:
    head = evidence.get("head_sha")
    identities = {
        "head": head,
        "upstream": evidence.get("upstream_sha"),
        "tag": evidence.get("tag_sha"),
        "remote_tag": evidence.get("remote_tag_sha"),
        "release_target": evidence.get("release", {}).get("target_sha"),
    }
    checks = [
        check("upstream", identities["upstream"] == head, f"HEAD={head} upstream={identities['upstream']}"),
        check("local_tag", identities["tag"] == head, f"HEAD={head} tag={identities['tag']}"),
        check(
            "immutable_tag",
            identities["remote_tag"] == identities["tag"] == head,
            f"HEAD={head} local={identities['tag']} remote={identities['remote_tag']}",
        ),
    ]
    release = evidence.get("release", {})
    checks.append(
        check(
            "release_target",
            release.get("exists") and identities["release_target"] == head,
            f"HEAD={head} release={identities['release_target']}",
        )
    )

    stale_workflows = False
    for required in contract["workflows"]:
        matching = [run for run in evidence.get("workflow_runs", []) if required in {run.get("name"), run.get("path")}]
        exact = [run for run in matching if run.get("head_sha") == head]
        passed = any(run.get("status") == "completed" and run.get("conclusion") == "success" for run in exact)
        stale_workflows = stale_workflows or bool(matching and not exact)
        checks.append(check(f"workflow:{required}", passed, f"exact_sha={head} successful={passed}"))

    assets = set(release.get("assets", []))
    for artifact in contract["artifacts"]:
        result = evidence.get("artifact_checks", {}).get(artifact["name"], {})
        passed = artifact["name"] in assets and result.get("valid") is True
        checks.append(check(f"artifact:{artifact['name']}", passed, result.get("detail", "missing validation")))

    missing_identity = any(value is None for value in identities.values()) or not release.get("exists")
    sha_mismatch = any(value is not None and value != head for key, value in identities.items() if key != "head")
    workflow_missing = any(not item["passed"] for item in checks if item["name"].startswith("workflow:"))
    artifact_failed = any(not item["passed"] for item in checks if item["name"].startswith("artifact:"))
    if missing_identity or (workflow_missing and not stale_workflows):
        state = "missing"
    elif release.get("draft") or not release.get("published_at"):
        state = "draft"
    elif sha_mismatch or stale_workflows:
        state = "stale"
    elif artifact_failed:
        state = "artifact-incomplete"
    else:
        state = "published"
    return {"schema": REPORT_SCHEMA, "tag": tag, "state": state, "identities": identities, "checks": checks}


def human(report: dict[str, Any]) -> str:
    lines = [f"release audit: {report['state'].upper()}", f"tag: {report['tag']}"]
    lines.extend(f"{name}: {value or 'missing'}" for name, value in report["identities"].items())
    lines.extend(
        f"{'PASS' if item['passed'] else 'FAIL'} {item['name']}: {item['detail']}" for item in report["checks"]
    )
    return "\n".join(lines)


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("tag")
    result.add_argument("--project", default=".")
    result.add_argument("--contract", default=".release-audit.json")
    result.add_argument("--evidence", help="offline evidence JSON for deterministic testing")
    result.add_argument("--json", action="store_true", dest="json_output")
    return result


def main() -> int:
    args = parser().parse_args()
    project = Path(args.project).resolve()
    try:
        contract = json.loads((project / args.contract).read_text(encoding="utf-8"))
        validate_contract(contract)
        evidence = (
            json.loads(Path(args.evidence).read_text(encoding="utf-8"))
            if args.evidence
            else collect(project, args.tag, contract)
        )
        report = evaluate(args.tag, contract, evidence)
    except (AuditError, OSError, ValueError, subprocess.TimeoutExpired) as error:
        print(f"error: {error}", file=sys.stderr)
        return 2
    print(json.dumps(report, indent=2) if args.json_output else human(report))
    return 0 if report["state"] == "published" else 1


if __name__ == "__main__":
    raise SystemExit(main())
