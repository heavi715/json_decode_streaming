# Shared Test Cases

All language implementations must produce exactly the same repaired output.

## Canonical examples

| Input | Output |
| --- | --- |
| `{"a":"b` | `{"a":"b"}` |
| `{"a":[1,2,` | `{"a":[1,2]}` |
| `{"a":"1","b":` | `{"a":"1"}` |
| `{"a":"1","b":"` | `{"a":"1","b":""}` |

## Additional cases

| Input | Output | Notes |
| --- | --- | --- |
| `{` | `{}` | close object |
| `[` | `[]` | close array |
| `{"a":1` | `{"a":1}` | close object |
| `{"a":[{"b":2}` | `{"a":[{"b":2}]}` | nested close |
| `{"a":"x","b"` | `{"a":"x"}` | drop unfinished key |
| `{"a":"x",` | `{"a":"x"}` | drop trailing comma segment |
| `{"a":true,"b":f` | `{"a":true}` | incomplete literal |
| `{"a":"he\\` | `{"a":"he\\"}` | escaped slash before EOF |
| `"hel` | `"hel"` | root string close |
| `12` | `12` | complete scalar no change |

## Complex multi-level cases

| Input | Output | Notes |
| --- | --- | --- |
| `{"meta":{"id":"u1","tags":["a","b"],"ext":{"score":9.5,"flags":[true,false,null]}` | `{"meta":{"id":"u1","tags":["a","b"],"ext":{"score":9.5,"flags":[true,false,null]}}}` | deep object closure (3 levels) |
| `{"a":[{"b":[{"c":"x"},{"d":[1,2,3]}` | `{"a":[{"b":[{"c":"x"},{"d":[1,2,3]}]}]}` | nested array/object mixed closure |
| `{"a":{"b":{"c":{"d":{"e":"v"}}` | `{"a":{"b":{"c":{"d":{"e":"v"}}}}}` | multi-level object tail repair |
| `{"events":[{"type":"start","ts":1},{"type":"data","payload":{"k1":"v1","k2":[1,2,{"k3":"v3` | `{"events":[{"type":"start","ts":1},{"type":"data","payload":{"k1":"v1","k2":[1,2,{"k3":"v3"}]}}]}` | string close + cascading bracket close |
| `{"cfg":{"path":"C:\\\\Users\\\\dev\\\\","enabled":true,"steps":[{"name":"init","ok":t` | `{"cfg":{"path":"C:\\\\Users\\\\dev\\\\","enabled":true,"steps":[{"name":"init"}]}}` | incomplete literal inside nested object |
| `{"m":[1,2,{"n":[3,4,{"o":"p"}],` | `{"m":[1,2,{"n":[3,4,{"o":"p"}]}]}` | drop trailing comma branch at deep level |
| `{"root":[{"child":{"arr":[{"leaf":"x\\` | `{"root":[{"child":{"arr":[{"leaf":"x\\"}]}}]}` | escaped backslash before EOF in deep string |
| `{"a":{"b":[{"c":1},{"d":2}],"e":{"f":[{"g":"h"}]}}, "x":"y` | `{"a":{"b":[{"c":1},{"d":2}],"e":{"f":[{"g":"h"}]}},"x":"y"}` | mixed complete/incomplete sibling segments |
| `[{"k":"v1"},{"k":"v2","inner":[{"x":1},{"y":2}` | `[{"k":"v1"},{"k":"v2","inner":[{"x":1},{"y":2}]}]` | root array with nested object completion |
| `{"a":[{"b":"c"}],"trail":{"x":1},"bad":` | `{"a":[{"b":"c"}],"trail":{"x":1}}` | remove unsafe trailing key-value |

## Boundary and prefix policy cases

| Input | Output | Notes |
| --- | --- | --- |
| `fals` | `` | incomplete root literal has no safe prefix |
| `nu` | `` | incomplete root null has no safe prefix |
| `{"a":1}xyz` | `{"a":1}` | stop at first complete root value |
| `[1,2]tail` | `[1,2]` | root array complete, drop trailing garbage |
| `12abc` | `12` | root number complete, drop trailing garbage |
| `{"a" 1` | `{}` | missing colon after key |
| `{"a\\` | `{}` | incomplete object key string cannot be salvaged |

