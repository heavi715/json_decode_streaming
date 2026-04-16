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

class RepairState {
  constructor() {
    this.text = "";
    this.stack = [];
    this.state = "root_value";
    this.inString = false;
    this.escapeNext = false;
    this.stringRole = "";
    this.lastSafe = -1;
    this.arrayWaitingValue = false;
    this.objectWaitingKey = false;
    this.i = 0;
    this.brokeEarly = false;
  }

  clone() {
    const cloned = new RepairState();
    cloned.text = this.text;
    cloned.stack = this.stack.slice();
    cloned.state = this.state;
    cloned.inString = this.inString;
    cloned.escapeNext = this.escapeNext;
    cloned.stringRole = this.stringRole;
    cloned.lastSafe = this.lastSafe;
    cloned.arrayWaitingValue = this.arrayWaitingValue;
    cloned.objectWaitingKey = this.objectWaitingKey;
    cloned.i = this.i;
    cloned.brokeEarly = this.brokeEarly;
    return cloned;
  }

  completeValue(idx) {
    this.arrayWaitingValue = false;
    this.objectWaitingKey = false;
    if (this.stack.length === 0) {
      this.state = "done";
      this.lastSafe = idx;
      return;
    }
    const top = this.stack[this.stack.length - 1];
    if (top === "object") {
      this.state = "object_comma_or_end";
      this.lastSafe = idx;
    } else {
      this.state = "array_comma_or_end";
      this.lastSafe = idx;
    }
  }

  feed(chunk) {
    if (!chunk) return;
    if (this.brokeEarly) {
      this.text += chunk;
      return;
    }
    this.text += chunk;

    while (this.i < this.text.length) {
      const ch = this.text[this.i];

      if (this.inString) {
        if (this.escapeNext) {
          if ("\"\\/bfnrt".includes(ch)) {
            this.escapeNext = false;
            this.i += 1;
            continue;
          }
          if (ch === "u") {
            if (this.i + 4 >= this.text.length) {
              this.brokeEarly = true;
              break;
            }
            if (!isHex4At(this.text, this.i + 1)) {
              this.brokeEarly = true;
              break;
            }
            this.escapeNext = false;
            this.i += 5;
            continue;
          }
          this.brokeEarly = true;
          break;
        }
        if (ch === "\\") {
          this.escapeNext = true;
          this.i += 1;
          continue;
        }
        if (ch === "\"") {
          this.inString = false;
          if (this.stringRole === "key") {
            this.state = "object_colon";
          } else {
            this.completeValue(this.i);
          }
        }
        this.i += 1;
        continue;
      }

      if (ch === " " || ch === "\t" || ch === "\r" || ch === "\n") {
        this.i += 1;
        continue;
      }

      if (this.state === "done") {
        this.brokeEarly = true;
        break;
      }

      if (this.state === "root_value" || this.state === "object_value" || this.state === "array_value_or_end") {
        if (ch === "{") {
          this.stack.push("object");
          this.state = "object_key_or_end";
          this.lastSafe = this.i;
          this.i += 1;
          continue;
        }
        if (ch === "[") {
          this.stack.push("array");
          this.state = "array_value_or_end";
          this.lastSafe = this.i;
          this.i += 1;
          continue;
        }
        if (ch === "\"") {
          this.inString = true;
          this.stringRole = "value";
          this.i += 1;
          continue;
        }
        if (ch === "-" || (ch >= "0" && ch <= "9")) {
          const end = scanNumberEnd(this.text, this.i);
          if (end < this.i) {
            this.brokeEarly = true;
            break;
          }
          this.i = end + 1;
          this.completeValue(end);
          continue;
        }
        if (this.text.startsWith("true", this.i)) {
          this.i += 4;
          this.completeValue(this.i - 1);
          continue;
        }
        if (this.text.startsWith("false", this.i)) {
          this.i += 5;
          this.completeValue(this.i - 1);
          continue;
        }
        if (this.text.startsWith("null", this.i)) {
          this.i += 4;
          this.completeValue(this.i - 1);
          continue;
        }
        if (this.state === "array_value_or_end" && ch === "]") {
          if (this.arrayWaitingValue) {
            this.brokeEarly = true;
            break;
          }
          this.stack.pop();
          this.completeValue(this.i);
          this.i += 1;
          continue;
        }
        this.brokeEarly = true;
        break;
      }

      if (this.state === "object_key_or_end") {
        if (ch === "}") {
          if (this.objectWaitingKey) {
            this.brokeEarly = true;
            break;
          }
          this.stack.pop();
          this.completeValue(this.i);
          this.i += 1;
          continue;
        }
        if (ch === "\"") {
          this.objectWaitingKey = false;
          this.inString = true;
          this.stringRole = "key";
          this.i += 1;
          continue;
        }
        this.brokeEarly = true;
        break;
      }

      if (this.state === "object_colon") {
        if (ch === ":") {
          this.state = "object_value";
          this.i += 1;
          continue;
        }
        this.brokeEarly = true;
        break;
      }

      if (this.state === "object_comma_or_end") {
        if (ch === ",") {
          this.state = "object_key_or_end";
          this.objectWaitingKey = true;
          this.i += 1;
          continue;
        }
        if (ch === "}") {
          this.stack.pop();
          this.completeValue(this.i);
          this.i += 1;
          continue;
        }
        this.brokeEarly = true;
        break;
      }

      if (this.state === "array_comma_or_end") {
        if (ch === ",") {
          this.state = "array_value_or_end";
          this.arrayWaitingValue = true;
          this.i += 1;
          continue;
        }
        if (ch === "]") {
          this.stack.pop();
          this.completeValue(this.i);
          this.i += 1;
          continue;
        }
        this.brokeEarly = true;
        break;
      }

      this.brokeEarly = true;
      break;
    }
  }

