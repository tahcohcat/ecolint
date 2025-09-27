# Auto-Discovery Feature for ecolint

## Overview

The auto-discovery feature eliminates the need to manually configure required environment variables by automatically scanning your project files to find which environment variables your code actually uses.

## Quick Start

### 1. Discover Variables in Your Project

```bash
# Scan current directory for environment variable usage
ecolint scan

# Scan with custom confidence threshold
ecolint scan --min-confidence 0.8

# Show detailed usage information
ecolint scan --show-usages

# Generate configuration file with discovered variables
ecolint scan --generate-config
```

### 2. Lint with Auto-Discovery

```bash
# Lint .env files using auto-discovered variables
ecolint lint --auto-discover

# Combine with other flags
ecolint lint --auto-discover --recursive --format json

# Scan specific directory for variables
ecolint lint --auto-discover --scan-path ./src --min-confidence 0.8
```

## How It Works

The scanner looks for common environment variable usage patterns across multiple languages:

### Supported Languages & Patterns

| Language | Pattern Examples | Code Example |
|----------|------------------|--------------|
| **Go** | `os.Getenv("VAR")` | `port := os.Getenv("PORT")` |
| **JavaScript/Node.js** | `process.env.VAR` | `const port = process.env.PORT` |
| **Python** | `os.environ["VAR"]`, `os.getenv("VAR")` | `port = os.environ["PORT"]` |
| **Shell/Bash** | `$VAR`, `${VAR}` | `echo "Port: $PORT"` |
| **Docker** | `ENV VAR` | `ENV PORT=8080` |
| **Java** | `System.getenv("VAR")` | `String port = System.getenv("PORT");` |
| **C#** | `Environment.GetEnvironmentVariable("VAR")` | `var port = Environment.GetEnvironmentVariable("PORT");` |
| **Ruby** | `ENV["VAR"]` | `port = ENV["PORT"]` |
| **PHP** | `getenv("VAR")`, `$_ENV["VAR"]` | `$port = getenv("PORT");` |
| **YAML** | `${VAR}` | `port: ${PORT}` |

### Confidence Scoring

Each discovered variable gets a confidence score (0.0-1.0) based on:

- **Variable naming**: ALL_CAPS variables get higher confidence
- **Context**: Usage in known environment variable functions increases confidence  
- **Common patterns**: Variables with underscores, common env var names
- **Length**: Very short names (1-2 chars) get lower confidence

### Example Scan Output

```bash
$ ecolint scan

🔍 Scanning . for environment variable usage...

📊 Scan Summary:
  • Scanned 45 files
  • Found 12 unique variables  
  • 8 variables meet criteria (confidence ≥ 0.7, usages ≥ 1)

🌿 Environment Variables Discovered:
──────────────────────────────────────────────────
🟢 DATABASE_URL (95.0% confidence, 3 usages across 2 files)
🟢 API_KEY (90.0% confidence, 2 usages across 2 files)  
🟢 PORT (85.0% confidence, 4 usages across 3 files)
🟡 REDIS_URL (75.0% confidence, 1 usage across 1 file)
🟡 LOG_LEVEL (70.0% confidence, 2 usages across 1 file)
🟢 JWT_SECRET (92.0% confidence, 1 usage across 1 file)
🟡 SMTP_HOST (73.0% confidence, 1 usage across 1 file)
🟡 SMTP_PORT (71.0% confidence, 1 usage across 1 file)

💡 Next Steps:
  • Review the discovered variables above
  • Add missing variables to your .env files
  • Run: ecolint scan --generate-config  # to create configuration
  • Run: ecolint lint --auto-discover    # to lint with discovered variables
```

## Configuration Options

### Command Line Flags

