#!/usr/bin/env python3
"""Validate test cases are synced between docs and json fixtures."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DOC_PATH = ROOT / "docs" / "test-cases.md"
JSON_PATH = ROOT / "test" / "cases.json"

ROW_RE = re.compile(
    r"^\|\s*`(?P<input>.*?)`\s*\|\s*`(?P<output>.*?)`\s*\|(?:\s*(?P<notes>.*?)\s*)?$"
)


def normalize_whitespace(value: str) -> str:
    return re.sub(r"\s+", "", value)


def parse_doc_cases(path: Path) -> list[dict[str, str]]:
    cases: list[dict[str, str]] = []
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line.startswith("|"):
            continue
        if "---" in line:
            continue
        matched = ROW_RE.match(line)
        if not matched:
            continue
        cases.append(
            {
                "input": matched.group("input"),
                "expected": matched.group("output"),
            }
        )
    return cases


def parse_json_cases(path: Path) -> list[dict[str, str]]:
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, list):
        raise ValueError("cases.json root must be a list")
    return data


def validate_expected_json(cases: list[dict[str, str]]) -> None:
    for idx, case in enumerate(cases):
        if case["expected"] == "":
            continue
        try:
            json.loads(case["expected"])
        except Exception as exc:  # noqa: BLE001
            raise ValueError(f"Invalid expected JSON at index {idx}: {exc}") from exc


def compare_cases(
    doc_cases: list[dict[str, str]], json_cases: list[dict[str, str]]
) -> list[str]:
    errors: list[str] = []

    if len(doc_cases) != len(json_cases):
        errors.append(
            f"Case count mismatch: docs={len(doc_cases)}, json={len(json_cases)}"
        )

    limit = min(len(doc_cases), len(json_cases))
    for idx in range(limit):
        doc_case = doc_cases[idx]
        json_case = json_cases[idx]
        if normalize_whitespace(doc_case["input"]) != normalize_whitespace(
            json_case["input"]
        ):
            errors.append(
                f"Input mismatch at index {idx}: "
                f"docs={doc_case['input']!r}, json={json_case['input']!r}"
            )
        if normalize_whitespace(doc_case["expected"]) != normalize_whitespace(
            json_case["expected"]
        ):
            errors.append(
                f"Expected mismatch at index {idx}: "
                f"docs={doc_case['expected']!r}, json={json_case['expected']!r}"
            )

    return errors


def main() -> int:
    doc_cases = parse_doc_cases(DOC_PATH)
    json_cases = parse_json_cases(JSON_PATH)
    validate_expected_json(json_cases)
    errors = compare_cases(doc_cases, json_cases)

    if errors:
        print("Case sync validation failed:")
        for err in errors:
            print(f"- {err}")
        return 1

    print(f"Case sync validation passed ({len(json_cases)} cases).")
    return 0


if __name__ == "__main__":
    sys.exit(main())
