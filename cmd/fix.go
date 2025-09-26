package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/tahcohcat/ecolint/internal/config"
)

var fixCmd = &cobra.Command{
	Use:   "fix [files...]",
	Short: "🔧 Auto-fix common issues in environment files",
	Long: `🔧 Auto-fix common issues in environment files

This command automatically fixes common issues that can be safely corrected:
• Convert variable names to UPPER_SNAKE_CASE
• Remove leading/trailing whitespace from values
• Remove duplicate variables (keeps the last occurrence)
• Quote values that contain spaces or special characters
• Fix malformed lines where possible

Examples:
  ecolint fix .env                    # fix .env file
  ecolint fix .env .env.local         # fix multiple files
  ecolint fix --dry-run .env          # preview changes without applying
  ecolint fix --backup .env           # create backup before fixing`,
	RunE: runFix,
}

var (
	dryRunFlag bool
	backupFlag bool
	fixAllFlag bool
)

func init() {
	rootCmd.AddCommand(fixCmd)

	fixCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "preview changes without applying them")
	fixCmd.Flags().BoolVar(&backupFlag, "backup", false, "create backup files before fixing")
	fixCmd.Flags().BoolVar(&fixAllFlag, "all", false, "fix all issues (including potentially unsafe ones)")
}

func runFix(cmd *cobra.Command, args []string) error {
	// Load configuration for determining which fixes to apply
	cfg := config.Load(configFlag)

	// Determine files to fix
	files, err := getFilesToLint(args, recursiveFlag)
	if err != nil {
		return fmt.Errorf("error finding files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("🤷 No .env files found to fix")
		return nil
	}

	totalFixed := 0
	for _, file := range files {
		fixed, err := fixFile(file, cfg)
		if err != nil {
			fmt.Printf("❌ Error fixing %s: %v\n", file, err)
			continue
		}
		totalFixed += fixed
	}

	if dryRunFlag {
		fmt.Printf("\n🔍 Dry run complete. Found %d fixable issues across %d files\n", totalFixed, len(files))
		fmt.Println("💡 Run without --dry-run to apply fixes")
	} else {
		fmt.Printf("\n✅ Fixed %d issues across %d files\n", totalFixed, len(files))
	}

	return nil
}

type FixResult struct {
	OriginalLine string
	FixedLine    string
	LineNumber   int
	Issue        string
}

func fixFile(filename string, cfg config.Config) (int, error) {
	// Read the original file
	file, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	var originalLines []string
	var fixes []FixResult
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		originalLines = append(originalLines, line)

		// Skip empty lines and comments
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Try to fix the line
		if fixed, issue := fixLine(line, cfg); fixed != line {
			fixes = append(fixes, FixResult{
				OriginalLine: line,
				FixedLine:    fixed,
				LineNumber:   lineNum,
				Issue:        issue,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading file: %w", err)
	}

	// Remove duplicates (keep last occurrence)
	if cfg.Rules.Duplicate {
		duplicateFixes := removeDuplicates(originalLines)
		fixes = append(fixes, duplicateFixes...)
	}

	if len(fixes) == 0 {
		if !dryRunFlag {
			fmt.Printf("✅ %s: no issues to fix\n", filename)
		}
		return 0, nil
	}

	// Apply fixes
	fixedLines := make([]string, len(originalLines))
	copy(fixedLines, originalLines)

	for _, fix := range fixes {
		if fix.LineNumber > 0 && fix.LineNumber <= len(fixedLines) {
			fixedLines[fix.LineNumber-1] = fix.FixedLine
		}
	}

	// Print what we're doing
	if dryRunFlag {
		fmt.Printf("🔍 %s (%d fixes would be applied):\n", filename, len(fixes))
		for _, fix := range fixes {
			fmt.Printf("  Line %d: %s\n", fix.LineNumber, fix.Issue)
			fmt.Printf("    - %s\n", fix.OriginalLine)
			fmt.Printf("    + %s\n", fix.FixedLine)
		}
	} else {
		// Create backup if requested
		if backupFlag {
			backupPath := filename + ".backup"
			if err := copyFile(filename, backupPath); err != nil {
				return 0, fmt.Errorf("failed to create backup: %w", err)
			}
			fmt.Printf("📋 Created backup: %s\n", backupPath)
		}

		// Write fixed content
		content := strings.Join(fixedLines, "\n")
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return 0, fmt.Errorf("failed to write fixed file: %w", err)
		}

		fmt.Printf("🔧 %s: fixed %d issues\n", filename, len(fixes))
		for _, fix := range fixes {
			fmt.Printf("  ✓ Line %d: %s\n", fix.LineNumber, fix.Issue)
		}
	}

	return len(fixes), nil
}

func fixLine(line string, cfg config.Config) (string, string) {
	trimmedLine := strings.TrimSpace(line)

	// Skip empty lines and comments
	if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
		return line, ""
	}

	// Check if line has equals sign
	if !strings.Contains(trimmedLine, "=") {
		// Can't fix malformed lines safely
		return line, ""
	}

	parts := strings.SplitN(trimmedLine, "=", 2)
	if len(parts) != 2 {
		return line, ""
	}

	originalKey := strings.TrimSpace(parts[0])
	originalValue := strings.TrimSpace(parts[1])

	if originalKey == "" {
		// Can't fix empty keys safely
		return line, ""
	}

	fixed := false
	issues := []string{}

	// Fix key naming convention
	fixedKey := originalKey
	if cfg.Rules.Convention {
		newKey := fixKeyConvention(originalKey)
		if newKey != originalKey {
			fixedKey = newKey
			fixed = true
			issues = append(issues, "fixed naming convention")
		}
	}

	// Fix value issues
	fixedValue := originalValue

	// Remove leading/trailing whitespace (this is already done by TrimSpace above)
	// but we should preserve the original spacing in the file
	actualValue := parts[1] // Get original value with spacing
	if actualValue != strings.TrimSpace(actualValue) {
		fixedValue = strings.TrimSpace(actualValue)
		fixed = true
		issues = append(issues, "removed leading/trailing whitespace")
	}

	// Quote values that need quoting
	if needsQuoting(fixedValue) && !isQuoted(fixedValue) {
		fixedValue = fmt.Sprintf("\"%s\"", fixedValue)
		fixed = true
		issues = append(issues, "added quotes")
	}

	if !fixed {
		return line, ""
	}

	fixedLine := fmt.Sprintf("%s=%s", fixedKey, fixedValue)
	return fixedLine, strings.Join(issues, ", ")
}

func fixKeyConvention(key string) string {
	// Remove redundant prefixes
	redundantPrefixes := []string{"ENV_", "ENVIRONMENT_", "CONFIG_", "CONF_", "SETTING_", "SETTINGS_"}
	for _, prefix := range redundantPrefixes {
		if strings.HasPrefix(key, prefix) {
			key = strings.TrimPrefix(key, prefix)
			break
		}
	}

	// Convert to UPPER_SNAKE_CASE
	var result strings.Builder
	for i, r := range key {
		if unicode.IsLetter(r) {
			result.WriteRune(unicode.ToUpper(r))
		} else if unicode.IsDigit(r) {
			result.WriteRune(r)
		} else if r == '-' || r == '.' || r == ' ' {
			// Convert hyphens, dots, and spaces to underscores
			result.WriteRune('_')
		} else if r == '_' {
			// Keep underscores, but avoid consecutive ones
			if i == 0 || result.Len() == 0 || result.String()[result.Len()-1] != '_' {
				result.WriteRune('_')
			}
		}
		// Skip other invalid characters
	}

	fixed := result.String()

	// Remove leading/trailing underscores
	fixed = strings.Trim(fixed, "_")

	// Replace multiple consecutive underscores with single ones
	for strings.Contains(fixed, "__") {
		fixed = strings.ReplaceAll(fixed, "__", "_")
	}

	// Ensure we don't return empty string
	if fixed == "" {
		return key
	}

	return fixed
}

func removeDuplicates(lines []string) []FixResult {
	var fixes []FixResult
	seen := make(map[string]int) // key -> last line number

	// First pass: find all variables and their last occurrence
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		if !strings.Contains(trimmedLine, "=") {
			continue
		}

		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		seen[key] = i + 1 // Store 1-based line number
	}

	// Second pass: mark earlier occurrences for removal
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		if !strings.Contains(trimmedLine, "=") {
			continue
		}

		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}

		lineNum := i + 1
		lastOccurrence := seen[key]

		// If this is not the last occurrence, mark for removal
		if lineNum != lastOccurrence {
			fixes = append(fixes, FixResult{
				OriginalLine: line,
				FixedLine:    "", // Remove the line
				LineNumber:   lineNum,
				Issue:        fmt.Sprintf("removed duplicate variable '%s' (kept line %d)", key, lastOccurrence),
			})
		}
	}

	return fixes
}

func needsQuoting(value string) bool {
	if value == "" {
		return false
	}

	// Values that need quoting
	return strings.Contains(value, " ") ||
		strings.HasPrefix(value, "#") ||
		strings.Contains(value, "$") ||
		strings.Contains(value, "`") ||
		strings.Contains(value, "\"") ||
		strings.Contains(value, "'")
}

func isQuoted(value string) bool {
	if len(value) < 2 {
		return false
	}
	return (value[0] == '"' && value[len(value)-1] == '"') ||
		(value[0] == '\'' && value[len(value)-1] == '\'')
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}
