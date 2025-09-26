package main

import (
	"fmt"
	"log"

	"github.com/tahcohcat/ecolint/lint"
	"github.com/tahcohcat/ecolint/parse"
	"github.com/tahcohcat/ecolint/rules"
)

var (
	files = []string{
		"examples/env/okay.env",
		"examples/env/duplicates.env",
	}
)

func main() {
	issues, err := lint.New(parse.NewEnhanced()).
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
