#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_cmd git
require_cmd python3
require_cmd uv
require_cmd npm
require_cmd composer
require_cmd go

echo "Checking version consistency..."
python3 - <<'PY' "${ROOT_DIR}"
import json
import re
import sys
from pathlib import Path

root = Path(sys.argv[1])
version = (root / "VERSION").read_text(encoding="utf-8").strip()
if not version:
    raise SystemExit("VERSION is empty")

pyproject = (root / "python" / "pyproject.toml").read_text(encoding="utf-8")
m = re.search(r'^version = "(.*)"$', pyproject, re.M)
if not m:
    raise SystemExit("python/pyproject.toml: version not found")
py_ver = m.group(1)

js_ver = json.loads((root / "javascript" / "package.json").read_text(encoding="utf-8"))["version"]

pairs = {"VERSION": version, "python": py_ver, "javascript": js_ver}
if len(set(pairs.values())) != 1:
    raise SystemExit(f"Version mismatch: {pairs}")

print(f"Version OK: {version}")
PY

echo "Checking git worktree..."
if [[ -n "$(git status --porcelain)" ]]; then
  echo "Git worktree is not clean. Commit or stash changes before release." >&2
  exit 1
fi

echo "Release check passed."
