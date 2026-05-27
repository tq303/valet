# Valet

![version](https://img.shields.io/github/v/release/tq303/valet) ![build](https://github.com/tq303/valet/actions/workflows/release.yml/badge.svg) ![language](https://img.shields.io/badge/built%20with-Go-00ADD8) ![license](https://img.shields.io/badge/license-none-lightgrey)

Manage and sync files across locations.

Keeps a canonical source of truth for files that need to exist in multiple places, whether that's shared configs, prompt files, scripts, or tooling.

Add any file, URL, git repo, or release archive once. Define locations in `valet.yaml`, run `val sync` to keep everything in step.

- Explicit yaml config vs stow's convention-based approach — handles non-standard paths and platform conditions cleanly
- Scales where stow doesn't — stow's simplicity is its ceiling
- URL files and git repos in the same sync, not just dotfiles
- Enterprise use case: keeping shared configs, CI templates, tooling in sync across a monorepo without drift
- Single source of truth, one-command sync

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

Or build from source:

```bash
make install
```

---

## Commands

```
val add <file>       Add a file, URL, git repo or archive — prompts for locations
val sync             Sync all configured files into their locations
val list             Show file coverage across all locations
val remove <file>    Remove a file from valet.yaml
```

| Flag               | Command    | Description                                |
| ------------------ | ---------- | ------------------------------------------ |
| `-f`               | sync       | Force overwrite even if up to date         |
| `-d`               | sync       | Dry run — preview changes without writing  |
| `--platform <tag>` | sync, list | Filter to rules matching the platform tag  |
| `-l <path>`        | add        | Location (skips prompt, repeatable)        |
| `--dest <folder>`  | add        | Destination subfolder within each location |
| `-i`               | add        | Sync once without writing to `valet.yaml`  |

---

## Dotfiles across machines

The most common use case — symlink config folders from a single source repo and sync them across mac, ubuntu desktop, and servers. Each machine passes its own `--platform` tag; rules without a `platforms` field apply everywhere.

```yaml
version: 1
rules:
  - files:
      - nvim
      - tmux
    link: true
    locations:
      - ~/.config

  - files:
      - ghostty
    link: true
    platforms:
      - mac
      - ubuntu
    locations:
      - ~/.config

  - files:
      - .zshrc
      - .zshsource
    link: true
    locations:
      - ~/
```

```bash
# mac or ubuntu desktop — syncs everything including ghostty
val sync --platform mac

# headless server — ghostty rule is skipped
val sync --platform server
```

`val list` shows the full picture with a PLATFORM column so you always know what applies where:

```
FILE                           LOCATION             PLATFORM             STATUS
----                           --------             --------             ------
nvim                           ~/.config            -                    ok
tmux                           ~/.config            -                    ok
ghostty                        ~/.config            mac,ubuntu           ok
.zshrc                         ~/                   -                    ok
.zshsource                     ~/                   -                    ok
```

---

## More examples

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
  - dest: .claude
    files:
      - CLAUDE.md
    locations:
      - packages/auth
      - packages/api
      - packages/ui
```

### Files from a git repo

Track files or folders from any git repo. On `val sync`, the repo is cloned to `/tmp/valet/repos/` and kept up to date:

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

Pull a specific binary out of a GitHub release tarball or zip. Supports `.tar.gz`, `.tar.xz`, `.tar.bz2`, and `.zip`. Cached locally and only re-downloaded when forced:

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
