package tools

import (
	"context"
	"fmt"
)

type ParseDiffTool struct{}

func (t *ParseDiffTool) Name() string { return "parse_diff" }

func (t *ParseDiffTool) Run(ctx context.Context, input map[string]any) (map[string]any, error) {
	pathVal, ok := input["diff_path"]
	if !ok {
		return nil, fmt.Errorf("missing diff_path")
	}

	path, ok := pathVal.(string)
	if !ok || path == "" {
		return nil, fmt.Errorf("invalid diff_path")
	}

	parsed, err := ParseDiffFile(path)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"files": parsed.Files,
		"lines": parsed.Lines,
	}, nil
}
