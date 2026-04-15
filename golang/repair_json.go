package jsonrepair

import "encoding/json"

func isHex4At(text string, start int) bool {
	if start+4 > len(text) {
		return false
	}
	for k := 0; k < 4; k++ {
		c := text[start+k]
		isDigit := c >= '0' && c <= '9'
		isLowerHex := c >= 'a' && c <= 'f'
		isUpperHex := c >= 'A' && c <= 'F'
		if !isDigit && !isLowerHex && !isUpperHex {
			return false
		}
	}
	return true
}

func scanNumberEnd(text string, start int) int {
	n := len(text)
	i := start

	if i < n && text[i] == '-' {
		i++
		if i >= n {
			return -1
		}
	}
	if i >= n {
		return -1
	}

	if text[i] == '0' {
		i++
	} else if text[i] >= '1' && text[i] <= '9' {
		i++
		for i < n && text[i] >= '0' && text[i] <= '9' {
			i++
		}
	} else {
		return -1
	}

	if i < n && text[i] == '.' {
		if i+1 >= n || !(text[i+1] >= '0' && text[i+1] <= '9') {
			return i - 1
		}
		i += 2
		for i < n && text[i] >= '0' && text[i] <= '9' {
			i++
		}
	}

	if i < n && (text[i] == 'e' || text[i] == 'E') {
		if i+1 >= n {
			return i - 1
		}
		j := i + 1
		if text[j] == '+' || text[j] == '-' {
			j++
		}
		if j >= n || !(text[j] >= '0' && text[j] <= '9') {
			return i - 1
		}
		i = j + 1
		for i < n && text[i] >= '0' && text[i] <= '9' {
			i++
		}
	}

	return i - 1
}

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
					if !isHex4At(text, i+1) {
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
				end := scanNumberEnd(text, i)
				if end < i {
					brokeEarly = true
					break
				}
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

func RepairJSONStrictPrefixWithOption(text string, returnObject bool) (any, error) {
	if !returnObject {
		return RepairJSONStrictPrefix(text), nil
	}
	var parsedOriginal any
	if err := json.Unmarshal([]byte(text), &parsedOriginal); err == nil {
		return parsedOriginal, nil
	}
	repaired := RepairJSONStrictPrefix(text)
	if repaired == "" {
		return nil, nil
	}
	var parsed any
	if err := json.Unmarshal([]byte(repaired), &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

func RepairJSONStrictPrefixWithAppendOption(text string, appendContent string, returnObject bool) (any, error) {
	return RepairJSONStrictPrefixWithOption(text+appendContent, returnObject)
}

func RepairJSONStrictPrefixBoth(text string) (string, any, error) {
	repaired := RepairJSONStrictPrefix(text)
	if repaired == "" {
		return repaired, nil, nil
	}
	var parsed any
	if err := json.Unmarshal([]byte(repaired), &parsed); err != nil {
		return repaired, nil, err
	}
	return repaired, parsed, nil
}

func RepairJSONStrictPrefixBothWithAppend(text string, appendContent string) (string, any, error) {
	return RepairJSONStrictPrefixBoth(text + appendContent)
}
