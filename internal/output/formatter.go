package output

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/tahcohcat/ecolint/domain/issues"
)

// Colors for pretty output
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Green  = "\033[32m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
	Bold   = "\033[1m"
)

type Formatter struct {
	format string
	quiet  bool
	color  bool
}

func NewFormatter(format string, quiet bool) *Formatter {
	return &Formatter{
		format: format,
		quiet:  quiet,
		color:  shouldUseColor(),
	}
}

func shouldUseColor() bool {
	// Check if output is a terminal and color is supported
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	// Add more sophisticated terminal detection if needed
	return true
}

func (f *Formatter) PrintResults(issues []issues.Issue, files []string) {
	switch f.format {
	case "json":
		f.printJSON(issues, files)
	case "github":
		f.printGitHub(issues)
	default:
		f.printPretty(issues, files)
	}
}

func (f *Formatter) printPretty(issueList []issues.Issue, files []string) {
	if len(issueList) == 0 {
		if !f.quiet {
			f.colorPrint(Green, "✅ No issues found! Your environment is squeaky clean!\n")
		}
		return
	}

	// Group issues by file
	fileIssues := make(map[string][]issues.Issue)
	for _, issue := range issueList {
		fileIssues[issue.File] = append(fileIssues[issue.File], issue)
	}

	// Sort files for consistent output
	sortedFiles := make([]string, 0, len(fileIssues))
	for file := range fileIssues {
		sortedFiles = append(sortedFiles, file)
	}
	sort.Strings(sortedFiles)

	// Print header
	f.colorPrint(Bold+Red, "🚨 Issues found:\n\n")

	totalIssues := len(issueList)
	for _, file := range sortedFiles {
		issues := fileIssues[file]

		// File header
		f.colorPrint(Bold+Blue, fmt.Sprintf("📁 %s\n", file))
		f.colorPrint(Gray, strings.Repeat("─", len(file)+4)+"\n")

		// Sort issues by line number
		sort.Slice(issues, func(i, j int) bool {
			return issues[i].FirstLine < issues[j].FirstLine
		})

		for _, issue := range issues {
			f.printIssue(issue)
		}
		fmt.Println()
	}

	// Summary
	f.colorPrint(Bold, fmt.Sprintf("Found %d issue(s) across %d file(s)\n", totalIssues, len(sortedFiles)))
}

func (f *Formatter) printIssue(issue issues.Issue) {
	// Icon based on issue type
	icon := f.getIssueIcon(issue.Name)
	color := f.getIssueColor(issue.Name)

	// Main issue line
	if issue.Line > 0 && issue.FirstLine > 0 && issue.Line != issue.FirstLine {
		f.colorPrint(color, fmt.Sprintf("  %s Line %d-%d: %s '%s'\n",
			icon, issue.FirstLine, issue.Line, issue.Name, issue.Key))
	} else if issue.Line > 0 || issue.FirstLine > 0 {
		lineNum := issue.Line
		if lineNum == 0 {
			lineNum = issue.FirstLine
		}
		f.colorPrint(color, fmt.Sprintf("  %s Line %d: %s '%s'\n",
			icon, lineNum, issue.Name, issue.Key))
	} else {
		f.colorPrint(color, fmt.Sprintf("  %s %s '%s'\n",
			icon, issue.Name, issue.Key))
	}

	// Recommendations
	if len(issue.Recommendations) > 0 {
		for _, rec := range issue.Recommendations {
			f.colorPrint(Gray, fmt.Sprintf("    💡 %s\n", rec))
		}
	}
}

func (f *Formatter) printJSON(issueList []issues.Issue, files []string) {
	output := struct {
		Issues []issues.Issue `json:"issues"`
		Files  []string       `json:"files"`
		Count  int            `json:"count"`
	}{
		Issues: issueList,
		Files:  files,
		Count:  len(issueList),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(output)
}

func (f *Formatter) printGitHub(issues []issues.Issue) {
	// GitHub Actions annotation format
	for _, issue := range issues {
		level := "error"
		if strings.Contains(strings.ToLower(issue.Name), "warning") ||
			strings.Contains(strings.ToLower(issue.Name), "convention") {
			level = "warning"
		}

		line := issue.FirstLine
		if line == 0 {
			line = issue.Line
		}
		if line == 0 {
			line = 1
		}

		fmt.Printf("::%s file=%s,line=%d::%s '%s'\n",
			level, issue.File, line, issue.Name, issue.Key)
	}
}

func (f *Formatter) colorPrint(color, text string) {
	if f.color {
		fmt.Print(color + text + Reset)
	} else {
		fmt.Print(text)
	}
}

func (f *Formatter) getIssueIcon(issueName string) string {
	switch {
	case strings.Contains(strings.ToLower(issueName), "duplicate"):
		return "🔄"
	case strings.Contains(strings.ToLower(issueName), "missing"):
		return "❓"
	case strings.Contains(strings.ToLower(issueName), "empty"):
		return "🗳️"
	case strings.Contains(strings.ToLower(issueName), "syntax") || strings.Contains(strings.ToLower(issueName), "malformed"):
		return "🔧"
	case strings.Contains(strings.ToLower(issueName), "security") || strings.Contains(strings.ToLower(issueName), "secret"):
		return "🔒"
	case strings.Contains(strings.ToLower(issueName), "convention") || strings.Contains(strings.ToLower(issueName), "format"):
		return "📐"
	default:
		return "⚠️"
	}
}

func (f *Formatter) getIssueColor(issueName string) string {
	if !f.color {
		return ""
	}

	switch {
	case strings.Contains(strings.ToLower(issueName), "security") || strings.Contains(strings.ToLower(issueName), "secret"):
		return Bold + Red
	case strings.Contains(strings.ToLower(issueName), "duplicate") || strings.Contains(strings.ToLower(issueName), "missing"):
		return Red
	case strings.Contains(strings.ToLower(issueName), "syntax") || strings.Contains(strings.ToLower(issueName), "malformed"):
		return Yellow
	case strings.Contains(strings.ToLower(issueName), "convention") || strings.Contains(strings.ToLower(issueName), "format"):
		return Blue
	default:
		return Yellow
	}
}
