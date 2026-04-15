const { execSync } = require("node:child_process");
const { repairJsonStrictPrefix } = require("../javascript/repairJson");

const apiKey = process.env.AI_STREAM_API_KEY;
if (!apiKey) {
  console.error("Missing AI_STREAM_API_KEY");
  process.exit(1);
}

const apiUrl = process.env.AI_STREAM_URL || "http://new-api.bangong.knowbox.cn/v1/chat/completions";
const model = process.env.AI_STREAM_MODEL || "claude-opus-4-20250514";
const prompt = process.env.AI_STREAM_PROMPT || '只返回json，格式: {"ok":true,"msg":"..."}';
const printSnapshots = process.env.AI_STREAM_PRINT_SNAPSHOTS !== "0";
const maxSnapshots = Number.parseInt(process.env.AI_STREAM_MAX_SNAPSHOTS || "20", 10);

const payload = JSON.stringify({
  model,
  messages: [{ role: "user", content: prompt }],
  stream: true,
});

const curlCmd =
  `curl -sS -N ${JSON.stringify(apiUrl)} ` +
  `-H ${JSON.stringify("Content-Type: application/json")} ` +
  `-H ${JSON.stringify(`Authorization: Bearer ${apiKey}`)} ` +
  `-d ${JSON.stringify(payload)}`;

let raw = "";
try {
  raw = execSync(curlCmd, { encoding: "utf-8", stdio: ["ignore", "pipe", "pipe"] });
} catch (error) {
  console.error("curl request failed");
  if (error.stderr) {
    console.error(String(error.stderr));
  }
  process.exit(1);
}

const lines = raw.split(/\r?\n/);
let accumulated = "";
let chunkCount = 0;
let snapshotCount = 0;
let printed = 0;
let eventCount = 0;
let skipped = 0;
const debugSamples = [];

for (const rawLine of lines) {
  const line = rawLine.trim();
  if (!line.startsWith("data:")) continue;
  const payloadLine = line.slice(5).trimStart();
  if (payloadLine === "[DONE]") break;
  eventCount += 1;

  let event;
  try {
    event = JSON.parse(payloadLine);
  } catch {
    skipped += 1;
    if (debugSamples.length < 3) debugSamples.push(payloadLine.slice(0, 200));
    continue;
  }

  const choices = event.choices || [];
  const delta = (choices[0] || {}).delta || {};
  const piece = typeof delta.content === "string" ? delta.content : "";
  if (!piece) {
    skipped += 1;
    if (debugSamples.length < 3) debugSamples.push(JSON.stringify(event).slice(0, 200));
    continue;
  }

  chunkCount += 1;
  const obj = repairJsonStrictPrefix(accumulated, true, piece);
  accumulated += piece;
  if (obj !== null) {
    snapshotCount += 1;
    if (printSnapshots && printed < maxSnapshots) {
      printed += 1;
      console.log(`snapshot#${snapshotCount}: ${JSON.stringify(obj)}`);
    }
  }
}

const final = repairJsonStrictPrefix(accumulated);
const finalObj = repairJsonStrictPrefix(accumulated, true);
if (finalObj === null) {
  console.log("Failed to parse streamed content as JSON.");
  console.log(`Repaired text: ${final}`);
  console.log(`events: ${eventCount}, content chunks: ${chunkCount}, skipped events: ${skipped}`);
  if (raw.trim()) {
    console.log("raw response sample:");
    console.log(raw.slice(0, 500));
  }
  if (debugSamples.length) {
    console.log("sample skipped payloads:");
    debugSamples.forEach((s) => console.log(s));
  }
  process.exit(1);
}

console.log(`content chunks: ${chunkCount}`);
console.log(`object snapshots: ${snapshotCount}`);
if (printSnapshots && snapshotCount > maxSnapshots) {
  console.log(`snapshot output truncated: printed ${maxSnapshots} of ${snapshotCount}`);
}
console.log(`events: ${eventCount}, skipped events: ${skipped}`);
console.log(`final repaired json: ${final}`);
console.log(`final object type: ${Array.isArray(finalObj) ? "array" : typeof finalObj}`);
