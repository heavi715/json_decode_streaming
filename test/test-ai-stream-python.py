import json
import os
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from python.repair_json import repair_json_strict_prefix


def main() -> None:
    api_key = os.getenv("AI_STREAM_API_KEY")
    if not api_key:
        raise SystemExit("Missing AI_STREAM_API_KEY")

    api_url = os.getenv(
        "AI_STREAM_URL",
        "http://new-api.bangong.knowbox.cn/v1/chat/completions",
    )
    model = os.getenv("AI_STREAM_MODEL", "claude-opus-4-20250514")
    prompt = os.getenv("AI_STREAM_PROMPT", "只返回json，格式: {\"ok\":true,\"msg\":\"...\"}")

    payload = {
        "model": model,
        "messages": [{"role": "user", "content": prompt}],
        "stream": True,
    }

    cmd = [
        "curl",
        "-sS",
        "-N",
        api_url,
        "-H",
        "Content-Type: application/json",
        "-H",
        f"Authorization: Bearer {api_key}",
        "-d",
        json.dumps(payload, ensure_ascii=False),
    ]

    proc = subprocess.Popen(
        cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        bufsize=1,
    )

    accumulated = ""
    snapshots = 0
    final_object = None

    assert proc.stdout is not None
    for line in proc.stdout:
        line = line.strip()
        if not line.startswith("data: "):
            continue
        data = line[6:]
        if data == "[DONE]":
            break
        try:
            event = json.loads(data)
        except json.JSONDecodeError:
            continue
        choices = event.get("choices") or []
        if not choices:
            continue
        delta = choices[0].get("delta") or {}
        piece = delta.get("content")
        if not piece:
            continue

        current_object = repair_json_strict_prefix(
            accumulated, return_object=True, append_content=piece
        )
        accumulated += piece
        snapshots += 1

        if current_object is not None:
            final_object = current_object

    return_code = proc.wait()
    stderr = proc.stderr.read() if proc.stderr else ""
    if return_code != 0:
        raise SystemExit(f"curl failed ({return_code}): {stderr}")

    repaired_text = repair_json_strict_prefix(accumulated)
    repaired_object = repair_json_strict_prefix(accumulated, return_object=True)
    if repaired_object is None:
        raise SystemExit(f"Streaming output could not be parsed as JSON: {repaired_text}")

    print(f"stream chunks with content: {snapshots}")
    print(f"repaired json: {repaired_text}")
    print(f"parsed object type: {type(repaired_object).__name__}")
    if final_object is not None:
        print("incremental object snapshots were parsed during stream.")


if __name__ == "__main__":
    main()
