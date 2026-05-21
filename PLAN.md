# Valet — Project Plan

## Problem
Config files for tools like Cursor, Claude Code, ESLint, and TypeScript need to be consistent across teams and projects. Without a tool to manage this, files drift, get forgotten, or have to be manually kept in sync.

## Solution
A CLI tool called Valet (`valet`) that manages and propagates config files across locations. You define what files go where in `valet.yaml`, and Valet keeps them in sync.

## Tech Stack
- Language: Go
- CLI framework: Cobra
- Config format: YAML (`valet.yaml`)
- Interactive prompts: huh (charmbracelet)

## Commands
- `valet add [file]` — add a file to be synced; creates `valet.yaml` if it doesn't exist
- `valet sync [file]` — sync all configured files; pass a path to promote that version as source
- `valet list` — show file coverage per location. CI-compatible exit codes

## valet.yaml shape
```yaml
version: 1
rules:
  - files:
      - CLAUDE.md
    dest: .claude
    locations:
      - packages/auth
      - packages/api
  - files:
      - .eslintrc.js
    locations:
      - packages/auth
      - packages/api
      - packages/web
```

## Phases

### Phase 1 — Scaffold ✅
- Set up Go project structure
- Wire up Cobra with commands
- Create basic `valet.yaml` config schema

### Phase 2 — Core Sync Logic ✅
- Read `valet.yaml` rules
- For each rule, copy files into each location (optionally into a dest subfolder)
- Symlink mode via `link: true`
- `--dry-run` support

### Phase 3 — Add Command ✅
- `valet add <file>` — prompts for location path and dest folder, writes to `valet.yaml`
- If file is already tracked, extends it to new locations
- Promote flow: if the file passed is from a known location, copy it back as the new source and re-sync everywhere

### Phase 4 — List Command ✅
- Tabular coverage report: file, location, status (ok / MISSING)
- CI-compatible exit codes

## Non-Goals (MVP)
- No monorepo auto-detection (can be added as an extension)
- No preset system (use `valet.yaml` directly)
- No marketplace or CDN discovery

## Future
- Monorepo workspace detection (npm/yarn/pnpm) to pre-fill location options
- Community presets via CDN
- MCP validation and health checks
