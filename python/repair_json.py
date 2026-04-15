from __future__ import annotations

import json


def _is_hex4(text: str, start: int) -> bool:
    if start + 4 > len(text):
        return False
    for ch in text[start : start + 4]:
        if not ("0" <= ch <= "9" or "a" <= ch <= "f" or "A" <= ch <= "F"):
            return False
    return True


def _scan_number_end(text: str, start: int) -> int:
    i = start
    n = len(text)

    if i < n and text[i] == "-":
        i += 1
        if i >= n:
            return -1

    if i >= n:
        return -1

    if text[i] == "0":
        i += 1
    elif "1" <= text[i] <= "9":
        i += 1
        while i < n and "0" <= text[i] <= "9":
            i += 1
    else:
        return -1

    if i < n and text[i] == ".":
        if i + 1 >= n or not ("0" <= text[i + 1] <= "9"):
            return i - 1
        i += 2
        while i < n and "0" <= text[i] <= "9":
            i += 1

    if i < n and (text[i] == "e" or text[i] == "E"):
        if i + 1 >= n:
            return i - 1
        j = i + 1
        if text[j] == "+" or text[j] == "-":
            j += 1
        if j >= n or not ("0" <= text[j] <= "9"):
            return i - 1
        i = j + 1
        while i < n and "0" <= text[i] <= "9":
            i += 1

    return i - 1


def _is_literal_prefix_at_eof(text: str, start: int, literal: str) -> bool:
    tail = text[start:]
    return len(tail) < len(literal) and literal.startswith(tail)


