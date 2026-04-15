# Code Review Report

## Summary

Found potential risks in code changes.

Findings: **3**

### 1. SQL_STRING_CONCAT
- File: `main.go:10`
- Severity: `high`
- Confidence: `0.90`
- Rule: SQL string concatenation risk
- Evidence:
  - `   sql := "select * from users where name = '" + user + "'"`
- Suggestion: Use parameterized query / prepared statement instead of string concatenation.

### 2. SENSITIVE_LOGGING
- File: `main.go:11`
- Severity: `medium`
- Confidence: `0.75`
- Rule: Sensitive data may be logged
- Evidence:
  - `   fmt.Println("debug user:", user)`
- Suggestion: Avoid logging sensitive fields; mask or remove them.

### 3. HARDCODED_SECRET
- File: `main.go:20`
- Severity: `high`
- Confidence: `0.80`
- Rule: Possible hardcoded secret
- Evidence:
  - `   apiKey := "sk_test_123456"`
- Suggestion: Move secrets to environment variables or secret manager.

