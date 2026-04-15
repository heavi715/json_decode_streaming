const { repairJsonStrictPrefix, repairJsonStrictPrefixBoth } = require("../javascript/repairJson");

function parseMode() {
  const modeArg = process.argv.find((arg) => arg.startsWith("--mode="));
  const mode = modeArg ? modeArg.slice("--mode=".length) : "both";
  if (!["string", "object", "both_return", "all"].includes(mode)) {
    console.error("Invalid --mode. Use string|object|both_return|all");
    process.exit(1);
  }
  return mode;
}

function buildSamples() {
  const small = '{"a":1,"b":[1,2,3],"c":"hello"}'.repeat(4);
  const medium = JSON.stringify({
    items: Array.from({ length: 200 }, (_, i) => ({
      id: i,
      name: "x".repeat(20),
      arr: Array.from({ length: 20 }, (_, j) => j),
    })),
  });
  const large = JSON.stringify({
    items: Array.from({ length: 2000 }, (_, i) => ({
      id: i,
      name: "x".repeat(40),
      arr: Array.from({ length: 40 }, (_, j) => j),
      obj: { k: "v".repeat(10), n: i },
    })),
  });
  return { small, medium, large };
}

function iterationsFor(name) {
  if (name === "small") return 2000;
  if (name === "medium") return 400;
  return 80;
}

function runBench(name, truncated, n, mode) {
  const t0 = process.hrtime.bigint();
  for (let i = 0; i < n; i += 1) {
    if (mode === "string") {
      repairJsonStrictPrefix(truncated, false);
    } else if (mode === "object") {
      repairJsonStrictPrefix(truncated, true);
    } else {
      repairJsonStrictPrefixBoth(truncated);
    }
  }
  const elapsedNs = Number(process.hrtime.bigint() - t0);
  const dt = elapsedNs / 1e9;
  const avgUs = (dt / n) * 1e6;
  const throughputMiBS = (truncated.length * n) / dt / 1024 / 1024;
  console.log(
    `${name}/${mode}: len=${truncated.length} n=${n} avg_us=${avgUs.toFixed(1)} throughput_mib_s=${throughputMiBS.toFixed(2)}`
  );
}

const mode = parseMode();

for (const [name, text] of Object.entries(buildSamples())) {
  const truncated = text.slice(0, -17);
  const n = iterationsFor(name);
  if (mode === "string" || mode === "all") {
    runBench(name, truncated, n, "string");
  }
  if (mode === "object" || mode === "all") {
    runBench(name, truncated, n, "object");
  }
  if (mode === "both_return" || mode === "all") {
    runBench(name, truncated, n, "both_return");
  }
}
