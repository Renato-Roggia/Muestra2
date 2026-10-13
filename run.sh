#!/usr/bin/env bash
set -euo pipefail

cd -- "$(dirname -- "${BASH_SOURCE[0]}")"

if [[ ! -x .venv/bin/python ]]; then
    echo "Install the dependencies first: bash install.sh" >&2
    exit 1
fi

if [[ -z "${DISPLAY:-}" && -z "${WAYLAND_DISPLAY:-}" ]]; then
    echo "Start the app from a graphical desktop session so its window can open." >&2
    exit 1
fi

exec .venv/bin/python main.py "$@"
