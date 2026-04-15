function isHex4At(text, start) {
  if (start + 4 > text.length) return false;
  for (let k = 0; k < 4; k += 1) {
    const ch = text[start + k];
    const isDigit = ch >= "0" && ch <= "9";
    const isLowerHex = ch >= "a" && ch <= "f";
    const isUpperHex = ch >= "A" && ch <= "F";
    if (!isDigit && !isLowerHex && !isUpperHex) return false;
  }
  return true;
}

function scanNumberEnd(text, start) {
  const n = text.length;
  let i = start;

  if (i < n && text[i] === "-") {
    i += 1;
    if (i >= n) return -1;
  }

  if (i >= n) return -1;
  if (text[i] === "0") {
    i += 1;
  } else if (text[i] >= "1" && text[i] <= "9") {
    i += 1;
    while (i < n && text[i] >= "0" && text[i] <= "9") i += 1;
  } else {
    return -1;
  }

  if (i < n && text[i] === ".") {
    if (i + 1 >= n || !(text[i + 1] >= "0" && text[i + 1] <= "9")) return i - 1;
    i += 2;
    while (i < n && text[i] >= "0" && text[i] <= "9") i += 1;
  }

  if (i < n && (text[i] === "e" || text[i] === "E")) {
    if (i + 1 >= n) return i - 1;
    let j = i + 1;
    if (text[j] === "+" || text[j] === "-") j += 1;
    if (j >= n || !(text[j] >= "0" && text[j] <= "9")) return i - 1;
    i = j + 1;
    while (i < n && text[i] >= "0" && text[i] <= "9") i += 1;
  }

  return i - 1;
}

function repairJsonStrictPrefix(text, returnObject = false, appendContent = "") {
  if (appendContent !== "") {
    text += appendContent;
  }
  if (returnObject) {
    try {
      return JSON.parse(text);
    } catch {
      // Fall back to repaired parse path.
    }
  }
  const stack = [];
  let state = "root_value";
  let inString = false;
  let escapeNext = false;
  let stringRole = "";
  let lastSafe = -1;
  let arrayWaitingValue = false;
  let objectWaitingKey = false;
  let i = 0;
  let brokeEarly = false;

  const completeValue = (idx) => {
    arrayWaitingValue = false;
    objectWaitingKey = false;
    if (stack.length === 0) {
      state = "done";
      lastSafe = idx;
      return;
    }
    const top = stack[stack.length - 1];
    if (top === "object") {
      state = "object_comma_or_end";
      lastSafe = idx;
    } else {
      state = "array_comma_or_end";
      lastSafe = idx;
    }
  };

  while (i < text.length) {
    const ch = text[i];

    if (inString) {
      if (escapeNext) {
        if ("\"\\/bfnrt".includes(ch)) {
          escapeNext = false;
          i += 1;
          continue;
        }
        if (ch === "u") {
          if (i + 4 >= text.length) {
            brokeEarly = true;
            break;
          }
          if (!isHex4At(text, i + 1)) {
            brokeEarly = true;
            break;
          }
          escapeNext = false;
          i += 5;
          continue;
        }
        brokeEarly = true;
        break;
      }
      if (ch === "\\") {
        escapeNext = true;
        i += 1;
        continue;
      }
      if (ch === "\"") {
        inString = false;
        if (stringRole === "key") {
          state = "object_colon";
        } else {
          completeValue(i);
        }
      }
      i += 1;
      continue;
    }

    if (ch === " " || ch === "\t" || ch === "\r" || ch === "\n") {
      i += 1;
      continue;
    }

    if (state === "done") {
      brokeEarly = true;
      break;
    }

    if (state === "root_value" || state === "object_value" || state === "array_value_or_end") {
      if (ch === "{") {
        stack.push("object");
        state = "object_key_or_end";
        lastSafe = i;
        i += 1;
        continue;
      }
      if (ch === "[") {
        stack.push("array");
        state = "array_value_or_end";
        lastSafe = i;
        i += 1;
        continue;
      }
      if (ch === "\"") {
        inString = true;
        stringRole = "value";
        i += 1;
        continue;
      }
      if (ch === "-" || (ch >= "0" && ch <= "9")) {
        const end = scanNumberEnd(text, i);
        if (end < i) {
          brokeEarly = true;
          break;
        }
        i = end + 1;
        completeValue(end);
        continue;
      }
      if (text.startsWith("true", i)) {
        i += 4;
        completeValue(i - 1);
        continue;
      }
      if (text.startsWith("false", i)) {
        i += 5;
        completeValue(i - 1);
        continue;
      }
      if (text.startsWith("null", i)) {
        i += 4;
        completeValue(i - 1);
        continue;
      }
      if (state === "array_value_or_end" && ch === "]") {
        if (arrayWaitingValue) {
          brokeEarly = true;
          break;
        }
        stack.pop();
        completeValue(i);
        i += 1;
        continue;
      }
      brokeEarly = true;
      break;
    }

    if (state === "object_key_or_end") {
      if (ch === "}") {
        if (objectWaitingKey) {
          brokeEarly = true;
          break;
        }
        stack.pop();
        completeValue(i);
        i += 1;
        continue;
      }
      if (ch === "\"") {
        objectWaitingKey = false;
        inString = true;
        stringRole = "key";
        i += 1;
        continue;
      }
      brokeEarly = true;
      break;
    }

    if (state === "object_colon") {
      if (ch === ":") {
        state = "object_value";
        i += 1;
        continue;
      }
      brokeEarly = true;
      break;
    }

    if (state === "object_comma_or_end") {
      if (ch === ",") {
        state = "object_key_or_end";
        objectWaitingKey = true;
        i += 1;
        continue;
      }
      if (ch === "}") {
        stack.pop();
        completeValue(i);
        i += 1;
        continue;
      }
      brokeEarly = true;
      break;
    }

    if (state === "array_comma_or_end") {
      if (ch === ",") {
        state = "array_value_or_end";
        arrayWaitingValue = true;
        i += 1;
        continue;
      }
      if (ch === "]") {
        stack.pop();
        completeValue(i);
        i += 1;
        continue;
      }
      brokeEarly = true;
      break;
    }

    brokeEarly = true;
    break;
  }

  let base = "";
  if (inString && !brokeEarly && !escapeNext && stringRole === "value") {
    base = `${text}"`;
    completeValue(text.length);
  } else {
    base = lastSafe >= 0 ? text.slice(0, lastSafe + 1) : "";
  }

  const closers = stack
    .slice()
    .reverse()
    .map((kind) => (kind === "object" ? "}" : "]"))
    .join("");

  const repaired = `${base}${closers}`;
  if (!returnObject) {
    return repaired;
  }
  if (repaired === "") {
    return null;
  }
  return JSON.parse(repaired);
}

function repairJsonStrictPrefixBoth(text, appendContent = "") {
  const repaired = repairJsonStrictPrefix(text, false, appendContent);
  if (repaired === "") {
    return [repaired, null];
  }
  return [repaired, JSON.parse(repaired)];
}

module.exports = { repairJsonStrictPrefix, repairJsonStrictPrefixBoth };
