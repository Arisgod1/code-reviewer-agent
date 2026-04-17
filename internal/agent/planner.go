package agent

func NextPlan(s *State) Plan {
	if s.Done {
		return Plan{Finish: true, Thought: "Already finished"}
	}

	// 还没解析 diff
	if s.ParsedLines == nil {
		return Plan{
			Thought:  "Need changed files and lines from diff",
			ToolName: "parse_diff",
			ToolInput: map[string]any{
				"diff_path": s.DiffPath,
			},
		}
	}

	// 还没扫描规则
	if s.Findings == nil {
		return Plan{
			Thought:  "Need to scan risk rules over changed lines",
			ToolName: "scan_risk_rules",
			ToolInput: map[string]any{
				"lines":    s.ParsedLines,
				"language": s.Language,
			},
		}
	}

	return Plan{
		Finish:  true,
		Thought: "Enough evidence collected; can finish",
	}
}
