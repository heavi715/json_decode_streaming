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
    if ($repaired !== '') {
        json_decode($repaired, true);
        if (json_last_error() !== JSON_ERROR_NONE) {
            $failures[] = [$idx, 'invalid json: ' . json_last_error_msg(), $repaired, $case['expected']];
        }
    }
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
