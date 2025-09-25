package cmd

import (
	"fmt"
	"os"

	"github.com/tahcohcat/ecolint/internal/config"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "🚀 Initialize ecolint configuration",
	Long: `🚀 Initialize ecolint configuration

This command creates a sample .ecolint.yaml configuration file
in your current directory with common settings and examples.

The configuration file allows you to:
• Define required environment variables
• Enable/disable specific rules
• Configure output formatting
• Set up custom linting rules`,
	RunE: runInit,
}

var (
	forceFlag bool
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "overwrite existing configuration file")
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := ".ecolint.yaml"

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil && !forceFlag {
		return fmt.Errorf("configuration file already exists at %s (use --force to overwrite)", configPath)
	}

	// Create the sample config
	if err := config.CreateSampleConfig(configPath); err != nil {
		return fmt.Errorf("failed to create configuration file: %w", err)
	}

	fmt.Printf("🌱 Successfully created %s\n", configPath)
	fmt.Println("\n📝 Configuration file created with the following features:")
	fmt.Println("  • Required variables validation")
	fmt.Println("  • Duplicate detection")
	fmt.Println("  • Syntax checking")
	fmt.Println("  • Empty value warnings")
	fmt.Println("  • Colored output")
	fmt.Println("\n✨ Edit the file to customize rules for your project!")

	return nil
}
