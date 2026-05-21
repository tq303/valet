# Valet

![version](https://img.shields.io/badge/version-0.2.0-blue) ![language](https://img.shields.io/badge/built%20with-Go-00ADD8) ![license](https://img.shields.io/badge/license-none-lightgrey)

One config. Any file. Anywhere it needs to be.

Define your files and locations in `valet.yaml`, run `valet sync` to keep everything in step.

---

## Install

```bash
go install github.com/tq303/valet@latest
```

## Usage

```bash
valet add <file>              # add a file or folder to sync across locations
valet add <git-repo-url>      # add files from a git repo
valet sync                    # sync changed files to their locations
valet sync -f                 # force overwrite all destinations
valet sync -d                 # dry run — show what would be written (* = would change)
valet sync <file-in-location> # promote that version as source and sync everywhere
valet list                    # show coverage across all locations
```

---

## Bootstrap flow

The typical setup:

```bash
# 1. Add a file — creates valet.yaml if it doesn't exist, prompts for location and destination folder
valet add .eslintrc.js

# 2. Sync — creates empty files at all locations if they don't exist yet
valet sync

# 3. Edit whichever copy you want, then promote it everywhere
valet sync packages/auth/.eslintrc.js
```

After step 2 you have empty placeholder files at every location. Edit the one you want to be the source of truth, then `valet sync <that-path>` to propagate it everywhere.

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

Symlink your config folders and dotfiles from a single rig repo:

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
