package jsonrepair

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
)

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

type repairState struct {
	text              string
	stack             []string
	state             string
	inString          bool
	escapeNext        bool
	stringRole        string
	lastSafe          int
	arrayWaitingValue bool
	objectWaitingKey  bool
	i                 int
	brokeEarly        bool
}

func newRepairState() *repairState {
	return &repairState{
		stack:    make([]string, 0),
		state:    "root_value",
		lastSafe: -1,
	}
}

func (s *repairState) clone() *repairState {
	cloned := *s
	cloned.stack = append([]string(nil), s.stack...)
	return &cloned
}

func (s *repairState) completeValue(idx int) {
	s.arrayWaitingValue = false
	s.objectWaitingKey = false
	if len(s.stack) == 0 {
		s.state = "done"
		s.lastSafe = idx
		return
	}
	top := s.stack[len(s.stack)-1]
	if top == "object" {
		s.state = "object_comma_or_end"
		s.lastSafe = idx
	} else {
		s.state = "array_comma_or_end"
		s.lastSafe = idx
	}
}

func (s *repairState) feed(chunk string) {
	if chunk == "" {
		return
	}
	if s.brokeEarly {
		s.text += chunk
		return
	}
	s.text += chunk

	for s.i < len(s.text) {
		ch := s.text[s.i]

		if s.inString {
			if s.escapeNext {
				if ch == '"' || ch == '\\' || ch == '/' || ch == 'b' || ch == 'f' || ch == 'n' || ch == 'r' || ch == 't' {
					s.escapeNext = false
					s.i++
					continue
				}
				if ch == 'u' {
					if s.i+4 >= len(s.text) {
						// Wait for remaining hex digits in next chunk.
						break
					}
					if !isHex4At(s.text, s.i+1) {
						s.brokeEarly = true
						break
					}
					s.escapeNext = false
					s.i += 5
					continue
				}
				s.brokeEarly = true
				break
			}
			if ch == '\\' {
				s.escapeNext = true
				s.i++
				continue
			}
			if ch == '"' {
				s.inString = false
				if s.stringRole == "key" {
					s.state = "object_colon"
				} else {
					s.completeValue(s.i)
				}
			}
			s.i++
			continue
		}

		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			s.i++
			continue
		}

		if s.state == "done" {
			s.brokeEarly = true
			break
		}

		if s.state == "root_value" || s.state == "object_value" || s.state == "array_value_or_end" {
			if ch == '{' {
				s.stack = append(s.stack, "object")
				s.state = "object_key_or_end"
				s.lastSafe = s.i
				s.i++
				continue
			}
			if ch == '[' {
				s.stack = append(s.stack, "array")
				s.state = "array_value_or_end"
				s.lastSafe = s.i
				s.i++
				continue
			}
			if ch == '"' {
				s.inString = true
				s.stringRole = "value"
				s.i++
				continue
			}
			if (ch >= '0' && ch <= '9') || ch == '-' {
				end := scanNumberEnd(s.text, s.i)
				if end < s.i {
					s.brokeEarly = true
					break
				}
				s.i = end + 1
				s.completeValue(end)
				continue
			}
			if len(s.text)-s.i >= 4 && s.text[s.i:s.i+4] == "true" {
				s.i += 4
				s.completeValue(s.i - 1)
				continue
			}
			if len(s.text)-s.i >= 5 && s.text[s.i:s.i+5] == "false" {
				s.i += 5
				s.completeValue(s.i - 1)
				continue
			}
			if len(s.text)-s.i >= 4 && s.text[s.i:s.i+4] == "null" {
				s.i += 4
				s.completeValue(s.i - 1)
				continue
			}
			if s.state == "array_value_or_end" && ch == ']' {
				if s.arrayWaitingValue {
					s.brokeEarly = true
					break
				}
				s.stack = s.stack[:len(s.stack)-1]
				s.completeValue(s.i)
				s.i++
				continue
			}
			s.brokeEarly = true
			break
		}

		if s.state == "object_key_or_end" {
			if ch == '}' {
				if s.objectWaitingKey {
					s.brokeEarly = true
					break
				}
				s.stack = s.stack[:len(s.stack)-1]
				s.completeValue(s.i)
				s.i++
				continue
			}
			if ch == '"' {
				s.objectWaitingKey = false
				s.inString = true
				s.stringRole = "key"
				s.i++
				continue
			}
			s.brokeEarly = true
			break
		}

		if s.state == "object_colon" {
			if ch == ':' {
				s.state = "object_value"
				s.i++
				continue
			}
			s.brokeEarly = true
			break
		}

		if s.state == "object_comma_or_end" {
			if ch == ',' {
				s.state = "object_key_or_end"
				s.objectWaitingKey = true
				s.i++
				continue
			}
			if ch == '}' {
				s.stack = s.stack[:len(s.stack)-1]
				s.completeValue(s.i)
				s.i++
				continue
			}
			s.brokeEarly = true
			break
		}

		if s.state == "array_comma_or_end" {
			if ch == ',' {
				s.state = "array_value_or_end"
				s.arrayWaitingValue = true
				s.i++
				continue
			}
			if ch == ']' {
				s.stack = s.stack[:len(s.stack)-1]
				s.completeValue(s.i)
				s.i++
				continue
			}
			s.brokeEarly = true
			break
		}

		s.brokeEarly = true
		break
	}
}

