package agent

import "context"

func RuleBasedNextPlan(s *State) Plan {
	if s.Done {
		return Plan{Finish: true, Thought: "Already finished"}
	}

	if !s.HasParsedDiff {
		return Plan{
			Thought:  "Need changed files and lines from diff",
			ToolName: "parse_diff",
			ToolInput: map[string]any{
				"diff_path": s.DiffPath,
			},
		}
	}

	if !s.HasScannedRules {
		return Plan{
			Thought:  "Need to scan risk rules over changed lines",
			ToolName: "scan_risk_rules",
			ToolInput: map[string]any{
				"lines":    s.ParsedLines,
				"language": s.Language,
			},
		}
	}

	if !s.HasFormatted {
		return Plan{
			Thought:  "Need to format final review report",
			ToolName: "format_report",
			ToolInput: map[string]any{
				"findings": s.Findings,
			},
		}
	}
	return Plan{
		Finish:  true,
		Thought: "Enough evidence collected; can finish",
	}
}
func NextPlanWithFallback(ctx context.Context, s *State, llmPlanner *LLMPlanner) (Plan, string, string) {
	if llmPlanner == nil {
		return RuleBasedNextPlan(s), "fallback", "llm planner is nil"
	}
	if llmPlanner.Client == nil {
		return RuleBasedNextPlan(s), "fallback", "llm client is nil"
	}

	p, err := llmPlanner.NextPlan(ctx, s)
	if err != nil {
		return RuleBasedNextPlan(s), "fallback", err.Error()
	}

	return p, "llm", ""
}
