# JSON Decode Streaming (strict_prefix)

A multi-language streaming JSON repair toolkit for AI outputs that may be truncated near the end.

[中文文档](docs/readme-zh.md)

Policy used in this project is `strict_prefix`:

- keep only the longest safe prefix from the beginning,
- auto-close missing string quote (value-only) and container delimiters,
- never guess missing non-empty values or inject new keys.

## Supported languages (phase 1)

- Python: `python/repair_json.py`
- JavaScript: `javascript/repairJson.js`
- Go: `golang/repair_json.go`
- PHP: `php/RepairJson.php`

## Install in other projects

Current published status:

- Python (PyPI): available
- JavaScript (npm): available
- Go module: available
- PHP (Packagist): may have indexing delay; VCS fallback is available

Direct remote install:

- Python: `pip install json-decode-streaming`
- JavaScript: `npm install json-decode-streaming`
- Go: `go get github.com/heavi715/json_decode_streaming@v0.1.6`
- PHP (when Packagist is indexed): `composer require heavi/json-decode-streaming:^0.1`

PHP fallback install (always works via VCS):

- `composer config repositories.json_decode_streaming vcs https://github.com/heavi715/json_decode_streaming.git`
- `composer require heavi/json-decode-streaming:v0.1.6`

## Build distributable artifacts locally

Build packages from source:

- Python wheel/sdist:
  - `make package-python`
  - output: `python/dist/`
- JavaScript npm tarball:
  - `make package-javascript`
  - output: `javascript/*.tgz`
- PHP Composer archive:
  - `make package-php`
  - output: `php/heavi-json-decode-streaming-*.zip`
- Go module check:
  - `make package-go` (checks module is importable/buildable)

Build everything:

- `make package-all`

Install from local artifacts:

- Python: `pip install python/dist/*.whl`
- JavaScript: `npm install ./javascript/*.tgz`
- PHP: `composer config repositories.local artifact ./php && composer require heavi/json-decode-streaming:*`

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

Append-cache tuning (optional):

- JavaScript: `setRepairJsonAppendCacheConfig({ maxEntries, maxTotalBytes, ttlMs, clear })`
- JavaScript preset helper: `applyRepairJsonAppendCachePreset("low_memory" | "high_throughput" | "default", clear?)`
- Go: `SetAppendCacheConfig(maxEntries, maxTotalBytes, ttl, clear)`
- Go preset helper: `ApplyAppendCachePreset(AppendCachePresetLowMemory|AppendCachePresetHighThroughput|AppendCachePresetDefault, clear)`
- Python: `set_repair_json_append_cache_config(max_entries=..., max_total_bytes=..., ttl_seconds=..., clear=...)`
- Python preset helper: `apply_repair_json_append_cache_preset("low_memory" | "high_throughput" | "default", clear=...)`
- PHP: `set_repair_json_append_cache_config($maxEntries, $maxTotalBytes, $ttlSeconds, $clear)`
- PHP preset helper: `apply_repair_json_append_cache_preset('low_memory'|'high_throughput'|'default', $clear)`
- Ready presets:
  - JavaScript: `node test/test-cache-config-javascript.js`
  - Go: `go run -tags cacheconfiggo test/test-cache-config-go.go`

Unified cache presets (all languages):

- `default`: `max_entries=256`, `max_total_bytes=4 MiB`, `ttl=120s`
- `low_memory`: `max_entries=64`, `max_total_bytes=512 KiB`, `ttl=15s`
- `high_throughput`: `max_entries=1024`, `max_total_bytes=16 MiB`, `ttl=600s`

Copy-ready examples:

- JavaScript:
  - `setRepairJsonAppendCacheConfig({ maxEntries: 64, maxTotalBytes: 512 * 1024, ttlMs: 15_000, clear: true })`
- Go:
  - `SetAppendCacheConfig(64, 512*1024, 15*time.Second, true)`
- Python:
  - `set_repair_json_append_cache_config(max_entries=64, max_total_bytes=512 * 1024, ttl_seconds=15.0, clear=True)`
- PHP:
  - `set_repair_json_append_cache_config(64, 512 * 1024, 15.0, true)`

## Piece convention for streaming

In this project, a `piece` means one incremental text fragment from the model stream (often from `delta.content`).

Recommended loop:

1. Extract `piece` from current event.
2. Skip when `piece` is empty.
3. Call repair API with `text=accumulated` and append parameter=`piece`.
4. Append piece locally: `accumulated += piece`.

Reference patterns in repository tests:

- Python: `test/test-ai-stream-python.py`
- JavaScript: `test/test-ai-stream-javascript.js`
- Go: `test/test-ai-stream-go.go`
- PHP: `test/test-ai-stream-php.php`
- curl+Python parser: `test/test-ai-stream-curl.sh`

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
  - `node test/test-bench-javascript-append.js` (appendContent incremental vs full recompute)
- Go: `go run -tags benchgo test/test-bench-go.go` (prints `string` and `object` modes)
  - `go run -tags benchgoappend test/test-bench-go-append.go` (appendContent incremental vs full recompute)

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