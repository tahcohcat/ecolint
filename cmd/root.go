package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ecolint",
	Short: "🌱 Ecolint - A linter for environment files",
	Long: `🌱 Ecolint checks your environment configuration (.env files)
for common issues such as duplicates, missing variables, empty values,
and consistency across your project.

Examples:
  ecolint init       # create a sample configuration
  ecolint lint .env  # run checks on your .env file
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommand is provided
		fmt.Println("🌱 Ecolint - use `ecolint --help` to see available commands")
	},
}

// Execute runs the root command.
// This is what `main.go` calls.
func Execute() error {
	return rootCmd.Execute()
}
