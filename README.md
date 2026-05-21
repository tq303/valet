# Valet

![version](https://img.shields.io/github/v/release/tq303/valet) ![build](https://github.com/tq303/valet/actions/workflows/release.yml/badge.svg) ![language](https://img.shields.io/badge/built%20with-Go-00ADD8) ![license](https://img.shields.io/badge/license-none-lightgrey)

Manage and sync files across locations. 

Add any file, URL or repo location once or cache to local file.

Define your files and locations in `valet.yaml`, run `valet sync` to keep everything in step.

---

## Install

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/tq303/valet/releases/latest/download/valet-darwin-arm64 -o /usr/local/bin/valet && chmod +x /usr/local/bin/valet
```


**Linux**
```bash
curl -L https://github.com/tq303/valet/releases/latest/download/valet-linux-amd64 -o /usr/local/bin/valet && chmod +x /usr/local/bin/valet
```

Or with Go:
```bash
go install github.com/tq303/valet@latest
```

## Usage

```
Valet manages and syncs files across locations. Add any file, URL or repo location once or cache to local file.

Usage:
  valet [command]

Available Commands:
  add         Add a file to be synced across locations
  list        Show configured files and coverage per package
  remove      Remove a file from valet.yaml
  sync        Sync all configured files into their locations

Use "valet [command] --help" for more information about a command.
```

---

## Bootstrap flow

```bash
# Add a file — prompts for locations and dest folder, creates valet.yaml if needed
valet add .eslintrc.js

# Sync — copies the file to all configured locations
valet sync

# Made a change in one of the locations? Promote it as the new source
valet sync packages/auth/.eslintrc.js
```

---

## Examples

Here are some `valet.yaml` examples for common scenarios.

### AI rules for a monorepo

Keep Claude Code and Cursor rules consistent across every package:

```yaml
version: 1
rules:
  - dest: .claude
    files:
      - CLAUDE.md
    locations:
      - packages/auth
      - packages/api
  - dest: .cursor/rules
    files:
      - api-standards.mdc
    locations:
      - packages/auth
      - packages/api
```

> Use `dest` when multiple locations share the same subfolder — it saves repeating the path.

### Dotfiles / rig

Symlink your config folders and dotfiles from a single rig repo. On Windows, `link: true` falls back to copy:

```yaml
version: 1
rules:
  - link: true
    files:
      - nvim
      - ghostty
    locations:
      - ~/.config
  - link: true
    files:
      - claude/CLAUDE.md
    locations:
      - ~/.claude
  - link: true
    files:
      - .zshsource
    locations:
      - ~/
```

### Files from a git repo

Track files or folders from any git repo as a source. On `valet sync`, the repo is cloned to `/tmp/valet/repos/` and kept up to date — folder structure and internal references are preserved.

```yaml
version: 1
rules:
  - repo: https://github.com/valyuAI/skills
    files:
      - valyu-search/valyu-best-practices
    locations:
      - ~/.claude/commands
```

---

### Shared tooling config across a monorepo

Keep ESLint, Prettier, and TypeScript config consistent across every package:

```yaml
version: 1
rules:
  - files:
      - .eslintrc.js
      - .prettierrc
      - tsconfig.json
    locations:
      - packages/auth
      - packages/api
      - packages/ui
```
