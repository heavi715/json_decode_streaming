<?php

declare(strict_types=1);

require_once __DIR__ . '/../php/RepairJson.php';

$cases = json_decode(file_get_contents(__DIR__ . '/cases.json'), true);
if (!is_array($cases)) {
    fwrite(STDERR, "Failed to parse test/cases.json\n");
    exit(1);
}
$failures = [];

foreach ($cases as $idx => $case) {
    $repaired = repair_json_strict_prefix($case['input']);
    if ($repaired !== $case['expected']) {
        $failures[] = [$idx, 'output mismatch', $repaired, $case['expected']];
        continue;
    }
    $repairedObject = repair_json_strict_prefix($case['input'], true);
    $expectedObject = ($case['expected'] !== '') ? json_decode($case['expected'], true) : null;
    if ($repairedObject !== $expectedObject) {
        $failures[] = [
            $idx,
            'object output mismatch',
            json_encode($repairedObject),
            json_encode($expectedObject),
        ];
        continue;
    }
    [$repairedBoth, $repairedBothObject] = repair_json_strict_prefix_both($case['input']);
    if ($repairedBoth !== $case['expected']) {
        $failures[] = [$idx, 'both output mismatch', $repairedBoth, $case['expected']];
        continue;
    }
    if ($repairedBothObject !== $expectedObject) {
        $failures[] = [
            $idx,
            'both object output mismatch',
            json_encode($repairedBothObject),
            json_encode($expectedObject),
        ];
        continue;
    }
    if ($repaired !== '') {
        json_decode($repaired, true);
        if (json_last_error() !== JSON_ERROR_NONE) {
            $failures[] = [$idx, 'invalid json: ' . json_last_error_msg(), $repaired, $case['expected']];
        }
    }
}

$base = '{"a":"1"';
$append = ',"b":2}';
$expectedAppend = '{"a":"1","b":2}';
apply_repair_json_append_cache_preset('low_memory', true);
$appended = repair_json_strict_prefix($base, false, $append);
if ($appended !== $expectedAppend) {
    $failures[] = ['append', 'append output mismatch', $appended, $expectedAppend];
}
$appendedObject = repair_json_strict_prefix($base, true, $append);
$expectedAppendedObject = json_decode($expectedAppend, true);
if ($appendedObject !== $expectedAppendedObject) {
    $failures[] = [
        'append',
        'append object mismatch',
        json_encode($appendedObject),
        json_encode($expectedAppendedObject),
    ];
}
$unicodeAccumulated = '';
$unicodeChunk1 = '{"a":"\u12';
$unicodeChunk2 = '34"}';
$unicodeStep1 = repair_json_strict_prefix($unicodeAccumulated, false, $unicodeChunk1);
$unicodeAccumulated .= $unicodeChunk1;
$unicodeStep2 = repair_json_strict_prefix($unicodeAccumulated, false, $unicodeChunk2);
$expectedUnicode = '{"a":"\u1234"}';
if ($unicodeStep1 !== '{}') {
    $failures[] = ['append-unicode-step1', 'append unicode intermediate mismatch', $unicodeStep1, '{}'];
}
if ($unicodeStep2 !== $expectedUnicode) {
    $failures[] = ['append-unicode-step2', 'append unicode final mismatch', $unicodeStep2, $expectedUnicode];
}

if (count($failures) > 0) {
    foreach ($failures as [$idx, $reason, $actual, $expected]) {
        echo "[FAIL] case #{$idx}: {$reason}\n";
        echo "  actual  : {$actual}\n";
        echo "  expected: {$expected}\n";
    }
    exit(1);
}

echo 'All ' . count($cases) . " PHP cases passed.\n";
