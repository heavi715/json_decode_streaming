import json
import argparse
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from python.repair_json import repair_json_strict_prefix, repair_json_strict_prefix_both


def build_samples() -> dict[str, str]:
    return {
        "small": '{"a":1,"b":[1,2,3],"c":"hello"}' * 4,
        "medium": json.dumps(
            {
                "items": [
                    {"id": i, "name": "x" * 20, "arr": list(range(20))}
                    for i in range(200)
                ]
            }
        ),
        "large": json.dumps(
            {
                "items": [
                    {
                        "id": i,
                        "name": "x" * 40,
                        "arr": list(range(40)),
                        "obj": {"k": "v" * 10, "n": i},
                    }
                    for i in range(2000)
                ]
            }
        ),
    }


def iterations_for(name: str) -> int:
    if name == "small":
        return 2000
    if name == "medium":
        return 400
    return 80


def run_bench(name: str, truncated: str, n: int, mode: str) -> None:
    t0 = time.perf_counter()
    for _ in range(n):
        if mode == "string":
            repair_json_strict_prefix(truncated, return_object=False)
        elif mode == "object":
            repair_json_strict_prefix(truncated, return_object=True)
        else:
            repair_json_strict_prefix_both(truncated)
    dt = time.perf_counter() - t0
    avg_us = (dt / n) * 1_000_000
    throughput_mib = (len(truncated) * n) / dt / 1024 / 1024
    print(
        f"{name}/{mode}: len={len(truncated)} n={n} avg_us={avg_us:.1f} throughput_mib_s={throughput_mib:.2f}"
    )


def parse_mode() -> str:
    parser = argparse.ArgumentParser()
    parser.add_argument("--mode", choices=("string", "object", "both_return", "all"), default="all")
    args = parser.parse_args()
    return args.mode


def main() -> None:
    mode = parse_mode()
    for name, text in build_samples().items():
        truncated = text[:-17]
        n = iterations_for(name)
        if mode in ("string", "all"):
            run_bench(name, truncated, n, mode="string")
        if mode in ("object", "all"):
            run_bench(name, truncated, n, mode="object")
        if mode in ("both_return", "all"):
            run_bench(name, truncated, n, mode="both_return")


if __name__ == "__main__":
    main()
