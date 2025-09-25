package rules

import (
	"github.com/tahcohcat/ecolint/domain/env"
	"github.com/tahcohcat/ecolint/domain/issues"
)

type Rule func(vars []env.Var, file string) []issues.Issue
