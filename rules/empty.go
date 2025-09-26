package rules

import (
	"github.com/tahcohcat/ecolint/domain/env"
	"github.com/tahcohcat/ecolint/domain/issues"
)

func EmptyValues(vars []env.Var, file string) []issues.Issue {
	var out []issues.Issue

	for _, v := range vars {
		if v.Value != "" {
			continue
		}

		out = append(out, issues.NewIssue(
			"empty variable",
			v.Key,
			file,
			v.Line,
			0,
			[]string{
				"Add the variable value to your .env file",
			},
		))
	}

	return out
}
