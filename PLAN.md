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
- `valet add [file]` — add a file to be synced; creates `valet.yaml` if it doesn't exist; supports `-l`/`--location`, `--dest`, `-i`/`--ignore` flags
- `valet sync [file]` — sync changed files; `-f` to force all, `-d` for dry run; pass a path to promote that version as source
- `valet remove [file]` — remove a file from `valet.yaml`; `-f` also deletes synced copies
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

### Phase 5 — URL Sources ✅
- `valet add <url>` tracks a remote file in `valet.yaml`
- `valet sync` re-fetches from the URL to stay current
- Works for any file type — skills, dotfiles, CI templates, editor config, etc.
- Claude Code skills just need to land in `.claude/` — no special install step

### Phase 6 — Repo Sources ✅
- Rules can declare a `repo:` git URL alongside `files:`
- `valet sync` clones the repo into `/tmp/valet/repos/<hash>/` on first run, pulls on subsequent runs
- Files and folders are resolved from the cached repo, preserving full directory structure
- Internal relative links within skill folders work because the whole repo is available
- `valet add <git-url>` starts the add flow with `repo:` pre-filled

### Phase 7 — Quality of Life ✅
- Content-aware sync — only writes files that have changed; `-f` to force
- Dry run (`-d`) shows `*` next to files that would be written
- `valet remove` with select prompt; `-f` deletes synced copies
- Multiple files and locations in a single `valet add` command
- `valet add -i` syncs once without writing to `valet.yaml`
- Windows support — `link: true` falls back to copy on Windows
- GitHub Actions release workflow — binaries for mac/linux/windows on tag push
- Dynamic version badge via shields.io

### Phase 8 — Archive Sources ✅
- Rules can declare `archive: true` alongside a URL in `files:` and an `extract:` list of paths within the archive
- `valet sync` downloads the archive, extracts the named files to a cache dir (`/tmp/valet/archives/<hash>/`), then copies them to each location
- Supported formats: `.tar.gz`, `.tar.xz`, `.tar.bz2`, `.tgz`, `.zip`
- Cache is keyed by URL hash — re-download only on `-f`; content-aware copy skips unchanged files
- `valet add <archive-url>` starts the add flow prompting for extract paths, location, and dest

### Phase 9 — Config Discovery
- Walk up from `os.Getwd()` looking for `valet.yaml`, stopping at `os.UserHomeDir()`
- Allows running valet from any subdirectory of the project root
- If no config is found, fall back to current directory (existing behaviour for `valet add` / first-time init)

## Non-Goals (MVP)
- No monorepo auto-detection
- No preset system (use `valet.yaml` directly)

## Future

### Monorepo workspace detection
Auto-detect packages from npm/yarn/pnpm workspace config to pre-fill location options during `valet add`. Lower priority — users can type paths directly and the yaml is easy to edit by hand.

### MCP validation and health checks
