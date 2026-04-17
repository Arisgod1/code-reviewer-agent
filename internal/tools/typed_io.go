package tools

import (
	"fmt"

	"github.com/Arisgod1/code-reviewer-agent/internal/types"
)

type ParseDiffOutput struct {
	Files []string
	Lines []ChangedLine
}

type ScanRulesOutput struct {
	Findings []types.Finding
}

type FormatReportOutput struct {
	ReviewReport types.ReviewReport
}

func DecodeParseDiffOutput(out map[string]any) (ParseDiffOutput, error) {
	files, ok := out["files"].([]string)
	if !ok {
		return ParseDiffOutput{}, fmt.Errorf("parse_diff output files type invalid")
	}
	lines, ok := out["lines"].([]ChangedLine)
	if !ok {
		return ParseDiffOutput{}, fmt.Errorf("parse_diff output lines type invalid")
	}
	return ParseDiffOutput{
		Files: files,
		Lines: lines,
	}, nil
}

func DecodeScanRulesOutput(out map[string]any) (ScanRulesOutput, error) {
	findings, ok := out["findings"].([]types.Finding)
	if !ok {
		return ScanRulesOutput{}, fmt.Errorf("scan_risk_rules output findings type invalid")
	}
	return ScanRulesOutput{
		Findings: findings,
	}, nil
}

func DecodeFormatReportOutput(out map[string]any) (FormatReportOutput, error) {
	r, ok := out["review_report"].(types.ReviewReport)
	if !ok {
		return FormatReportOutput{}, fmt.Errorf("format_report output review_report type invalid")
	}
	return FormatReportOutput{
		ReviewReport: r,
	}, nil
}
