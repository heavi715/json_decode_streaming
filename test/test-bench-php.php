<?php

declare(strict_types=1);

require_once __DIR__ . '/../php/RepairJson.php';

function parse_mode(array $argv): string
{
    $mode = 'both';
    foreach ($argv as $arg) {
        if (strpos($arg, '--mode=') === 0) {
            $mode = substr($arg, 7);
            break;
        }
    }
    if (!in_array($mode, ['string', 'object', 'both_return', 'all'], true)) {
        fwrite(STDERR, "Invalid --mode. Use string|object|both_return|all\n");
        exit(1);
    }
    return $mode;
}

function build_samples(): array
{
    $small = str_repeat('{"a":1,"b":[1,2,3],"c":"hello"}', 4);

    $mediumItems = [];
    for ($i = 0; $i < 200; $i++) {
        $mediumItems[] = [
            'id' => $i,
            'name' => str_repeat('x', 20),
            'arr' => range(0, 19),
        ];
    }
    $medium = json_encode(['items' => $mediumItems], JSON_UNESCAPED_UNICODE);

    $largeItems = [];
    for ($i = 0; $i < 2000; $i++) {
        $largeItems[] = [
            'id' => $i,
            'name' => str_repeat('x', 40),
            'arr' => range(0, 39),
            'obj' => ['k' => str_repeat('v', 10), 'n' => $i],
        ];
    }
    $large = json_encode(['items' => $largeItems], JSON_UNESCAPED_UNICODE);

    return [
        'small' => $small,
        'medium' => $medium,
        'large' => $large,
    ];
}

function iterations_for(string $name): int
{
    if ($name === 'small') {
        return 2000;
    }
    if ($name === 'medium') {
        return 400;
    }
    return 80;
}

function run_bench(string $name, string $truncated, int $n, string $mode): void
{
    $t0 = microtime(true);
    for ($i = 0; $i < $n; $i++) {
        if ($mode === 'string') {
            repair_json_strict_prefix($truncated, false);
        } elseif ($mode === 'object') {
            repair_json_strict_prefix($truncated, true);
        } else {
            repair_json_strict_prefix_both($truncated);
        }
    }
    $dt = microtime(true) - $t0;
    $avgUs = ($dt / $n) * 1000000.0;
    $throughputMiB = (strlen($truncated) * $n) / $dt / 1024 / 1024;
    echo sprintf(
        "%s/%s: len=%d n=%d avg_us=%.1f throughput_mib_s=%.2f\n",
        $name,
        $mode,
        strlen($truncated),
        $n,
        $avgUs,
        $throughputMiB
    );
}

$mode = parse_mode($argv);

foreach (build_samples() as $name => $text) {
    $truncated = substr($text, 0, -17);
    $n = iterations_for($name);
    if ($mode === 'string' || $mode === 'all') {
        run_bench($name, $truncated, $n, 'string');
    }
    if ($mode === 'object' || $mode === 'all') {
        run_bench($name, $truncated, $n, 'object');
    }
    if ($mode === 'both_return' || $mode === 'all') {
        run_bench($name, $truncated, $n, 'both_return');
    }
}
