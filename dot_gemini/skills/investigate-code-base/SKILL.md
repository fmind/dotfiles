---
name: investigate-code-base
description:
  Comprehensive repository analysis for best practices, security, and idiomatic
  recommendations.
---

# Investigate Code Base

This skill provides a structured workflow for auditing and improving a codebase.
It focuses on three main pillars: **Best Practices**, **Security**, and
**Idiomatic Recommendations**.

## Core Workflow

To investigate a repository, follow these sequential phases:

### Phase 1: Contextual Mapping

1. **Repository Structure:** Run `list_directory` on the root and key source
   directories to understand the project layout.
2. **Tooling & Stack:** Check for configuration files (`package.json`,
   `Cargo.toml`, `requirements.txt`, `mise.toml`, `Dockerfile`, etc.) to
   identify the tech stack and development tools.
3. **Local Conventions:** Read `README.md`, `CONTRIBUTING.md`, `AGENTS.md`, or
   any project-specific documentation to understand established patterns.
4. **Initial Scan:** Use `codebase_investigator` for an architectural overview
   if the codebase is large or complex.

### Phase 2: Targeted Analysis

Analyze the codebase through three lenses:

#### 1. Best Practices

- **Modularity:** Identify large, monolithic files that could be refactored into
  smaller, more focused modules.
- **Consistency:** Ensure naming conventions, directory structure, and code
  style are consistent across the project.
- **Documentation:** Check for missing or outdated docstrings, comments, and
  README instructions.
- **Error Handling:** Look for generic `try-except` blocks or missing error
  management.

#### 2. Security

- **Secrets:** Scan for hardcoded API keys, tokens, or credentials (especially
  in config files or git-tracked `.env` files).
- **Permissions:** Review file permissions and sensitive configuration files.
- **Dependencies:** Identify outdated or vulnerable dependencies.
- **Input Validation:** Check for potential injection vulnerabilities in areas
  handling external input.

#### 3. Idiomatic Recommendations

- **Patterns:** Identify where established design patterns (e.g., Factory,
  Strategy, Observer) could simplify logic.
- **Language Features:** Suggest modern language features (e.g., async/await,
  type hints, modern syntax) where appropriate.
- **Library Usage:** Recommend more efficient or standard libraries if
  non-standard ones are used.

### Phase 3: Reporting & Action

1. **Summary:** Present a categorized list of findings, ranked by priority
   (Critical, Recommended, Optional).
2. **Rationale:** For each recommendation, provide a brief technical
   justification ("Why this change?").
3. **Action Plan:** Propose specific, surgical changes to address the most
   critical findings.
4. **Wait for Directive:** Present the report and wait for user approval before
   making any modifications.

## Guidelines

- **Surgicality:** Focus on meaningful improvements rather than cosmetic
  "cleanup" that doesn't add value.
- **Compatibility:** Ensure recommendations are compatible with the existing
  stack and environment.
- **Empiricism:** Validate assumptions by reading the code before suggesting a
  change.
- **Online Check:** Use `web_search` to check for up-to-date information on best
  practices, vulnerabilities, or library usage.
