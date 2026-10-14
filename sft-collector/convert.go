package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Envelope struct {
	Model        string          `json:"model"`
	FinishReason string          `json:"finish_reason,omitempty"`
	Messages     []ChatMessage   `json:"messages"`
	Tools        json.RawMessage `json:"tools,omitempty"`
}

type ChatMessage struct {
	Role       string          `json:"role"`
	Content    string          `json:"content,omitempty"`
	ToolCalls  json.RawMessage `json:"tool_calls,omitempty"`
	ToolCallId string          `json:"tool_call_id,omitempty"`
}

type ShareGPTRecord struct {
	Conversations []ShareGPTTurn `json:"conversations"`
	System        string         `json:"system,omitempty"`
	Tools         string         `json:"tools,omitempty"`
}

type ShareGPTTurn struct {
	From  string `json:"from"`
	Value string `json:"value"`
}

type openaiToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

func toShareGPT(env Envelope) (ShareGPTRecord, bool) {
	var rec ShareGPTRecord
	var systems []string
	for _, msg := range env.Messages {
		switch strings.ToLower(msg.Role) {
		case "system":
			if text := strings.TrimSpace(msg.Content); text != "" {
				systems = append(systems, text)
			}
		case "user":
			if text := strings.TrimSpace(msg.Content); text == "" {
				return ShareGPTRecord{}, false
			}
			rec.Conversations = append(rec.Conversations, ShareGPTTurn{From: "human", Value: msg.Content})
		case "assistant":
			if text := strings.TrimSpace(msg.Content); text != "" {
				rec.Conversations = append(rec.Conversations, ShareGPTTurn{From: "gpt", Value: msg.Content})
			}
			for _, call := range parseToolCalls(msg.ToolCalls) {
				value, err := json.Marshal(map[string]any{
					"name":      call.Function.Name,
					"arguments": parseArguments(call.Function.Arguments),
				})
				if err != nil {
					return ShareGPTRecord{}, false
				}
				rec.Conversations = append(rec.Conversations, ShareGPTTurn{From: "function_call", Value: string(value)})
			}
		case "tool", "function":
			rec.Conversations = append(rec.Conversations, ShareGPTTurn{From: "observation", Value: msg.Content})
		}
	}
	if !shareGPTTrainable(rec) {
		return ShareGPTRecord{}, false
	}
	if len(systems) > 0 {
		rec.System = strings.Join(systems, "\n")
	}
	if len(bytes.TrimSpace(env.Tools)) > 0 {
		rec.Tools = string(env.Tools)
	}
	return rec, true
}

func shareGPTTrainable(rec ShareGPTRecord) bool {
	hasHuman := false
	hasReply := false
	for _, turn := range rec.Conversations {
		if turn.From == "human" && strings.TrimSpace(turn.Value) != "" {
			hasHuman = true
		}
		if (turn.From == "gpt" || turn.From == "function_call") && strings.TrimSpace(turn.Value) != "" {
			hasReply = true
		}
	}
	return hasHuman && hasReply
}

func parseToolCalls(raw json.RawMessage) []openaiToolCall {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}
	var calls []openaiToolCall
	if err := json.Unmarshal(raw, &calls); err != nil {
		return nil
	}
	return calls
}

func parseArguments(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]any{}
	}
	var asObject any
	if err := json.Unmarshal([]byte(trimmed), &asObject); err == nil {
		return asObject
	}
	return raw
}

func encodeShareGPT(w io.Writer, rec ShareGPTRecord) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(rec)
}

func packJSONArray(jsonlPath, outPath string) error {
	f, err := os.Open(jsonlPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	tmp := outPath + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = out.Close()
		if !ok {
			_ = os.Remove(tmp)
		}
	}()

	if _, err := out.Write([]byte("[\n")); err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 2<<20)
	first := true
	count := 0
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		if !json.Valid(line) {
			continue
		}
		if !first {
			if _, err := out.Write([]byte(",\n")); err != nil {
				return err
			}
		}
		if _, err := out.Write(line); err != nil {
			return err
		}
		first = false
		count++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if count == 0 {
		_ = os.Remove(tmp)
		return nil
	}
	if _, err := out.Write([]byte("\n]\n")); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	ok = true
	return os.Rename(tmp, outPath)
}

func loadAllowlist(fromEnv, fromFile string) map[string]struct{} {
	set := map[string]struct{}{}
	add := func(raw string) {
		for _, part := range strings.Split(raw, ",") {
			name := strings.TrimSpace(part)
			if name == "" || strings.HasPrefix(name, "#") {
				continue
			}
			set[name] = struct{}{}
		}
	}
	add(fromEnv)
	if fromFile == "" {
		return set
	}
	data, err := os.ReadFile(fromFile)
	if err != nil {
		return set
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		set[line] = struct{}{}
	}
	return set
}

func modelAllowed(model string, allowlist map[string]struct{}) bool {
	if len(allowlist) == 0 {
		return true
	}
	_, ok := allowlist[model]
	return ok
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

func workPath(dir, day string) string {
	return filepath.Join(dir, day+".sharegpt.jsonl")
}

func exportPath(dir, day string) string {
	return filepath.Join(dir, "sharegpt-"+day+".json")
}
