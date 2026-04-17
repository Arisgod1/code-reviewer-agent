package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Arisgod1/code-reviewer-agent/internal/llm"
)

type LLMPlanner struct {
	Client *llm.Client
}

type llmPlanResponse struct {
	Thought   string         `json:"thought"`
	ToolName  string         `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
	Finish    bool           `json:"finish"`
}

func (p *LLMPlanner) NextPlan(ctx context.Context, s *State) (Plan, error) {
	if p == nil || p.Client == nil {
		return Plan{}, fmt.Errorf("llm planner client is nil")
	}

	prompt := fmt.Sprintf(`
You are a code-review agent planner.
Return ONLY JSON with keys:
thought, tool_name, tool_input, finish

State:
- has_parsed_diff: %v
- has_scanned_rules: %v
- has_formatted: %v
- diff_path: %s
- language: %s
- parsed_lines_count: %d
- findings_count: %d

Allowed tools:
- parse_diff (input: {"diff_path":"..."})
- scan_risk_rules (input: {"lines":[...], "language":"go|java"})
- format_report (input: {"findings":[...]})

Rules:
- If not parsed -> parse_diff
- Else if not scanned -> scan_risk_rules
- Else if not formatted -> format_report
- Else finish=true
`, s.HasParsedDiff, s.HasScannedRules, s.HasFormatted, s.DiffPath, s.Language, len(s.ParsedLines), len(s.Findings))

	msgs := []llm.ChatMessage{
		{Role: "system", Content: "You must output strict JSON only."},
		{Role: "user", Content: prompt},
	}

	raw, err := p.Client.Chat(ctx, msgs)
	if err != nil {
		return Plan{}, err
	}

	r, err := decodeLLMPlanResponse(raw)
	if err != nil {
		return Plan{}, fmt.Errorf("invalid planner json: %w; raw=%s", err, raw)
	}
	if !r.Finish && r.ToolName == "" {
		r.ToolName = inferNextTool(s)
		if r.ToolName == "" {
			r.Finish = true
		}
		if r.ToolInput == nil {
			r.ToolInput = defaultToolInput(s, r.ToolName)
		}
	}
	if r.Finish {
		r.ToolName = ""
		r.ToolInput = nil
	}

	return Plan{
		Thought:   r.Thought,
		ToolName:  r.ToolName,
		ToolInput: r.ToolInput,
		Finish:    r.Finish,
	}, nil
}

func decodeLLMPlanResponse(raw string) (llmPlanResponse, error) {
	text := strings.TrimSpace(raw)

	var out llmPlanResponse
	if err := json.Unmarshal([]byte(text), &out); err == nil {
		return out, nil
	}

	if candidate := llm.ExtractJSONObject(text); candidate != "" {
		if err := json.Unmarshal([]byte(candidate), &out); err == nil {
			return out, nil
		}
	}

	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	if candidate := llm.ExtractJSONObject(text); candidate != "" {
		if err := json.Unmarshal([]byte(candidate), &out); err == nil {
			return out, nil
		}
	}

	return llmPlanResponse{}, fmt.Errorf("planner response is not valid JSON object")
}

func inferNextTool(s *State) string {
	if !s.HasParsedDiff {
		return "parse_diff"
	}
	if !s.HasScannedRules {
		return "scan_risk_rules"
	}
	if !s.HasFormatted {
		return "format_report"
	}
	return ""
}

func defaultToolInput(s *State, tool string) map[string]any {
	switch tool {
	case "parse_diff":
		return map[string]any{"diff_path": s.DiffPath}
	case "scan_risk_rules":
		return map[string]any{"lines": s.ParsedLines, "language": s.Language}
	case "format_report":
		return map[string]any{"findings": s.Findings}
	default:
		return nil
	}
}
