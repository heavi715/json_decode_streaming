#!/usr/bin/env bash
set -euo pipefail

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_cmd gh

if ! gh auth status >/dev/null 2>&1; then
  echo "GitHub CLI is not authenticated. Run: gh auth login" >&2
  exit 1
fi

echo "Configure GitHub Actions release secrets for current repository."
echo "Press Enter to skip an optional secret."
echo

read -r -p "TWINE_USERNAME [default: __token__]: " TWINE_USERNAME_INPUT
TWINE_USERNAME="${TWINE_USERNAME_INPUT:-__token__}"

read -r -s -p "TWINE_PASSWORD (PyPI token): " TWINE_PASSWORD
echo
read -r -s -p "NPM_TOKEN: " NPM_TOKEN
echo
read -r -p "PACKAGIST_UPDATE_URL (optional): " PACKAGIST_UPDATE_URL

if [[ -n "${TWINE_USERNAME}" ]]; then
  printf "%s" "${TWINE_USERNAME}" | gh secret set TWINE_USERNAME
  echo "Set TWINE_USERNAME"
fi

if [[ -n "${TWINE_PASSWORD}" ]]; then
  printf "%s" "${TWINE_PASSWORD}" | gh secret set TWINE_PASSWORD
  echo "Set TWINE_PASSWORD"
fi

if [[ -n "${NPM_TOKEN}" ]]; then
  printf "%s" "${NPM_TOKEN}" | gh secret set NPM_TOKEN
  echo "Set NPM_TOKEN"
fi

if [[ -n "${PACKAGIST_UPDATE_URL}" ]]; then
  printf "%s" "${PACKAGIST_UPDATE_URL}" | gh secret set PACKAGIST_UPDATE_URL
  echo "Set PACKAGIST_UPDATE_URL"
fi

echo
echo "All provided secrets are configured."
