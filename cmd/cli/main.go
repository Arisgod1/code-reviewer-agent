package main

import (
	"fmt"
	"os"

	"github.com/Arisgod1/code-reviewer-agent/internal/tools"
	"github.com/spf13/cobra"
)

var diffPath string
var language string

func main() {
	rootCmd := &cobra.Command{
		Use:   "review",
		Short: "CodeReviewer-Agent CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if diffPath == "" {
				return fmt.Errorf("please provide --diff path, e.g. review --diff testdata/sample.patch")
			}

			fmt.Println("== CodeReviewer-Agent ==")
			fmt.Println("diff path:", diffPath)
			fmt.Println("language :", language)
			fmt.Println("next step: parse diff -> scan rules -> format report")
			parsed, err := tools.ParseDiffFile(diffPath)
			if err != nil {
				return err
			}
			fmt.Println("parsed files:", parsed.Files)
			fmt.Println("changed lines:", len(parsed.Lines))

			for i, l := range parsed.Lines {
				if i >= 5 {
					break
				}
				fmt.Printf("line[%d] file=%s:%d content=%s\n", i, l.File, l.Line, l.Content)
			}
			findings := tools.ScanRiskRules(parsed, language)
			fmt.Println("findings:", len(findings))
			for i, f := range findings {
				fmt.Printf("[%d] %s %s:%d severity=%s\n", i, f.ID, f.File, f.Line, f.Severity)
			}
			return nil
		},
	}

	rootCmd.Flags().StringVar(&diffPath, "diff", "", "path to patch/diff file")
	rootCmd.Flags().StringVar(&language, "lang", "go", "language: go/java")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
