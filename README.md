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

## Package for direct use

You can build distributable packages for each language from this repo:

- Python wheel/sdist:
  - `make package-python`
  - output: `python/dist/`
- JavaScript npm tarball:
  - `make package-javascript`
  - output: `javascript/*.tgz`
- PHP Composer archive:
  - `make package-php`
  - output: `php/heavi-json-decode-streaming-*.zip`
- Go module:
  - `make package-go` (checks module is importable/buildable)

Build everything:

- `make package-all`

Quick usage after packaging:

- Python: `pip install python/dist/*.whl`
- JavaScript: `npm install ./javascript/*.tgz`
- PHP: `composer config repositories.local artifact ./php && composer require heavi/json-decode-streaming:0.1.0`
- Go: publish module path first, then `go get <your-module-path>`

## Simplest integration in other projects

If your goal is "use immediately with minimal setup", use this order:

1) **Single-file copy (simplest, no package registry needed)**  
Copy one file into your target project:

- Python: copy `python/repair_json.py`
- JavaScript: copy `javascript/repairJson.js`
- PHP: copy `php/RepairJson.php`
- Go: copy `golang/repair_json.go`

Then call the same APIs described in this README:

- Python: `repair_json_strict_prefix(...)`
- JavaScript: `repairJsonStrictPrefix(...)`
- PHP: `repair_json_strict_prefix(...)`
- Go: `RepairJSONStrictPrefix(...)`

2) **Install from built artifacts (still simple, more standardized)**

- Python:
  - in this repo: `make package-python`
  - in target project: `pip install /path/to/json_decode_streaming/python/dist/*.whl`
- JavaScript:
  - in this repo: `make package-javascript`
  - in target project: `npm install /path/to/json_decode_streaming/javascript/json-decode-streaming-0.1.0.tgz`
- PHP:
  - in this repo: `make package-php`
  - in target project:
    - `composer config repositories.local artifact /path/to/json_decode_streaming/php`
    - `composer require heavi/json-decode-streaming:0.1.0`

3) **Source install from Git (Python only, one command)**

- `pip install "git+<your-repo-url>.git#subdirectory=python"`

## Example behavior

- `{"a":"b` -> `{"a":"b"}`
- `{"a":[1,2,` -> `{"a":[1,2]}`
- `{"a":"1","b":` -> `{"a":"1"}`
- `{"a":"1","b":"` -> `{"a":"1","b":""}`

## Return parsed object directly

- Python: `repair_json_strict_prefix(text, return_object=True)` returns parsed object (`dict`/`list`/primitive), and returns `None` when repaired output is empty.
- Python: `repair_json_strict_prefix_both(text)` returns `(repaired_string, parsed_object_or_none)`.
- JavaScript: `repairJsonStrictPrefix(text, true)` returns parsed object (`object`/`array`/primitive), and returns `null` when repaired output is empty.
- JavaScript: `repairJsonStrictPrefixBoth(text)` returns `[repairedString, parsedObjectOrNull]`.
- Go: `RepairJSONStrictPrefixWithOption(text, true)` returns parsed object (`any`) and returns `nil` when repaired output is empty.
- Go: `RepairJSONStrictPrefixBoth(text)` returns `(repairedString, parsedObjectOrNil, err)`.
- PHP: `repair_json_strict_prefix($text, true)` returns parsed array/primitive, and returns `null` when repaired output is empty.
- PHP: `repair_json_strict_prefix_both($text)` returns `[$repaired, $parsedOrNull]`.

## Append content for streaming

When AI responses arrive chunk-by-chunk, pass existing content as `text` and delta as append parameter:

- Python: `repair_json_strict_prefix(text, return_object=True, append_content=chunk)`
- JavaScript: `repairJsonStrictPrefix(text, true, chunk)`
- Go: `RepairJSONStrictPrefixWithAppendOption(text, chunk, true)`
- PHP: `repair_json_strict_prefix($text, true, $chunk)`

## Run tests

Shared test vectors are in `test/cases.json`.

- Sync check:
  - `make test-cases-sync`
  - `make test-cases-stats`
- Python:
  - `python3 test/test-python.py`
  - `python3 test/test-fuzz-python.py`
  - `python3 test/test-python-incremental.py`
  - `python3 test/test-ai-stream-python.py` (requires env vars)
  - `./test/test-ai-stream-curl.sh` (requires env vars)
- JavaScript:
  - `export PATH="/Users/heavi/.nvm/versions/node/v22.14.0/bin:$PATH"`
  - `node test/test-ai-stream-javascript.js` (requires env vars)
- Go:
  - `go run -tags aistream test/test-ai-stream-go.go` (requires env vars)
- PHP:
  - `php test/test-ai-stream-php.php` (requires env vars)
- JavaScript:
  - `export PATH="/Users/heavi/.nvm/versions/node/v22.14.0/bin:$PATH"`
  - `node test/test-javascript.js`
- Go:
  - `go run test/test-go.go`
- PHP:
  - `php test/test-php.php`

## Run benchmarks

- Python: `python3 test/test-bench-python.py` (prints `string` and `object` modes)
- Python incremental memory: `python3 test/test-bench-python-incremental-memory.py`
- PHP: `php test/test-bench-php.php` (prints `string` and `object` modes)
- JavaScript:
  - `export PATH="/Users/heavi/.nvm/versions/node/v22.14.0/bin:$PATH"`
  - `node test/test-bench-javascript.js` (prints `string` and `object` modes)
- Go: `go run -tags benchgo test/test-bench-go.go` (prints `string` and `object` modes)

Optional mode selection for all benchmark scripts:

- `--mode=string` (or Go: `-mode=string`)
- `--mode=object` (or Go: `-mode=object`)
- `--mode=both_return` (or Go: `-mode=both_return`) for single-call `(repaired, object)` path
- default is `all`

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
- Incremental mode design (Python): `docs/incremental-mode.md`
- Standardized publish workflow: `docs/publish.md`
- GitHub auto release workflow: `.github/workflows/release.yml`

## Complexity

- Time complexity: `O(n)` (single linear pass).
- Space complexity: `O(depth)` for container stack.

## Notes

- This project targets broken/truncated suffixes.
- It does not attempt to fix deeply corrupted prefixes.
- Stream test env vars: `AI_STREAM_API_KEY`, optional `AI_STREAM_URL`, `AI_STREAM_MODEL`, `AI_STREAM_PROMPT`.
- Stream snapshot printing: `AI_STREAM_PRINT_SNAPSHOTS=1` (default), limit with `AI_STREAM_MAX_SNAPSHOTS` (default `20`).

