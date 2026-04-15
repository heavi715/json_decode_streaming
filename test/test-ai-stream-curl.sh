#!/usr/bin/env sh
set -eu

if [ -z "${AI_STREAM_API_KEY:-}" ]; then
  echo "Missing AI_STREAM_API_KEY" >&2
  exit 1
fi

API_URL="${AI_STREAM_URL:-http://new-api.bangong.knowbox.cn/v1/chat/completions}"
MODEL="${AI_STREAM_MODEL:-claude-opus-4-20250514}"
PROMPT="${AI_STREAM_PROMPT:-只返回json，格式: {\"ok\":true,\"msg\":\"...\"}}"
PRINT_SNAPSHOTS="${AI_STREAM_PRINT_SNAPSHOTS:-1}"
MAX_SNAPSHOTS="${AI_STREAM_MAX_SNAPSHOTS:-20}"

PAYLOAD="$(python3 - <<'PY'
import json
import os

payload = {
    "model": os.environ.get("AI_STREAM_MODEL", "claude-opus-4-20250514"),
    "messages": [{"role": "user", "content": os.environ.get("AI_STREAM_PROMPT", "只返回json，格式: {\"ok\":true,\"msg\":\"...\"}")}],
    "stream": True,
}
print(json.dumps(payload, ensure_ascii=False))
PY
)"

RESPONSE_FILE="$(mktemp)"
HTTP_CODE_FILE="$(mktemp)"
cleanup() {
  rm -f "${RESPONSE_FILE}" "${HTTP_CODE_FILE}"
}
trap cleanup EXIT

if ! curl -sS -N "${API_URL}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${AI_STREAM_API_KEY}" \
  -d "${PAYLOAD}" \
  -o "${RESPONSE_FILE}" \
  -w "%{http_code}" > "${HTTP_CODE_FILE}"; then
  echo "curl request failed" >&2
  exit 1
fi

HTTP_CODE="$(cat "${HTTP_CODE_FILE}")"
if [ "${HTTP_CODE}" -lt 200 ] || [ "${HTTP_CODE}" -ge 300 ]; then
  echo "HTTP status: ${HTTP_CODE}" >&2
  echo "response sample:" >&2
  python3 - "${RESPONSE_FILE}" <<'PY'
from pathlib import Path
import sys

raw = Path(sys.argv[1]).read_text(encoding="utf-8", errors="replace")
print(raw[:500])
PY
  exit 1
fi

python3 - "${RESPONSE_FILE}" <<'PY'
import json
import os
import sys
from pathlib import Path

root = Path.cwd()
if str(root) not in sys.path:
    sys.path.insert(0, str(root))

from python.repair_json import repair_json_strict_prefix

accumulated = ""
chunk_count = 0
snapshot_count = 0
printed_snapshots = 0
event_count = 0
skipped_events = 0
debug_samples = []
print_snapshots = os.environ.get("AI_STREAM_PRINT_SNAPSHOTS", "1") != "0"
max_snapshots = int(os.environ.get("AI_STREAM_MAX_SNAPSHOTS", "20"))


def extract_piece(event):
    choices = event.get("choices") or []
    if not choices:
        return ""
    choice0 = choices[0] or {}
    delta = choice0.get("delta") or {}

    content = delta.get("content")
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        parts = []
        for item in content:
            if isinstance(item, str):
                parts.append(item)
                continue
            if not isinstance(item, dict):
                continue
            text = item.get("text")
            if isinstance(text, str):
                parts.append(text)
        return "".join(parts)

    # Some providers may place text directly under choice or delta.
    for key in ("text", "reasoning_content"):
        value = delta.get(key)
        if isinstance(value, str):
            return value
    value = choice0.get("text")
    if isinstance(value, str):
        return value
    return ""

raw_path = Path(sys.argv[1])
raw_response = raw_path.read_text(encoding="utf-8", errors="replace")

for raw_line in raw_response.splitlines():
    line = raw_line.strip()
    if not line.startswith("data:"):
        continue
    payload = line[5:].lstrip()
    if payload == "[DONE]":
        break
    event_count += 1
    try:
        event = json.loads(payload)
    except json.JSONDecodeError:
        skipped_events += 1
        if len(debug_samples) < 3:
            debug_samples.append(payload[:200])
        continue
    piece = extract_piece(event)
    if not piece:
        skipped_events += 1
        if len(debug_samples) < 3:
            debug_samples.append(json.dumps(event, ensure_ascii=False)[:200])
        continue

    chunk_count += 1
    obj = repair_json_strict_prefix(accumulated, return_object=True, append_content=piece)
    accumulated += piece
    if obj is not None:
        snapshot_count += 1
        if print_snapshots and printed_snapshots < max_snapshots:
            printed_snapshots += 1
            print(f"snapshot#{snapshot_count}: {json.dumps(obj, ensure_ascii=False)}")

if event_count == 0 and raw_response.strip():
    # Fallback: some gateways may ignore stream=true and return a single JSON object.
    try:
        plain = json.loads(raw_response)
        choices = plain.get("choices") or []
        if choices:
            msg = choices[0].get("message") or {}
            content = msg.get("content")
            if isinstance(content, str):
                accumulated = content
    except json.JSONDecodeError:
        pass

repaired = repair_json_strict_prefix(accumulated)
parsed = repair_json_strict_prefix(accumulated, return_object=True)
if parsed is None:
    print("Failed to parse streamed content as JSON.")
    print(f"Repaired text: {repaired}")
    print(f"events: {event_count}, content chunks: {chunk_count}, skipped events: {skipped_events}")
    if raw_response.strip():
        print("raw response sample:")
        print(raw_response[:500])
    if debug_samples:
        print("sample skipped payloads:")
        for item in debug_samples:
            print(item)
    raise SystemExit(1)

print(f"content chunks: {chunk_count}")
print(f"object snapshots: {snapshot_count}")
if print_snapshots and snapshot_count > max_snapshots:
    print(f"snapshot output truncated: printed {max_snapshots} of {snapshot_count}")
print(f"events: {event_count}, skipped events: {skipped_events}")
print(f"final repaired json: {repaired}")
print(f"final object type: {type(parsed).__name__}")
PY
