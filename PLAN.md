# Valet — Project Plan

## Problem
In a monorepo, AI coding rules for tools like Cursor, Claude Code, and GitHub Copilot often exist at the root but are never wired into individual packages. This causes AI agents to generate code without the correct context or standards applied.

## Solution
A CLI tool called Valet (`val`) that discovers monorepo structure, checks rule coverage across packages, and installs rules into each package automatically.

## Tech Stack
- Language: Go
- CLI framework: Cobra
- Config format: YAML (`valet.yaml`)

## Commands
- `val init` — initialise Valet in a monorepo, create `valet.yaml`, detect monorepo structure
- `val check` — health check across all packages, report rule coverage, flag conflicts. CI-compatible exit codes
- `val install` — apply root rules into each package, generating correct config files per detected AI tool. Supports `--dry-run`

## Phases

### Phase 1 — Scaffold ✅
- Set up Go project structure
- Wire up Cobra with the three commands as stubs
- Create basic `valet.yaml` config schema

### Phase 2 — Monorepo Discovery ✅
- Parse root `package.json` for `workspaces` field (npm/yarn)
- Handle both glob patterns and explicit paths
- Build internal package map: name, path, detected tooling

### Phase 3 — Rule Scanning & Selection
- Scan `.rules` folder for `.md` files (tool-agnostic rule content)
- If no `.rules` folder exists, create it and tell the user to add rules
- Interactively prompt the user to select which rules to include
- For each selected rule, prompt which tools it applies to (Claude Code, Cursor)
- Write selections into `valet.yaml`
- Requires an interactive prompt library (e.g. huh)

### Phase 4 — Install Logic
- For each discovered package, generate appropriate config files based on detected tools
- Apply root rules into each package directory
- `--dry-run` flag to preview changes before applying

### Phase 5 — Validation Report
- `val check` outputs structured health check
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

