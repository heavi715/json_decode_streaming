from __future__ import annotations

import json
import time
from collections import OrderedDict


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

    def clone(self) -> "RepairState":
        cloned = RepairState(
            compact_prefix=self.compact_prefix,
            conservative_eof=self.conservative_eof,
        )
        cloned.text = self.text
        cloned.prefix_parts = self.prefix_parts[:]
        cloned.stack = self.stack[:]
        cloned.state = self.state
        cloned.in_string = self.in_string
        cloned.escape_next = self.escape_next
        cloned.string_role = self.string_role
        cloned.last_safe = self.last_safe
        cloned.array_waiting_value = self.array_waiting_value
        cloned.object_waiting_key = self.object_waiting_key
        cloned.i = self.i
        cloned.broke_early = self.broke_early
        return cloned


_append_cache: "OrderedDict[str, dict]" = OrderedDict()
_append_cache_max_entries = 256
_append_cache_max_total_bytes = 4 * 1024 * 1024
_append_cache_ttl_seconds = 120.0
_append_cache_total_bytes = 0
APPEND_CACHE_PRESETS: dict[str, dict[str, float | int]] = {
    "default": {
        "max_entries": 256,
        "max_total_bytes": 4 * 1024 * 1024,
        "ttl_seconds": 120.0,
    },
    "low_memory": {
        "max_entries": 64,
        "max_total_bytes": 512 * 1024,
        "ttl_seconds": 15.0,
    },
    "high_throughput": {
        "max_entries": 1024,
        "max_total_bytes": 16 * 1024 * 1024,
        "ttl_seconds": 600.0,
    },
}


def _estimate_key_bytes(text: str) -> int:
    return len(text.encode("utf-8"))


def _prune_append_cache(now: float) -> None:
    global _append_cache_total_bytes
    while _append_cache:
        key, entry = next(iter(_append_cache.items()))
        over_limit = (
            len(_append_cache) > _append_cache_max_entries
            or _append_cache_total_bytes > _append_cache_max_total_bytes
        )
        expired = entry["expires_at"] <= now
        if not over_limit and not expired:
            break
        _append_cache.pop(key, None)
        _append_cache_total_bytes -= entry["key_bytes"]


def _cache_append_state(text: str, state: RepairState) -> None:
    global _append_cache_total_bytes
    now = time.time()
    key_bytes = _estimate_key_bytes(text)
    existing = _append_cache.pop(text, None)
    if existing is not None:
        _append_cache_total_bytes -= existing["key_bytes"]
    _append_cache[text] = {
        "state": state.clone(),
        "key_bytes": key_bytes,
        "expires_at": now + _append_cache_ttl_seconds,
    }
    _append_cache_total_bytes += key_bytes
    _prune_append_cache(now)


def _get_cached_append_state(text: str) -> RepairState | None:
    entry = _append_cache.get(text)
    if entry is None:
        return None
    now = time.time()
    if entry["expires_at"] <= now:
        _append_cache.pop(text, None)
        global _append_cache_total_bytes
        _append_cache_total_bytes -= entry["key_bytes"]
        return None
    _append_cache.pop(text, None)
    entry["expires_at"] = now + _append_cache_ttl_seconds
    _append_cache[text] = entry
    return entry["state"].clone()


def set_repair_json_append_cache_config(
    *,
    max_entries: int | None = None,
    max_total_bytes: int | None = None,
    ttl_seconds: float | None = None,
    clear: bool = False,
) -> None:
    global _append_cache_max_entries, _append_cache_max_total_bytes, _append_cache_ttl_seconds, _append_cache_total_bytes
    if max_entries is not None:
        if max_entries < 1:
            raise ValueError("max_entries must be >= 1")
        _append_cache_max_entries = max_entries
    if max_total_bytes is not None:
        if max_total_bytes < 1024:
            raise ValueError("max_total_bytes must be >= 1024")
        _append_cache_max_total_bytes = max_total_bytes
    if ttl_seconds is not None:
        if ttl_seconds < 0.1:
            raise ValueError("ttl_seconds must be >= 0.1")
        _append_cache_ttl_seconds = ttl_seconds
    if clear:
        _append_cache.clear()
        _append_cache_total_bytes = 0
    _prune_append_cache(time.time())


def apply_repair_json_append_cache_preset(preset: str, *, clear: bool = True) -> None:
    picked = APPEND_CACHE_PRESETS.get(preset)
    if picked is None:
        raise ValueError(f"unknown cache preset: {preset}")
    set_repair_json_append_cache_config(
        max_entries=int(picked["max_entries"]),
        max_total_bytes=int(picked["max_total_bytes"]),
        ttl_seconds=float(picked["ttl_seconds"]),
        clear=clear,
    )


def repair_json_strict_prefix(
    text: str, return_object: bool = False, append_content: str = ""
):
    full_text = text + append_content if append_content else text
    if return_object:
        try:
            return json.loads(full_text)
        except json.JSONDecodeError:
            pass
    if append_content:
        state = _get_cached_append_state(text)
        if state is not None:
            state.feed(append_content)
            repaired = state.finalize()
            _cache_append_state(full_text, state)
        else:
            state = RepairState(compact_prefix=False, conservative_eof=False)
            state.feed(full_text)
            repaired = state.finalize()
            _cache_append_state(full_text, state)
    else:
        state = RepairState(compact_prefix=False, conservative_eof=False)
        state.feed(full_text)
        repaired = state.finalize()
        _cache_append_state(full_text, state)
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