  snapshot() {
    let base = "";
    if (this.inString && !this.brokeEarly && !this.escapeNext && this.stringRole === "value") {
      base = `${this.text}"`;
    } else {
      base = this.lastSafe >= 0 ? this.text.slice(0, this.lastSafe + 1) : "";
    }
    const closers = this.stack
      .slice()
      .reverse()
      .map((kind) => (kind === "object" ? "}" : "]"))
      .join("");
    return `${base}${closers}`;
  }
}

const appendStateCache = new Map();
const APPEND_CACHE_MAX_ENTRIES = 256;

function cacheAppendState(text, state) {
  appendStateCache.set(text, state);
  if (appendStateCache.size <= APPEND_CACHE_MAX_ENTRIES) return;
  const oldestKey = appendStateCache.keys().next().value;
  appendStateCache.delete(oldestKey);
}

function parseAndRepairFromScratch(text) {
  const state = new RepairState();
  state.feed(text);
  return state.snapshot();
}

function repairJsonStrictPrefix(text, returnObject = false, appendContent = "") {
  const fullText = appendContent !== "" ? text + appendContent : text;
  if (returnObject) {
    try {
      return JSON.parse(fullText);
    } catch {
      // Fall back to repaired parse path.
    }
  }
  let repaired;
  if (appendContent !== "") {
    const cachedBaseState = appendStateCache.get(text);
    if (cachedBaseState) {
      const nextState = cachedBaseState.clone();
      nextState.feed(appendContent);
      repaired = nextState.snapshot();
      cacheAppendState(fullText, nextState);
    } else {
      repaired = parseAndRepairFromScratch(fullText);
      const fullState = new RepairState();
      fullState.feed(fullText);
      cacheAppendState(fullText, fullState);
    }
  } else {
    repaired = parseAndRepairFromScratch(fullText);
    const baseState = new RepairState();
    baseState.feed(fullText);
    cacheAppendState(fullText, baseState);
  }
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
