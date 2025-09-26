package rules

import (
	"regexp"
	"strings"

	"github.com/tahcohcat/ecolint/domain/env"
	"github.com/tahcohcat/ecolint/domain/issues"
)

// Security checks for potential secrets and sensitive data in plaintext
func Security(vars []env.Var, file string) []issues.Issue {
	var out []issues.Issue

	// Patterns that might indicate secrets
	secretKeyPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(password|pwd|pass)$`),
		regexp.MustCompile(`(?i)(secret|key|token)$`),
		regexp.MustCompile(`(?i)(private|priv)_key$`),
		regexp.MustCompile(`(?i)api_(key|secret|token)$`),
		regexp.MustCompile(`(?i)(auth|oauth)_(key|secret|token)$`),
		regexp.MustCompile(`(?i)(access|refresh)_token$`),
		regexp.MustCompile(`(?i)jwt_(secret|key)$`),
		regexp.MustCompile(`(?i)(db|database)_(password|pass|pwd)$`),
		regexp.MustCompile(`(?i)(smtp|email)_(password|pass|pwd)$`),
		regexp.MustCompile(`(?i)(aws|gcp|azure)_(secret|key)$`),
	}

	// Patterns that might indicate actual secret values (not just keys)
	secretValuePatterns := []*regexp.Regexp{
		// JWT tokens (base64 with dots)
		regexp.MustCompile(`^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`),
		// API keys (long alphanumeric strings)
		regexp.MustCompile(`^[A-Za-z0-9]{32,}$`),
		// Base64 encoded data (longer than 20 chars)
		regexp.MustCompile(`^[A-Za-z0-9+/]{20,}={0,2}$`),
		// Hex encoded keys (even length, 16+ chars)
		regexp.MustCompile(`^[a-fA-F0-9]{16,}$`),
		// AWS-style keys
		regexp.MustCompile(`^AKIA[0-9A-Z]{16}$`),
		// Google API keys
		regexp.MustCompile(`^AIza[0-9A-Za-z_-]{35}$`),
	}

	for _, v := range vars {
		// Skip empty values
		if v.Value == "" {
			continue
		}

		// Check if the variable name suggests it might contain a secret
		isSecretKey := false
		for _, pattern := range secretKeyPatterns {
			if pattern.MatchString(v.Key) {
				isSecretKey = true
				break
			}
		}

		// Check if the value looks like a secret
		looksLikeSecret := false
		for _, pattern := range secretValuePatterns {
			if pattern.MatchString(v.Value) {
				looksLikeSecret = true
				break
			}
		}

		// Check for common placeholder values that are safe
		safePlaceholders := []string{
			"changeme", "placeholder", "your_key_here", "your_secret_here",
			"example", "sample", "dummy", "test", "localhost", "127.0.0.1",
			"true", "false", "development", "production", "staging",
		}

		isSafePlaceholder := false
		lowerValue := strings.ToLower(v.Value)
		for _, placeholder := range safePlaceholders {
			if lowerValue == placeholder || strings.Contains(lowerValue, placeholder) {
				isSafePlaceholder = true
				break
			}
		}

		// Skip if it's a safe placeholder
		if isSafePlaceholder {
			continue
		}

		// Report issue if key suggests secret OR value looks like secret
		if isSecretKey || looksLikeSecret {
			recommendations := []string{
				"Consider using a secret management system (e.g., HashiCorp Vault, AWS Secrets Manager)",
				"Use placeholder values in committed files (e.g., 'your_api_key_here')",
				"Add this file to .gitignore if it contains real secrets",
				"Use environment-specific files (.env.local) for sensitive data",
			}

			if isSecretKey && !looksLikeSecret {
				recommendations = append([]string{"Variable name suggests it may contain sensitive data"}, recommendations...)
			}

			if looksLikeSecret {
				recommendations = append([]string{"Value appears to be a secret or API key"}, recommendations...)
			}

			out = append(out, issues.NewIssue(
				"potential secret in plaintext",
				v.Key,
				file,
				v.Line,
				0,
				recommendations,
			))
		}
	}

	return out
}
