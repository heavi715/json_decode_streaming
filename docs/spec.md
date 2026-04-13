# Streaming JSON Repair Specification (strict_prefix)

## Goal

Repair truncated or partially streamed JSON text by preserving only the longest syntactically safe prefix from the beginning, then auto-closing required delimiters.

The repairer must be deterministic and avoid semantic guessing.

## Core policy

- Input is a UTF-8 string that is *intended* to be JSON but may be truncated or broken near the end.
- Output must be valid JSON text.
- The algorithm only keeps content from index `0..lastSafeIndex`.
- If the parser reaches an unrecoverable state, it rolls back to `lastSafeIndex`.
- After rollback/cutoff, the algorithm may:
  - close an unterminated string only when in a valid value context,
  - append missing `]` / `}` according to open containers.
- The algorithm must not invent non-empty values or new keys.

## High-level state machine

Track:

- `stack`: open containers; item is `object` or `array`.
- `ctx.state`: expected token in current scope:
  - `root_value`
  - `object_key_or_end`
  - `object_colon`
  - `object_value`
  - `object_comma_or_end`
  - `array_value_or_end`
  - `array_comma_or_end`
- `inString`: whether currently scanning inside a JSON string literal.
- `escapeNext`: whether next char inside string is escaped.
- `stringRole`: `key` or `value`.
- `lastSafeIndex`: last character index where prefix can be safely emitted.

## Safe-boundary rule

`lastSafeIndex` can be updated only when a full structural/value unit is complete, such as:

- root scalar fully parsed,
- root object/array fully closed,
- object member fully complete (`"k": <value>`),
- array element fully complete (`<value>`).

Do not update safe boundary while waiting for:

- object key after comma,
- colon after key,
- value after colon,
- next element after array comma.

## Truncation handling

### 1) Unterminated string

If EOF occurs inside string and `stringRole == value`, append `"` and continue closing containers.

If EOF occurs inside key string or in invalid context, rollback to `lastSafeIndex` (drop unfinished fragment).

### 2) Missing closing delimiters

After prefix cutoff and optional string close, append closers for all still-open containers in reverse order.

### 3) Invalid token at current state

When token is not valid for current state:

- stop scanning,
- trim to `lastSafeIndex`,
- close containers.

## Behavior examples

- Input: `{"a":"b` -> Output: `{"a":"b"}`
- Input: `{"a":[1,2,` -> Output: `{"a":[1,2]}`
- Input: `{"a":"1","b":` -> Output: `{"a":"1"}`
- Input: `{"a":"1","b":"` -> Output: `{"a":"1","b":""}`

## Non-goals (v1)

- Repairing text with invalid prefix (errors near the beginning).
- Supporting comments, trailing commas, or JSON5 extensions.
- Semantic merge/guess recovery beyond strict syntax prefix.
