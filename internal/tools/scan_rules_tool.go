package tools

import (
	"context"
	"fmt"
)

type ScanRulesTool struct{}

func (t *ScanRulesTool) Name() string { return "scan_risk_rules" }

func (t *ScanRulesTool) Run(ctx context.Context, input map[string]any) (map[string]any, error) {
	linesVal, ok := input["lines"]
	if !ok {
		return nil, fmt.Errorf("missing lines")
	}
	langVal, _ := input["language"]
	lang, _ := langVal.(string)

	lines, ok := linesVal.([]ChangedLine)
	if !ok {
		return nil, fmt.Errorf("invalid lines type")
	}

	parsed := ParseDiffResult{
		Files: []string{},
		Lines: lines,
	}

	findings := ScanRiskRules(parsed, lang)

	return map[string]any{
		"findings": findings,
	}, nil
}
