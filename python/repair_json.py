from __future__ import annotations

import re


NUMBER_RE = re.compile(r"-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?")
HEX4_RE = re.compile(r"^[0-9a-fA-F]{4}$")


def repair_json_strict_prefix(text: str) -> str:
    stack: list[str] = []
    state = "root_value"
    in_string = False
    escape_next = False
    string_role = ""
    last_safe = -1
    array_waiting_value = False
    object_waiting_key = False

    i = 0
    n = len(text)
    broke_early = False

    def complete_value(idx: int) -> None:
        nonlocal state, last_safe, array_waiting_value, object_waiting_key
        array_waiting_value = False
        object_waiting_key = False
        if not stack:
            state = "done"
            last_safe = idx
            return
        top = stack[-1]
        if top == "object":
            state = "object_comma_or_end"
            last_safe = idx
        else:
            state = "array_comma_or_end"
            last_safe = idx

    while i < n:
        ch = text[i]

        if in_string:
            if escape_next:
                if ch in '"\\/bfnrt':
                    escape_next = False
                    i += 1
                    continue
                if ch == "u":
                    if i + 4 >= n:
                        broke_early = True
                        break
                    if not HEX4_RE.match(text[i + 1 : i + 5]):
                        broke_early = True
                        break
                    escape_next = False
                    i += 5
                    continue
                broke_early = True
                break
            if ch == "\\":
                escape_next = True
                i += 1
                continue
            if ch == '"':
                in_string = False
                if string_role == "key":
                    state = "object_colon"
                else:
                    complete_value(i)
                i += 1
                continue
            i += 1
            continue

        if ch in " \t\r\n":
            i += 1
            continue

        if state == "done":
            broke_early = True
            break

        if state in ("root_value", "object_value", "array_value_or_end"):
            if ch == "{":
                stack.append("object")
                state = "object_key_or_end"
                last_safe = i
                i += 1
                continue
            if ch == "[":
                stack.append("array")
                state = "array_value_or_end"
                last_safe = i
                i += 1
                continue
            if ch == '"':
                in_string = True
                string_role = "value"
                i += 1
                continue
            if ch in "-0123456789":
                m = NUMBER_RE.match(text[i:])
                if not m:
                    broke_early = True
                    break
                end = i + len(m.group(0)) - 1
                i = end + 1
                complete_value(end)
                continue
            if text.startswith("true", i):
                i += 4
                complete_value(i - 1)
                continue
            if text.startswith("false", i):
                i += 5
                complete_value(i - 1)
                continue
            if text.startswith("null", i):
                i += 4
                complete_value(i - 1)
                continue
            if state == "array_value_or_end" and ch == "]":
                if array_waiting_value:
                    broke_early = True
                    break
                stack.pop()
                complete_value(i)
                i += 1
                continue
            broke_early = True
            break

        if state == "object_key_or_end":
            if ch == "}":
                if object_waiting_key:
                    broke_early = True
                    break
                stack.pop()
                complete_value(i)
                i += 1
                continue
            if ch == '"':
                object_waiting_key = False
                in_string = True
                string_role = "key"
                i += 1
                continue
            broke_early = True
            break

        if state == "object_colon":
            if ch == ":":
                state = "object_value"
                i += 1
                continue
            broke_early = True
            break

        if state == "object_comma_or_end":
            if ch == ",":
                state = "object_key_or_end"
                object_waiting_key = True
                i += 1
                continue
            if ch == "}":
                stack.pop()
                complete_value(i)
                i += 1
                continue
            broke_early = True
            break

        if state == "array_comma_or_end":
            if ch == ",":
                state = "array_value_or_end"
                array_waiting_value = True
                i += 1
                continue
            if ch == "]":
                stack.pop()
                complete_value(i)
                i += 1
                continue
            broke_early = True
            break

        broke_early = True
        break

    if in_string and not broke_early and string_role == "value" and not escape_next:
        base = text + '"'
        complete_value(n)
    elif broke_early:
        base = text[: last_safe + 1] if last_safe >= 0 else ""
    else:
        base = text[: last_safe + 1] if last_safe >= 0 else ""

    closers = "".join("}" if kind == "object" else "]" for kind in reversed(stack))
    return base + closers


if __name__ == "__main__":
    samples = [
        '{"a":"b',
        '{"a":[1,2,',
        '{"a":"1","b":',
        '{"a":"1","b":"',
    ]
    for sample in samples:
        print(sample, "=>", repair_json_strict_prefix(sample))