## Number token edge cases

| Input | Output | Notes |
| --- | --- | --- |
| `-` | `` | incomplete root minus has no safe prefix |
| `1.` | `1` | incomplete fraction part rolls back to integer prefix |
| `1e` | `1` | incomplete exponent rolls back to integer prefix |
| `1e-` | `1` | incomplete exponent sign rolls back to integer prefix |
| `{"n":-` | `{}` | incomplete number after object colon drops member |
| `{"n":1.` | `{"n":1}` | object value keeps safe numeric prefix |
| `[1e` | `[1]` | array element keeps safe numeric prefix |
| `[1e,2]` | `[1]` | invalid token after partial exponent truncates tail |

## Unicode and whitespace cases

| Input | Output | Notes |
| --- | --- | --- |
| `   {"a":1` | `   {"a":1}` | leading whitespace preserved, then close object |
| `{"u":"\\u4f60` | `{"u":"\\u4f60"}` | close value string with complete unicode escape |
| `{"u":"\\u4f60\\u597d` | `{"u":"\\u4f60\\u597d"}` | multiple complete unicode escapes in truncated string |
| `{"emoji":"\\ud83d\\ude00` | `{"emoji":"\\ud83d\\ude00"}` | surrogate pair escapes inside truncated value string |
| `[true,false,null` | `[true,false,null]` | close root array after complete literals |
| `{"a":1}   x` | `{"a":1}` | drop trailing garbage after complete root with whitespace |
| `   ` | `` | whitespace-only input has no safe JSON value prefix |

## Escape validation cases

| Input | Output | Notes |
| --- | --- | --- |
| `{"s":"\q` | `{}` | invalid escape in value string drops unfinished member |
| `"\q` | `` | invalid escape in root string has no safe prefix |
| `{"s":"\u12` | `{}` | incomplete unicode escape in value string |
| `{"s":"\u12xz` | `{}` | invalid unicode hex digits in value string |

## Object-key escape cases

| Input | Output | Notes |
| --- | --- | --- |
| `{"\q":1` | `{}` | invalid escape in object key string |
| `{"\u12":1` | `{}` | incomplete unicode escape in object key |
| `{"\u12xz":1` | `{}` | invalid unicode hex digits in object key |
| `{"ok\\n":1` | `{"ok\\n":1}` | valid escaped key should be preserved |

## Cross-state combination cases

| Input | Output | Notes |
| --- | --- | --- |
| `{"k"` | `{}` | complete key string but missing colon/value |
| `{"k":` | `{}` | colon seen but value missing |
| `{"k":"v","x":"\q` | `{"k":"v"}` | invalid escape in later sibling value drops trailing fragment |
| `{"k":"v","x":` | `{"k":"v"}` | second member missing value keeps previous safe member |
| `["ok","\q` | `["ok"]` | invalid escape in second array element drops tail |
| `[{"a":1},{"b":"\u12` | `[{"a":1},{}]` | second object already opened, so it is safely auto-closed |

## Array comma-state cases

| Input | Output | Notes |
| --- | --- | --- |
| `[1,]` | `[1]` | trailing comma before array end drops pending element |
| `[1,,2` | `[1]` | second comma is invalid while waiting for next value |
| `[{},,{}]` | `[{}]` | invalid empty element after comma truncates tail |
| `[1,{"a":2},]` | `[1,{"a":2}]` | completed elements kept, trailing comma dropped |
| `[1,\q` | `[1]` | invalid token for next array value truncates tail |
| `{"a":[1,2,]}` | `{"a":[1,2]}` | nested array trailing comma repaired via rollback |

## Object comma-state cases

| Input | Output | Notes |
| --- | --- | --- |
| `{"a":1,}` | `{"a":1}` | trailing comma before object end drops pending member |
| `{"a":1,,"b":2}` | `{"a":1}` | invalid empty member after comma truncates tail |
| `{"a":1,` | `{"a":1}` | object ends while waiting for next key |
| `{"a":1,\q` | `{"a":1}` | invalid token for next key truncates tail |
| `[{"a":1,}]` | `[{"a":1}]` | nested object trailing comma repaired via rollback |
| `{"a":{"b":2,},"c":3` | `{"a":{"b":2}}` | nested trailing comma invalidates following outer sibling fragment |

