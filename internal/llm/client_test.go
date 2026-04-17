package llm

import (
	"encoding/json"
	"testing"
)

func TestDecodeContentText_String(t *testing.T) {
	raw := json.RawMessage(`"hello"`)
	if got := decodeContentText(raw); got != "hello" {
		t.Fatalf("decodeContentText()=%q, want hello", got)
	}
}

func TestDecodeContentText_ArrayParts(t *testing.T) {
	raw := json.RawMessage(`[{"type":"text","text":"a"},{"type":"text","text":"b"}]`)
	if got := decodeContentText(raw); got != "a\nb" {
		t.Fatalf("decodeContentText()=%q, want a\\nb", got)
	}
}

func TestExtractChatContent_OutputTextFirst(t *testing.T) {
	out := chatResp{OutputText: "from output_text"}
	got, err := extractChatContent(out)
	if err != nil {
		t.Fatalf("extractChatContent error: %v", err)
	}
	if got != "from output_text" {
		t.Fatalf("extractChatContent()=%q, want from output_text", got)
	}
}

func TestExtractChatContent_ChoiceTextFallback(t *testing.T) {
	var out chatResp
	if err := json.Unmarshal([]byte(`{"choices":[{"text":"from text field"}]}`), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got, err := extractChatContent(out)
	if err != nil {
		t.Fatalf("extractChatContent error: %v", err)
	}
	if got != "from text field" {
		t.Fatalf("extractChatContent()=%q, want from text field", got)
	}
}

func TestExtractChatContent_Empty(t *testing.T) {
	_, err := extractChatContent(chatResp{})
	if err == nil {
		t.Fatalf("expected error for empty response")
	}
}

func TestExtractJSONObject(t *testing.T) {
	got := ExtractJSONObject("prefix ```json\n{\"thought\":\"x\",\"tool_name\":\"parse_diff\",\"tool_input\":{},\"finish\":false}\n``` suffix")
	if got != "{\"thought\":\"x\",\"tool_name\":\"parse_diff\",\"tool_input\":{},\"finish\":false}" {
		t.Fatalf("extractJSONObject()=%q", got)
	}
}

func TestExtractChatContent_ReasoningContentFallback(t *testing.T) {
	var out chatResp
	if err := json.Unmarshal([]byte(`{"choices":[{"message":{"reasoning_content":"from reasoning content"}}]}`), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got, err := extractChatContent(out)
	if err != nil {
		t.Fatalf("extractChatContent error: %v", err)
	}
	if got != "from reasoning content" {
		t.Fatalf("extractChatContent()=%q, want from reasoning content", got)
	}
}

