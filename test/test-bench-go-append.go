//go:build benchgoappend
// +build benchgoappend

package main

import (
	"encoding/json"
	"fmt"
	"time"

	jsonrepair "github.com/heavi715/json_decode_streaming/golang"
)

func buildJSONText(size int) string {
	items := make([]map[string]any, 0, size)
	for i := 0; i < size; i++ {
		items = append(items, map[string]any{
			"id":   i,
			"name": fmt.Sprintf("name-%d", i),
			"tags": []string{fmt.Sprintf("t%d", i%5), fmt.Sprintf("g%d", i%7)},
			"ok":   i%2 == 0,
		})
	}
	payload := map[string]any{
		"items": items,
		"meta": map[string]any{
			"count":  size,
			"source": "bench-append-go",
		},
	}
	raw, _ := json.Marshal(payload)
	return string(raw)
}

func splitIntoPieces(text string, pieceSize int) []string {
	out := make([]string, 0, len(text)/pieceSize+1)
	for i := 0; i < len(text); i += pieceSize {
		end := i + pieceSize
		if end > len(text) {
			end = len(text)
		}
		out = append(out, text[i:end])
	}
	return out
}

func runOneStream(pieces []string, mode string) {
	accumulated := ""
	for _, piece := range pieces {
		if mode == "append" {
			_, _ = jsonrepair.RepairJSONStrictPrefixWithAppendOption(accumulated, piece, false)
		} else {
			_ = jsonrepair.RepairJSONStrictPrefix(accumulated + piece)
		}
		accumulated += piece
	}
}

func bench(name string, pieces []string, rounds int) {
	totalPieceBytes := 0
	for _, p := range pieces {
		totalPieceBytes += len(p)
	}
	totalBytes := float64(totalPieceBytes * rounds)

	startAppend := time.Now()
	for i := 0; i < rounds; i++ {
		runOneStream(pieces, "append")
	}
	appendSec := time.Since(startAppend).Seconds()

	startFull := time.Now()
	for i := 0; i < rounds; i++ {
		runOneStream(pieces, "full")
	}
	fullSec := time.Since(startFull).Seconds()

	appendThroughput := totalBytes / appendSec / 1024 / 1024
	fullThroughput := totalBytes / fullSec / 1024 / 1024
	speedup := fullSec / appendSec
	fmt.Printf(
		"%s: pieces=%d rounds=%d append_s=%.3f full_s=%.3f append_mib_s=%.2f full_mib_s=%.2f speedup=%.2fx\n",
		name, len(pieces), rounds, appendSec, fullSec, appendThroughput, fullThroughput, speedup,
	)
}

func main() {
	scenarios := []struct {
		name      string
		size      int
		pieceSize int
		rounds    int
	}{
		{name: "small", size: 120, pieceSize: 24, rounds: 200},
		{name: "medium", size: 600, pieceSize: 32, rounds: 80},
		{name: "large", size: 1600, pieceSize: 48, rounds: 30},
	}

	for _, scenario := range scenarios {
		text := buildJSONText(scenario.size)
		pieces := splitIntoPieces(text, scenario.pieceSize)
		bench(scenario.name, pieces, scenario.rounds)
	}
}
