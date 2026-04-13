<?php

declare(strict_types=1);

function repair_json_strict_prefix(string $text): string
{
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
                    $hex = substr($text, $i + 1, 4);
                    if (!preg_match('/^[0-9a-fA-F]{4}$/', $hex)) {
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

        if (preg_match('/\s/', $ch)) {
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
            if (strpos('-0123456789', $ch) !== false) {
                if (!preg_match('/^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?/', substr($text, $i), $m)) {
                    $brokeEarly = true;
                    break;
                }
                $end = $i + strlen($m[0]) - 1;
                $i = $end + 1;
                $completeValue($end);
                continue;
            }
            if (substr($text, $i, 4) === 'true') {
                $i += 4;
                $completeValue($i - 1);
                continue;
            }
            if (substr($text, $i, 5) === 'false') {
                $i += 5;
                $completeValue($i - 1);
                continue;
            }
            if (substr($text, $i, 4) === 'null') {
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

    return $base . $closers;
}
