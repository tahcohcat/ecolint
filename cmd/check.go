package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tahcohcat/ecolint/parse"
)

var checkCmd = &cobra.Command{
	Use:   "check [files...]",
	Short: "✅ Quick syntax check for environment files",
	Long: `✅ Quick syntax check for environment files

This command performs a basic syntax validation of .env files
without running the full linting rules. Useful for quick validation
in CI/CD pipelines.

Examples:
  ecolint check .env
  ecolint check .env .env.local .env.production`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("please specify at least one file to check")
	}

	parser := parse.NewEnhanced()
	hasErrors := false

	for _, file := range args {
		// Check if file exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("❌ %s: file not found\n", file)
			hasErrors = true
			continue
		}

		// Try to parse the file
		_, err := parser.Parse(file)
		if err != nil {
			fmt.Printf("❌ %s: %v\n", file, err)
			hasErrors = true
		} else {
			fmt.Printf("✅ %s: syntax OK\n", file)
		}
	}

	if hasErrors {
		os.Exit(1)
	}

	return nil
}
