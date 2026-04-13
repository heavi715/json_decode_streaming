package jsonrepair

import "regexp"

var numberRe = regexp.MustCompile(`^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?`)
var hex4Re = regexp.MustCompile(`^[0-9a-fA-F]{4}$`)

func RepairJSONStrictPrefix(text string) string {
	stack := make([]string, 0)
	state := "root_value"
	inString := false
	escapeNext := false
	stringRole := ""
	lastSafe := -1
	arrayWaitingValue := false
	objectWaitingKey := false
	i := 0
	brokeEarly := false

	completeValue := func(idx int) {
		arrayWaitingValue = false
		objectWaitingKey = false
		if len(stack) == 0 {
			state = "done"
			lastSafe = idx
			return
		}
		top := stack[len(stack)-1]
		if top == "object" {
			state = "object_comma_or_end"
			lastSafe = idx
		} else {
			state = "array_comma_or_end"
			lastSafe = idx
		}
	}

	for i < len(text) {
		ch := text[i]

		if inString {
			if escapeNext {
				if ch == '"' || ch == '\\' || ch == '/' || ch == 'b' || ch == 'f' || ch == 'n' || ch == 'r' || ch == 't' {
					escapeNext = false
					i++
					continue
				}
				if ch == 'u' {
					if i+4 >= len(text) {
						brokeEarly = true
						break
					}
					hex := text[i+1 : i+5]
					if !hex4Re.MatchString(hex) {
						brokeEarly = true
						break
					}
					escapeNext = false
					i += 5
					continue
				}
				brokeEarly = true
				break
			}
			if ch == '\\' {
				escapeNext = true
				i++
				continue
			}
			if ch == '"' {
				inString = false
				if stringRole == "key" {
					state = "object_colon"
				} else {
					completeValue(i)
				}
			}
			i++
			continue
		}

		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			i++
			continue
		}

		if state == "done" {
			brokeEarly = true
			break
		}

		if state == "root_value" || state == "object_value" || state == "array_value_or_end" {
			if ch == '{' {
				stack = append(stack, "object")
				state = "object_key_or_end"
				lastSafe = i
				i++
				continue
			}
			if ch == '[' {
				stack = append(stack, "array")
				state = "array_value_or_end"
				lastSafe = i
				i++
				continue
			}
			if ch == '"' {
				inString = true
				stringRole = "value"
				i++
				continue
			}
			if (ch >= '0' && ch <= '9') || ch == '-' {
				match := numberRe.FindString(text[i:])
				if match == "" {
					brokeEarly = true
					break
				}
				end := i + len(match) - 1
				i = end + 1
				completeValue(end)
				continue
			}
			if len(text)-i >= 4 && text[i:i+4] == "true" {
				i += 4
				completeValue(i - 1)
				continue
			}
			if len(text)-i >= 5 && text[i:i+5] == "false" {
				i += 5
				completeValue(i - 1)
				continue
			}
			if len(text)-i >= 4 && text[i:i+4] == "null" {
				i += 4
				completeValue(i - 1)
				continue
			}
			if state == "array_value_or_end" && ch == ']' {
				if arrayWaitingValue {
					brokeEarly = true
					break
				}
				stack = stack[:len(stack)-1]
				completeValue(i)
				i++
				continue
			}
			brokeEarly = true
			break
		}

		if state == "object_key_or_end" {
			if ch == '}' {
				if objectWaitingKey {
					brokeEarly = true
					break
				}
				stack = stack[:len(stack)-1]
				completeValue(i)
				i++
				continue
			}
			if ch == '"' {
				objectWaitingKey = false
				inString = true
				stringRole = "key"
				i++
				continue
			}
			brokeEarly = true
			break
		}

		if state == "object_colon" {
			if ch == ':' {
				state = "object_value"
				i++
				continue
			}
			brokeEarly = true
			break
		}

		if state == "object_comma_or_end" {
			if ch == ',' {
				state = "object_key_or_end"
				objectWaitingKey = true
				i++
				continue
			}
			if ch == '}' {
				stack = stack[:len(stack)-1]
				completeValue(i)
				i++
				continue
			}
			brokeEarly = true
			break
		}

		if state == "array_comma_or_end" {
			if ch == ',' {
				state = "array_value_or_end"
				arrayWaitingValue = true
				i++
				continue
			}
			if ch == ']' {
				stack = stack[:len(stack)-1]
				completeValue(i)
				i++
				continue
			}
			brokeEarly = true
			break
		}

		brokeEarly = true
		break
	}

	base := ""
	if inString && !brokeEarly && !escapeNext && stringRole == "value" {
		base = text + `"`
		completeValue(len(text))
	} else if lastSafe >= 0 {
		base = text[:lastSafe+1]
	}

	closers := ""
	for idx := len(stack) - 1; idx >= 0; idx-- {
		if stack[idx] == "object" {
			closers += "}"
		} else {
			closers += "]"
		}
	}

	return base + closers
}
