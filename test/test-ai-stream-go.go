//go:build aistream
// +build aistream

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	jsonrepair "github.com/heavi715/json_decode_streaming/golang"
)

type reqBody struct {
	Model    string        `json:"model"`
	Messages []messageItem `json:"messages"`
	Stream   bool          `json:"stream"`
}

type messageItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func main() {
	apiKey := os.Getenv("AI_STREAM_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Missing AI_STREAM_API_KEY")
		os.Exit(1)
	}

	apiURL := os.Getenv("AI_STREAM_URL")
	if apiURL == "" {
		apiURL = "http://new-api.bangong.knowbox.cn/v1/chat/completions"
	}
	model := os.Getenv("AI_STREAM_MODEL")
	if model == "" {
		model = "claude-opus-4-20250514"
	}
	prompt := os.Getenv("AI_STREAM_PROMPT")
	if prompt == "" {
		prompt = `只返回json，格式: {"ok":true,"msg":"..."}`
	}
	printSnapshots := os.Getenv("AI_STREAM_PRINT_SNAPSHOTS") != "0"
	maxSnapshots := 20
	if v := os.Getenv("AI_STREAM_MAX_SNAPSHOTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxSnapshots = n
		}
	}

	body, err := json.Marshal(reqBody{
		Model: model,
		Messages: []messageItem{
			{Role: "user", Content: prompt},
		},
		Stream: true,
	})
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "request failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	rawBytes := new(bytes.Buffer)
	if _, err := rawBytes.ReadFrom(resp.Body); err != nil {
		fmt.Fprintln(os.Stderr, "read response failed:", err)
		os.Exit(1)
	}
	raw := rawBytes.String()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintln(os.Stderr, "HTTP status:", resp.StatusCode)
		fmt.Fprintln(os.Stderr, "response sample:")
		if len(raw) > 500 {
			fmt.Fprintln(os.Stderr, raw[:500])
		} else {
			fmt.Fprintln(os.Stderr, raw)
		}
		os.Exit(1)
	}

	lines := strings.Split(raw, "\n")
	accumulated := ""
	chunkCount := 0
	snapshotCount := 0
	printed := 0
	eventCount := 0
	skipped := 0
	debugSamples := make([]string, 0, 3)

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		eventCount++

		var event map[string]any
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			skipped++
			if len(debugSamples) < 3 {
				debugSamples = append(debugSamples, truncate(payload, 200))
			}
			continue
		}

		piece := extractDeltaContent(event)
		if piece == "" {
			skipped++
			if len(debugSamples) < 3 {
				eventRaw, _ := json.Marshal(event)
				debugSamples = append(debugSamples, truncate(string(eventRaw), 200))
			}
			continue
		}

		chunkCount++
		obj, err := jsonrepair.RepairJSONStrictPrefixWithAppendOption(accumulated, piece, true)
		if err != nil {
			continue
		}
		accumulated += piece
		if obj != nil {
			snapshotCount++
			if printSnapshots && printed < maxSnapshots {
				printed++
				objRaw, _ := json.Marshal(obj)
				fmt.Printf("snapshot#%d: %s\n", snapshotCount, string(objRaw))
			}
		}
	}

	final := jsonrepair.RepairJSONStrictPrefix(accumulated)
	finalObj, err := jsonrepair.RepairJSONStrictPrefixWithOption(accumulated, true)
	if err != nil || finalObj == nil {
		fmt.Println("Failed to parse streamed content as JSON.")
		fmt.Println("Repaired text:", final)
		fmt.Printf("events: %d, content chunks: %d, skipped events: %d\n", eventCount, chunkCount, skipped)
		if strings.TrimSpace(raw) != "" {
			fmt.Println("raw response sample:")
			fmt.Println(truncate(raw, 500))
		}
		if len(debugSamples) > 0 {
			fmt.Println("sample skipped payloads:")
			for _, s := range debugSamples {
				fmt.Println(s)
			}
		}
		os.Exit(1)
	}

	fmt.Printf("content chunks: %d\n", chunkCount)
	fmt.Printf("object snapshots: %d\n", snapshotCount)
	if printSnapshots && snapshotCount > maxSnapshots {
		fmt.Printf("snapshot output truncated: printed %d of %d\n", maxSnapshots, snapshotCount)
	}
	fmt.Printf("events: %d, skipped events: %d\n", eventCount, skipped)
	fmt.Printf("final repaired json: %s\n", final)
	fmt.Printf("final object type: %T\n", finalObj)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func extractDeltaContent(event map[string]any) string {
	choicesAny, ok := event["choices"]
	if !ok {
		return ""
	}
	choices, ok := choicesAny.([]any)
	if !ok || len(choices) == 0 {
		return ""
	}
	choice0, ok := choices[0].(map[string]any)
	if !ok {
		return ""
	}
	deltaAny, ok := choice0["delta"]
	if !ok {
		return ""
	}
	delta, ok := deltaAny.(map[string]any)
	if !ok {
		return ""
	}
	contentAny, ok := delta["content"]
	if !ok {
		return ""
	}
	content, ok := contentAny.(string)
	if !ok {
		return ""
	}
	return content
}
