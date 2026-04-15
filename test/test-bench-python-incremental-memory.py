import json
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from python.repair_json import RepairState


def build_large_json_text() -> str:
    payload = {
        "session": {
            "id": "stream-1",
            "chunks": [
                {
                    "i": i,
                    "data": "x" * 128,
                    "meta": {
                        "ok": True,
                        "arr": list(range(20)),
                    },
                }
                for i in range(8000)
            ],
        }
    }
    return json.dumps(payload, separators=(",", ":"))


def iter_chunks(text: str, chunk_size: int):
    for i in range(0, len(text), chunk_size):
        yield text[i : i + chunk_size]


def run_once(text: str, chunk_size: int) -> tuple[float, int, int, int]:
    st = RepairState(compact_prefix=True, conservative_eof=True)
    peak_tail = 0
    peak_prefix_parts = 0
    start = time.perf_counter()
    for chunk in iter_chunks(text, chunk_size):
        st.feed(chunk)
        peak_tail = max(peak_tail, len(st.text))
        peak_prefix_parts = max(peak_prefix_parts, len(st.prefix_parts))
    out = st.finalize()
    elapsed_ms = (time.perf_counter() - start) * 1000
    return elapsed_ms, peak_tail, peak_prefix_parts, len(out)


def main() -> None:
    text = build_large_json_text()
    for chunk_size in (16, 64, 256, 1024):
        elapsed_ms, peak_tail, peak_prefix_parts, out_len = run_once(text, chunk_size)
        print(
            "chunk_size=%d input_len=%d output_len=%d elapsed_ms=%.1f peak_tail_bytes=%d peak_prefix_parts=%d"
            % (chunk_size, len(text), out_len, elapsed_ms, peak_tail, peak_prefix_parts)
        )


if __name__ == "__main__":
    main()
