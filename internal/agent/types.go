package agent

type State struct {
	DiffPath string
	Language string

	ParsedFiles []string
	ParsedLines any // 先用 any，后面再收紧类型
	Findings    any

	StepCount int
	MaxSteps  int
	Done      bool
}
type Plan struct {
	Thought   string
	ToolName  string
	ToolInput map[string]any
	Finish    bool
}
