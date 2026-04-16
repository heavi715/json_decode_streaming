<?php

declare(strict_types=1);

function is_hex4_at(string $text, int $start): bool
{
    if ($start + 4 > strlen($text)) {
        return false;
    }
    for ($k = 0; $k < 4; $k++) {
        $c = $text[$start + $k];
        $isDigit = ($c >= '0' && $c <= '9');
        $isLowerHex = ($c >= 'a' && $c <= 'f');
        $isUpperHex = ($c >= 'A' && $c <= 'F');
        if (!$isDigit && !$isLowerHex && !$isUpperHex) {
            return false;
        }
    }
    return true;
}

function scan_number_end(string $text, int $start): int
{
    $n = strlen($text);
    $i = $start;

    if ($i < $n && $text[$i] === '-') {
        $i++;
        if ($i >= $n) {
            return -1;
        }
    }

    if ($i >= $n) {
        return -1;
    }

    if ($text[$i] === '0') {
        $i++;
    } elseif ($text[$i] >= '1' && $text[$i] <= '9') {
        $i++;
        while ($i < $n && $text[$i] >= '0' && $text[$i] <= '9') {
            $i++;
        }
    } else {
        return -1;
    }

    if ($i < $n && $text[$i] === '.') {
        if ($i + 1 >= $n || !($text[$i + 1] >= '0' && $text[$i + 1] <= '9')) {
            return $i - 1;
        }
        $i += 2;
        while ($i < $n && $text[$i] >= '0' && $text[$i] <= '9') {
            $i++;
        }
    }

    if ($i < $n && ($text[$i] === 'e' || $text[$i] === 'E')) {
        if ($i + 1 >= $n) {
            return $i - 1;
        }
        $j = $i + 1;
        if ($text[$j] === '+' || $text[$j] === '-') {
            $j++;
        }
        if ($j >= $n || !($text[$j] >= '0' && $text[$j] <= '9')) {
            return $i - 1;
        }
        $i = $j + 1;
        while ($i < $n && $text[$i] >= '0' && $text[$i] <= '9') {
            $i++;
        }
    }

    return $i - 1;
}

final class RepairState
{
    public $text = '';
    /** @var list<string> */
    public $stack = [];
    public $state = 'root_value';
    public $inString = false;
    public $escapeNext = false;
    public $stringRole = '';
    public $lastSafe = -1;
    public $arrayWaitingValue = false;
    public $objectWaitingKey = false;
    public $i = 0;
    public $brokeEarly = false;

    public function cloneState(): self
    {
        $next = new self();
        $next->text = $this->text;
        $next->stack = $this->stack;
        $next->state = $this->state;
        $next->inString = $this->inString;
        $next->escapeNext = $this->escapeNext;
        $next->stringRole = $this->stringRole;
        $next->lastSafe = $this->lastSafe;
        $next->arrayWaitingValue = $this->arrayWaitingValue;
        $next->objectWaitingKey = $this->objectWaitingKey;
        $next->i = $this->i;
        $next->brokeEarly = $this->brokeEarly;
        return $next;
    }

    private function completeValue(int $idx): void
    {
        $this->arrayWaitingValue = false;
        $this->objectWaitingKey = false;
        if (count($this->stack) === 0) {
            $this->state = 'done';
            $this->lastSafe = $idx;
            return;
        }
        $top = $this->stack[count($this->stack) - 1];
        if ($top === 'object') {
            $this->state = 'object_comma_or_end';
            $this->lastSafe = $idx;
        } else {
            $this->state = 'array_comma_or_end';
            $this->lastSafe = $idx;
        }
    }

