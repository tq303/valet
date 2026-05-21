# Valet

![version](https://img.shields.io/badge/version-0.1.0-blue) ![language](https://img.shields.io/badge/built%20with-Go-00ADD8) ![license](https://img.shields.io/badge/license-none-lightgrey)

One config. Any file. Anywhere it needs to be.

Define your files and locations in `valet.yaml`, run `valet sync` to keep everything in step.

---

## Install

```bash
go install github.com/tq303/valet@latest
```

## Usage

```bash
valet add <file>              # add a file to sync across locations
valet sync                    # sync all configured files to their locations
valet sync <file-in-location> # promote that version as source and sync everywhere
valet list                    # show coverage across all locations
```

---

## Bootstrap flow

The typical setup:

```bash
# 1. Add a file — creates valet.yaml if it doesn't exist, prompts for location and destination folder
valet add CLAUDE.md

# 2. Sync — creates empty files at all locations if they don't exist yet
valet sync

# 3. Edit whichever copy you want, then promote it everywhere
valet sync packages/auth/CLAUDE.md
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

### Dotfiles / rig

Symlink your config folders to `~/.config` from a single location repo:

```yaml
version: 1
rules:
  - link: true
    files:
      - nvim
      - ghostty
    locations:
      - ~/.config
```

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
