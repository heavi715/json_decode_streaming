const fs = require("node:fs");
const path = require("node:path");

const { repairJsonStrictPrefix, repairJsonStrictPrefixBoth } = require("../javascript/repairJson");

function main() {
  const root = path.resolve(__dirname, "..");
  const cases = JSON.parse(fs.readFileSync(path.join(root, "test", "cases.json"), "utf-8"));
  const failures = [];

  cases.forEach((testCase, idx) => {
    const repaired = repairJsonStrictPrefix(testCase.input);
    if (repaired !== testCase.expected) {
      failures.push({ idx, reason: "output mismatch", repaired, expected: testCase.expected });
      return;
    }
    const repairedObject = repairJsonStrictPrefix(testCase.input, true);
    const expectedObject = testCase.expected !== "" ? JSON.parse(testCase.expected) : null;
    if (JSON.stringify(repairedObject) !== JSON.stringify(expectedObject)) {
      failures.push({
        idx,
        reason: "object output mismatch",
        repaired: JSON.stringify(repairedObject),
        expected: JSON.stringify(expectedObject),
      });
      return;
    }
    const [repairedBoth, repairedBothObject] = repairJsonStrictPrefixBoth(testCase.input);
    if (repairedBoth !== testCase.expected) {
      failures.push({ idx, reason: "both output mismatch", repaired: repairedBoth, expected: testCase.expected });
      return;
    }
    if (JSON.stringify(repairedBothObject) !== JSON.stringify(expectedObject)) {
      failures.push({
        idx,
        reason: "both object output mismatch",
        repaired: JSON.stringify(repairedBothObject),
        expected: JSON.stringify(expectedObject),
      });
      return;
    }
    if (repaired !== "") {
      try {
        JSON.parse(repaired);
      } catch (error) {
        failures.push({ idx, reason: `invalid json: ${error.message}`, repaired, expected: testCase.expected });
      }
    }
  });

  const base = '{"a":"1"';
  const append = ',"b":2}';
  const expectedAppend = '{"a":"1","b":2}';
  const appended = repairJsonStrictPrefix(base, false, append);
  if (appended !== expectedAppend) {
    failures.push({ idx: "append", reason: "append output mismatch", repaired: appended, expected: expectedAppend });
  }
  const appendedObject = repairJsonStrictPrefix(base, true, append);
  const expectedAppendedObject = JSON.parse(expectedAppend);
  if (JSON.stringify(appendedObject) !== JSON.stringify(expectedAppendedObject)) {
    failures.push({
      idx: "append",
      reason: "append object mismatch",
      repaired: JSON.stringify(appendedObject),
      expected: JSON.stringify(expectedAppendedObject),
    });
  }

  if (failures.length > 0) {
    failures.forEach((item) => {
      console.log(`[FAIL] case #${item.idx}: ${item.reason}`);
      console.log(`  actual  : ${item.repaired}`);
      console.log(`  expected: ${item.expected}`);
    });
    process.exit(1);
  }

  console.log(`All ${cases.length} JavaScript cases passed.`);
}

main();