    public function feed(string $chunk): void
    {
        if ($chunk === '') {
            return;
        }
        if ($this->brokeEarly) {
            $this->text .= $chunk;
            return;
        }
        $this->text .= $chunk;
        $n = strlen($this->text);
        while ($this->i < $n) {
            $ch = $this->text[$this->i];
            if ($this->inString) {
                if ($this->escapeNext) {
                    if (strpos('"\\/bfnrt', $ch) !== false) {
                        $this->escapeNext = false;
                        $this->i++;
                        continue;
                    }
                    if ($ch === 'u') {
                        if ($this->i + 4 >= $n) {
                            break;
                        }
                        if (!is_hex4_at($this->text, $this->i + 1)) {
                            $this->brokeEarly = true;
                            break;
                        }
                        $this->escapeNext = false;
                        $this->i += 5;
                        continue;
                    }
                    $this->brokeEarly = true;
                    break;
                }
                if ($ch === '\\') {
                    $this->escapeNext = true;
                    $this->i++;
                    continue;
                }
                if ($ch === '"') {
                    $this->inString = false;
                    if ($this->stringRole === 'key') {
                        $this->state = 'object_colon';
                    } else {
                        $this->completeValue($this->i);
                    }
                }
                $this->i++;
                continue;
            }
            if ($ch === ' ' || $ch === "\t" || $ch === "\r" || $ch === "\n") {
                $this->i++;
                continue;
            }
            if ($this->state === 'done') {
                $this->brokeEarly = true;
                break;
            }
            if ($this->state === 'root_value' || $this->state === 'object_value' || $this->state === 'array_value_or_end') {
                if ($ch === '{') {
                    $this->stack[] = 'object';
                    $this->state = 'object_key_or_end';
                    $this->lastSafe = $this->i;
                    $this->i++;
                    continue;
                }
                if ($ch === '[') {
                    $this->stack[] = 'array';
                    $this->state = 'array_value_or_end';
                    $this->lastSafe = $this->i;
                    $this->i++;
                    continue;
                }
                if ($ch === '"') {
                    $this->inString = true;
                    $this->stringRole = 'value';
                    $this->i++;
                    continue;
                }
                if ($ch === '-' || ($ch >= '0' && $ch <= '9')) {
                    $end = scan_number_end($this->text, $this->i);
                    if ($end < $this->i) {
                        $this->brokeEarly = true;
                        break;
                    }
                    $this->i = $end + 1;
                    $this->completeValue($end);
                    continue;
                }
                if ($ch === 't' && $this->i + 4 <= $n && $this->text[$this->i + 1] === 'r' && $this->text[$this->i + 2] === 'u' && $this->text[$this->i + 3] === 'e') {
                    $this->i += 4;
                    $this->completeValue($this->i - 1);
                    continue;
                }
                if ($ch === 'f' && $this->i + 5 <= $n && $this->text[$this->i + 1] === 'a' && $this->text[$this->i + 2] === 'l' && $this->text[$this->i + 3] === 's' && $this->text[$this->i + 4] === 'e') {
                    $this->i += 5;
                    $this->completeValue($this->i - 1);
                    continue;
                }
                if ($ch === 'n' && $this->i + 4 <= $n && $this->text[$this->i + 1] === 'u' && $this->text[$this->i + 2] === 'l' && $this->text[$this->i + 3] === 'l') {
                    $this->i += 4;
                    $this->completeValue($this->i - 1);
                    continue;
                }
                if ($this->state === 'array_value_or_end' && $ch === ']') {
                    if ($this->arrayWaitingValue) {
                        $this->brokeEarly = true;
                        break;
                    }
                    array_pop($this->stack);
                    $this->completeValue($this->i);
                    $this->i++;
                    continue;
                }
                $this->brokeEarly = true;
                break;
            }
            if ($this->state === 'object_key_or_end') {
                if ($ch === '}') {
                    if ($this->objectWaitingKey) {
                        $this->brokeEarly = true;
                        break;
                    }
                    array_pop($this->stack);
                    $this->completeValue($this->i);
                    $this->i++;
                    continue;
                }
                if ($ch === '"') {
                    $this->objectWaitingKey = false;
                    $this->inString = true;
                    $this->stringRole = 'key';
                    $this->i++;
                    continue;
                }
                $this->brokeEarly = true;
                break;
            }
            if ($this->state === 'object_colon') {
                if ($ch === ':') {
                    $this->state = 'object_value';
                    $this->i++;
                    continue;
                }
                $this->brokeEarly = true;
                break;
            }
            if ($this->state === 'object_comma_or_end') {
                if ($ch === ',') {
                    $this->state = 'object_key_or_end';
                    $this->objectWaitingKey = true;
                    $this->i++;
                    continue;
                }
                if ($ch === '}') {
                    array_pop($this->stack);
                    $this->completeValue($this->i);
                    $this->i++;
                    continue;
                }
                $this->brokeEarly = true;
                break;
            }
            if ($this->state === 'array_comma_or_end') {
                if ($ch === ',') {
                    $this->state = 'array_value_or_end';
                    $this->arrayWaitingValue = true;
                    $this->i++;
                    continue;
                }
                if ($ch === ']') {
                    array_pop($this->stack);
                    $this->completeValue($this->i);
                    $this->i++;
                    continue;
                }
                $this->brokeEarly = true;
                break;
            }
            $this->brokeEarly = true;
            break;
        }
    }

    public function snapshot(): string
    {
        if ($this->inString && !$this->brokeEarly && !$this->escapeNext && $this->stringRole === 'value') {
            $base = $this->text . '"';
        } else {
            $base = $this->lastSafe >= 0 ? substr($this->text, 0, $this->lastSafe + 1) : '';
        }
        $closers = '';
        for ($idx = count($this->stack) - 1; $idx >= 0; $idx--) {
            $closers .= ($this->stack[$idx] === 'object') ? '}' : ']';
        }
        return $base . $closers;
    }
}

