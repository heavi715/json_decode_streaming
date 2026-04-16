const {
  applyRepairJsonAppendCachePreset,
  repairJsonStrictPrefix,
} = require("../javascript/repairJson");

function runStream(label, preset) {
  applyRepairJsonAppendCachePreset(preset, true);
  let accumulated = "";
  const pieces = ['{"items":[', '{"id":1},', '{"id":2},', '{"id":3}', ']}'];
  let last = "";
  for (const piece of pieces) {
    last = repairJsonStrictPrefix(accumulated, false, piece);
    accumulated += piece;
  }
  console.log(`${label}: ${last}`);
}

function main() {
  runStream("preset-low-memory", "low_memory");
  runStream("preset-high-throughput", "high_throughput");
}

main();
