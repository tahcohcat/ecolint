package scan

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProjectScanner discovers environment variable usage across a project
type ProjectScanner struct {
	patterns     []VariablePattern
	excludePaths []string
	includeExts  []string
}

type VariablePattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Description string
	Language    string
}

type UsageResult struct {
	Variable   string
	File       string
	Line       int
	Context    string
	Pattern    string
	Confidence float64 // 0.0 - 1.0 confidence this is actually an env var
}

type ScanResult struct {
	Variables map[string][]UsageResult // var name -> usages
	Files     []string                 // files scanned
	Errors    []error                  // scanning errors
}

// NewProjectScanner creates a scanner with common environment variable patterns
func NewProjectScanner() *ProjectScanner {
	scanner := &ProjectScanner{
		patterns: getCommonPatterns(),
		excludePaths: []string{
			"node_modules", "vendor", "dist", "build", ".git",
			"target", "bin", "obj", ".next", ".nuxt",
		},
		includeExts: []string{
			".go", ".js", ".ts", ".jsx", ".tsx", ".py", ".java",
			".rb", ".php", ".cs", ".cpp", ".c", ".rs", ".kt",
			".scala", ".sh", ".bash", ".zsh", ".fish", ".ps1",
			".yml", ".yaml", ".json", ".toml", ".ini", ".conf",
			".dockerfile", "Dockerfile", ".env", ".env.example",
		},
	}
	return scanner
}

// WithCustomPatterns adds custom regex patterns for finding env vars
func (ps *ProjectScanner) WithCustomPatterns(patterns []VariablePattern) *ProjectScanner {
	ps.patterns = append(ps.patterns, patterns...)
	return ps
}

// WithExcludePaths sets directories to skip during scanning
func (ps *ProjectScanner) WithExcludePaths(paths []string) *ProjectScanner {
	ps.excludePaths = paths
	return ps
}

// WithIncludeExtensions sets file extensions to scan
func (ps *ProjectScanner) WithIncludeExtensions(exts []string) *ProjectScanner {
	ps.includeExts = exts
	return ps
}

// ScanProject scans the entire project for environment variable usage
func (ps *ProjectScanner) ScanProject(rootPath string) (*ScanResult, error) {
	result := &ScanResult{
		Variables: make(map[string][]UsageResult),
		Files:     []string{},
		Errors:    []error{},
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error accessing %s: %w", path, err))
			return nil // Continue walking
		}

		// Skip directories we don't want to scan
		if info.IsDir() {
			for _, exclude := range ps.excludePaths {
				if strings.Contains(path, exclude) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Skip files with extensions we don't care about
		if !ps.shouldScanFile(path) {
			return nil
		}

		// Scan the file
		fileResult, err := ps.scanFile(path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("error scanning %s: %w", path, err))
			return nil
		}

		result.Files = append(result.Files, path)

		// Merge results
		for varName, usages := range fileResult.Variables {
			result.Variables[varName] = append(result.Variables[varName], usages...)
		}

		return nil
	})

	return result, err
}

// ScanFile scans a single file for environment variable usage
func (ps *ProjectScanner) scanFile(filePath string) (*ScanResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := &ScanResult{
		Variables: make(map[string][]UsageResult),
		Files:     []string{filePath},
		Errors:    []error{},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Apply all patterns to this line
		for _, pattern := range ps.patterns {
			matches := pattern.Pattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) < 2 {
					continue
				}

				varName := match[1] // First capture group should be the variable name
				confidence := ps.calculateConfidence(varName, line, pattern)

				usage := UsageResult{
					Variable:   varName,
					File:       filePath,
					Line:       lineNum,
					Context:    strings.TrimSpace(line),
					Pattern:    pattern.Name,
					Confidence: confidence,
				}

				result.Variables[varName] = append(result.Variables[varName], usage)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		result.Errors = append(result.Errors, err)
	}

	return result, nil
}

// GetRequiredVariables returns a list of likely required environment variables
// based on confidence scores and usage frequency
func (sr *ScanResult) GetRequiredVariables(minConfidence float64, minUsages int) []string {
	var required []string

	for varName, usages := range sr.Variables {
		if len(usages) < minUsages {
			continue
		}

		// Calculate average confidence
		totalConfidence := 0.0
		for _, usage := range usages {
			totalConfidence += usage.Confidence
		}
		avgConfidence := totalConfidence / float64(len(usages))

		if avgConfidence >= minConfidence {
			required = append(required, varName)
		}
	}

	return required
}

// shouldScanFile determines if a file should be scanned based on extension
func (ps *ProjectScanner) shouldScanFile(path string) bool {
	ext := filepath.Ext(path)
	fileName := filepath.Base(path)

	// Check exact filename matches (like Dockerfile)
	for _, include := range ps.includeExts {
		if fileName == include || ext == include {
			return true
		}
	}

	return false
}

