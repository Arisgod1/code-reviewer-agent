package types

type ReviewRequest struct {
	DiffPath string            `json:"diff_path"`
	Language string            `json:"language"`
	Options  map[string]string `json:"options"`
}

type Finding struct {
	ID         string   `json:"id"`
	File       string   `json:"file"`
	Line       int      `json:"line"`
	Rule       string   `json:"rule"`
	Severity   string   `json:"severity"`   // low/medium/high
	Confidence float64  `json:"confidence"` // 0~1
	Evidence   []string `json:"evidence"`
	Suggestion string   `json:"suggestion"`
}

type ToolCall struct {
	Name    string `json:"name"`
	CostMs  int64  `json:"cost_ms"`
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

type TraceStep struct {
	Step        int    `json:"step"`
	Thought     string `json:"thought"`
	Action      string `json:"action"`
	Observation string `json:"observation"`
	Timestamp   string `json:"timestamp"`
}

type ReviewReport struct {
	Summary  string         `json:"summary"`
	Findings []Finding      `json:"findings"`
	Metrics  map[string]any `json:"metrics"`
	Trace    []TraceStep    `json:"trace"`
}
