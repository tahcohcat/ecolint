package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tahcohcat/ecolint/internal/config"
	"github.com/tahcohcat/ecolint/internal/output"
	"github.com/tahcohcat/ecolint/internal/scan"
	"github.com/tahcohcat/ecolint/lint"
	"github.com/tahcohcat/ecolint/parse"
	"github.com/tahcohcat/ecolint/rules"
)

var lintCmd = &cobra.Command{
	Use:   "lint [files...]",
	Short: "🔍 Lint environment files",
	Long: `🔍 Lint environment files for common issues

This command checks your .env files for:
• Duplicate variable definitions
• Missing required variables (from config or auto-discovered)
• Syntax errors
• Empty values
• Security issues (potential secrets)
• Naming conventions

Auto-Discovery Mode:
When --auto-discover is used, ecolint will scan your project files to
automatically find environment variables that your code uses, eliminating
the need to manually configure required variables.

Examples:
  ecolint lint                        # lint .env in current directory
  ecolint lint .env .env.local        # lint specific files
  ecolint lint --recursive .          # recursively find and lint all .env files
  ecolint lint --auto-discover        # auto-discover required variables
  ecolint lint --auto-discover --scan-path ./src  # scan specific directory
  ecolint lint --format json          # output in JSON format`,
	RunE: runLint,
}

var (
	recursiveFlag     bool
	formatFlag        string
	quietFlag         bool
	configFlag        string
	autoDiscoverFlag  bool
	scanPathFlag      string
	minConfidenceFlag float64
	minUsagesFlag     int
)

func init() {
	rootCmd.AddCommand(lintCmd)

	lintCmd.Flags().BoolVarP(&recursiveFlag, "recursive", "r", false, "recursively search for .env files")
	lintCmd.Flags().StringVarP(&formatFlag, "format", "f", "", "output format (pretty, json, github)")
	lintCmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress output when no issues found")
	lintCmd.Flags().StringVarP(&configFlag, "config", "c", "", "path to configuration file")
	lintCmd.Flags().BoolVar(&autoDiscoverFlag, "auto-discover", false, "automatically discover required variables by scanning project")
	lintCmd.Flags().StringVar(&scanPathFlag, "scan-path", ".", "path to scan for auto-discovery (default: current directory)")
	lintCmd.Flags().Float64Var(&minConfidenceFlag, "min-confidence", 0.7, "minimum confidence for auto-discovered variables (0.0-1.0)")
	lintCmd.Flags().IntVar(&minUsagesFlag, "min-usages", 1, "minimum usages for auto-discovered variables")
}

func runLint(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg := config.Load(configFlag)

	// Override format from command line if provided
	if formatFlag != "" {
		cfg.Output.Format = formatFlag
	}

	// Auto-discover required variables if requested
	if autoDiscoverFlag {
		discoveredVars, err := autoDiscoverRequiredVars()
		if err != nil {
			return fmt.Errorf("auto-discovery failed: %w", err)
		}

		if !quietFlag && len(discoveredVars) > 0 {
			fmt.Printf("🔍 Auto-discovered %d required variables from project scan\n", len(discoveredVars))
		}

		// Merge with configured required vars (auto-discovered takes precedence)
		cfg.RequiredVars = mergeLists(cfg.RequiredVars, discoveredVars)
	}

	// Determine files to lint
	files, err := getFilesToLint(args, recursiveFlag)
	if err != nil {
		return fmt.Errorf("error finding files: %w", err)
	}

	if len(files) == 0 {
		if !quietFlag {
			fmt.Println("🤷 No .env files found to lint")
		}
		return nil
	}

	// Create linter with appropriate rules
	linter := lint.New(parse.NewEnhanced())

	// Add rules based on configuration
	if cfg.Rules.Duplicate {
		linter.WithRule(rules.Duplicate)
	}
	if cfg.Rules.Missing && len(cfg.RequiredVars) > 0 {
		linter.WithRule(rules.Missing(cfg.RequiredVars))
	}
	if cfg.Rules.Security {
		linter.WithRule(rules.Security)
	}
	if cfg.Rules.Convention {
		linter.WithRule(rules.Convention)
	}

	// Run linting
	issues, err := linter.Lint(files)
	if err != nil {
		return fmt.Errorf("linting failed: %w", err)
	}

	// Format and print results
	formatter := output.NewFormatter(cfg.Output.Format, quietFlag)
	formatter.PrintResults(issues, files)

	// Exit with error code if issues found
	if len(issues) > 0 {
		os.Exit(1)
	}

	return nil
}

func autoDiscoverRequiredVars() ([]string, error) {
	// Create scanner
	scanner := scan.NewProjectScanner()

	// Perform scan
	result, err := scanner.ScanProject(scanPathFlag)
	if err != nil {
		return nil, err
	}

	// Get required variables based on confidence and usage thresholds
	required := result.GetRequiredVariables(minConfidenceFlag, minUsagesFlag)

	return required, nil
}

func mergeLists(existing, discovered []string) []string {
	// Create a map to track unique variables
	unique := make(map[string]bool)
	var merged []string

	// Add discovered variables first (they take precedence)
	for _, v := range discovered {
		if !unique[v] {
			unique[v] = true
			merged = append(merged, v)
		}
	}

	// Add existing configured variables if not already present
	for _, v := range existing {
		if !unique[v] {
			unique[v] = true
			merged = append(merged, v)
		}
	}

	return merged
}

func getFilesToLint(args []string, recursive bool) ([]string, error) {
	var files []string

	if len(args) == 0 {
		// No arguments provided, look for default .env file or search recursively
		if recursive {
			return findEnvFilesRecursively(".")
		} else {
			// Look for common .env file names
			candidates := []string{".env", ".env.local", ".env.development", ".env.production"}
			for _, candidate := range candidates {
				if _, err := os.Stat(candidate); err == nil {
					files = append(files, candidate)
				}
			}
			return files, nil
		}
	}

	// Process provided arguments
	for _, arg := range args {
		if recursive {
			found, err := findEnvFilesRecursively(arg)
			if err != nil {
				return nil, err
			}
			files = append(files, found...)
		} else {
			// Check if it's a file or directory
			info, err := os.Stat(arg)
			if err != nil {
				return nil, fmt.Errorf("cannot access %s: %w", arg, err)
			}

			if info.IsDir() {
				return nil, fmt.Errorf("%s is a directory (use --recursive to search directories)", arg)
			}

			files = append(files, arg)
		}
	}

	return files, nil
}

func findEnvFilesRecursively(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip hidden directories and common build/dependency directories
			name := info.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			if name == "node_modules" || name == "vendor" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file matches .env patterns
		filename := info.Name()
		if isEnvFile(filename) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func isEnvFile(filename string) bool {
	// Match various .env file patterns
	envPatterns := []string{
		".env",
		".env.local",
		".env.development",
		".env.production",
		".env.staging",
		".env.test",
	}

	for _, pattern := range envPatterns {
		if filename == pattern {
			return true
		}
	}

	// Match .env.* pattern
	return strings.HasPrefix(filename, ".env.")
}
