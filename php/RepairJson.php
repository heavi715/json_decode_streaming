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

function repair_json_strict_prefix(string $text, bool $returnObject = false, string $appendContent = '')
{
    if ($appendContent !== '') {
        $text .= $appendContent;
    }
    if ($returnObject) {
        $parsedOriginal = json_decode($text, true);
        if (json_last_error() === JSON_ERROR_NONE) {
            return $parsedOriginal;
        }
    }

    $stack = [];
    $state = 'root_value';
    $inString = false;
    $escapeNext = false;
    $stringRole = '';
    $lastSafe = -1;
    $arrayWaitingValue = false;
    $objectWaitingKey = false;
    $i = 0;
    $n = strlen($text);
    $brokeEarly = false;

    $completeValue = function (int $idx) use (&$stack, &$state, &$lastSafe, &$arrayWaitingValue, &$objectWaitingKey): void {
        $arrayWaitingValue = false;
        $objectWaitingKey = false;
        if (count($stack) === 0) {
            $state = 'done';
            $lastSafe = $idx;
            return;
        }

        $top = $stack[count($stack) - 1];
        if ($top === 'object') {
            $state = 'object_comma_or_end';
            $lastSafe = $idx;
        } else {
            $state = 'array_comma_or_end';
            $lastSafe = $idx;
        }
    };

    while ($i < $n) {
        $ch = $text[$i];

        if ($inString) {
            if ($escapeNext) {
                if (strpos('"\\/bfnrt', $ch) !== false) {
                    $escapeNext = false;
                    $i++;
                    continue;
                }
                if ($ch === 'u') {
                    if ($i + 4 >= $n) {
                        $brokeEarly = true;
                        break;
                    }
                    if (!is_hex4_at($text, $i + 1)) {
                        $brokeEarly = true;
                        break;
                    }
                    $escapeNext = false;
                    $i += 5;
                    continue;
                }
                $brokeEarly = true;
                break;
            }
            if ($ch === '\\') {
                $escapeNext = true;
                $i++;
                continue;
            }
            if ($ch === '"') {
                $inString = false;
                if ($stringRole === 'key') {
                    $state = 'object_colon';
                } else {
                    $completeValue($i);
                }
            }
            $i++;
            continue;
        }

        if ($ch === ' ' || $ch === "\t" || $ch === "\r" || $ch === "\n") {
            $i++;
            continue;
        }

        if ($state === 'done') {
            $brokeEarly = true;
            break;
        }

        if ($state === 'root_value' || $state === 'object_value' || $state === 'array_value_or_end') {
            if ($ch === '{') {
                $stack[] = 'object';
                $state = 'object_key_or_end';
                $lastSafe = $i;
                $i++;
                continue;
            }
            if ($ch === '[') {
                $stack[] = 'array';
                $state = 'array_value_or_end';
                $lastSafe = $i;
                $i++;
                continue;
            }
            if ($ch === '"') {
                $inString = true;
                $stringRole = 'value';
                $i++;
                continue;
            }
            if ($ch === '-' || ($ch >= '0' && $ch <= '9')) {
                $end = scan_number_end($text, $i);
                if ($end < $i) {
                    $brokeEarly = true;
                    break;
                }
                $i = $end + 1;
                $completeValue($end);
                continue;
            }
            if ($ch === 't' && $i + 4 <= $n && $text[$i + 1] === 'r' && $text[$i + 2] === 'u' && $text[$i + 3] === 'e') {
                $i += 4;
                $completeValue($i - 1);
                continue;
            }
            if ($ch === 'f' && $i + 5 <= $n && $text[$i + 1] === 'a' && $text[$i + 2] === 'l' && $text[$i + 3] === 's' && $text[$i + 4] === 'e') {
                $i += 5;
                $completeValue($i - 1);
                continue;
            }
            if ($ch === 'n' && $i + 4 <= $n && $text[$i + 1] === 'u' && $text[$i + 2] === 'l' && $text[$i + 3] === 'l') {
                $i += 4;
                $completeValue($i - 1);
                continue;
            }
            if ($state === 'array_value_or_end' && $ch === ']') {
                if ($arrayWaitingValue) {
                    $brokeEarly = true;
                    break;
                }
                array_pop($stack);
                $completeValue($i);
                $i++;
                continue;
            }
            $brokeEarly = true;
            break;
        }

        if ($state === 'object_key_or_end') {
            if ($ch === '}') {
                if ($objectWaitingKey) {
                    $brokeEarly = true;
                    break;
                }
                array_pop($stack);
                $completeValue($i);
                $i++;
                continue;
            }
            if ($ch === '"') {
                $objectWaitingKey = false;
                $inString = true;
                $stringRole = 'key';
                $i++;
                continue;
            }
            $brokeEarly = true;
            break;
        }

        if ($state === 'object_colon') {
            if ($ch === ':') {
                $state = 'object_value';
                $i++;
                continue;
            }
            $brokeEarly = true;
            break;
        }

        if ($state === 'object_comma_or_end') {
            if ($ch === ',') {
                $state = 'object_key_or_end';
                $objectWaitingKey = true;
                $i++;
                continue;
            }
            if ($ch === '}') {
                array_pop($stack);
                $completeValue($i);
                $i++;
                continue;
            }
            $brokeEarly = true;
            break;
        }

        if ($state === 'array_comma_or_end') {
            if ($ch === ',') {
                $state = 'array_value_or_end';
                $arrayWaitingValue = true;
                $i++;
                continue;
            }
            if ($ch === ']') {
                array_pop($stack);
                $completeValue($i);
                $i++;
                continue;
            }
            $brokeEarly = true;
            break;
        }

        $brokeEarly = true;
        break;
    }

    if ($inString && !$brokeEarly && !$escapeNext && $stringRole === 'value') {
        $base = $text . '"';
        $completeValue($n);
    } else {
        $base = $lastSafe >= 0 ? substr($text, 0, $lastSafe + 1) : '';
    }

    $closers = '';
    for ($idx = count($stack) - 1; $idx >= 0; $idx--) {
        $closers .= ($stack[$idx] === 'object') ? '}' : ']';
    }

    $repaired = $base . $closers;
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
