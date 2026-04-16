package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Arisgod1/code-reviewer-agent/internal/agent"
	"github.com/Arisgod1/code-reviewer-agent/internal/report"
	"github.com/Arisgod1/code-reviewer-agent/internal/tools"
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
			fmt.Println("== CodeReviewer-Agent ==")
			fmt.Println("diff path:", diffPath)
			fmt.Println("language :", language)

			fmt.Println("next step: parse diff -> scan rules -> format report")

			loopRes, err := agent.RunReviewLoop(ctx, reg, diffPath, language, 4)
			if err != nil {
				return err
			}

			findings := loopRes.Findings
			trace := loopRes.Trace
			toolCalls := loopRes.ToolCalls

			fmt.Println("findings:", len(findings))
			for i, f := range findings {
				fmt.Printf("[%d] %s %s:%d severity=%s\n", i, f.ID, f.File, f.Line, f.Severity)
			}
			reviewReport := tools.BuildReport(findings)
			reviewReport.Trace = trace
			reviewReport.Metrics["total_costƒ_ms"] = time.Since(startAll).Milliseconds()
			reviewReport.Metrics["trace_steps"] = len(trace)
			reviewReport.Metrics["tool_timeout_ms"] = 2000
			if err := tools.WriteReportJSON(reviewReport, outputPath); err != nil {
				return err
			}
			fmt.Println("report written to:", outputPath)
			if err := report.WriteMarkdownReport(reviewReport, mdOutputPath); err != nil {
				return err
			}
			fmt.Println("markdown report written to:", mdOutputPath)
			fmt.Println("tool calls:")
			for _, tc := range toolCalls {
				fmt.Printf("- %s cost=%dms success=%v\n", tc.Name, tc.CostMs, tc.Success)
			}
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