class RepairState:
    def __init__(self, compact_prefix: bool = True, conservative_eof: bool = True) -> None:
        self.text = ""
        self.prefix_parts: list[str] = []
        self.compact_prefix = compact_prefix
        self.conservative_eof = conservative_eof
        self.stack: list[str] = []
        self.state = "root_value"
        self.in_string = False
        self.escape_next = False
        self.string_role = ""
        self.last_safe = -1
        self.array_waiting_value = False
        self.object_waiting_key = False
        self.i = 0
        self.broke_early = False

    def _append_prefix(self, segment: str) -> None:
        if not segment:
            return
        self.prefix_parts.append(segment)
        # Avoid very large list bookkeeping overhead on long streams.
        if len(self.prefix_parts) >= 4096:
            self.prefix_parts = ["".join(self.prefix_parts)]

    def _complete_value(self, idx: int) -> None:
        self.array_waiting_value = False
        self.object_waiting_key = False
        if not self.stack:
            self.state = "done"
            self.last_safe = idx
            return
        top = self.stack[-1]
        if top == "object":
            self.state = "object_comma_or_end"
        else:
            self.state = "array_comma_or_end"
        self.last_safe = idx

    def _maybe_compact_prefix(self) -> int:
        # Compact confirmed safe prefix to keep memory bounded in long streams.
        if not self.compact_prefix:
            return 0
        if self.last_safe < 0:
            return 0
        if self.in_string or self.escape_next:
            return 0
        cut = self.last_safe + 1
        if cut <= 0:
            return 0
        self._append_prefix(self.text[:cut])
        self.text = self.text[cut:]
        self.i -= cut
        self.last_safe = -1
        return cut

    def feed(self, chunk: str) -> None:
        if not chunk:
            return
        if self.broke_early:
            self.text += chunk
            return
        self.text += chunk
        n = len(self.text)

        while self.i < n:
            ch = self.text[self.i]

            if self.in_string:
                if self.escape_next:
                    if ch in '"\\/bfnrt':
                        self.escape_next = False
                        self.i += 1
                        continue
                    if ch == "u":
                        if self.i + 4 >= n:
                            # In incremental mode we may be waiting for remaining hex digits.
                            break
                        if not _is_hex4(self.text, self.i + 1):
                            self.broke_early = True
                            break
                        self.escape_next = False
                        self.i += 5
                        continue
                    self.broke_early = True
                    break
                if ch == "\\":
                    self.escape_next = True
                    self.i += 1
                    continue
                if ch == '"':
                    self.in_string = False
                    if self.string_role == "key":
                        self.state = "object_colon"
                    else:
                        self._complete_value(self.i)
                        n -= self._maybe_compact_prefix()
                    self.i += 1
                    continue
                self.i += 1
                continue

            if ch in " \t\r\n":
                self.i += 1
                continue

            if self.state == "done":
                self.broke_early = True
                break

            if self.state in ("root_value", "object_value", "array_value_or_end"):
                if ch == "{":
                    self.stack.append("object")
                    self.state = "object_key_or_end"
                    self.last_safe = self.i
                    self.i += 1
                    continue
                if ch == "[":
                    self.stack.append("array")
                    self.state = "array_value_or_end"
                    self.last_safe = self.i
                    self.i += 1
                    continue
                if ch == '"':
                    self.in_string = True
                    self.string_role = "value"
                    self.i += 1
                    continue
                if ch == "-" or ("0" <= ch <= "9"):
                    end = _scan_number_end(self.text, self.i)
                    if end < self.i:
                        if ch == "-" and self.i == n - 1:
                            # Wait for more digits in next chunk.
                            break
                        self.broke_early = True
                        break
                    if self.conservative_eof and end == n - 1:
                        break
                    self.i = end + 1
                    self._complete_value(end)
                    n -= self._maybe_compact_prefix()
                    continue
                if _is_literal_prefix_at_eof(self.text, self.i, "true"):
                    break
                if self.text.startswith("true", self.i):
                    if self.conservative_eof and self.i + 4 == n:
                        break
                    self.i += 4
                    self._complete_value(self.i - 1)
                    n -= self._maybe_compact_prefix()
                    continue
                if _is_literal_prefix_at_eof(self.text, self.i, "false"):
                    break
                if self.text.startswith("false", self.i):
                    if self.conservative_eof and self.i + 5 == n:
                        break
                    self.i += 5
                    self._complete_value(self.i - 1)
                    n -= self._maybe_compact_prefix()
                    continue
                if _is_literal_prefix_at_eof(self.text, self.i, "null"):
                    break
                if self.text.startswith("null", self.i):
                    if self.conservative_eof and self.i + 4 == n:
                        break
                    self.i += 4
                    self._complete_value(self.i - 1)
                    n -= self._maybe_compact_prefix()
                    continue
                if self.state == "array_value_or_end" and ch == "]":
                    if self.array_waiting_value:
                        self.broke_early = True
                        break
                    self.stack.pop()
                    self._complete_value(self.i)
                    n -= self._maybe_compact_prefix()
                    self.i += 1
                    continue
                self.broke_early = True
                break

            if self.state == "object_key_or_end":
                if ch == "}":
                    if self.object_waiting_key:
                        self.broke_early = True
                        break
                    self.stack.pop()
                    self._complete_value(self.i)
                    n -= self._maybe_compact_prefix()
                    self.i += 1
                    continue
                if ch == '"':
                    self.object_waiting_key = False
                    self.in_string = True
                    self.string_role = "key"
                    self.i += 1
                    continue
                self.broke_early = True
                break

            if self.state == "object_colon":
                if ch == ":":
                    self.state = "object_value"
                    self.i += 1
                    continue
                self.broke_early = True
                break

            if self.state == "object_comma_or_end":
                if ch == ",":
                    self.state = "object_key_or_end"
                    self.object_waiting_key = True
                    self.i += 1
                    continue
                if ch == "}":
                    self.stack.pop()
                    self._complete_value(self.i)
                    n -= self._maybe_compact_prefix()
                    self.i += 1
                    continue
                self.broke_early = True
                break

            if self.state == "array_comma_or_end":
                if ch == ",":
                    self.state = "array_value_or_end"
                    self.array_waiting_value = True
                    self.i += 1
                    continue
                if ch == "]":
                    self.stack.pop()
                    self._complete_value(self.i)
                    n -= self._maybe_compact_prefix()
                    self.i += 1
                    continue
                self.broke_early = True
                break

            self.broke_early = True
            break

    def snapshot(self) -> str:
        prefix = "".join(self.prefix_parts)
        base = self.text[: self.last_safe + 1] if self.last_safe >= 0 else ""
        if (
            self.in_string
            and not self.broke_early
            and self.string_role == "value"
            and not self.escape_next
        ):
            base = prefix + self.text + '"'
            closers = "".join("}" if kind == "object" else "]" for kind in reversed(self.stack))
            if not self.stack:
                return base
            return base + closers
        base = prefix + base
        closers = "".join("}" if kind == "object" else "]" for kind in reversed(self.stack))
        return base + closers

    def finalize(self) -> str:
        if not self.conservative_eof:
            return self.snapshot()
        full_text = "".join(self.prefix_parts) + self.text
        fresh = RepairState(compact_prefix=False, conservative_eof=False)
        fresh.feed(full_text)
        return fresh.snapshot()


def repair_json_strict_prefix(
    text: str, return_object: bool = False, append_content: str = ""
):
    if append_content:
        text += append_content
    if return_object:
        try:
            return json.loads(text)
        except json.JSONDecodeError:
            pass
    state = RepairState(compact_prefix=False, conservative_eof=False)
    state.feed(text)
    repaired = state.finalize()
    if not return_object:
        return repaired
    if repaired == "":
        return None
    return json.loads(repaired)


def repair_json_strict_prefix_both(
    text: str, append_content: str = ""
) -> tuple[str, object | None]:
    repaired = repair_json_strict_prefix(text, append_content=append_content)
    if repaired == "":
        return repaired, None
    return repaired, json.loads(repaired)


if __name__ == "__main__":
    samples = [
        '{"a":"b',
        '{"a":[1,2,',
        '{"a":"1","b":',
        '{"a":"1","b":"',
    ]
    for sample in samples:
        print(sample, "=>", repair_json_strict_prefix(sample))
