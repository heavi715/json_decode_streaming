const { repairJsonStrictPrefix } = require("../javascript/repairJson");

function buildJsonText(size) {
  const items = Array.from({ length: size }, (_, i) => ({
    id: i,
    name: `name-${i}`,
    tags: [`t${i % 5}`, `g${i % 7}`],
    ok: i % 2 === 0,
  }));
  return JSON.stringify({ items, meta: { count: size, source: "bench-append-js" } });
}

function splitIntoPieces(text, pieceSize) {
  const out = [];
  for (let i = 0; i < text.length; i += pieceSize) {
    out.push(text.slice(i, i + pieceSize));
  }
  return out;
}

function runOneStream(pieces, mode) {
  let accumulated = "";
  for (const piece of pieces) {
    if (mode === "append") {
      repairJsonStrictPrefix(accumulated, false, piece);
    } else {
      repairJsonStrictPrefix(accumulated + piece, false);
    }
    accumulated += piece;
  }
}

function bench(name, pieces, rounds) {
  const totalBytes = pieces.reduce((sum, p) => sum + p.length, 0) * rounds;

  const t0 = process.hrtime.bigint();
  for (let i = 0; i < rounds; i += 1) runOneStream(pieces, "append");
  const appendSec = Number(process.hrtime.bigint() - t0) / 1e9;

  const t1 = process.hrtime.bigint();
  for (let i = 0; i < rounds; i += 1) runOneStream(pieces, "full");
  const fullSec = Number(process.hrtime.bigint() - t1) / 1e9;

  const appendThroughput = totalBytes / appendSec / 1024 / 1024;
  const fullThroughput = totalBytes / fullSec / 1024 / 1024;
  const speedup = fullSec / appendSec;
  console.log(
    `${name}: pieces=${pieces.length} rounds=${rounds} ` +
      `append_s=${appendSec.toFixed(3)} full_s=${fullSec.toFixed(3)} ` +
      `append_mib_s=${appendThroughput.toFixed(2)} full_mib_s=${fullThroughput.toFixed(2)} speedup=${speedup.toFixed(2)}x`
  );
}

function main() {
  const scenarios = [
    { name: "small", size: 120, pieceSize: 24, rounds: 200 },
    { name: "medium", size: 600, pieceSize: 32, rounds: 80 },
    { name: "large", size: 1600, pieceSize: 48, rounds: 30 },
  ];

  for (const scenario of scenarios) {
    const text = buildJsonText(scenario.size);
    const pieces = splitIntoPieces(text, scenario.pieceSize);
    bench(scenario.name, pieces, scenario.rounds);
  }
}

main();