func (s *repairState) snapshot() string {
	base := ""
	if s.inString && !s.brokeEarly && !s.escapeNext && s.stringRole == "value" {
		base = s.text + `"`
	} else if s.lastSafe >= 0 {
		base = s.text[:s.lastSafe+1]
	}
	closers := ""
	for idx := len(s.stack) - 1; idx >= 0; idx-- {
		if s.stack[idx] == "object" {
			closers += "}"
		} else {
			closers += "]"
		}
	}
	return base + closers
}

const defaultAppendCacheMaxEntries = 256
const defaultAppendCacheMaxTotalBytes = 4 * 1024 * 1024
const defaultAppendCacheTTL = 2 * time.Minute

var (
	appendStateCacheMu sync.Mutex
	appendStateCache   = map[string]appendCacheEntry{}
	appendCacheOrder   = make([]string, 0, defaultAppendCacheMaxEntries)
	appendCacheBytes   = 0
	appendCacheConfig  = appendCacheConfigState{
		maxEntries:    defaultAppendCacheMaxEntries,
		maxTotalBytes: defaultAppendCacheMaxTotalBytes,
		ttl:           defaultAppendCacheTTL,
	}
)

type appendCacheEntry struct {
	state    *repairState
	keyBytes int
	expires  time.Time
}

type appendCacheConfigState struct {
	maxEntries    int
	maxTotalBytes int
	ttl           time.Duration
}

type AppendCachePreset string

const (
	AppendCachePresetDefault        AppendCachePreset = "default"
	AppendCachePresetLowMemory      AppendCachePreset = "low_memory"
	AppendCachePresetHighThroughput AppendCachePreset = "high_throughput"
)

// SetAppendCacheConfig configures append-state cache behavior.
// Pass zero values to keep existing settings.
// Set clear=true to drop current cache entries immediately.
func SetAppendCacheConfig(maxEntries int, maxTotalBytes int, ttl time.Duration, clear bool) error {
	appendStateCacheMu.Lock()
	defer appendStateCacheMu.Unlock()

	if maxEntries < 0 {
		return errors.New("maxEntries must be >= 0")
	}
	if maxTotalBytes < 0 {
		return errors.New("maxTotalBytes must be >= 0")
	}
	if ttl < 0 {
		return errors.New("ttl must be >= 0")
	}
	if maxEntries > 0 {
		appendCacheConfig.maxEntries = maxEntries
	}
	if maxTotalBytes > 0 {
		appendCacheConfig.maxTotalBytes = maxTotalBytes
	}
	if ttl > 0 {
		appendCacheConfig.ttl = ttl
	}
	if clear {
		appendStateCache = map[string]appendCacheEntry{}
		appendCacheOrder = appendCacheOrder[:0]
		appendCacheBytes = 0
		return nil
	}
	pruneAppendCacheLocked(time.Now())
	return nil
}

