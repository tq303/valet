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
valet init        # detect monorepo structure and create valet.yaml
valet add <file>  # add a file to sync across locations
valet sync        # sync all configured files to their locations
valet list        # show coverage across all locations
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
      - path: packages/auth
      - path: packages/api
  - dest: .cursor/rules
    files:
      - api-standards.mdc
    locations:
      - path: packages/auth
      - path: packages/api
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
      - path: ~/.config
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
      - path: packages/auth
      - path: packages/api
      - path: packages/ui
```
