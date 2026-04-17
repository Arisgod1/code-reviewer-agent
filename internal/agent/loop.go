package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/Arisgod1/code-reviewer-agent/internal/tools"
	"github.com/Arisgod1/code-reviewer-agent/internal/types"
)

func limitText(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

type LoopResult struct {
	Report    types.ReviewReport
	Trace     []types.TraceStep
	ToolCalls []types.ToolCall
}

func RunReviewLoop(
	ctx context.Context,
	reg *tools.Registry,
	diffPath string,
	language string,
	maxSteps int,
	llmPlanner *LLMPlanner,
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
		plan, plannerSource, plannerReason := NextPlanWithFallback(ctx, state, llmPlanner)
		observation := fmt.Sprintf("source=%s tool=%s finish=%v", plannerSource, plan.ToolName, plan.Finish)
		if plannerSource != "llm" && plannerReason != "" {
			observation += " reason=" + limitText(plannerReason, 180)
		}
		addTrace(
			"Planner selected next step",
			"plan",
			observation,
		)
		state.StepCount++

		if plan.Finish {
			state.Done = true
			addTrace(plan.Thought, "finish", "loop completed")
			break
		}

		out, err := callTool(plan.ToolName, 2*time.Second, plan.ToolInput, plan.Thought)
		if err != nil {
			return LoopResult{}, err
		}

		switch plan.ToolName {
		case "parse_diff":
			decoded, err := tools.DecodeParseDiffOutput(out)
			if err != nil {
				return LoopResult{}, err
			}
			state.ParsedFiles = decoded.Files
			state.ParsedLines = decoded.Lines

			addTrace(
				"Summarize parse_diff result",
				"parse_diff summary",
				fmt.Sprintf("files=%d, changed_lines=%d", len(state.ParsedFiles), len(state.ParsedLines)),
			)
			state.HasParsedDiff = true
		case "scan_risk_rules":
			decoded, err := tools.DecodeScanRulesOutput(out)
			if err != nil {
				return LoopResult{}, err
			}
			state.Findings = decoded.Findings

			addTrace(
				"Summarize scan_risk_rules result",
				"scan_risk_rules summary",
				fmt.Sprintf("findings=%d", len(state.Findings)),
			)
			state.HasScannedRules = true
		case "format_report":
			r, ok := out["review_report"].(types.ReviewReport)
			if !ok {
				return LoopResult{}, fmt.Errorf("format_report output review_report type invalid")
			}
			state.Report = r
			state.HasFormatted = true

			addTrace(
				"Summarize format_report result",
				"format_report summary",
				fmt.Sprintf("findings=%d", len(state.Report.Findings)),
			)
		default:
			return LoopResult{}, fmt.Errorf("unknown tool in plan: %s", plan.ToolName)
		}
	}

	return LoopResult{
		Report:    state.Report,
		Trace:     trace,
		ToolCalls: toolCalls,
	}, nil
}
