package tools

import (
	"os"
	"strconv"
	"strings"
)

type ChangedLine struct {
	File    string
	Line    int
	Content string
}

type ParseDiffResult struct {
	Files []string
	Lines []ChangedLine
}

func ParseDiffFile(path string) (ParseDiffResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ParseDiffResult{}, err
	}

	return ParseDiffContent(string(data)), nil
}

func ParseDiffContent(diff string) ParseDiffResult {
	result := ParseDiffResult{
		Files: []string{},
		Lines: []ChangedLine{},
	}

	lines := strings.Split(diff, "\n")

	currentFile := ""
	currentNewLine := 0

	seenFile := map[string]bool{}

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")

		// 1) 识别文件：+++ b/xxx.go
		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			if !seenFile[currentFile] {
				result.Files = append(result.Files, currentFile)
				seenFile[currentFile] = true
			}
			continue
		}

		// 2) 识别 hunk 头：@@ -10,6 +10,8 @@
		if strings.HasPrefix(line, "@@") {
			// 找 "+10,8" 这一段
			parts := strings.Split(line, " ")
			for _, p := range parts {
				if strings.HasPrefix(p, "+") {
					// p 可能是 +10,8 或 +10
					p = strings.TrimPrefix(p, "+")
					p = strings.TrimSuffix(p, "@@")
					p = strings.TrimSpace(p)

					num := p
					if strings.Contains(p, ",") {
						num = strings.Split(p, ",")[0]
					}
					if n, err := strconv.Atoi(num); err == nil {
						currentNewLine = n
					}
					break
				}
			}
			continue
		}

		// 3) 新增行：以 "+" 开头，但排除 "+++"
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			result.Lines = append(result.Lines, ChangedLine{
				File:    currentFile,
				Line:    currentNewLine,
				Content: strings.TrimPrefix(line, "+"),
			})
			currentNewLine++
			continue
		}

		// 4) 删除行："-" 开头（不推进新文件行号）
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			continue
		}

		// 5) 上下文行：推进新文件行号
		if currentNewLine > 0 {
			currentNewLine++
		}
	}

	return result
}
