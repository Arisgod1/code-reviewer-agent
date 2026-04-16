package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/Arisgod1/code-reviewer-agent/internal/tools"
	"github.com/Arisgod1/code-reviewer-agent/internal/types"
)

type LoopResult struct {
	Findings  []types.Finding
	Trace     []types.TraceStep
	ToolCalls []types.ToolCall
}

func RunReviewLoop(
	ctx context.Context,
	reg *tools.Registry,
	diffPath string,
	language string,
	maxSteps int,
) (LoopResult, error) {
	state := &State{
		DiffPath: diffPath,
		Language: language,
		MaxSteps: maxSteps,
	}

	trace := make([]types.TraceStep, 0)
	toolCalls := make([]types.ToolCall, 0)

	step := 1
	addTrace := func(thought, action, observation string) {
		trace = append(trace, types.TraceStep{
			Step:        step,
			Thought:     thought,
			Action:      action,
			Observation: observation,
			Timestamp:   time.Now().Format(time.RFC3339),
		})
		step++
	}

	for !state.Done && state.StepCount < state.MaxSteps {
		action := NextAction(state)
		state.StepCount++

		switch action {
		case "parse_diff":
			t, err := reg.Get("parse_diff")
			if err != nil {
				return LoopResult{}, err
			}

			start := time.Now()
			out, err := t.Run(ctx, map[string]any{"diff_path": state.DiffPath})
			cost := time.Since(start).Milliseconds()

			tc := types.ToolCall{Name: "parse_diff", CostMs: cost, Success: err == nil}
			if err != nil {
				tc.Output = "error: " + err.Error()
				toolCalls = append(toolCalls, tc)
				return LoopResult{}, err
			}
			tc.Output = "ok"
			toolCalls = append(toolCalls, tc)

			files, _ := out["files"].([]string)
			lines, ok := out["lines"].([]tools.ChangedLine)
			if !ok {
				return LoopResult{}, fmt.Errorf("parse_diff output lines type invalid")
			}

			state.ParsedFiles = files
			state.ParsedLines = lines

			addTrace("Need changed files and lines", "parse_diff", fmt.Sprintf("files=%d, changed_lines=%d", len(files), len(lines)))

		case "scan_risk_rules":
			t, err := reg.Get("scan_risk_rules")
			if err != nil {
				return LoopResult{}, err
			}

			lines, ok := state.ParsedLines.([]tools.ChangedLine)
			if !ok {
				return LoopResult{}, fmt.Errorf("state ParsedLines type invalid")
			}

			start := time.Now()
			out, err := t.Run(ctx, map[string]any{
				"lines":    lines,
				"language": state.Language,
			})
			cost := time.Since(start).Milliseconds()

			tc := types.ToolCall{Name: "scan_risk_rules", CostMs: cost, Success: err == nil}
			if err != nil {
				tc.Output = "error: " + err.Error()
				toolCalls = append(toolCalls, tc)
				return LoopResult{}, err
			}
			tc.Output = "ok"
			toolCalls = append(toolCalls, tc)

			findings, ok := out["findings"].([]types.Finding)
			if !ok {
				return LoopResult{}, fmt.Errorf("scan_risk_rules output findings type invalid")
			}
			state.Findings = findings

			addTrace("Need risk findings from changed lines", "scan_risk_rules", fmt.Sprintf("findings=%d", len(findings)))

		case "finish":
			state.Done = true
			addTrace("Enough evidence collected", "finish", "loop completed")

		default:
			return LoopResult{}, fmt.Errorf("unknown action: %s", action)
		}
	}

	finalFindings, _ := state.Findings.([]types.Finding)
	return LoopResult{
		Findings:  finalFindings,
		Trace:     trace,
		ToolCalls: toolCalls,
	}, nil
}
