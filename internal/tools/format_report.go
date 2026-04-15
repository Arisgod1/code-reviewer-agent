package tools

import (
	"encoding/json"
	"os"
	"time"

	"github.com/Arisgod1/code-reviewer-agent/internal/types"
)

func BuildReport(findings []types.Finding) types.ReviewReport {
	summary := "No risks found."
	if len(findings) > 0 {
		summary = "Found potential risks in code changes."
	}

	return types.ReviewReport{
		Summary:  summary,
		Findings: findings,
		Metrics: map[string]any{
			"findings_count": len(findings),
			"generated_at":   time.Now().Format(time.RFC3339),
		},
		Trace: []types.TraceStep{},
	}
}

func WriteReportJSON(report types.ReviewReport, outputPath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0644)
}
