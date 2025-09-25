package lint

import (
	"ecolint/domain/issues"
	"ecolint/parse"
	"ecolint/rules"
)

type Linter struct {
	rules  []rules.Rule
	parser *parse.Parser
}

func New(p *parse.Parser) *Linter {
	return &Linter{
		rules:  make([]rules.Rule, 0),
		parser: p,
	}
}

func (l *Linter) WithRule(rule rules.Rule) *Linter {
	l.rules = append(l.rules, rule)
	return l
}

func (l *Linter) Lint(files []string) ([]issues.Issue, error) {

	var out []issues.Issue
	for _, file := range files {
		vars, err := l.parser.Parse(file)
		if err != nil {
			return nil, err
		}

		for _, rule := range l.rules {
			out = append(out, rule(vars, file)...)
		}
	}

	return out, nil
}
