package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Arisgod1/code-reviewer-agent/internal/agent"
	"github.com/Arisgod1/code-reviewer-agent/internal/config"
	"github.com/Arisgod1/code-reviewer-agent/internal/llm"
	"github.com/Arisgod1/code-reviewer-agent/internal/report"
	"github.com/Arisgod1/code-reviewer-agent/internal/tools"
	"github.com/spf13/cobra"
)

var diffPath string
var language string
var outputPath string
var mdOutputPath string
var useLLMPlanner bool

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
			appCfg := config.Load()
			var llmPlanner *agent.LLMPlanner
			if useLLMPlanner {
				client := llm.NewClient(llm.Config{
					BaseURL: appCfg.LLM.BaseURL,
					APIKey:  appCfg.LLM.APIKey,
					Model:   appCfg.LLM.Model,
					Timeout: appCfg.LLM.Timeout,
					Retries: appCfg.LLM.Retries,
				})
				llmPlanner = &agent.LLMPlanner{Client: client}

				keyPresent := "no"
				if appCfg.LLM.APIKey != "" {
					keyPresent = "yes"
				}

				fmt.Println("planner  : LLM enabled")
				fmt.Println("llm base :", appCfg.LLM.BaseURL)
				fmt.Println("llm model:", appCfg.LLM.Model)
				fmt.Println("llm key? :", keyPresent)
			} else {
				fmt.Println("planner  : rule-based")
			}

			reg := tools.NewRegistry()
			reg.Register(&tools.ParseDiffTool{})
			reg.Register(&tools.ScanRulesTool{})
			reg.Register(&tools.FormatReportTool{})

			startAll := time.Now()
			fmt.Println("== CodeReviewer-Agent ==")
			fmt.Println("diff path:", diffPath)
			fmt.Println("language :", language)

			fmt.Println("next step: parse diff -> scan rules -> format report")

			loopRes, err := agent.RunReviewLoop(ctx, reg, diffPath, language, 6, llmPlanner)
			if err != nil {
				return err
			}

			reviewReport := loopRes.Report
			trace := loopRes.Trace
			toolCalls := loopRes.ToolCalls
			findings := reviewReport.Findings
			plannerUsedLLM := false
			plannerFallbackReasons := make([]string, 0)
			for _, step := range trace {
				if step.Action == "plan" && step.Observation != "" {
					if strings.Contains(step.Observation, "source=llm") {
						plannerUsedLLM = true
					}
					if strings.Contains(step.Observation, "source=fallback") {
						plannerFallbackReasons = append(plannerFallbackReasons, step.Observation)
					}
				}
			}

			fmt.Println("findings:", len(findings))
			for i, f := range findings {
				fmt.Printf("[%d] %s %s:%d severity=%s\n", i, f.ID, f.File, f.Line, f.Severity)
			}
			fmt.Println("planner used llm:", plannerUsedLLM)
			if len(plannerFallbackReasons) > 0 {
				fmt.Println("planner fallback observations:")
				for _, obs := range plannerFallbackReasons {
					fmt.Println("-", obs)
				}
			}
			reviewReport.Trace = trace
			reviewReport.Metrics["total_cost_ms"] = time.Since(startAll).Milliseconds()
			reviewReport.Metrics["trace_steps"] = len(trace)
			reviewReport.Metrics["tool_timeout_ms"] = 2000
			reviewReport.Metrics["tool_calls_count"] = len(toolCalls)
			reviewReport.Metrics["tool_calls"] = toolCalls
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
	rootCmd.Flags().BoolVar(&useLLMPlanner, "use-llm-planner", false, "enable LLM-based planning")
	rootCmd.Flags().StringVar(&outputPath, "out", "report.json", "output json report path")
	rootCmd.Flags().StringVar(&mdOutputPath, "mdout", "report.md", "output markdown report path")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
