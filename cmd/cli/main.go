package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Arisgod1/code-reviewer-agent/internal/report"
	"github.com/Arisgod1/code-reviewer-agent/internal/tools"
	"github.com/Arisgod1/code-reviewer-agent/internal/types"
	"github.com/spf13/cobra"
)

var diffPath string
var language string
var outputPath string
var mdOutputPath string

func main() {
	rootCmd := &cobra.Command{
		Use:   "review",
		Short: "CodeReviewer-Agent CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if diffPath == "" {
				return fmt.Errorf("please provide --diff path, e.g. review --diff testdata/sample.patch")
			}
			//添加上下文
			ctx := cmd.Context()

			reg := tools.NewRegistry()
			reg.Register(&tools.ParseDiffTool{})
			reg.Register(&tools.ScanRulesTool{})

			startAll := time.Now()
			trace := make([]types.TraceStep, 0)
			step := 1
			addTrace := func(thought, action, observation string) {
				trace = append(trace, types.TraceStep{
					Step:        step,
					Thought:     thought,
					Action:      action,
					Observation: observation,
					Timestamp:   time.Now().Format(time.RFC3339),
				})
				step++
			}
			fmt.Println("== CodeReviewer-Agent ==")
			fmt.Println("diff path:", diffPath)
			fmt.Println("language :", language)
			addTrace(
				"Need in ingest user input",
				"read CLI flags",
				"diff="+diffPath+",lang="+language,
			)

			fmt.Println("next step: parse diff -> scan rules -> format report")

			parseTool, err := reg.Get("parse_diff")
			if err != nil {
				return err
			}
			parseOut, err := parseTool.Run(ctx, map[string]any{
				"diff_path": diffPath,
			})
			if err != nil {
				return err
			}

			files, _ := parseOut["files"].([]string)
			lines, ok := parseOut["lines"].([]tools.ChangedLine)
			if !ok {
				return fmt.Errorf("parse_diff output lines type invalid")
			}

			parsed := tools.ParseDiffResult{
				Files: files,
				Lines: lines,
			}
			addTrace(
				"Need changed files and lines",
				"call tool:parse_diff",
				fmt.Sprintf("files=%d, changed_lines=%d", len(parsed.Files), len(parsed.Lines)),
			)
			fmt.Println("parsed files:", parsed.Files)
			fmt.Println("changed lines:", len(parsed.Lines))

			for i, l := range parsed.Lines {
				if i >= 5 {
					break
				}
				fmt.Printf("line[%d] file=%s:%d content=%s\n", i, l.File, l.Line, l.Content)
			}
			scanTool, err := reg.Get("scan_risk_rules")
			if err != nil {
				return err
			}
			scanOut, err := scanTool.Run(ctx, map[string]any{
				"lines":    parsed.Lines,
				"language": language,
			})
			if err != nil {
				return err
			}

			findings, ok := scanOut["findings"].([]types.Finding)
			if !ok {
				return fmt.Errorf("scan_risk_rules output findings type invalid")
			}
			addTrace(
				"Need risk findings from changed lines",
				"call tool: scan_risk_rules",
				fmt.Sprintf("findings=%d", len(findings)),
			)
			fmt.Println("findings:", len(findings))
			for i, f := range findings {
				fmt.Printf("[%d] %s %s:%d severity=%s\n", i, f.ID, f.File, f.Line, f.Severity)
			}
			addTrace(
				"Need final structured outputs",
				"call tool: format_report",
				"json="+outputPath+", markdown="+mdOutputPath,
			)
			reviewReport := tools.BuildReport(findings)
			reviewReport.Trace = trace
			reviewReport.Metrics["total_costƒ_ms"] = time.Since(startAll).Milliseconds()
			reviewReport.Metrics["trace_steps"] = len(trace)
			if err := tools.WriteReportJSON(reviewReport, outputPath); err != nil {
				return err
			}
			fmt.Println("report written to:", outputPath)
			if err := report.WriteMarkdownReport(reviewReport, mdOutputPath); err != nil {
				return err
			}
			fmt.Println("markdown report written to:", mdOutputPath)
			return nil
		},
	}

	rootCmd.Flags().StringVar(&diffPath, "diff", "", "path to patch/diff file")
	rootCmd.Flags().StringVar(&language, "lang", "go", "language: go/java")
	rootCmd.Flags().StringVar(&outputPath, "out", "report.json", "output json report path")
	rootCmd.Flags().StringVar(&mdOutputPath, "mdout", "report.md", "output markdown report path")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
