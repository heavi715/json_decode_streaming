#!/usr/bin/env python3
"""Print per-section case stats from docs and compare with fixture count."""

from __future__ import annotations

import json
import re
import sys
from collections import OrderedDict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DOC_PATH = ROOT / "docs" / "test-cases.md"
JSON_PATH = ROOT / "test" / "cases.json"

SECTION_RE = re.compile(r"^##\s+(.+)$")
ROW_RE = re.compile(r"^\|\s*`.*?`\s*\|\s*`.*?`")


def read_doc_section_counts(path: Path) -> OrderedDict[str, int]:
    counts: OrderedDict[str, int] = OrderedDict()
    section_name = ""
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        section = SECTION_RE.match(line)
        if section:
            section_name = section.group(1)
            counts.setdefault(section_name, 0)
            continue
        if not section_name:
            continue
        if ROW_RE.match(line):
            counts[section_name] += 1
    return counts


def main() -> int:
    section_counts = read_doc_section_counts(DOC_PATH)
    fixture_count = len(json.loads(JSON_PATH.read_text(encoding="utf-8")))
    doc_total = sum(section_counts.values())

    print("Case stats by section:")
    for section, count in section_counts.items():
        if count > 0:
            print(f"- {section}: {count}")
    print(f"- TOTAL (docs): {doc_total}")
    print(f"- TOTAL (fixture): {fixture_count}")

    if doc_total != fixture_count:
        print("ERROR: docs total does not match fixture total.", file=sys.stderr)
        return 1

    print("Counts are consistent.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
