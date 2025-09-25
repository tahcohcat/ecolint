package rules

import (
	"ecolint/domain/env"
	"ecolint/domain/issues"
)

type Rule func(vars []env.Var, file string) []issues.Issue
