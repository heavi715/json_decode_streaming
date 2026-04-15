//go:build benchgo
// +build benchgo

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	jsonrepair "github.com/heavi715/json_decode_streaming/golang"
)

func buildSamples() map[string]string {
	small := `{"a":1,"b":[1,2,3],"c":"hello"}{"a":1,"b":[1,2,3],"c":"hello"}{"a":1,"b":[1,2,3],"c":"hello"}{"a":1,"b":[1,2,3],"c":"hello"}`

	mediumItems := make([]map[string]any, 0, 200)
	for i := 0; i < 200; i++ {
		arr := make([]int, 20)
		for j := 0; j < 20; j++ {
			arr[j] = j
		}
		mediumItems = append(mediumItems, map[string]any{
			"id":   i,
			"name": "xxxxxxxxxxxxxxxxxxxx",
			"arr":  arr,
		})
	}
	mediumBytes, _ := json.Marshal(map[string]any{"items": mediumItems})

	largeItems := make([]map[string]any, 0, 2000)
	for i := 0; i < 2000; i++ {
		arr := make([]int, 40)
		for j := 0; j < 40; j++ {
			arr[j] = j
		}
		largeItems = append(largeItems, map[string]any{
			"id":   i,
			"name": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			"arr":  arr,
			"obj": map[string]any{
				"k": "vvvvvvvvvv",
				"n": i,
			},
		})
	}
	largeBytes, _ := json.Marshal(map[string]any{"items": largeItems})

	return map[string]string{
		"small":  small,
		"medium": string(mediumBytes),
		"large":  string(largeBytes),
	}
}

func iterationsFor(name string) int {
	if name == "small" {
		return 2000
	}
	if name == "medium" {
		return 400
	}
	return 80
}

func main() {
	mode := flag.String("mode", "all", "benchmark mode: string|object|both_return|all")
	flag.Parse()
	if *mode != "string" && *mode != "object" && *mode != "both_return" && *mode != "all" {
		panic("invalid -mode, use string|object|both_return|all")
	}

	for name, text := range buildSamples() {
		truncated := text[:len(text)-17]
		n := iterationsFor(name)
		if *mode == "string" || *mode == "all" {
			runBench(name, truncated, n, "string")
		}
		if *mode == "object" || *mode == "all" {
			runBench(name, truncated, n, "object")
		}
		if *mode == "both_return" || *mode == "all" {
			runBench(name, truncated, n, "both_return")
		}
	}
}

func runBench(name string, truncated string, n int, mode string) {
	start := time.Now()
	for i := 0; i < n; i++ {
		if mode == "object" {
			_, _ = jsonrepair.RepairJSONStrictPrefixWithOption(truncated, true)
		} else if mode == "both_return" {
			_, _, _ = jsonrepair.RepairJSONStrictPrefixBoth(truncated)
		} else {
			_ = jsonrepair.RepairJSONStrictPrefix(truncated)
		}
	}
	dt := time.Since(start).Seconds()
	avgUs := (dt / float64(n)) * 1_000_000
	throughput := (float64(len(truncated)) * float64(n)) / dt / 1024 / 1024
	fmt.Printf("%s/%s: len=%d n=%d avg_us=%.1f throughput_mib_s=%.2f\n", name, mode, len(truncated), n, avgUs, throughput)
}
