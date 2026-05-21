# Valet

![version](https://img.shields.io/badge/version-0.1.0-blue) ![language](https://img.shields.io/badge/built%20with-Go-00ADD8) ![license](https://img.shields.io/badge/license-none-lightgrey)

Sync any file to any location. Add it once, keep it everywhere.

Valet tracks files and syncs them across locations — monorepo packages, projects, or anywhere on your filesystem. Built-in support for AI coding tools like Claude Code and Cursor.

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

## valet.yaml

```yaml
version: 1
rules:
  - dest: .claude
    files:
      - CLAUDE.md
    locations:
      - path: packages/auth
      - path: packages/api
  - files:
      - .eslintrc.js
    locations:
      - path: packages/auth
      - path: packages/api
      - path: ../other-project
```
