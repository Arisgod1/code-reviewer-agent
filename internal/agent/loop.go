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
	callTool := func(name string, timeout time.Duration, input map[string]any, thought string) (map[string]any, error) {
		t, err := reg.Get(name)
		if err != nil {
			toolCalls = append(toolCalls, types.ToolCall{
				Name:    name,
				CostMs:  0,
				Success: false,
				Output:  "error: " + err.Error(),
			})
			return nil, err
		}

		toolCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		start := time.Now()
		out, err := t.Run(toolCtx, input)
		cost := time.Since(start).Milliseconds()

		tc := types.ToolCall{
			Name:    name,
			CostMs:  cost,
			Success: err == nil,
		}
		if err != nil {
			tc.Output = "error: " + err.Error()
			toolCalls = append(toolCalls, tc)

			addTrace(thought, "call tool: "+name, "failed: "+err.Error())
			return nil, err
		}

		tc.Output = "ok"
		toolCalls = append(toolCalls, tc)

		addTrace(thought, "call tool: "+name, "ok")
		return out, nil
	}
	for !state.Done && state.StepCount < state.MaxSteps {
		action := NextAction(state)
		state.StepCount++

		switch action {
		case "parse_diff":
			out, err := callTool(
				"parse_diff",
				2*time.Second,
				map[string]any{"diff_path": state.DiffPath},
				"Need changed files and lines",
			)
			if err != nil {
				return LoopResult{}, err
			}

			files, _ := out["files"].([]string)
			lines, ok := out["lines"].([]tools.ChangedLine)
			if !ok {
				return LoopResult{}, fmt.Errorf("parse_diff output lines type invalid")
			}

			state.ParsedFiles = files
			state.ParsedLines = lines

			// 这里补一条更有信息量的 trace（覆盖上面 callTool 的 'ok'）
			addTrace(
				"Summarize parse_diff result",
				"parse_diff summary",
				fmt.Sprintf("files=%d, changed_lines=%d", len(files), len(lines)),
			)
		case "scan_risk_rules":
			lines, ok := state.ParsedLines.([]tools.ChangedLine)
			if !ok {
				return LoopResult{}, fmt.Errorf("state ParsedLines type invalid")
			}

			out, err := callTool(
				"scan_risk_rules",
				2*time.Second,
				map[string]any{
					"lines":    lines,
					"language": state.Language,
				},
				"Need risk findings from changed lines",
			)
			if err != nil {
				return LoopResult{}, err
			}

			findings, ok := out["findings"].([]types.Finding)
			if !ok {
				return LoopResult{}, fmt.Errorf("scan_risk_rules output findings type invalid")
			}
			state.Findings = findings

			addTrace(
				"Summarize scan_risk_rules result",
				"scan_risk_rules summary",
				fmt.Sprintf("findings=%d", len(findings)),
			)
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
