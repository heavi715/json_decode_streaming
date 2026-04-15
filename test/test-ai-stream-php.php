<?php

declare(strict_types=1);

require_once __DIR__ . '/../php/RepairJson.php';

function extract_piece(array $event): string
{
    $choices = $event['choices'] ?? [];
    if (!is_array($choices) || !isset($choices[0]) || !is_array($choices[0])) {
        return '';
    }
    $choice0 = $choices[0];
    $delta = $choice0['delta'] ?? [];
    if (!is_array($delta)) {
        return '';
    }

    $content = $delta['content'] ?? '';
    if (is_string($content)) {
        return $content;
    }
    if (is_array($content)) {
        $parts = [];
        foreach ($content as $item) {
            if (is_string($item)) {
                $parts[] = $item;
                continue;
            }
            if (!is_array($item)) {
                continue;
            }
            $text = $item['text'] ?? '';
            if (is_string($text)) {
                $parts[] = $text;
            }
        }
        return implode('', $parts);
    }

    foreach (['text', 'reasoning_content'] as $key) {
        $value = $delta[$key] ?? '';
        if (is_string($value)) {
            return $value;
        }
    }
    $choiceText = $choice0['text'] ?? '';
    if (is_string($choiceText)) {
        return $choiceText;
    }
    return '';
}

$apiKey = getenv('AI_STREAM_API_KEY');
if ($apiKey === false || $apiKey === '') {
    fwrite(STDERR, "Missing AI_STREAM_API_KEY\n");
    exit(1);
}

$apiUrl = getenv('AI_STREAM_URL') ?: 'http://new-api.bangong.knowbox.cn/v1/chat/completions';
$model = getenv('AI_STREAM_MODEL') ?: 'claude-opus-4-20250514';
$prompt = getenv('AI_STREAM_PROMPT') ?: '只返回json，格式: {"ok":true,"msg":"..."}';
$printSnapshots = getenv('AI_STREAM_PRINT_SNAPSHOTS');
$printSnapshots = ($printSnapshots === false || $printSnapshots !== '0');
$maxSnapshots = intval(getenv('AI_STREAM_MAX_SNAPSHOTS') ?: '20');

$payload = json_encode([
    'model' => $model,
    'messages' => [['role' => 'user', 'content' => $prompt]],
    'stream' => true,
], JSON_UNESCAPED_UNICODE);

if ($payload === false) {
    fwrite(STDERR, "Failed to encode payload\n");
    exit(1);
}

$cmd = sprintf(
    'curl -sS -N %s -H %s -H %s -d %s',
    escapeshellarg($apiUrl),
    escapeshellarg('Content-Type: application/json'),
    escapeshellarg('Authorization: Bearer ' . $apiKey),
    escapeshellarg($payload)
);

$raw = shell_exec($cmd);
if ($raw === null) {
    fwrite(STDERR, "curl request failed\n");
    exit(1);
}

$lines = preg_split("/\r\n|\n|\r/", $raw);
if (!is_array($lines)) {
    $lines = [];
}

$accumulated = '';
$chunkCount = 0;
$snapshotCount = 0;
$printed = 0;
$eventCount = 0;
$skipped = 0;
$debugSamples = [];

foreach ($lines as $line) {
    $line = trim($line);
    if (strpos($line, 'data:') !== 0) {
        continue;
    }
    $payloadLine = ltrim(substr($line, 5));
    if ($payloadLine === '[DONE]') {
        break;
    }
    $eventCount++;
    $event = json_decode($payloadLine, true);
    if (!is_array($event)) {
        $skipped++;
        if (count($debugSamples) < 3) {
            $debugSamples[] = substr($payloadLine, 0, 200);
        }
        continue;
    }

    $piece = extract_piece($event);
    if ($piece === '') {
        $skipped++;
        if (count($debugSamples) < 3) {
            $debugSamples[] = substr(json_encode($event, JSON_UNESCAPED_UNICODE) ?: '', 0, 200);
        }
        continue;
    }

    $chunkCount++;
    $obj = repair_json_strict_prefix($accumulated, true, $piece);
    $accumulated .= $piece;
    if ($obj !== null) {
        $snapshotCount++;
        if ($printSnapshots && $printed < $maxSnapshots) {
            $printed++;
            echo 'snapshot#' . $snapshotCount . ': ' . json_encode($obj, JSON_UNESCAPED_UNICODE) . "\n";
        }
    }
}

$final = repair_json_strict_prefix($accumulated);
$finalObj = repair_json_strict_prefix($accumulated, true);
if ($finalObj === null) {
    echo "Failed to parse streamed content as JSON.\n";
    echo "Repaired text: {$final}\n";
    echo "events: {$eventCount}, content chunks: {$chunkCount}, skipped events: {$skipped}\n";
    if (trim($raw) !== '') {
        echo "raw response sample:\n";
        echo substr($raw, 0, 500) . "\n";
    }
    if (count($debugSamples) > 0) {
        echo "sample skipped payloads:\n";
        foreach ($debugSamples as $sample) {
            echo $sample . "\n";
        }
    }
    exit(1);
}

echo "content chunks: {$chunkCount}\n";
echo "object snapshots: {$snapshotCount}\n";
if ($printSnapshots && $snapshotCount > $maxSnapshots) {
    echo "snapshot output truncated: printed {$maxSnapshots} of {$snapshotCount}\n";
}
echo "events: {$eventCount}, skipped events: {$skipped}\n";
echo "final repaired json: {$final}\n";
echo "final object type: " . gettype($finalObj) . "\n";
