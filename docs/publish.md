# Standardized Release Guide

This project supports standardized publishing for Python (PyPI), JavaScript (npm), PHP (Packagist), and Go modules.

## 1) One-time setup

- Python:
  - `uv pip install twine`
  - configure PyPI token in `~/.pypirc` or `TWINE_USERNAME` / `TWINE_PASSWORD`
- JavaScript:
  - `npm login`
- PHP:
  - register this VCS repository on Packagist (or Private Packagist/Satis)
  - optional webhook: set `PACKAGIST_UPDATE_URL`
- Go:
  - ensure `go.mod` module path is the real VCS module path (already configured to GitHub path in this repo)

## 2) Release workflow

1. Bump version:
   - edit root `VERSION`
2. Sync versions into package metadata:
   - `make release-sync-version`
3. Commit and make sure worktree is clean.
4. Run checks:
   - `make release-check`
5. Build artifacts:
   - `make release-build`
6. Dry-run publish commands:
   - `DRY_RUN=1 make release-publish-all`
7. Publish:
   - `make release-publish-python`
   - `make release-publish-javascript`
   - `make release-publish-php`
   - `make release-publish-go`

## 3) Notes by ecosystem

- Python: uploads from `python/dist/*` via `twine upload`.
- JavaScript: publishes from `javascript/` via `npm publish --access public`.
- PHP: Composer package is discovered from VCS by Packagist; webhook can trigger immediate re-index.
- Go: release is based on `git tag vX.Y.Z` and `git push origin vX.Y.Z`.

## 4) Quick commands

- Full check/build:
  - `make release-sync-version && make release-check && make release-build`
- Full dry-run:
  - `DRY_RUN=1 make release-publish-all`

## 5) GitHub Actions auto release

This repo includes `.github/workflows/release.yml`:

- trigger: push tag `v*` (for example `v0.1.1`)
- actions:
  - build and publish Python package (if PyPI secrets are configured)
  - build and publish npm package (if npm token is configured)
  - validate PHP package and optionally trigger Packagist webhook
  - verify Go module layout

Required repository secrets:

- `TWINE_USERNAME` (usually `__token__`)
- `TWINE_PASSWORD` (PyPI API token)
- `NPM_TOKEN`
- optional `PACKAGIST_UPDATE_URL`

Quick interactive setup (recommended, avoids putting tokens in shell history):

- `make release-setup-secrets`

Release from local:

1. `echo "0.1.1" > VERSION`
2. `make release-sync-version`
3. commit and push branch
4. `git tag v0.1.1`
5. `git push origin v0.1.1`
