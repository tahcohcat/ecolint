package rules

import (
	"ecolint/domain/env"
	"ecolint/domain/issues"
)

func Duplicate(vars []env.Var, file string) []issues.Issue {
	var out []issues.Issue

	seen := make(map[string]issues.Issue) // key -> first line number

	for _, v := range vars {
		if currentIssue, ok := seen[v.Key]; ok {
			seen[v.Key] = issues.NewIssue("duplicate variable", v.Key, file, currentIssue.FirstLine, v.Line, []string{})
		} else {
			seen[v.Key] = issues.NewIssue("duplicate variable", v.Key, file, v.Line, 0, []string{})
		}
	}

	for _, issue := range seen {
		if issue.Line != 0 {
			out = append(out, issue)
		}
	}

	return out
}