## Delimiter mismatch cases

| Input | Output | Notes |
| --- | --- | --- |
| `{]` | `{}` | wrong closer in object context triggers rollback and close |
| `[}` | `[]` | wrong closer in array context triggers rollback and close |
| `{"a"]` | `{}` | object expects colon after key, got wrong closer |
| `[1}` | `[1]` | array expects comma/end, got object closer |
| `{"a":[1,2}}` | `{"a":[1,2]}` | mismatched object closer while in array context |
| `[{"a":1]]` | `[{"a":1}]` | extra array closer after complete nested object |

## Escaped control-sequence cases

| Input | Output | Notes |
| --- | --- | --- |
| `{"s":"line1\\nline2` | `{"s":"line1\\nline2"}` | escaped newline sequence preserved in repaired string |
| `{"s":"a\\rb` | `{"s":"a\\rb"}` | escaped carriage-return sequence preserved |
| `{"s":"x\\ty` | `{"s":"x\\ty"}` | escaped tab sequence preserved |
| `["ok","bad\\nstr` | `["ok","bad\\nstr"]` | escaped control sequence in second element + string close |
| `"bad\\nstr` | `"bad\\nstr"` | root string with escaped sequence closes normally |

## Invalid number-format cases

| Input | Output | Notes |
| --- | --- | --- |
| `01` | `0` | leading-zero integer keeps first safe digit only |
| `-01` | `-0` | signed leading-zero integer keeps safe prefix |
| `00` | `0` | second zero is invalid continuation at root |
| `[01,2]` | `[0]` | invalid number continuation in array truncates tail |
| `{"n":01}` | `{"n":0}` | invalid number continuation in object value truncates member tail |
| `{"n":00,"m":1}` | `{"n":0}` | invalid number in first member truncates following sibling |

## Root-done and multi-root cases

| Input | Output | Notes |
| --- | --- | --- |
| `12 34` | `12` | root scalar complete, later root token dropped |
| `{"a":1} {"b":2}` | `{"a":1}` | second root object must be dropped |
| `[] []` | `[]` | second root array dropped after first complete root |
| `"x" "y"` | `"x"` | second root string dropped |
| `null true` | `null` | second root literal dropped |
| `{"a":1}\n\t  ` | `{"a":1}` | trailing whitespace after complete root is truncated |

## Exponent and fraction edge cases

| Input | Output | Notes |
| --- | --- | --- |
| `-0.` | `-0` | incomplete fraction rolls back to integer prefix |
| `-0e` | `-0` | incomplete exponent rolls back to integer prefix |
| `-0e+` | `-0` | incomplete exponent sign rolls back to integer prefix |
| `1.0e` | `1.0` | incomplete exponent after fraction keeps safe prefix |
| `1.0e-` | `1.0` | incomplete signed exponent keeps safe prefix |
| `{"n":-0e` | `{"n":-0}` | object value uses safe numeric prefix |
| `[1.0e,2]` | `[1.0]` | invalid exponent continuation truncates array tail |
| `{"a":[-0e+,3]}` | `{"a":[-0]}` | nested array invalid exponent truncates remaining tail |

## High-noise mixed-tail cases

| Input | Output | Notes |
| --- | --- | --- |
| `{"a":[1,2,]}]garbage` | `{"a":[1,2]}` | trailing comma repaired, extra closer and garbage dropped |
| `[{"a":1,}],,tail` | `[{"a":1}]` | nested object trailing comma repaired, then invalid comma tail dropped |
| `{"a":{"b":[1,2,],},"c":3}xyz` | `{"a":{"b":[1,2]}}` | inner invalid comma forces rollback before outer sibling |
| `[1,2]]]]` | `[1,2]` | repeated extra closers after complete root are dropped |
| `{"k":"v"}},,` | `{"k":"v"}` | complete root object kept, noisy suffix dropped |
| `[{"x":"y\u12xz"}],]` | `[{}]` | invalid unicode escape drops member but opened object is safely auto-closed |

## Noise replay baseline cases

