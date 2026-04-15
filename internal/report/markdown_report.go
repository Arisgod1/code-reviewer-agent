package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/Arisgod1/code-reviewer-agent/internal/types"
)

func WriteMarkdownReport(r types.ReviewReport, path string) error {
	var b strings.Builder

	b.WriteString("# Code Review Report\n\n")
	b.WriteString(fmt.Sprintf("## Summary\n\n%s\n\n", r.Summary))
	b.WriteString(fmt.Sprintf("Findings: **%d**\n\n", len(r.Findings)))

	for i, f := range r.Findings {
		b.WriteString(fmt.Sprintf("### %d. %s\n", i+1, f.ID))
		b.WriteString(fmt.Sprintf("- File: `%s:%d`\n", f.File, f.Line))
		b.WriteString(fmt.Sprintf("- Severity: `%s`\n", f.Severity))
		b.WriteString(fmt.Sprintf("- Confidence: `%.2f`\n", f.Confidence))
		b.WriteString(fmt.Sprintf("- Rule: %s\n", f.Rule))
		if len(f.Evidence) > 0 {
			b.WriteString("- Evidence:\n")
			for _, e := range f.Evidence {
				b.WriteString(fmt.Sprintf("  - `%s`\n", e))
			}
		}
		b.WriteString(fmt.Sprintf("- Suggestion: %s\n\n", f.Suggestion))
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}
