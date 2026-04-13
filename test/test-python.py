import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from python.repair_json import repair_json_strict_prefix


def main() -> None:
    cases = json.loads((ROOT / "test" / "cases.json").read_text(encoding="utf-8"))
    failures = []

    for idx, case in enumerate(cases):
        repaired = repair_json_strict_prefix(case["input"])
        if repaired != case["expected"]:
            failures.append((idx, "output mismatch", repaired, case["expected"]))
            continue
        if repaired != "":
            try:
                json.loads(repaired)
            except json.JSONDecodeError as exc:
                failures.append((idx, f"invalid json: {exc}", repaired, case["expected"]))

    if failures:
        for idx, reason, actual, expected in failures:
            print(f"[FAIL] case #{idx}: {reason}")
            print(f"  actual  : {actual}")
            print(f"  expected: {expected}")
        raise SystemExit(1)

    print(f"All {len(cases)} Python cases passed.")


if __name__ == "__main__":
    main()