| Input | Output | Notes |
| --- | --- | --- |
| `{"a":"b]]]` | `{"a":"b]]]"}` | noise inside unfinished value string is preserved as content |
| `{"a":[1,2},,` | `{"a":[1,2]}` | object closer and commas in array context are dropped |
| `[{"k":"v"}}tail` | `[{"k":"v"}]` | extra closers/tail after completed first element are dropped |
| `{"n":1.0e]` | `{"n":1.0}` | invalid exponent continuation rolls back to safe numeric prefix |
| `{"x":truefalse` | `{"x":true}` | second literal in object context is unsafe tail |
| `["ok",,` | `["ok"]` | duplicate comma while waiting for next element truncates tail |
| `{"o":{"k":1,},"z":2` | `{"o":{"k":1}}` | inner invalid trailing comma invalidates following outer sibling |
| `[1.0e-x` | `[1.0]` | broken exponent + noise keeps only safe numeric prefix |
| `{"a":1}\n\njunk` | `{"a":1}` | done state drops all non-whitespace tail content |
| `{"bad":"\q123` | `{}` | invalid escape in first member means no completed member survives |

## Extreme stress cases

| Input | Output | Notes |
| --- | --- | --- |
| `{"l1":{"l2":{"l3":{"l4":{"l5":{"l6":{"l7":{"l8":{"v":"x` | `{"l1":{"l2":{"l3":{"l4":{"l5":{"l6":{"l7":{"l8":{"v":"x"}}}}}}}}}` | 8-level object + unterminated deep string |
| `[{"a":[{"b":[{"c":[{"d":[{"e":[{"f":[1,2,3` | `[{"a":[{"b":[{"c":[{"d":[{"e":[{"f":[1,2,3]}]}]}]}]}]}]` | 7-level alternating array/object closure |
| `{"text":"line1\\nline2\\tpath=C:\\\\tmp\\\\","meta":{"ok":true,"msg":"partial` | `{"text":"line1\\nline2\\tpath=C:\\\\tmp\\\\","meta":{"ok":true,"msg":"partial"}}` | escaped chars + nested string close |
| `{"u":"\\u4f60\\u597d","arr":[{"k":"v"}, {"k2":"v2"}], "tail":{"x":1},"z` | `{"u":"\\u4f60\\u597d","arr":[{"k":"v"},{"k2":"v2"}],"tail":{"x":1}}` | unicode escape with unfinished trailing key |
| `{"a":[1,2,3],"b":{"c":[4,5,{"d":"e"}]},"f":false,"g":n` | `{"a":[1,2,3],"b":{"c":[4,5,{"d":"e"}]},"f":false}` | incomplete null literal dropped |
| `{"session":{"id":"s1","chunks":[{"i":0,"data":"abc"},{"i":1,"data":"def"},{"i":2,"data":"ghi` | `{"session":{"id":"s1","chunks":[{"i":0,"data":"abc"},{"i":1,"data":"def"},{"i":2,"data":"ghi"}]}}` | long sibling list, last element incomplete |
| `{"safe":{"k":"v"},"unsafe":{"x":1},"end":` | `{"safe":{"k":"v"},"unsafe":{"x":1}}` | preserve maximal safe prefix only |
| `[{"id":1,"p":{"q":{"r":[{"s":"t"}]}}},{"id":2,"p":{"q":{"r":[{"s":"u` | `[{"id":1,"p":{"q":{"r":[{"s":"t"}]}}},{"id":2,"p":{"q":{"r":[{"s":"u"}]}}}]` | multi-item root array, partial last node |

## Validation requirements

- If repaired output is non-empty, it must be parseable by each language's standard JSON parser.
- Empty output (`""`) is allowed when no safe prefix exists (for example: incomplete root literal like `fals` or `nu`).
- Repair must be idempotent:
  - `repair(repair(input)) == repair(input)`.
- Prefix policy must hold:
  - output must not include characters from any unsafe trailing fragment.

## Random truncation validation

- Generate valid random JSON trees with nested objects/arrays and scalar leaves.
- Truncate each generated JSON at a random byte/char index.
- Repair the truncated text and verify:
  - repaired text is valid JSON,
  - repair is idempotent.
- Script:
  - `test/test-fuzz-python.py`
