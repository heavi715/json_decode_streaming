import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from python.repair_json import repair_json_strict_prefix, repair_json_strict_prefix_both


def main() -> None:
    cases = json.loads((ROOT / "test" / "cases.json").read_text(encoding="utf-8"))
    failures = []

    for idx, case in enumerate(cases):
        repaired = repair_json_strict_prefix(case["input"])
        if repaired != case["expected"]:
            failures.append((idx, "output mismatch", repaired, case["expected"]))
            continue
        repaired_object = repair_json_strict_prefix(case["input"], return_object=True)
        expected_object = json.loads(case["expected"]) if case["expected"] != "" else None
        if repaired_object != expected_object:
            failures.append(
                (
                    idx,
                    "object output mismatch",
                    repr(repaired_object),
                    repr(expected_object),
                )
            )
            continue
        repaired_both, repaired_both_object = repair_json_strict_prefix_both(case["input"])
        if repaired_both != case["expected"]:
            failures.append((idx, "both output mismatch", repaired_both, case["expected"]))
            continue
        if repaired_both_object != expected_object:
            failures.append(
                (
                    idx,
                    "both object output mismatch",
                    repr(repaired_both_object),
                    repr(expected_object),
                )
            )
            continue
        if repaired != "":
            try:
                json.loads(repaired)
            except json.JSONDecodeError as exc:
                failures.append((idx, f"invalid json: {exc}", repaired, case["expected"]))

    base = '{"a":"1"'
    append = ',"b":2}'
    expected_append = '{"a":"1","b":2}'
    appended = repair_json_strict_prefix(base, append_content=append)
    if appended != expected_append:
        failures.append(("append", "append output mismatch", appended, expected_append))
    appended_object = repair_json_strict_prefix(base, return_object=True, append_content=append)
    expected_appended_object = json.loads(expected_append)
    if appended_object != expected_appended_object:
        failures.append(
            (
                "append",
                "append object mismatch",
                repr(appended_object),
                repr(expected_appended_object),
            )
        )

    if failures:
        for idx, reason, actual, expected in failures:
            print(f"[FAIL] case #{idx}: {reason}")
            print(f"  actual  : {actual}")
            print(f"  expected: {expected}")
        raise SystemExit(1)

    print(f"All {len(cases)} Python cases passed.")


if __name__ == "__main__":
    main()
