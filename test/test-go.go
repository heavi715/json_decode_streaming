package main

import (
	"encoding/json"
	"fmt"
	"os"

	jsonrepair "json_decode_streaming/golang"
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

	if failed {
		os.Exit(1)
	}

	fmt.Printf("All %d Go cases passed.\n", len(cases))
}
