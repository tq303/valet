# Valet

![version](https://img.shields.io/github/v/release/tq303/valet) ![build](https://github.com/tq303/valet/actions/workflows/release.yml/badge.svg) ![language](https://img.shields.io/badge/built%20with-Go-00ADD8) ![license](https://img.shields.io/badge/license-none-lightgrey)

Manage and sync files across locations.

Add any file, URL, git repo, or release archive once. Define locations in `valet.yaml`, run `val sync` to keep everything in step.

---

## Install

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/tq303/valet/releases/latest/download/val-darwin-arm64 -o /usr/local/bin/val && chmod +x /usr/local/bin/val
```

**Linux**
```bash
curl -L https://github.com/tq303/valet/releases/latest/download/val-linux-amd64 -o /usr/local/bin/val && chmod +x /usr/local/bin/val
```

Or with Go:
```bash
go install github.com/tq303/valet@latest
```

## Usage

```
Valet manages and syncs files across locations. Add any file, URL or repo location once or cache to local file.

Usage:
  val [command]

Available Commands:
  add         Add a file to be synced across locations
  list        Show configured files and coverage per package
  remove      Remove a file from valet.yaml
  sync        Sync all configured files into their locations

Use "val [command] --help" for more information about a command.
```

---

## Bootstrap flow

```bash
# Add a file — prompts for locations and dest folder, creates valet.yaml if needed
val add config.js

# Sync — copies the file to all configured locations
val sync

# Made a change in one of the locations? Promote it as the new source
val sync packages/auth/config.js
```

---

## Examples

Here are some `valet.yaml` examples for common scenarios.

### Shared config across a monorepo

Keep linting, formatting, and TypeScript config consistent across every package:

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

### AI agent rules

Keep AI editor rules consistent across every package, using `dest` to place them in the right subfolder:

```yaml
version: 1
rules:
  - dest: .agent/rules
    files:
      - coding-standards.md
      - api-conventions.md
    locations:
      - packages/auth
      - packages/api
```

### Dotfiles

Symlink config folders and dotfiles from a single source. On Windows, `link: true` falls back to copy:

```yaml
version: 1
rules:
  - link: true
    files:
      - editor
      - terminal
    locations:
      - ~/.config
  - link: true
    files:
      - .shellrc
    locations:
      - ~/
```

### Files from a git repo

Track files or folders from any git repo. On `val sync`, the repo is cloned to `/tmp/valet/repos/` and kept up to date — folder structure and internal references are preserved:

```yaml
version: 1
rules:
  - repo: https://github.com/your-org/shared-configs
    files:
      - agent-skills/
      - ci-templates/
    locations:
      - ~/.config/agent
```

### Binaries from a release archive

Pull a specific binary out of a GitHub release tarball or zip. Supports `.tar.gz`, `.tar.xz`, `.tar.bz2`, and `.zip`. The archive is cached locally and only re-downloaded when forced:

```yaml
version: 1
rules:
  - files:
      - https://github.com/your-org/tool/releases/download/v1.0.0/tool-v1.0.0.tar.gz
    archive: true
    extract:
      - tool
    locations:
      - ~/bin
```
