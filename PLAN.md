# Valet — Project Plan

## Problem
In a monorepo, config files for tools like Cursor, Claude Code, ESLint, and TypeScript often exist at the root but are never consistently applied across individual packages. This causes drift, inconsistency, and wasted time.

## Solution
A CLI tool called Valet (`val`) that manages and propagates config files across monorepo packages. It supports built-in presets (e.g. AI rules) and ad-hoc file syncing for any file type.

## Tech Stack
- Language: Go
- CLI framework: Cobra
- Config format: YAML (`valet.yaml`)
- Interactive prompts: huh (charmbracelet)

## Commands
- `valet init` — initialise Valet, detect monorepo structure, configure rules via preset or ad-hoc
- `valet add` — ad-hoc add a file to be synced across packages (e.g. `valet add .eslintrc.js`)
- `valet install` — apply all configured rules into each package. Supports `--dry-run`
- `valet add` — ad-hoc add a file to be synced across packages (e.g. `valet add .eslintrc.js`)
- `valet list` — show what's configured and coverage per package. CI-compatible exit codes

## Modes
### Preset mode
`valet init --preset ai` walks through AI-specific rule setup. Knows destination conventions for Claude Code (`.claude/CLAUDE.md`) and Cursor (`.cursor/rules/*.mdc`). More presets can be added over time.

### Ad-hoc mode
`valet add .eslintrc.js` — select a file, choose which packages it applies to, and valet tracks and syncs it. No preset needed.

Both modes write to `valet.yaml` and are applied by `valet install`.

## valet.yaml shape
```yaml
version: 1
rules:
  - file: .rules/.claude/auth-rules.md
    preset: claude                    # resolves destination automatically
  - file: .eslintrc.js
    dest: .eslintrc.js                # explicit destination path in each package
packages:
  - path: packages/auth
  - path: packages/api
```

## Phases

### Phase 1 — Scaffold ✅
- Set up Go project structure
- Wire up Cobra with the three commands as stubs
- Create basic `valet.yaml` config schema

### Phase 2 — Monorepo Discovery ✅
- Parse root `package.json` for `workspaces` field (npm/yarn)
- Parse `pnpm-workspace.yaml` for pnpm workspaces
- Fall back to single-package root if no workspaces declared
- Build internal package map: name, path, detected tooling

### Phase 3 — Rule Scanning & Selection ✅
- Scan `.rules` folder for `.md` files (tool-agnostic rule content)
- Files in `.rules/.claude/` and `.rules/.cursor/` are pre-assigned to their tool
- Interactively prompt the user to select which rules to include
- For each unassigned rule, prompt which tools it applies to
- Write selections into `valet.yaml`

### Phase 4 — Install Logic ✅
- Read `valet.yaml` for rules and packages
- For each package, resolve destination path per rule (preset or explicit `dest`)
- Generate tool-specific formats where needed (e.g. `.mdc` frontmatter for Cursor)
- `--dry-run` flag to preview changes before applying

### Phase 5 — Ad-hoc File Syncing (`valet add`) ✅
- `valet add <file>` — pick a file, select target packages, write to `valet.yaml`
- Works for any file type (ESLint, Prettier, tsconfig, etc.)
- `valet install` applies it the same as preset rules

### Phase 6 — List Command ✅
- `valet list` outputs structured coverage report
- Shows rule/file coverage per package
- Flags missing or outdated files
- CI-compatible exit codes

## Non-Goals (MVP)
- No marketplace or CDN preset discovery
- No MCP validation
- No Cargo/Rust monorepo support

## Future
- `valet search` and `valet add <preset>` — search and install community presets from a CDN
- Cargo/Rust monorepo support
- MCP validation and health checks
- Pre-execution agent orchestration layer
