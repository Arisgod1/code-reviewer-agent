package tools

import (
	"strings"

	"github.com/Arisgod1/code-reviewer-agent/internal/types"
)

func ScanRiskRules(parsed ParseDiffResult, language string) []types.Finding {
	findings := make([]types.Finding, 0)

	for _, l := range parsed.Lines {
		c := strings.ToLower(l.Content)

		// 规则1：SQL 字符串拼接
		// 简化判断：包含 select/insert/update/delete 且包含 '+' 拼接
		if (strings.Contains(c, "select ") ||
			strings.Contains(c, "insert ") ||
			strings.Contains(c, "update ") ||
			strings.Contains(c, "delete ")) &&
			strings.Contains(c, "+") {
			findings = append(findings, types.Finding{
				ID:         "SQL_STRING_CONCAT",
				File:       l.File,
				Line:       l.Line,
				Rule:       "SQL string concatenation risk",
				Severity:   "high",
				Confidence: 0.90,
				Evidence:   []string{l.Content},
				Suggestion: "Use parameterized query / prepared statement instead of string concatenation.",
			})
		}

		// 规则2：敏感日志打印（MVP启发式）
		if (strings.Contains(c, "fmt.println") || strings.Contains(c, "log.") || strings.Contains(c, "print(")) &&
			(strings.Contains(c, "user") || strings.Contains(c, "token") || strings.Contains(c, "password") || strings.Contains(c, "secret")) {
			findings = append(findings, types.Finding{
				ID:         "SENSITIVE_LOGGING",
				File:       l.File,
				Line:       l.Line,
				Rule:       "Sensitive data may be logged",
				Severity:   "medium",
				Confidence: 0.75,
				Evidence:   []string{l.Content},
				Suggestion: "Avoid logging sensitive fields; mask or remove them.",
			})
		}
		// 规则3：硬编码密钥（MVP 启发式）
		// 命中关键词 + 赋值 + 引号，减少一点误报
		if (strings.Contains(c, "apikey") ||
			strings.Contains(c, "api_key") ||
			strings.Contains(c, "secret") ||
			strings.Contains(c, "password") ||
			strings.Contains(c, "token")) &&
			strings.Contains(c, "=") &&
			(strings.Contains(c, "\"") || strings.Contains(c, "'")) {
			findings = append(findings, types.Finding{
				ID:         "HARDCODED_SECRET",
				File:       l.File,
				Line:       l.Line,
				Rule:       "Possible hardcoded secret",
				Severity:   "high",
				Confidence: 0.80,
				Evidence:   []string{l.Content},
				Suggestion: "Move secrets to environment variables or secret manager.",
			})
		}
	}

	return findings
}