// calculateConfidence returns a confidence score for whether this is actually an env var
func (ps *ProjectScanner) calculateConfidence(varName, line string, pattern VariablePattern) float64 {
	confidence := 0.5 // Base confidence

	// Higher confidence for ALL_CAPS variables
	if strings.ToUpper(varName) == varName && len(varName) > 1 {
		confidence += 0.3
	}

	// Higher confidence for variables with underscores
	if strings.Contains(varName, "_") {
		confidence += 0.2
	}

	// Lower confidence for very short names
	if len(varName) <= 2 {
		confidence -= 0.3
	}

	// Higher confidence for common env var names
	commonEnvVars := []string{
		"PORT", "HOST", "DATABASE_URL", "API_KEY", "SECRET_KEY",
		"NODE_ENV", "ENVIRONMENT", "DEBUG", "LOG_LEVEL", "TIMEOUT",
	}
	for _, common := range commonEnvVars {
		if strings.Contains(strings.ToUpper(varName), common) {
			confidence += 0.2
			break
		}
	}

	// Adjust based on context
	lowerLine := strings.ToLower(line)
	if strings.Contains(lowerLine, "process.env") ||
		strings.Contains(lowerLine, "os.getenv") ||
		strings.Contains(lowerLine, "os.environ") ||
		strings.Contains(lowerLine, "$env:") ||
		strings.Contains(lowerLine, "${") {
		confidence += 0.2
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	// Floor at 0.0
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// getCommonPatterns returns regex patterns for common environment variable usage patterns
func getCommonPatterns() []VariablePattern {
	return []VariablePattern{
		{
			Name:        "Go os.Getenv",
			Pattern:     regexp.MustCompile(`os\.Getenv\(["']([A-Z][A-Z0-9_]*)["']\)`),
			Description: "Go os.Getenv() calls",
			Language:    "go",
		},
		{
			Name:        "Node.js process.env",
			Pattern:     regexp.MustCompile(`process\.env\.([A-Z][A-Z0-9_]*)`),
			Description: "Node.js process.env access",
			Language:    "javascript",
		},
		{
			Name:        "Python os.environ",
			Pattern:     regexp.MustCompile(`os\.environ\[["']([A-Z][A-Z0-9_]*)["']\]`),
			Description: "Python os.environ access",
			Language:    "python",
		},
		{
			Name:        "Python os.getenv",
			Pattern:     regexp.MustCompile(`os\.getenv\(["']([A-Z][A-Z0-9_]*)["']\)`),
			Description: "Python os.getenv() calls",
			Language:    "python",
		},
		{
			Name:        "Shell variable expansion",
			Pattern:     regexp.MustCompile(`\$\{([A-Z][A-Z0-9_]*)\}`),
			Description: "Shell ${VAR} expansion",
			Language:    "shell",
		},
		{
			Name:        "Shell variable simple",
			Pattern:     regexp.MustCompile(`\$([A-Z][A-Z0-9_]{2,})`),
			Description: "Shell $VAR expansion",
			Language:    "shell",
		},
		{
			Name:        "Docker ENV",
			Pattern:     regexp.MustCompile(`ENV\s+([A-Z][A-Z0-9_]*)`),
			Description: "Dockerfile ENV declarations",
			Language:    "docker",
		},
		{
			Name:        "Java System.getenv",
			Pattern:     regexp.MustCompile(`System\.getenv\(["']([A-Z][A-Z0-9_]*)["']\)`),
			Description: "Java System.getenv() calls",
			Language:    "java",
		},
		{
			Name:        "C# Environment.GetEnvironmentVariable",
			Pattern:     regexp.MustCompile(`Environment\.GetEnvironmentVariable\(["']([A-Z][A-Z0-9_]*)["']\)`),
			Description: "C# Environment.GetEnvironmentVariable() calls",
			Language:    "csharp",
		},
		{
			Name:        "Ruby ENV",
			Pattern:     regexp.MustCompile(`ENV\[["']([A-Z][A-Z0-9_]*)["']\]`),
			Description: "Ruby ENV hash access",
			Language:    "ruby",
		},
		{
			Name:        "PHP getenv",
			Pattern:     regexp.MustCompile(`getenv\(["']([A-Z][A-Z0-9_]*)["']\)`),
			Description: "PHP getenv() calls",
			Language:    "php",
		},
		{
			Name:        "PHP $_ENV",
			Pattern:     regexp.MustCompile(`\$_ENV\[["']([A-Z][A-Z0-9_]*)["']\]`),
			Description: "PHP $_ENV superglobal access",
			Language:    "php",
		},
		{
			Name:        "YAML environment reference",
			Pattern:     regexp.MustCompile(`\$\{([A-Z][A-Z0-9_]*)\}`),
			Description: "YAML environment variable references",
			Language:    "yaml",
		},
		{
			Name:        "Generic string literal",
			Pattern:     regexp.MustCompile(`["']([A-Z][A-Z0-9_]{3,})["']`),
			Description: "Environment variable names in strings (lower confidence)",
			Language:    "generic",
		},
	}
}
