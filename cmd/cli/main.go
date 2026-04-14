package main

import (
	"fmt"
	"os"

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
