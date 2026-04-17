package tools

import (
	"context"
	"fmt"

	"github.com/Arisgod1/code-reviewer-agent/internal/types"
)

type FormatReportTool struct{}

func (t *FormatReportTool) Name() string { return "format_report" }

func (t *FormatReportTool) Run(ctx context.Context, input map[string]any) (map[string]any, error) {
	val, ok := input["findings"]
	if !ok {
		return nil, fmt.Errorf("missing findings")
	}

	findings, ok := val.([]types.Finding)
	if !ok {
		return nil, fmt.Errorf("invalid findings type")
	}

	r := BuildReport(findings)

	return map[string]any{
		"review_report": r,
	}, nil
}
