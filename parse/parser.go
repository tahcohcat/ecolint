package parse

import (
	"bufio"
	"ecolint/domain/env"
	"fmt"
	"os"
	"strings"
)

type Parser struct {
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(filename string) ([]env.Var, error) {
	return p.parseFile(filename)
}

func (p *Parser) parseFile(path string) ([]env.Var, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open .env file: %w", err)
	}
	defer file.Close()

	var vars []env.Var
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // malformed, maybe later warn
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		vars = append(vars, env.Var{Key: key, Value: value, Line: lineNum})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return vars, nil
}
