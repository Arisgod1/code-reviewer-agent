package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
	Retries int
}

type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
	Retries int
}

func NewClient(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	retries := cfg.Retries
	if retries < 0 {
		retries = 0
	}
	return &Client{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		HTTP:    &http.Client{Timeout: timeout},
		Retries: retries,
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatReq struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatResp struct {
	Choices []struct {
		Message struct {
			Role             string          `json:"role"`
			Content          json.RawMessage `json:"content"`
			ReasoningContent json.RawMessage `json:"reasoning_content,omitempty"`
			Reasoning        json.RawMessage `json:"reasoning,omitempty"`
		} `json:"message"`
		Text         string `json:"text"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	OutputText string `json:"output_text"`
	Error      *struct {
		Message string `json:"message"`
		Type    string `json:"type,omitempty"`
	} `json:"error,omitempty"`
}

func (c *Client) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is empty")
	}
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 30 * time.Second}
	}

	attempts := c.Retries + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		out, err := c.chatOnce(ctx, messages)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if !isRetryableError(err) || i == attempts-1 {
			break
		}

		backoff := time.Duration(i+1) * 300 * time.Millisecond
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	return "", lastErr
}

func (c *Client) chatOnce(ctx context.Context, messages []ChatMessage) (string, error) {

	body := chatReq{
		Model:       c.Model,
		Messages:    messages,
		Temperature: 0,
		Stream:      false,
		MaxTokens:   512,
	}
	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("llm http status=%d body=%s", resp.StatusCode, shortText(string(bodyBytes), 300))
	}

	var out chatResp
	if err := json.Unmarshal(bodyBytes, &out); err != nil {
		return "", fmt.Errorf("llm decode failed: %w body=%s", err, shortText(string(bodyBytes), 300))
	}

	if out.Error != nil {
		return "", fmt.Errorf("llm api error: %s", out.Error.Message)
	}

	content, err := extractChatContent(out)
	if err != nil {
		return "", fmt.Errorf("%w body=%s", err, shortText(string(bodyBytes), 300))
	}
	return content, nil
}

func extractChatContent(out chatResp) (string, error) {
	if strings.TrimSpace(out.OutputText) != "" {
		return out.OutputText, nil
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("llm api empty choices")
	}

	for _, ch := range out.Choices {
		if strings.TrimSpace(ch.Text) != "" {
			return ch.Text, nil
		}
		text := decodeContentText(ch.Message.Content)
		if strings.TrimSpace(text) != "" {
			return text, nil
		}
		text = decodeContentText(ch.Message.ReasoningContent)
		if strings.TrimSpace(text) != "" {
			return text, nil
		}
		text = decodeContentText(ch.Message.Reasoning)
		if strings.TrimSpace(text) != "" {
			return text, nil
		}
	}
	return "", fmt.Errorf("llm api empty content")
}

func decodeContentText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	var parts []map[string]any
	if err := json.Unmarshal(raw, &parts); err == nil {
		var b strings.Builder
		for _, p := range parts {
			v, _ := p["text"].(string)
			if v != "" {
				if b.Len() > 0 {
					b.WriteByte('\n')
				}
				b.WriteString(v)
			}
		}
		return b.String()
	}

	return ""
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if netErr, ok := errors.AsType[net.Error](err); ok && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "deadline exceeded") || strings.Contains(msg, "timeout") {
		return true
	}
	for _, code := range []string{"status=429", "status=500", "status=502", "status=503", "status=504"} {
		if strings.Contains(msg, code) {
			return true
		}
	}
	return false
}

func shortText(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func ExtractJSONObject(text string) string {
	text = strings.TrimSpace(text)
	start := strings.Index(text, "{")
	if start < 0 {
		return ""
	}

	inString := false
	escaped := false
	depth := 0
	for i := start; i < len(text); i++ {
		ch := text[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			switch ch {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}
			continue
		}

		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return strings.TrimSpace(text[start : i+1])
			}
		}
	}
	return ""
}
