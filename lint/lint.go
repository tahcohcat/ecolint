package lint

import (
	"github.com/tahcohcat/ecolint/domain/env"
	"github.com/tahcohcat/ecolint/domain/issues"
	"github.com/tahcohcat/ecolint/parse"
	"github.com/tahcohcat/ecolint/rules"
)

// Linter provides better error handling and parsing issues integration
type Linter struct {
	rules              []rules.Rule
	parser             *parse.EnhancedParser
	includeParseIssues bool
}

func New(p *parse.EnhancedParser) *Linter {
	return &Linter{
		rules:              make([]rules.Rule, 0),
		parser:             p,
		includeParseIssues: true,
	}
}

func (l *Linter) WithRule(rule rules.Rule) *Linter {
	l.rules = append(l.rules, rule)
	return l
}

func (l *Linter) WithParseIssues(include bool) *Linter {
	l.includeParseIssues = include
	return l
}

func (l *Linter) Lint(files []string) ([]issues.Issue, error) {
	var allIssues []issues.Issue

	for _, file := range files {
		// Parse with detailed error reporting
		result, err := l.parser.ParseWithIssues(file)
		if err != nil {
			return nil, err
		}

		// Include parsing issues if enabled
		if l.includeParseIssues {
			allIssues = append(allIssues, result.IssueList...)
		}

		// Apply rules to successfully parsed variables
		for _, rule := range l.rules {
			ruleIssues := rule(result.Vars, file)
			allIssues = append(allIssues, ruleIssues...)
		}
	}

	return allIssues, nil
}

// LintSingle lints a single file and returns detailed results
func (l *Linter) LintSingle(file string) (Result, error) {
	result, err := l.parser.ParseWithIssues(file)
	if err != nil {
		return Result{}, err
	}

	var ruleIssues []issues.Issue
	for _, rule := range l.rules {
		ruleIssues = append(ruleIssues, rule(result.Vars, file)...)
	}

	return Result{
		File:        file,
		Vars:        result.Vars,
		ParseIssues: result.IssueList,
		RuleIssues:  ruleIssues,
		TotalIssues: len(result.IssueList) + len(ruleIssues),
	}, nil
}

type Result struct {
	File        string
	Vars        []env.Var
	ParseIssues []issues.Issue
	RuleIssues  []issues.Issue
	TotalIssues int
}

func (lr Result) AllIssues() []issues.Issue {
	all := make([]issues.Issue, 0, lr.TotalIssues)
	all = append(all, lr.ParseIssues...)
	all = append(all, lr.RuleIssues...)
	return all
}
