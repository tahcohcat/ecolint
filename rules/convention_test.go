package rules

import (
	"strings"
	"testing"

	"github.com/tahcohcat/ecolint/domain/env"
)

func TestConvention(t *testing.T) {
	tests := []struct {
		name     string
		vars     []env.Var
		expected int
	}{
		{
			name: "valid names",
			vars: []env.Var{
				{Key: "DATABASE_URL", Value: "postgres://localhost", Line: 1},
				{Key: "API_KEY", Value: "secret", Line: 2},
				{Key: "PORT", Value: "8080", Line: 3},
				{Key: "LOG_LEVEL", Value: "info", Line: 4},
			},
			expected: 0,
		},
		{
			name: "camelCase",
			vars: []env.Var{
				{Key: "databaseUrl", Value: "postgres://localhost", Line: 1},
				{Key: "apiKey", Value: "secret", Line: 2},
			},
			expected: 2,
		},
		{
			name: "kebab-case",
			vars: []env.Var{
				{Key: "database-url", Value: "postgres://localhost", Line: 1},
				{Key: "api-key", Value: "secret", Line: 2},
			},
			expected: 2,
		},
		{
			name: "spaces in names",
			vars: []env.Var{
				{Key: "DATABASE URL", Value: "postgres://localhost", Line: 1},
				{Key: "API KEY", Value: "secret", Line: 2},
			},
			expected: 2,
		},
		{
			name: "lowercase",
			vars: []env.Var{
				{Key: "database_url", Value: "postgres://localhost", Line: 1},
				{Key: "api_key", Value: "secret", Line: 2},
			},
			expected: 2,
		},
		{
			name: "starts with number",
			vars: []env.Var{
				{Key: "3RD_PARTY_API", Value: "value", Line: 1},
				{Key: "123_CONFIG", Value: "config", Line: 2},
			},
			expected: 2,
		},
		{
			name: "single letter",
			vars: []env.Var{
				{Key: "A", Value: "value", Line: 1},
				{Key: "X", Value: "value", Line: 2},
			},
			expected: 2,
		},
		{
			name: "system variables",
			vars: []env.Var{
				{Key: "PATH", Value: "/custom/path", Line: 1},
				{Key: "HOME", Value: "/home/user", Line: 2},
				{Key: "USER", Value: "customuser", Line: 3},
			},
			expected: 3,
		},
		{
			name: "generic names",
			vars: []env.Var{
				{Key: "CONFIG", Value: "config.json", Line: 1},
				{Key: "DATA", Value: "data.json", Line: 2},
				{Key: "SETTINGS", Value: "settings.json", Line: 3},
			},
			expected: 3,
		},
		{
			name: "redundant prefixes",
			vars: []env.Var{
				{Key: "ENV_DATABASE_URL", Value: "postgres://localhost", Line: 1},
				{Key: "ENVIRONMENT_API_KEY", Value: "secret", Line: 2},
				{Key: "VAR_PORT", Value: "8080", Line: 3},
			},
			expected: 3,
		},
		{
			name: "abbreviations",
			vars: []env.Var{
				{Key: "DB", Value: "postgres://localhost", Line: 1},
				{Key: "DB_PWD", Value: "secret", Line: 2},
				{Key: "SVR", Value: "8080", Line: 3},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := Convention(tt.vars, "test.env")

			if len(issues) != tt.expected {
				t.Errorf("Convention() = %d issues, want %d", len(issues), tt.expected)

				// Debug output
				for i, issue := range issues {
					t.Logf("Issue %d: %s (%s)", i+1, issue.Key, issue.Name)
				}
			}
		})
	}
}

func TestConventionRecommendations(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		expectedInRec []string // Strings that should be in recommendations
	}{
		{
			name:          "camelCase",
			key:           "databaseUrl",
			expectedInRec: []string{"camelCase", "DATABASE_URL"},
		},
		{
			name:          "kebab-case",
			key:           "database-url",
			expectedInRec: []string{"underscores", "DATABASE_URL"},
		},
		{
			name:          "spaces",
			key:           "DATABASE URL",
			expectedInRec: []string{"spaces", "underscores"},
		},
		{
			name:          "lowercase",
			key:           "database_url",
			expectedInRec: []string{"UPPERCASE", "DATABASE_URL"},
		},
		{
			name:          "system variable",
			key:           "PATH",
			expectedInRec: []string{"system", "MYAPP_PATH"},
		},
		{
			name:          "abbreviation",
			key:           "DB_URL",
			expectedInRec: []string{"abbreviations", "DATABASE_URL"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := []env.Var{{Key: tt.key, Value: "value", Line: 1}}
			issues := Convention(vars, "test.env")

			if len(issues) == 0 {
				t.Fatalf("Expected at least one issue for key '%s'", tt.key)
			}

			issue := issues[0]
			allRecommendations := strings.Join(issue.Recommendations, " ")

			for _, expected := range tt.expectedInRec {
				if !strings.Contains(allRecommendations, expected) {
					t.Errorf("Expected recommendation to contain '%s', but got: %v", expected, issue.Recommendations)
				}
			}
		})
	}
}

func TestConvertCamelToSnake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"camelCase", "camel_case"},
		{"PascalCase", "pascal_case"},
		{"databaseUrl", "database_url"},
		{"HTTPSProxy", "https_proxy"},
		{"XMLHttpRequest", "xml_http_request"},
		{"iPhone", "i_phone"},
		{"URLPath", "url_path"},
		{"simpleword", "simpleword"},
		{"ALLCAPS", "allcaps"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertCamelToSnake(tt.input)
			if result != tt.expected {
				t.Errorf("convertCamelToSnake(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}
