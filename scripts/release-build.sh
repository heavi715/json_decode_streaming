#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "${ROOT_DIR}"
make package-all

echo "Release artifacts built:"
echo "  - python/dist/"
echo "  - javascript/*.tgz"
echo "  - php/*.zip"
