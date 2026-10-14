package sftcollect

import (
	"bufio"
	"bytes"
	"encoding/json"
	"maps"
	"slices"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
)

type Message struct {
	Role       string          `json:"role"`
	Content    string          `json:"content,omitempty"`
	ToolCalls  json.RawMessage `json:"tool_calls,omitempty"`
	ToolCallId string          `json:"tool_call_id,omitempty"`
}

type Envelope struct {
	Model        string          `json:"model"`
	FinishReason string          `json:"finish_reason,omitempty"`
	Messages     []Message       `json:"messages"`
	Tools        json.RawMessage `json:"tools,omitempty"`
}

func SnapshotMessages(msgs []dto.Message) ([]Message, bool) {
	if len(msgs) == 0 {
		return nil, false
	}
	out := make([]Message, 0, len(msgs))
	hasUser := false
	for i := range msgs {
		msg := &msgs[i]
		if isMultimodal(msg) {
			return nil, false
		}
		copied := Message{
			Role:       msg.Role,
			Content:    msg.StringContent(),
			ToolCallId: msg.ToolCallId,
		}
		if len(msg.ToolCalls) > 0 {
			copied.ToolCalls = append(json.RawMessage(nil), msg.ToolCalls...)
		}
		if msg.Role == "user" && copied.Content != "" {
			hasUser = true
		}
		out = append(out, copied)
	}
	if !hasUser {
		return nil, false
	}
	return out, true
}

func AcceptableFinish(reason string) bool {
	switch reason {
	case "", types.FinishReasonStop, types.FinishReasonToolCalls, types.FinishReasonFunctionCall:
		return true
	default:
		return false
	}
}

func ParseAssistant(body []byte, isStream bool) (Message, string, bool) {
	if len(bytes.TrimSpace(body)) == 0 {
		return Message{}, "", false
	}
	if isStream || looksLikeSSE(body) {
		return parseStreamAssistant(body)
	}
	return parseJSONAssistant(body)
}

func isMultimodal(msg *dto.Message) bool {
	parts := msg.ParseContent()
	for _, part := range parts {
		if part.Type != "" && part.Type != dto.ContentTypeText {
			return true
		}
		if part.ImageUrl != nil || part.InputAudio != nil || part.File != nil || part.VideoUrl != nil {
			return true
		}
	}
	return false
}

func looksLikeSSE(body []byte) bool {
	trimmed := bytes.TrimSpace(body)
	return bytes.HasPrefix(trimmed, []byte("data:")) || bytes.Contains(body, []byte("\ndata:"))
}

func parseJSONAssistant(body []byte) (Message, string, bool) {
	var resp dto.OpenAITextResponse
	if err := common.Unmarshal(body, &resp); err != nil || len(resp.Choices) == 0 {
		return Message{}, "", false
	}
	choice := resp.Choices[0]
	msg := snapshotAssistant(&choice.Message)
	if !hasTrainableAssistant(msg) {
		return Message{}, "", false
	}
	return msg, choice.FinishReason, true
}

func parseStreamAssistant(body []byte) (Message, string, bool) {
	var content strings.Builder
	tools := map[int]*dto.ToolCallResponse{}
	finish := ""
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 0, 64*1024), maxCaptureBytes)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		payload, ok := sseDataPayload(line)
		if !ok || payload == "[DONE]" {
			continue
		}
		var chunk dto.ChatCompletionsStreamResponse
		if err := common.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		for i := range chunk.Choices {
			choice := &chunk.Choices[i]
			if choice.FinishReason != nil && *choice.FinishReason != "" {
				finish = *choice.FinishReason
			}
			content.WriteString(choice.Delta.GetContentString())
			for _, tool := range choice.Delta.ToolCalls {
				idx := 0
				if tool.Index != nil {
					idx = *tool.Index
				}
				acc, exists := tools[idx]
				if !exists {
					copied := tool
					if copied.Type == "" {
						copied.Type = "function"
					}
					tools[idx] = &copied
					continue
				}
				if tool.ID != "" {
					acc.ID = tool.ID
				}
				if tool.Type != "" {
					acc.Type = tool.Type
				}
				if tool.Function.Name != "" {
					acc.Function.Name = tool.Function.Name
				}
				acc.Function.Arguments += tool.Function.Arguments
			}
		}
	}
	msg := Message{Role: "assistant", Content: content.String()}
	if len(tools) > 0 {
		indices := slices.Sorted(maps.Keys(tools))
		ordered := make([]dto.ToolCallResponse, 0, len(tools))
		for _, index := range indices {
			ordered = append(ordered, *tools[index])
		}
		if raw, err := common.Marshal(ordered); err == nil {
			msg.ToolCalls = raw
		}
	}
	if !hasTrainableAssistant(msg) {
		return Message{}, "", false
	}
	return msg, finish, true
}

func sseDataPayload(line string) (string, bool) {
	if !strings.HasPrefix(line, "data:") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(line, "data:")), true
}

func snapshotAssistant(msg *dto.Message) Message {
	out := Message{
		Role:    "assistant",
		Content: msg.StringContent(),
	}
	if len(msg.ToolCalls) > 0 {
		out.ToolCalls = append(json.RawMessage(nil), msg.ToolCalls...)
	}
	return out
}

func hasTrainableAssistant(msg Message) bool {
	return strings.TrimSpace(msg.Content) != "" || len(bytes.TrimSpace(msg.ToolCalls)) > 0
}
