import json
import random
import string
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from python.repair_json import repair_json_strict_prefix


def rand_str(n: int) -> str:
    alphabet = string.ascii_letters + string.digits + " _-:/\\"
    return "".join(random.choice(alphabet) for _ in range(n))


def rand_value(depth: int):
    if depth <= 0:
        leaf = random.choice(
            [
                None,
                True,
                False,
                random.randint(-1000, 1000),
                random.random() * 1000,
                rand_str(random.randint(0, 20)),
            ]
        )
        return leaf

    t = random.choice(["obj", "arr", "leaf"])
    if t == "leaf":
        return rand_value(0)
    if t == "arr":
        return [rand_value(depth - 1) for _ in range(random.randint(0, 5))]
    d = {}
    for _ in range(random.randint(0, 5)):
        d[rand_str(random.randint(1, 10))] = rand_value(depth - 1)
    return d


def main() -> None:
    random.seed(42)
    rounds = 500
    failures = []

    for idx in range(rounds):
        source = rand_value(depth=4)
        text = json.dumps(source, ensure_ascii=False)
        if not text:
            continue
        cut = random.randint(0, len(text))
        truncated = text[:cut]
        repaired = repair_json_strict_prefix(truncated)

        if repaired:
            try:
                json.loads(repaired)
            except json.JSONDecodeError as exc:
                failures.append((idx, truncated, repaired, str(exc)))
                continue

        # Idempotence: repairing repaired output should not change it.
        repaired2 = repair_json_strict_prefix(repaired)
        if repaired2 != repaired:
            failures.append((idx, truncated, repaired, "not idempotent"))

    if failures:
        for idx, truncated, repaired, reason in failures[:10]:
            print(f"[FAIL] round #{idx}: {reason}")
            print(f"  truncated: {truncated}")
            print(f"  repaired : {repaired}")
        raise SystemExit(1)

    print(f"Fuzz passed: {rounds} random truncation cases.")


if __name__ == "__main__":
    main()
