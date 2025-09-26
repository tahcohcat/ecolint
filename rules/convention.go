package rules

import (
	"regexp"
	"strings"

	"github.com/tahcohcat/ecolint/domain/env"
	"github.com/tahcohcat/ecolint/domain/issues"
)

// Convention checks for proper naming conventions in environment variables
// Enforces UPPER_SNAKE_CASE and other best practices
func Convention(vars []env.Var, file string) []issues.Issue {
	var out []issues.Issue

	// Valid environment variable name pattern: UPPER_SNAKE_CASE
	// Must start with letter, contain only letters, numbers, and underscores
	validPattern := regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

	for _, v := range vars {
		var recommendations []string
		issueFound := false

		// Check basic UPPER_SNAKE_CASE pattern
		if !validPattern.MatchString(v.Key) {
			issueFound = true

			// Provide specific recommendations based on the type of issue

			// Check for spaces or tabs
			if strings.Contains(v.Key, " ") || strings.Contains(v.Key, "\t") {
				recommendations = append(recommendations, "Remove spaces and tabs from variable names")
				recommendations = append(recommendations, "Use underscores (_) to separate words")
			}

			// Check for lowercase
			if strings.ToLower(v.Key) == v.Key {
				recommendations = append(recommendations, "Use UPPERCASE for environment variables")
				recommendations = append(recommendations, "Try: "+strings.ToUpper(v.Key))
			}

			// Check for mixed case but not proper UPPER_SNAKE_CASE
			if v.Key != strings.ToUpper(v.Key) && v.Key != strings.ToLower(v.Key) {
				recommendations = append(recommendations, "Use consistent UPPER_SNAKE_CASE")
				recommendations = append(recommendations, "Try: "+strings.ToUpper(v.Key))
			}

			// Check for hyphens (common mistake)
			if strings.Contains(v.Key, "-") {
				recommendations = append(recommendations, "Use underscores (_) instead of hyphens (-)")
				fixed := strings.ReplaceAll(strings.ToUpper(v.Key), "-", "_")
				recommendations = append(recommendations, "Try: "+fixed)
			}

			// Check for camelCase
			if regexp.MustCompile(`[a-z][A-Z]`).MatchString(v.Key) {
				recommendations = append(recommendations, "Convert camelCase to UPPER_SNAKE_CASE")
				converted := convertCamelToSnake(v.Key)
				recommendations = append(recommendations, "Try: "+strings.ToUpper(converted))
			}

			// Check for leading numbers
			if regexp.MustCompile(`^[0-9]`).MatchString(v.Key) {
				recommendations = append(recommendations, "Variable names cannot start with numbers")
				recommendations = append(recommendations, "Prefix with a descriptive word (e.g., ITEM_"+v.Key+")")
			}

			// Check for special characters
			if regexp.MustCompile(`[^A-Za-z0-9_]`).MatchString(v.Key) {
				recommendations = append(recommendations, "Only use letters, numbers, and underscores")
				recommendations = append(recommendations, "Remove or replace special characters")
			}

			// Default recommendations if no specific issues detected
			if len(recommendations) == 0 {
				recommendations = append(recommendations, "Use UPPER_SNAKE_CASE convention (e.g., DATABASE_URL)")
				recommendations = append(recommendations, "Start with a letter, use only letters, numbers, and underscores")
			}
		}

		// Check for overly short names (even if they match the pattern)
		if len(v.Key) == 1 {
			issueFound = true
			recommendations = append(recommendations, "Avoid single-letter variable names")
			recommendations = append(recommendations, "Use descriptive names (e.g., PORT instead of P)")
		}

		// Check for overly long names
		if len(v.Key) > 50 {
			issueFound = true
			recommendations = append(recommendations, "Consider shorter, more concise variable names")
			recommendations = append(recommendations, "Break down complex names into logical parts")
		}

		// Check for reserved keywords or potentially confusing names
		systemVars := []string{
			"PATH", "HOME", "USER", "SHELL", "PWD", "TERM", "LANG", "LC_ALL",
			"TMPDIR", "TMP", "TEMP", "HOSTNAME", "HOSTTYPE", "MACHTYPE",
		}

		for _, reserved := range systemVars {
			if v.Key == reserved {
				issueFound = true
				recommendations = append(recommendations, "Avoid overriding system environment variables")
				recommendations = append(recommendations, "Consider prefixing with your app name (e.g., MYAPP_"+v.Key+")")
				recommendations = append(recommendations, "This could cause unexpected behavior in scripts and tools")
				break
			}
		}

		// Check for common naming anti-patterns
		antiPatterns := map[string]string{
			"CONFIG":   "Be more specific (e.g., DATABASE_CONFIG, APP_CONFIG)",
			"SETTINGS": "Be more specific (e.g., USER_SETTINGS, APP_SETTINGS)",
			"DATA":     "Be more specific (e.g., USER_DATA, CACHE_DATA)",
			"INFO":     "Be more specific (e.g., USER_INFO, DEBUG_INFO)",
			"TEMP":     "Use TMPDIR or TMP_PATH instead",
			"TEST":     "Be more specific (e.g., TEST_DATABASE_URL)",
		}

		if suggestion, isAntiPattern := antiPatterns[v.Key]; isAntiPattern {
			issueFound = true
			recommendations = append(recommendations, "Variable name is too generic")
			recommendations = append(recommendations, suggestion)
		}

		// Check for redundant prefixes/suffixes
		redundantPrefixes := []string{"ENV_", "ENVIRONMENT_", "VAR_", "VARIABLE_"}
		for _, prefix := range redundantPrefixes {
			if strings.HasPrefix(v.Key, prefix) {
				issueFound = true
				recommendations = append(recommendations, "Remove redundant prefix '"+prefix+"'")
				suggestions := strings.TrimPrefix(v.Key, prefix)
				if suggestions != "" {
					recommendations = append(recommendations, "Try: "+suggestions)
				}
				break
			}
		}

		// Suggest improvements for common abbreviations
		abbreviationSuggestions := map[string]string{
			"DB":  "DATABASE",
			"PWD": "PASSWORD",
			"USR": "USER",
			"SVR": "SERVER",
			"CFG": "CONFIG",
			"STG": "STAGING",
			"PRD": "PRODUCTION",
			"DEV": "DEVELOPMENT",
		}

		for abbrev, full := range abbreviationSuggestions {
			if strings.Contains(v.Key, abbrev) && !strings.Contains(v.Key, full) {
				// Only suggest if it's not already part of a longer word
				pattern := regexp.MustCompile(`\b` + abbrev + `\b`)
				if pattern.MatchString(v.Key) {
					if !issueFound {
						issueFound = true
					}
					expanded := strings.ReplaceAll(v.Key, abbrev, full)
					recommendations = append(recommendations, "Consider using full words instead of abbreviations")
					recommendations = append(recommendations, "Try: "+expanded+" (instead of "+abbrev+")")
				}
			}
		}

		// Create issue if any problems were found
		if issueFound {
			out = append(out, issues.NewIssue(
				"naming convention violation",
				v.Key,
				file,
				v.Line,
				v.Line,
				recommendations,
			))
		}
	}

	return out
}

// convertCamelToSnake converts camelCase and PascalCase to snake_case
func convertCamelToSnake(s string) string {
	// Insert underscore before uppercase letters that follow lowercase letters or numbers
	re1 := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	result := re1.ReplaceAllString(s, "${1}_${2}")

	// Handle sequences of uppercase letters (e.g., HTTPSProxy -> HTTPS_Proxy)
	re2 := regexp.MustCompile(`([A-Z])([A-Z][a-z])`)
	result = re2.ReplaceAllString(result, "${1}_${2}")

	return strings.ToLower(result)
}
