package main

import (
	"ecolint/lint"
	"ecolint/parse"
	"ecolint/rules"
	"fmt"
	"log"
)

var (
	files = []string{
		"examples/env/okay.env",
		"examples/env/duplicates.env",
	}
)

func main() {
	issues, err := lint.New(parse.NewParser()).
		WithRule(rules.Duplicate).
		Lint(files)

	if err != nil {
		log.Fatal(err)
	}

	if len(issues) == 0 {
		fmt.Println("No issues found.")
		return
	}

	fmt.Println("Issues:")
	for _, issue := range issues {
		fmt.Printf("⚠️ %s\n", issue.String())
	}
}
