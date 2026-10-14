package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToShareGPTWithTools(t *testing.T) {
	env := Envelope{
		Model: "gpt-4.1",
		Messages: []ChatMessage{
			{Role: "system", Content: "be concise"},
			{Role: "user", Content: "weather in beijing?"},
			{Role: "assistant", ToolCalls: json.RawMessage(`[{"id":"call_1","type":"function","function":{"name":"get_weather","arguments":"{\"city\":\"beijing\"}"}}]`)},
			{Role: "tool", ToolCallId: "call_1", Content: `{"temp":22}`},
			{Role: "assistant", Content: "It is 22C."},
		},
		Tools: json.RawMessage(`[{"type":"function","function":{"name":"get_weather"}}]`),
	}
	rec, ok := toShareGPT(env)
	if !ok {
		t.Fatal("expected convertible sample")
	}
	if rec.System != "be concise" {
		t.Fatalf("system=%q", rec.System)
	}
	if rec.Tools == "" {
		t.Fatal("expected tools json string")
	}
	if len(rec.Conversations) != 4 {
		t.Fatalf("turns=%d %#v", len(rec.Conversations), rec.Conversations)
	}
	if rec.Conversations[0].From != "human" || rec.Conversations[1].From != "function_call" {
		t.Fatalf("unexpected turns: %#v", rec.Conversations)
	}
	if rec.Conversations[2].From != "observation" || rec.Conversations[3].From != "gpt" {
		t.Fatalf("unexpected turns: %#v", rec.Conversations)
	}
	var call map[string]any
	if err := json.Unmarshal([]byte(rec.Conversations[1].Value), &call); err != nil {
		t.Fatal(err)
	}
	if call["name"] != "get_weather" {
		t.Fatalf("name=%v", call["name"])
	}
	args, ok := call["arguments"].(map[string]any)
	if !ok || args["city"] != "beijing" {
		t.Fatalf("arguments=%v", call["arguments"])
	}
}

func TestRedactSecrets(t *testing.T) {
	rec := redactRecord(ShareGPTRecord{
		System: "contact admin@example.com",
		Conversations: []ShareGPTTurn{
			{From: "human", Value: "my key is sk-abcdefghijklmnop"},
			{From: "gpt", Value: "Authorization: Bearer abcdefghijk"},
		},
	})
	if rec.System != "contact [REDACTED_EMAIL]" {
		t.Fatalf("system=%q", rec.System)
	}
	if rec.Conversations[0].Value != "my key is [REDACTED_KEY]" {
		t.Fatalf("human=%q", rec.Conversations[0].Value)
	}
	if rec.Conversations[1].Value != "Authorization: Bearer [REDACTED_TOKEN]" {
		t.Fatalf("gpt=%q", rec.Conversations[1].Value)
	}
}

func TestModelAllowlist(t *testing.T) {
	allow := map[string]struct{}{"gpt-4.1": {}}
	if !modelAllowed("gpt-4.1", allow) {
		t.Fatal("expected allow")
	}
	if modelAllowed("gpt-4o", allow) {
		t.Fatal("expected deny")
	}
	if !modelAllowed("anything", nil) {
		t.Fatal("empty allowlist should collect all")
	}
}

func TestPackJSONArray(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.sharegpt.jsonl")
	dst := filepath.Join(dir, "a.json")
	content := "{\"conversations\":[{\"from\":\"human\",\"value\":\"hi\"},{\"from\":\"gpt\",\"value\":\"yo\"}]}\n"
	if err := os.WriteFile(src, []byte(content+content), 0644); err != nil {
		t.Fatal(err)
	}
	if err := packJSONArray(src, dst); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	var arr []ShareGPTRecord
	if err := json.Unmarshal(raw, &arr); err != nil {
		t.Fatal(err)
	}
	if len(arr) != 2 {
		t.Fatalf("len=%d body=%s", len(arr), raw)
	}
	if !strings.Contains(string(raw), `"from": "human"`) && !strings.Contains(string(raw), `"from":"human"`) {
		t.Fatalf("unexpected json: %s", raw)
	}
}
