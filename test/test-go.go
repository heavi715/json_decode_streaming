package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	jsonrepair "github.com/heavi715/json_decode_streaming/golang"
)

type testCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

func main() {
	raw, err := os.ReadFile("test/cases.json")
	if err != nil {
		panic(err)
	}

	var cases []testCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		panic(err)
	}

	failed := false
	for idx, tc := range cases {
		repaired := jsonrepair.RepairJSONStrictPrefix(tc.Input)
		if repaired != tc.Expected {
			fmt.Printf("[FAIL] case #%d: output mismatch\n", idx)
			fmt.Printf("  actual  : %s\n", repaired)
			fmt.Printf("  expected: %s\n", tc.Expected)
			failed = true
			continue
		}
		repairedObject, err := jsonrepair.RepairJSONStrictPrefixWithOption(tc.Input, true)
		if err != nil {
			fmt.Printf("[FAIL] case #%d: object parse failed: %v\n", idx, err)
			failed = true
			continue
		}
		var expectedObject any
		if tc.Expected != "" {
			if err := json.Unmarshal([]byte(tc.Expected), &expectedObject); err != nil {
				fmt.Printf("[FAIL] case #%d: expected parse failed: %v\n", idx, err)
				failed = true
				continue
			}
		}
		if !reflect.DeepEqual(repairedObject, expectedObject) {
			fmt.Printf("[FAIL] case #%d: object output mismatch\n", idx)
			fmt.Printf("  actual  : %#v\n", repairedObject)
			fmt.Printf("  expected: %#v\n", expectedObject)
			failed = true
			continue
		}
		repairedBoth, repairedBothObject, err := jsonrepair.RepairJSONStrictPrefixBoth(tc.Input)
		if err != nil {
			fmt.Printf("[FAIL] case #%d: both parse failed: %v\n", idx, err)
			failed = true
			continue
		}
		if repairedBoth != tc.Expected {
			fmt.Printf("[FAIL] case #%d: both output mismatch\n", idx)
			fmt.Printf("  actual  : %s\n", repairedBoth)
			fmt.Printf("  expected: %s\n", tc.Expected)
			failed = true
			continue
		}
		if !reflect.DeepEqual(repairedBothObject, expectedObject) {
			fmt.Printf("[FAIL] case #%d: both object output mismatch\n", idx)
			fmt.Printf("  actual  : %#v\n", repairedBothObject)
			fmt.Printf("  expected: %#v\n", expectedObject)
			failed = true
			continue
		}
		if repaired != "" {
			var parsed any
			if err := json.Unmarshal([]byte(repaired), &parsed); err != nil {
				fmt.Printf("[FAIL] case #%d: invalid json: %v\n", idx, err)
				fmt.Printf("  actual  : %s\n", repaired)
				fmt.Printf("  expected: %s\n", tc.Expected)
				failed = true
			}
		}
	}

	appendBase := `{"a":"1"`
	appendTail := `,"b":2}`
	expectedAppend := `{"a":"1","b":2}`
	appendOut, err := jsonrepair.RepairJSONStrictPrefixWithAppendOption(appendBase, appendTail, false)
	if err != nil {
		fmt.Printf("[FAIL] append case: append string failed: %v\n", err)
		failed = true
	} else if appendOut.(string) != expectedAppend {
		fmt.Printf("[FAIL] append case: append output mismatch\n")
		fmt.Printf("  actual  : %s\n", appendOut.(string))
		fmt.Printf("  expected: %s\n", expectedAppend)
		failed = true
	}
	appendObj, err := jsonrepair.RepairJSONStrictPrefixWithAppendOption(appendBase, appendTail, true)
	if err != nil {
		fmt.Printf("[FAIL] append case: append object failed: %v\n", err)
		failed = true
	} else {
		var expectedObj any
		if err := json.Unmarshal([]byte(expectedAppend), &expectedObj); err != nil {
			fmt.Printf("[FAIL] append case: expected parse failed: %v\n", err)
			failed = true
		} else if !reflect.DeepEqual(appendObj, expectedObj) {
			fmt.Printf("[FAIL] append case: append object mismatch\n")
			fmt.Printf("  actual  : %#v\n", appendObj)
			fmt.Printf("  expected: %#v\n", expectedObj)
			failed = true
		}
	}

	if failed {
		os.Exit(1)
	}

	fmt.Printf("All %d Go cases passed.\n", len(cases))
}
