#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION_FILE="${ROOT_DIR}/VERSION"

if [[ ! -f "${VERSION_FILE}" ]]; then
  echo "VERSION file not found: ${VERSION_FILE}" >&2
  exit 1
fi

VERSION="$(tr -d '[:space:]' < "${VERSION_FILE}")"
if [[ -z "${VERSION}" ]]; then
  echo "VERSION is empty." >&2
  exit 1
fi

echo "Syncing version -> ${VERSION}"

python3 - <<'PY' "${ROOT_DIR}" "${VERSION}"
import json
import re
import sys
from pathlib import Path

root = Path(sys.argv[1])
version = sys.argv[2]

pyproject = root / "python" / "pyproject.toml"
text = pyproject.read_text(encoding="utf-8")
text = re.sub(r'^version = ".*"$', f'version = "{version}"', text, flags=re.M)
pyproject.write_text(text, encoding="utf-8")

package_json = root / "javascript" / "package.json"
obj = json.loads(package_json.read_text(encoding="utf-8"))
obj["version"] = version
package_json.write_text(json.dumps(obj, indent=2, ensure_ascii=True) + "\n", encoding="utf-8")

composer_json = root / "php" / "composer.json"
obj = json.loads(composer_json.read_text(encoding="utf-8"))
obj.pop("version", None)
composer_json.write_text(json.dumps(obj, indent=2, ensure_ascii=True) + "\n", encoding="utf-8")
PY

echo "Version sync complete."
