#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET="${1:-all}"
DRY_RUN="${DRY_RUN:-0}"
VERSION="$(tr -d '[:space:]' < "${ROOT_DIR}/VERSION")"
MODULE_PATH="$(awk '/^module / {print $2}' "${ROOT_DIR}/go.mod")"

run_cmd() {
  if [[ "${DRY_RUN}" == "1" ]]; then
    echo "[dry-run] $*"
  else
    eval "$@"
  fi
}

publish_python() {
  if [[ "${DRY_RUN}" != "1" ]]; then
    command -v twine >/dev/null 2>&1 || {
      echo "twine not found. Install with: uv pip install twine" >&2
      exit 1
    }
  fi
  run_cmd "cd \"${ROOT_DIR}\" && twine upload python/dist/*"
}

publish_javascript() {
  run_cmd "cd \"${ROOT_DIR}/javascript\" && npm publish --access public"
}

publish_php() {
  echo "PHP package uses VCS + Packagist indexing."
  if [[ -n "${PACKAGIST_UPDATE_URL:-}" ]]; then
    run_cmd "curl -X POST \"${PACKAGIST_UPDATE_URL}\""
  else
    echo "Set PACKAGIST_UPDATE_URL to trigger automatic Packagist update webhook."
  fi
}

publish_go() {
  if [[ "${DRY_RUN}" == "1" ]]; then
    if [[ -z "${MODULE_PATH}" || "${MODULE_PATH}" != */* ]]; then
      echo "[dry-run] go.mod module path looks non-publishable: ${MODULE_PATH}"
    fi
    echo "[dry-run] git tag v${VERSION}"
    echo "[dry-run] git push origin v${VERSION}"
    return
  fi
  if [[ -z "${MODULE_PATH}" || "${MODULE_PATH}" != */* ]]; then
    echo "go.mod module path is not publishable: ${MODULE_PATH}" >&2
    echo "Set module path to your VCS location, for example github.com/your-org/json_decode_streaming" >&2
    exit 1
  fi
  git tag "v${VERSION}"
  git push origin "v${VERSION}"
}

case "${TARGET}" in
  all)
    publish_python
    publish_javascript
    publish_php
    publish_go
    ;;
  python)
    publish_python
    ;;
  javascript)
    publish_javascript
    ;;
  php)
    publish_php
    ;;
  go)
    publish_go
    ;;
  *)
    echo "Unknown target: ${TARGET}" >&2
    echo "Usage: $0 [all|python|javascript|php|go]" >&2
    exit 1
    ;;
esac
