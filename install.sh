#!/bin/sh
set -e

DEST="${HOME}/.local/bin/val"

go build -o "$DEST" .
chmod +x "$DEST"

echo "Installed to $DEST"
