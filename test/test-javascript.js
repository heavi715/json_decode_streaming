const fs = require("node:fs");
const path = require("node:path");

const { repairJsonStrictPrefix } = require("../javascript/repairJson");

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
    if (repaired !== "") {
      try {
        JSON.parse(repaired);
      } catch (error) {
        failures.push({ idx, reason: `invalid json: ${error.message}`, repaired, expected: testCase.expected });
      }
    }
  });

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