const APPEND_CACHE_MAX_ENTRIES_DEFAULT = 256;
const APPEND_CACHE_MAX_TOTAL_BYTES_DEFAULT = 4194304;
const APPEND_CACHE_TTL_SECONDS_DEFAULT = 120.0;
/** @var array<string,array{maxEntries:int,maxTotalBytes:int,ttlSeconds:float}> */
const APPEND_CACHE_PRESETS = [
    'default' => ['maxEntries' => 256, 'maxTotalBytes' => 4194304, 'ttlSeconds' => 120.0],
    'low_memory' => ['maxEntries' => 64, 'maxTotalBytes' => 524288, 'ttlSeconds' => 15.0],
    'high_throughput' => ['maxEntries' => 1024, 'maxTotalBytes' => 16777216, 'ttlSeconds' => 600.0],
];
/** @var array<string,array{state:RepairState,keyBytes:int,expiresAt:float}> */
$GLOBALS['REPAIR_JSON_APPEND_CACHE'] = [];
/** @var list<string> */
$GLOBALS['REPAIR_JSON_APPEND_CACHE_ORDER'] = [];
$GLOBALS['REPAIR_JSON_APPEND_CACHE_TOTAL_BYTES'] = 0;
$GLOBALS['REPAIR_JSON_APPEND_CACHE_MAX_ENTRIES'] = APPEND_CACHE_MAX_ENTRIES_DEFAULT;
$GLOBALS['REPAIR_JSON_APPEND_CACHE_MAX_TOTAL_BYTES'] = APPEND_CACHE_MAX_TOTAL_BYTES_DEFAULT;
$GLOBALS['REPAIR_JSON_APPEND_CACHE_TTL_SECONDS'] = APPEND_CACHE_TTL_SECONDS_DEFAULT;

function set_repair_json_append_cache_config(?int $maxEntries = null, ?int $maxTotalBytes = null, ?float $ttlSeconds = null, bool $clear = false): void
{
    if ($maxEntries !== null) {
        if ($maxEntries < 1) {
            throw new InvalidArgumentException('maxEntries must be >= 1');
        }
        $GLOBALS['REPAIR_JSON_APPEND_CACHE_MAX_ENTRIES'] = $maxEntries;
    }
    if ($maxTotalBytes !== null) {
        if ($maxTotalBytes < 1024) {
            throw new InvalidArgumentException('maxTotalBytes must be >= 1024');
        }
        $GLOBALS['REPAIR_JSON_APPEND_CACHE_MAX_TOTAL_BYTES'] = $maxTotalBytes;
    }
    if ($ttlSeconds !== null) {
        if ($ttlSeconds < 0.1) {
            throw new InvalidArgumentException('ttlSeconds must be >= 0.1');
        }
        $GLOBALS['REPAIR_JSON_APPEND_CACHE_TTL_SECONDS'] = $ttlSeconds;
    }
    if ($clear) {
        $GLOBALS['REPAIR_JSON_APPEND_CACHE'] = [];
        $GLOBALS['REPAIR_JSON_APPEND_CACHE_ORDER'] = [];
        $GLOBALS['REPAIR_JSON_APPEND_CACHE_TOTAL_BYTES'] = 0;
        return;
    }
    prune_repair_json_append_cache(microtime(true));
}

function apply_repair_json_append_cache_preset(string $preset, bool $clear = true): void
{
    if (!isset(APPEND_CACHE_PRESETS[$preset])) {
        throw new InvalidArgumentException('unknown cache preset: ' . $preset);
    }
    $picked = APPEND_CACHE_PRESETS[$preset];
    set_repair_json_append_cache_config($picked['maxEntries'], $picked['maxTotalBytes'], $picked['ttlSeconds'], $clear);
}

function repair_json_cache_key_bytes(string $text): int
{
    return strlen($text);
}

function repair_json_cache_touch(string $key): void
{
    $order = $GLOBALS['REPAIR_JSON_APPEND_CACHE_ORDER'];
    $idx = array_search($key, $order, true);
    if ($idx !== false) {
        unset($order[$idx]);
        $order = array_values($order);
    }
    $order[] = $key;
    $GLOBALS['REPAIR_JSON_APPEND_CACHE_ORDER'] = $order;
}

