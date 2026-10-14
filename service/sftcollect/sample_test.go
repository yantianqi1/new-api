package sftcollect

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnapshotMessagesSkipsMultimodal(t *testing.T) {
	msgs := []dto.Message{
		{Role: "user", Content: []any{
			map[string]any{"type": "text", "text": "what is this"},
			map[string]any{"type": "image_url", "image_url": map[string]any{"url": "https://example.com/a.png"}},
		}},
	}
	_, ok := SnapshotMessages(msgs)
	assert.False(t, ok)
}

func TestSnapshotMessagesKeepsTextAndTools(t *testing.T) {
	toolCalls, err := common.Marshal([]dto.ToolCallRequest{{
		ID:   "call_1",
		Type: "function",
		Function: dto.FunctionRequest{
			Name:      "get_weather",
			Arguments: `{"city":"beijing"}`,
		},
	}})
	require.NoError(t, err)
	msgs := []dto.Message{
		{Role: "system", Content: "you are helpful"},
		{Role: "user", Content: "weather?"},
		{Role: "assistant", ToolCalls: toolCalls},
		{Role: "tool", ToolCallId: "call_1", Content: `{"ok":true}`},
	}
	got, ok := SnapshotMessages(msgs)
	require.True(t, ok)
	require.Len(t, got, 4)
	assert.Equal(t, "system", got[0].Role)
	assert.Equal(t, "you are helpful", got[0].Content)
	assert.Equal(t, "user", got[1].Role)
	assert.Equal(t, "call_1", got[3].ToolCallId)
	assert.NotEmpty(t, got[2].ToolCalls)
}

func TestParseJSONAssistant(t *testing.T) {
	body := []byte(`{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"hello there"}}]}`)
	msg, finish, ok := ParseAssistant(body, false)
	require.True(t, ok)
	assert.Equal(t, "stop", finish)
	assert.Equal(t, "hello there", msg.Content)
}

func TestParseStreamAssistantAssemblesToolCalls(t *testing.T) {
	body := []byte(`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"lookup","arguments":""}}]}}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"city\":\"beijing\"}"}}]}}]}

data: {"choices":[{"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]

`)
	msg, finish, ok := ParseAssistant(body, true)
	require.True(t, ok)
	assert.Equal(t, "tool_calls", finish)
	require.NotEmpty(t, msg.ToolCalls)
	var tools []dto.ToolCallResponse
	require.NoError(t, common.Unmarshal(msg.ToolCalls, &tools))
	require.Len(t, tools, 1)
	assert.Equal(t, "call_1", tools[0].ID)
	assert.Equal(t, "lookup", tools[0].Function.Name)
	assert.Equal(t, `{"city":"beijing"}`, tools[0].Function.Arguments)
}

func TestParseStreamAssistantConcatenatesText(t *testing.T) {
	body := []byte("data: {\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}]}\n\n" +
		"data: [DONE]\n\n")
	msg, finish, ok := ParseAssistant(body, true)
	require.True(t, ok)
	assert.Equal(t, "stop", finish)
	assert.Equal(t, "Hello", msg.Content)
}

func TestAcceptableFinish(t *testing.T) {
	assert.True(t, AcceptableFinish(""))
	assert.True(t, AcceptableFinish("stop"))
	assert.True(t, AcceptableFinish("tool_calls"))
	assert.False(t, AcceptableFinish("length"))
	assert.False(t, AcceptableFinish("content_filter"))
}

func TestParseAssistantRejectsEmpty(t *testing.T) {
	body := []byte(`{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":""}}]}`)
	_, _, ok := ParseAssistant(body, false)
	assert.False(t, ok)
}

func TestParseStreamAssistantPreservesSparseToolIndices(t *testing.T) {
	body := []byte(`data: {"choices":[{"delta":{"tool_calls":[{"index":2147483647,"id":"last","type":"function","function":{"name":"last","arguments":"{}"}},{"index":2,"id":"first","type":"function","function":{"name":"first","arguments":"{}"}}]},"finish_reason":"tool_calls"}]}

`)
	msg, finish, ok := ParseAssistant(body, true)
	require.True(t, ok)
	assert.Equal(t, "tool_calls", finish)
	var tools []dto.ToolCallResponse
	require.NoError(t, common.Unmarshal(msg.ToolCalls, &tools))
	require.Len(t, tools, 2)
	assert.Equal(t, "first", tools[0].ID)
	assert.Equal(t, "last", tools[1].ID)
}
