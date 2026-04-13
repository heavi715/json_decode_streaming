# JSON Decode Streaming (strict_prefix)

A multi-language streaming JSON repair toolkit for AI outputs that may be truncated near the end.

Policy used in this project is `strict_prefix`:

- keep only the longest safe prefix from the beginning,
- auto-close missing string quote (value-only) and container delimiters,
- never guess missing non-empty values or inject new keys.

## Supported languages (phase 1)

- Python: `python/repair_json.py`
- JavaScript: `javascript/repairJson.js`
- Go: `golang/repair_json.go`
- PHP: `php/RepairJson.php`

## Example behavior

- `{"a":"b` -> `{"a":"b"}`
- `{"a":[1,2,` -> `{"a":[1,2]}`
- `{"a":"1","b":` -> `{"a":"1"}`
- `{"a":"1","b":"` -> `{"a":"1","b":""}`

## Run tests

Shared test vectors are in `test/cases.json`.

- Sync check:
  - `make test-cases-sync`
  - `make test-cases-stats`
- Python:
  - `python3 test/test-python.py`
  - `python3 test/test-fuzz-python.py`
- JavaScript:
  - `export PATH="/Users/heavi/.nvm/versions/node/v22.14.0/bin:$PATH"`
  - `node test/test-javascript.js`
- Go:
  - `go run test/test-go.go`
- PHP:
  - `php test/test-php.php`

## Coverage snapshot

Current shared deterministic cases: see `make test-cases-stats` (reads `docs/test-cases.md` + `test/cases.json` and prints per-section counts).

Covered scenario groups include:

- Canonical truncation examples
- Additional basic closures and tail drop behavior
- Complex multi-level nested object/array closures
- Boundary and prefix-policy rollback behavior
- Number token and exponent/fraction edge cases
- Unicode/escape handling in keys and values
- Array/object comma-state repairs
- Delimiter mismatch handling
- Root-done and multi-root truncation behavior
- High-noise mixed-tail replay baselines

## Design docs

- Spec: `docs/spec.md`
- Test cases and validation rules: `docs/test-cases.md`

## Complexity

- Time complexity: `O(n)` (single linear pass).
- Space complexity: `O(depth)` for container stack.

## Notes

- This project targets broken/truncated suffixes.
- It does not attempt to fix deeply corrupted prefixes.