function prune_repair_json_append_cache(float $now): void
{
    while (count($GLOBALS['REPAIR_JSON_APPEND_CACHE_ORDER']) > 0) {
        $oldest = $GLOBALS['REPAIR_JSON_APPEND_CACHE_ORDER'][0];
        if (!isset($GLOBALS['REPAIR_JSON_APPEND_CACHE'][$oldest])) {
            array_shift($GLOBALS['REPAIR_JSON_APPEND_CACHE_ORDER']);
            continue;
        }
        $entry = $GLOBALS['REPAIR_JSON_APPEND_CACHE'][$oldest];
        $overLimit = count($GLOBALS['REPAIR_JSON_APPEND_CACHE']) > $GLOBALS['REPAIR_JSON_APPEND_CACHE_MAX_ENTRIES']
            || $GLOBALS['REPAIR_JSON_APPEND_CACHE_TOTAL_BYTES'] > $GLOBALS['REPAIR_JSON_APPEND_CACHE_MAX_TOTAL_BYTES'];
        $expired = $entry['expiresAt'] <= $now;
        if (!$overLimit && !$expired) {
            break;
        }
        unset($GLOBALS['REPAIR_JSON_APPEND_CACHE'][$oldest]);
        array_shift($GLOBALS['REPAIR_JSON_APPEND_CACHE_ORDER']);
        $GLOBALS['REPAIR_JSON_APPEND_CACHE_TOTAL_BYTES'] -= $entry['keyBytes'];
    }
}

function get_repair_json_cached_state(string $text): ?RepairState
{
    if (!isset($GLOBALS['REPAIR_JSON_APPEND_CACHE'][$text])) {
        return null;
    }
    $entry = $GLOBALS['REPAIR_JSON_APPEND_CACHE'][$text];
    $now = microtime(true);
    if ($entry['expiresAt'] <= $now) {
        unset($GLOBALS['REPAIR_JSON_APPEND_CACHE'][$text]);
        $GLOBALS['REPAIR_JSON_APPEND_CACHE_TOTAL_BYTES'] -= $entry['keyBytes'];
        $order = $GLOBALS['REPAIR_JSON_APPEND_CACHE_ORDER'];
        $idx = array_search($text, $order, true);
        if ($idx !== false) {
            unset($order[$idx]);
            $GLOBALS['REPAIR_JSON_APPEND_CACHE_ORDER'] = array_values($order);
        }
        return null;
    }
    $entry['expiresAt'] = $now + $GLOBALS['REPAIR_JSON_APPEND_CACHE_TTL_SECONDS'];
    $GLOBALS['REPAIR_JSON_APPEND_CACHE'][$text] = $entry;
    repair_json_cache_touch($text);
    return $entry['state']->cloneState();
}

function cache_repair_json_state(string $text, RepairState $state): void
{
    $keyBytes = repair_json_cache_key_bytes($text);
    if (isset($GLOBALS['REPAIR_JSON_APPEND_CACHE'][$text])) {
        $existing = $GLOBALS['REPAIR_JSON_APPEND_CACHE'][$text];
        $GLOBALS['REPAIR_JSON_APPEND_CACHE_TOTAL_BYTES'] -= $existing['keyBytes'];
    }
    $GLOBALS['REPAIR_JSON_APPEND_CACHE'][$text] = [
        'state' => $state->cloneState(),
        'keyBytes' => $keyBytes,
        'expiresAt' => microtime(true) + $GLOBALS['REPAIR_JSON_APPEND_CACHE_TTL_SECONDS'],
    ];
    $GLOBALS['REPAIR_JSON_APPEND_CACHE_TOTAL_BYTES'] += $keyBytes;
    repair_json_cache_touch($text);
    prune_repair_json_append_cache(microtime(true));
}

function repair_json_strict_prefix(string $text, bool $returnObject = false, string $appendContent = '')
{
    $fullText = $appendContent !== '' ? $text . $appendContent : $text;
    if ($returnObject) {
        $parsedOriginal = json_decode($fullText, true);
        if (json_last_error() === JSON_ERROR_NONE) {
            return $parsedOriginal;
        }
    }
    if ($appendContent !== '') {
        $state = get_repair_json_cached_state($text);
        if ($state !== null) {
            $state->feed($appendContent);
            $repaired = $state->snapshot();
            cache_repair_json_state($fullText, $state);
        } else {
            $state = new RepairState();
            $state->feed($fullText);
            $repaired = $state->snapshot();
            cache_repair_json_state($fullText, $state);
        }
    } else {
        $state = new RepairState();
        $state->feed($fullText);
        $repaired = $state->snapshot();
        cache_repair_json_state($fullText, $state);
    }
    if (!$returnObject) {
        return $repaired;
    }
    if ($repaired === '') {
        return null;
    }
    return json_decode($repaired, true);
}

function repair_json_strict_prefix_both(string $text, string $appendContent = ''): array
{
    $repaired = repair_json_strict_prefix($text, false, $appendContent);
    if ($repaired === '') {
        return [$repaired, null];
    }
    return [$repaired, json_decode($repaired, true)];
}
