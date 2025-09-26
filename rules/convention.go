package rules

import (
	"github.com/tahcohcat/ecolint/domain/env"
	"github.com/tahcohcat/ecolint/domain/issues"
)

func Convention(vars []env.Var, file string) []issues.Issue {
	return []issues.Issue{}
}
