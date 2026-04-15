# Incremental Mode Design (strict_prefix)

## Goal

Provide a Python-side streaming API that consumes JSON text chunk-by-chunk and avoids rescanning already processed bytes.

## API

- `RepairState.feed(chunk: str) -> None`
  - Append a new chunk and parse only new bytes.
- `RepairState.snapshot() -> str`
  - Return repaired output at current boundary.
- `RepairState.finalize() -> str`
  - Alias of `snapshot()` for end-of-stream.
- `repair_json_strict_prefix(text: str) -> str`
  - Existing one-shot API remains unchanged and now delegates to `RepairState`.

## State Model

`RepairState` keeps parser state across chunks:

- `stack`: object/array nesting.
- `state`: FSM state (`root_value`, `object_key_or_end`, etc.).
- `in_string`, `escape_next`, `string_role`.
- `last_safe`: latest safe prefix boundary.
- `i`: cursor index in the accumulated input.
- `broke_early`: unrecoverable unsafe tail marker.

This allows each `feed()` call to continue from prior cursor and state.

## Memory Strategy

- Incremental mode now uses a compacting buffer:
  - Confirmed safe prefix segments are moved into `prefix_parts`.
  - Active parsing continues on a shortened `text` tail buffer.
- This keeps working memory bounded by "unresolved tail + nesting context"
  instead of growing with total stream length.

## Semantics

- Same `strict_prefix` policy as one-shot mode.
- No semantic guessing, no synthetic non-empty values or keys.
- If current chunk ends in value-string context, `snapshot()` may close quote and open containers, matching existing truncation behavior.
- After unrecoverable break (`broke_early = true`), additional chunks are ignored by parser state and output remains anchored at last safe prefix.

## Current Scope

- Implemented in Python only.
- PHP/JS/Go remain one-shot for now.
- Shared deterministic behavior remains validated by `test/cases.json`.

## Follow-ups

- Add incremental conformance tests for all shared cases under random chunk boundaries.
- Mirror incremental API in PHP/JS/Go after Python API is stable.
