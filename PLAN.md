# Valiant — Project Plan

## Problem
In a monorepo, AI coding rules for tools like Cursor, Claude Code, and GitHub Copilot often exist at the root but are never wired into individual packages. This causes AI agents to generate code without the correct context or standards applied.

## Solution
A CLI tool called Valiant (`val`) that discovers monorepo structure, validates rule coverage across packages, and installs rules into each package automatically.

## Tech Stack
- Language: Go
- CLI framework: Cobra
- Config format: YAML (`valiant.yaml`)

## Commands
- `val init` — initialise Valiant in a monorepo, create `valiant.yaml`, detect monorepo structure
- `val validate` — health check across all packages, report rule coverage, flag conflicts. CI-compatible exit codes
- `val install` — apply root rules into each package, generating correct config files per detected AI tool. Supports `--dry-run`

## Phases

### Phase 1 — Scaffold
- Set up Go project structure
- Wire up Cobra with the three commands as stubs
- Create basic `valiant.yaml` config schema

### Phase 2 — Monorepo Discovery
- Parse root `package.json` for `workspaces` field (npm/yarn)
- Handle both glob patterns and explicit paths
- Build internal package map: name, path, detected tooling

### Phase 3 — Rule Scanning
- Scan configurable rules folder (default `.valiant/rules`)
- Detect rule file formats:
  - `.mdc` — Cursor
  - `CLAUDE.md` — Claude Code
  - `.github/copilot-instructions.md` — GitHub Copilot
- Validate files are parseable and non-empty

### Phase 4 — Install Logic
- For each discovered package, generate appropriate config files based on detected tools
- Apply root rules into each package directory
- `--dry-run` flag to preview changes before applying

### Phase 5 — Validation Report
- `val validate` outputs structured health check
- Shows rule coverage per package
- Flags missing or conflicting rules

## Non-Goals (MVP)
- No marketplace or CDN skill discovery
- No MCP validation
- No Cargo/Rust monorepo support

## Future
- `val search` and `val add` — search and install rule packs from a CDN
- Cargo/Rust monorepo support
- MCP validation and health checks
- Pre-execution agent orchestration layer

