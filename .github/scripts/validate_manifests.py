#!/usr/bin/env python3
"""Shim: delegates to the hermetic Go validator. Prefer `go run ./cmd/check`."""
import subprocess, sys, pathlib
root = pathlib.Path(__file__).resolve().parents[2]
result = subprocess.run(["go", "run", "./cmd/check"] + sys.argv[1:], cwd=root)
sys.exit(result.returncode)