func ApplyAppendCachePreset(preset AppendCachePreset, clear bool) error {
	switch preset {
	case AppendCachePresetDefault:
		return SetAppendCacheConfig(256, 4*1024*1024, 120*time.Second, clear)
	case AppendCachePresetLowMemory:
		return SetAppendCacheConfig(64, 512*1024, 15*time.Second, clear)
	case AppendCachePresetHighThroughput:
		return SetAppendCacheConfig(1024, 16*1024*1024, 600*time.Second, clear)
	default:
		return errors.New("unknown append cache preset")
	}
}

func removeKeyFromOrder(key string) {
	for i := range appendCacheOrder {
		if appendCacheOrder[i] == key {
			appendCacheOrder = append(appendCacheOrder[:i], appendCacheOrder[i+1:]...)
			return
		}
	}
}

func removeCacheEntryLocked(key string) {
	entry, ok := appendStateCache[key]
	if !ok {
		return
	}
	delete(appendStateCache, key)
	appendCacheBytes -= entry.keyBytes
	removeKeyFromOrder(key)
}

func touchCacheKeyLocked(key string) {
	removeKeyFromOrder(key)
	appendCacheOrder = append(appendCacheOrder, key)
}

func pruneAppendCacheLocked(now time.Time) {
	for len(appendCacheOrder) > 0 {
		oldestKey := appendCacheOrder[0]
		entry, ok := appendStateCache[oldestKey]
		if !ok {
			appendCacheOrder = appendCacheOrder[1:]
			continue
		}
		overLimit := len(appendStateCache) > appendCacheConfig.maxEntries || appendCacheBytes > appendCacheConfig.maxTotalBytes
		expired := !entry.expires.After(now)
		if !overLimit && !expired {
			break
		}
		removeCacheEntryLocked(oldestKey)
	}
}

func getAppendState(key string) *repairState {
	appendStateCacheMu.Lock()
	defer appendStateCacheMu.Unlock()

	now := time.Now()
	state, ok := appendStateCache[key]
	if !ok {
		return nil
	}
	if !state.expires.After(now) {
		removeCacheEntryLocked(key)
		return nil
	}
	state.expires = now.Add(appendCacheConfig.ttl)
	appendStateCache[key] = state
	touchCacheKeyLocked(key)
	return state.state.clone()
}

func putAppendState(key string, state *repairState) {
	appendStateCacheMu.Lock()
	defer appendStateCacheMu.Unlock()

	now := time.Now()
	if _, exists := appendStateCache[key]; exists {
		removeCacheEntryLocked(key)
	}
	appendStateCache[key] = appendCacheEntry{
		state:    state.clone(),
		keyBytes: len(key),
		expires:  now.Add(appendCacheConfig.ttl),
	}
	appendCacheBytes += len(key)
	appendCacheOrder = append(appendCacheOrder, key)
	pruneAppendCacheLocked(now)
}

func repairFromScratchWithState(text string) (string, *repairState) {
	state := newRepairState()
	state.feed(text)
	return state.snapshot(), state
}

func RepairJSONStrictPrefix(text string) string {
	repaired, state := repairFromScratchWithState(text)
	putAppendState(text, state)
	return repaired
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
	fullText := text + appendContent
	if appendContent == "" {
		return RepairJSONStrictPrefixWithOption(text, returnObject)
	}
	if returnObject {
		var parsedOriginal any
		if err := json.Unmarshal([]byte(fullText), &parsedOriginal); err == nil {
			return parsedOriginal, nil
		}
	}
	cached := getAppendState(text)
	var repaired string
	var state *repairState
	if cached != nil {
		cached.feed(appendContent)
		repaired = cached.snapshot()
		state = cached
	} else {
		repaired, state = repairFromScratchWithState(fullText)
	}
	putAppendState(fullText, state)
	if !returnObject {
		return repaired, nil
	}
	if repaired == "" {
		return nil, nil
	}
	var parsed any
	if err := json.Unmarshal([]byte(repaired), &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
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
	repairedAny, err := RepairJSONStrictPrefixWithAppendOption(text, appendContent, false)
	if err != nil {
		return "", nil, err
	}
	repaired := repairedAny.(string)
	if repaired == "" {
		return repaired, nil, nil
	}
	var parsed any
	if err := json.Unmarshal([]byte(repaired), &parsed); err != nil {
		return repaired, nil, err
	}
	return repaired, parsed, nil
}
