const NUMBER_RE = /^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?/;

function repairJsonStrictPrefix(text) {
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
          const hex = text.slice(i + 1, i + 5);
          if (!/^[0-9a-fA-F]{4}$/.test(hex)) {
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

    if (/\s/.test(ch)) {
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
      if ("-0123456789".includes(ch)) {
        const m = text.slice(i).match(NUMBER_RE);
        if (!m) {
          brokeEarly = true;
          break;
        }
        const end = i + m[0].length - 1;
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

  return `${base}${closers}`;
}

module.exports = { repairJsonStrictPrefix };