```bash
# Scanning
ecolint scan --min-confidence 0.8    # Higher confidence threshold
ecolint scan --min-usages 2          # Must be used at least twice  
ecolint scan --exclude vendor        # Additional paths to exclude
ecolint scan --include-ext .config   # Additional extensions to scan

# Linting with auto-discovery
ecolint lint --auto-discover --scan-path ./src    # Scan specific directory
ecolint lint --auto-discover --min-confidence 0.9 # High confidence only
```

### Configuration File

```yaml
# .ecolint.yaml
scan:
  min_confidence: 0.7      # Minimum confidence for discovered variables
  min_usages: 1            # Minimum usage count to consider required
  exclude_paths:           # Additional paths to exclude
    - "build"
    - "dist"  
    - "coverage"
  include_extensions:      # Additional file extensions to scan
    - ".config.js"
    - ".local.yml"
```

## Advanced Usage

### Custom Patterns

You can extend the scanner with custom patterns for your specific use cases:

```go
// Custom pattern for your framework
customPatterns := []scan.VariablePattern{
    {
        Name:        "Custom Framework Config",
        Pattern:     regexp.MustCompile(`Config\.Get\(["']([A-Z][A-Z0-9_]*)["']\)`),
        Description: "Custom framework configuration access",
        Language:    "go",
    },
}

scanner := scan.NewProjectScanner().WithCustomPatterns(customPatterns)
```

### Integration with CI/CD

```yaml
# .github/workflows/env-check.yml
name: Environment Variable Check

on: [push, pull_request]

jobs:
  check-env:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install ecolint
        run: |
          curl -L https://github.com/tahcohcat/ecolint/releases/latest/download/ecolint-linux-amd64 -o ecolint
          chmod +x ecolint
      
      - name: Check environment variables
        run: |
          ./ecolint lint --auto-discover --format github
```

## Best Practices

### 1. **Use Confidence Thresholds Wisely**
- Start with 0.7 for initial discovery
- Increase to 0.8-0.9 for production environments  
- Use 0.5-0.6 for comprehensive discovery during development

### 2. **Review Auto-Discovered Variables**
Always review the results before adding to production:
```bash
# Review first, then generate config
ecolint scan --show-usages
ecolint scan --generate-config
```

### 3. **Combine with Manual Configuration**
Auto-discovery works alongside manual configuration:
```yaml
required_vars:
  # Manually specified critical variables
  - CRITICAL_SECRET_KEY
  # Auto-discovered variables will be added when using --auto-discover
```

### 4. **Environment-Specific Scanning**
Scan different parts of your codebase for different environments:
```bash
# Development environment
ecolint lint --auto-discover --scan-path ./src

# Production deployment scripts  
ecolint lint --auto-discover --scan-path ./deploy
```

## Troubleshooting

### No Variables Found
```bash
# Try lower confidence threshold
ecolint scan --min-confidence 0.5

# Include more file types
ecolint scan --include-ext .config --include-ext .local

# Check excluded paths
ecolint scan --show-usages  # See what files are being scanned
```

### False Positives
```bash  
# Increase confidence threshold
ecolint scan --min-confidence 0.9

# Require multiple usages
ecolint scan --min-usages 2

# Exclude specific paths
ecolint scan --exclude tests --exclude examples
```

### Integration Issues
- Ensure your project structure follows common patterns
- Check that environment variables follow UPPER_SNAKE_CASE convention
- Verify file extensions are in the scan list

## Examples

### Node.js Project
```bash
# Typical Node.js project scan
ecolint scan
# Discovers: PORT, NODE_ENV, DATABASE_URL, API_KEY, etc.

ecolint lint --auto-discover
# Checks .env files against discovered variables
```

### Go Project  
```bash
# Scan Go project with custom confidence
ecolint scan --min-confidence 0.8 --min-usages 2

# Generate config for Go project
ecolint scan --generate-config
```

### Multi-Language Project
```bash  
# Comprehensive scan across all supported languages
ecolint scan --min-confidence 0.6 --show-usages

# Lint with auto-discovery for polyglot projects
ecolint lint --auto-discover --recursive
```