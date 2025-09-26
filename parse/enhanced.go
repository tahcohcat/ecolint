package parse

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/tahcohcat/ecolint/domain/env"
	"github.com/tahcohcat/ecolint/domain/issues"
)

type EnhancedResult struct {
	IssueList []issues.Issue
	Vars      []env.Var
}

type EnhancedParser struct {
}

func NewEnhanced() *EnhancedParser {
	return &EnhancedParser{}
}

func (e *EnhancedParser) Parse(filename string) ([]env.Var, error) {
	result, err := e.ParseWithIssues(filename)
	if err != nil {
		return nil, err
	}
	return result.Vars, nil
}

func (e *EnhancedParser) ParseWithIssues(filename string) (EnhancedResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return EnhancedResult{}, fmt.Errorf("cannot open .env file: %w", err)
	}
	defer file.Close()

	var vars []env.Var
	var issueList []issues.Issue
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Check for malformed lines (no equals sign)
		if !strings.Contains(trimmedLine, "=") {
			issueList = append(issueList, issues.NewIssue(
				"malformed line",
				trimmedLine,
				filename,
				lineNum,
				lineNum,
				[]string{
					"Each line should be in KEY=VALUE format",
					"Use # for comments",
					"Check for missing equals sign",
				},
			))
			continue
		}

		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) != 2 {
			issueList = append(issueList, issues.NewIssue(
				"malformed line",
				trimmedLine,
				filename,
				lineNum,
				lineNum,
				[]string{
					"Each line should be in KEY=VALUE format",
					"Check for multiple equals signs without proper quoting",
				},
			))
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Validate key format
		if key == "" {
			issueList = append(issueList, issues.NewIssue(
				"empty key",
				line,
				filename,
				lineNum,
				lineNum,
				[]string{
					"Variable names cannot be empty",
					"Use descriptive variable names",
				},
			))
			continue
		}

		// Check for invalid characters in key
		if strings.Contains(key, " ") || strings.Contains(key, "\t") {
			issueList = append(issueList, issues.NewIssue(
				"invalid key format",
				key,
				filename,
				lineNum,
				lineNum,
				[]string{
					"Variable names should not contain spaces or tabs",
					"Use underscores instead of spaces",
					"Follow UPPER_SNAKE_CASE convention",
				},
			))
		}

		// Check for empty values (warning, not error)
		if value == "" {
			issueList = append(issueList, issues.NewIssue(
				"empty value",
				key,
				filename,
				lineNum,
				lineNum,
				[]string{
					"Consider if this variable should have a default value",
					"Use quotes for intentionally empty strings: KEY=\"\"",
					"Document why this value is empty",
				},
			))
		}

		vars = append(vars, env.Var{Key: key, Value: value, Line: lineNum})
	}

	if err := scanner.Err(); err != nil {
		return EnhancedResult{}, fmt.Errorf("error reading file: %w", err)
	}

	return EnhancedResult{
		IssueList: issueList,
		Vars:      vars,
	}, nil
}
