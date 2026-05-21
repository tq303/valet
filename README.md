# Valet

![version](https://img.shields.io/badge/version-0.1.0-blue) ![language](https://img.shields.io/badge/built%20with-Go-00ADD8) ![license](https://img.shields.io/badge/license-none-lightgrey)

Sync config files across your monorepo packages. Add any file once, install it everywhere.

Valet manages rule files, editor config, linting config, or any file that should be consistent across packages — with built-in support for AI coding tools like Claude Code and Cursor.

---

## Install

```bash
go install github.com/tq303/valet@latest
```

## Usage

```bash
valet init        # detect monorepo structure and configure rules
valet add <file>  # add a file to sync across packages
valet install     # apply everything in valet.yaml
valet list        # show coverage across packages
```
