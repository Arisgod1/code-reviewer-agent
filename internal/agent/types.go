package agent

import (
	"github.com/Arisgod1/code-reviewer-agent/internal/tools"
	"github.com/Arisgod1/code-reviewer-agent/internal/types"
)

type State struct {
	DiffPath string
	Language string

	ParsedFiles []string
	ParsedLines []tools.ChangedLine
	Findings    []types.Finding

	HasParsedDiff   bool
	HasScannedRules bool

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
