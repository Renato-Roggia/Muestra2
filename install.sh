#!/usr/bin/env bash
set -Eeuo pipefail

cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
trap 'printf "\nInstallation failed. Fix the error above and run: bash install.sh\n" >&2' ERR

if [[ "$EUID" -eq 0 ]]; then
    echo "Run this script as your normal user: bash install.sh" >&2
    echo "It will request sudo only for system dependencies." >&2
    exit 1
fi

if ! command -v apt-get >/dev/null 2>&1; then
    echo "This installer requires a Linux system with apt-get." >&2
    exit 1
fi
if ! command -v python3 >/dev/null 2>&1; then
    echo "Install Python 3.11 or newer, then run this script again." >&2
    exit 1
fi

python3 - <<'PY'
import platform
import struct
import sys

if sys.version_info < (3, 11):
    sys.exit("Python 3.11 or newer is required.")
if struct.calcsize("P") != 8 or platform.machine().lower() not in ("aarch64", "arm64", "x86_64", "amd64"):
    sys.exit("Prebuilt dependencies require 64-bit ARM or x86 Python and Linux.")
PY

if ! command -v sudo >/dev/null 2>&1; then
    echo "sudo is required to install system dependencies." >&2
    exit 1
fi

echo "Installing system dependencies..."
sudo apt-get update
glib_package=libglib2.0-0
if apt-cache show libglib2.0-0t64 >/dev/null 2>&1; then
    glib_package=libglib2.0-0t64
fi
sudo apt-get install -y --no-install-recommends \
    python3-venv python3-pip ca-certificates \
    libgl1 libegl1 libgles2 "$glib_package" libportaudio2 \
    libxcb-cursor0 libxcb-xinerama0 libxkbcommon-x11-0

if [[ -e .venv && ! -x .venv/bin/python ]]; then
    echo ".venv already exists but is not a usable Linux virtual environment." >&2
    echo "Move it aside and rerun this script. Nothing has been deleted." >&2
    exit 1
fi
if [[ ! -d .venv ]]; then
    python3 -m venv .venv
fi

echo "Installing Python dependencies in .venv..."
.venv/bin/python -m pip install --only-binary=:all: --upgrade pip
.venv/bin/python -m pip install --only-binary=:all: -r requirements.txt
.venv/bin/python -m pip check
.venv/bin/python - <<'PY'
import cv2
import mediapipe
import numpy

print(f"OpenCV {cv2.__version__}, MediaPipe {mediapipe.__version__}, NumPy {numpy.__version__}")
if not any(line.strip().startswith("GUI:") and line.split(":", 1)[1].strip() not in ("", "NONE")
           for line in cv2.getBuildInformation().splitlines()):
    raise SystemExit("OpenCV has no GUI support. A desktop-enabled OpenCV package is required.")
PY

echo "Installation complete. Start the app with: bash run.sh"
