package rules

import (
	"github.com/tahcohcat/ecolint/domain/env"
	"github.com/tahcohcat/ecolint/domain/issues"
)

// Missing checks for required variables that are not defined
func Missing(requiredVars []string) Rule {
	return func(vars []env.Var, file string) []issues.Issue {
		var out []issues.Issue

		// Create a set of existing variable names for quick lookup
		existing := make(map[string]bool)
		for _, v := range vars {
			existing[v.Key] = true
		}

		// Check each required variable
		for _, required := range requiredVars {
			if !existing[required] {
				out = append(out, issues.NewIssue(
					"missing required variable",
					required,
					file,
					0, // No line number since it's missing
					0,
					[]string{
						"Add the missing variable to your .env file",
						"Check your configuration for required variables",
						"Ensure the variable name is spelled correctly",
						"Consider if this should be optional instead",
					},
				))
			}
		}

		return out
	}
}
