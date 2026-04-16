//go:build cacheconfiggo
// +build cacheconfiggo

package main

import (
	"fmt"

	jsonrepair "github.com/heavi715/json_decode_streaming/golang"
)

func runStream(label string, preset jsonrepair.AppendCachePreset) {
	if err := jsonrepair.ApplyAppendCachePreset(preset, true); err != nil {
		fmt.Printf("%s: config error: %v\n", label, err)
		return
	}
	accumulated := ""
	pieces := []string{`{"items":[`, `{"id":1},`, `{"id":2},`, `{"id":3}`, `]}`}
	last := ""
	for _, piece := range pieces {
		out, err := jsonrepair.RepairJSONStrictPrefixWithAppendOption(accumulated, piece, false)
		if err != nil {
			fmt.Printf("%s: stream error: %v\n", label, err)
			return
		}
		last = out.(string)
		accumulated += piece
	}
	fmt.Printf("%s: %s\n", label, last)
}

func main() {
	runStream("preset-low-memory", jsonrepair.AppendCachePresetLowMemory)
	runStream("preset-high-throughput", jsonrepair.AppendCachePresetHighThroughput)
}
