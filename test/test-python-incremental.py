import json
import random
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from python.repair_json import RepairState, repair_json_strict_prefix


def split_random(text: str) -> list[str]:
    if not text:
        return [""]
    parts: list[str] = []
    i = 0
    n = len(text)
    while i < n:
        step = random.randint(1, 7)
        parts.append(text[i : i + step])
        i += step
    return parts


def main() -> None:
    random.seed(7)
    cases = json.loads((ROOT / "test" / "cases.json").read_text(encoding="utf-8"))
    failures = []

    for idx, case in enumerate(cases):
        text = case["input"]
        expected = repair_json_strict_prefix(text)
        st = RepairState()
        for chunk in split_random(text):
            st.feed(chunk)
        actual = st.finalize()
        if actual != expected:
            failures.append((idx, text, expected, actual))

    if failures:
        for idx, text, expected, actual in failures[:10]:
            print(f"[FAIL] case #{idx}")
            print(f"  input   : {text}")
            print(f"  expected: {expected}")
            print(f"  actual  : {actual}")
        raise SystemExit(1)

    print(f"Incremental mode passed: {len(cases)} cases.")


if __name__ == "__main__":
    main()
