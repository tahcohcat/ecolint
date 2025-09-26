# 🌱 ecolint

[![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)](https://golang.org/dl/)
[![Build Status][![CI Status](https://img.shields.io/github/actions/workflow/status/tahcohcat/ecolint/ci.yml?branch=main)](https://github.com/tahcohcat/ecolint/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/tahcohcat/ecolint)](https://goreportcard.com/report/github.com/tahcohcat/ecolint)
[![codecov](https://codecov.io/gh/yourusername/ecolint/branch/main/graph/badge.svg)](https://codecov.io/gh/tahcohcat/ecolint)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

> **Cultivating clean environments** 🌿

A delightfully fast and extensible linter for environment files that helps you maintain squeaky clean `.env` files across your projects. Because messy environment files are the root of all evil! 😈

## ✨ Features

- 🔍 **Smart Detection**: Finds duplicates, missing variables, and syntax errors
- 🎨 **Beautiful Output**: Colorful, emoji-rich terminal output that actually makes you happy
- ⚡ **Lightning Fast**: Written in Go for maximum performance
- 🔧 **Highly Configurable**: YAML configuration with sensible defaults
- 📦 **Zero Dependencies**: Single binary, no runtime requirements
- 🎯 **Multiple Formats**: Pretty, JSON, and GitHub Actions output
- 🔒 **Security Aware**: Detects potential secrets in plaintext
- 📏 **Convention Checking**: Enforces naming conventions and best practices

## 🚀 Quick Start

### Installation

#### Homebrew (macOS/Linux)
```bash
brew install tahcohcat/tap/ecolint
```

#### Go Install
```bash
go install github.com/tahcohcat/ecolint/cmd/ecolint@latest
```

#### Download Binary
Grab the latest binary from the [releases page](https://github.com/tahcohcat/ecolint/releases).

### Basic Usage

```bash
# Lint your .env files
ecolint lint

# Check specific files
ecolint lint .env .env.production

# Recursive search for all .env files
ecolint lint --recursive ./configs

# Initialize configuration
ecolint init

# Quick syntax check
ecolint check .env
```

## 🎪 Demo

```bash
$ ecolint lint examples/
🚨 Issues found:

📁 examples/duplicates.env
────────────────────────────────
  🔄 Line 2-3: duplicate variable 'VAR'
    💡 Remove one of the duplicate definitions
    💡 Use different variable names if both are needed
    💡 Check if this is a copy-paste error

📁 examples/security.env
──────────────────────────────
  🔒 Line 5: potential secret in plaintext 'API_SECRET'
    💡 Consider using a secret management system
    💡 Use placeholder values in committed files
    💡 Add this file to .gitignore if it contains real secrets

Found 2 issue(s) across 2 file(s)
```

## 🛠️ Configuration

Create a `.ecolint.yaml` file in your project root:

```yaml
# 🌱 ecolint configuration
required_vars:
  - DATABASE_URL
  - API_KEY
  - PORT

rules:
  duplicate: true      # Check for duplicate variables
  missing: true        # Check for missing required variables  
  syntax: true         # Validate .env syntax
  empty_values: true   # Warn about empty values
  security: true       # Check for potential secrets
  convention: true     # Enforce naming conventions

output:
  format: "pretty"     # pretty, json, github
  color: true          # Enable colors
```

## 📋 Rules

| Rule | Description | Example |
|------|-------------|---------|
| **duplicate** | Detects duplicate variable definitions | `VAR=1` and `VAR=2` in same file |
| **missing** | Finds missing required variables | `API_KEY` not defined but required |
| **syntax** | Validates .env file syntax | `INVALID LINE WITHOUT EQUALS` |
| **empty_values** | Warns about empty variable values | `DATABASE_URL=` |
| **security** | Detects potential secrets in plaintext | `PASSWORD=supersecret123` |
| **convention** | Enforces naming conventions | `CamelCase` instead of `UPPER_SNAKE_CASE` |

## 🎨 Output Formats

### Pretty (Default)
Beautiful, colorful terminal output with emojis and helpful suggestions.

### JSON
Perfect for CI/CD integration and programmatic processing:
```bash
ecolint lint --format json
```

### GitHub Actions
Native GitHub Actions annotations:
```bash
ecolint lint --format github
```

## 🔧 Advanced Usage

### CI/CD Integration

#### GitHub Actions
```yaml
- name: Lint Environment Files
  run: |
    ecolint lint --format github --recursive .
```

#### Docker
```dockerfile
FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY ecolint /usr/local/bin/
ENTRYPOINT ["ecolint"]
```

### Recursive Scanning
```bash
# Find all .env files in project
ecolint lint --recursive .

# Custom patterns
ecolint lint --recursive --include "*.env*" --exclude "*.example*" ./configs
```

### Configuration Discovery
ecolint automatically looks for configuration in:
- `.ecolint.yaml`
- `.ecolint.yml`
- `ecolint.yaml`
- `ecolint.yml`

## 🧪 Examples

Check out the [`examples/`](examples/) directory for sample files and configurations:

- [`examples/env/okay.env`](examples/env/okay.env) - A clean, well-formatted file
- [`examples/env/duplicates.env`](examples/env/duplicates.env) - Contains duplicate variables
- [`examples/config/`](examples/config/) - Sample configuration files

## 🤝 Contributing

We love contributions! Here's how to get started:

1. 🍴 Fork the repository
2. 🌟 Create a feature branch: `git checkout -b awesome-new-feature`
3. 💡 Make your changes and add tests
4. ✅ Run the tests: `go test ./...`
5. 📝 Commit your changes: `git commit -am 'Add awesome feature'`
6. 🚀 Push to the branch: `git push origin awesome-new-feature`
7. 🎉 Create a Pull Request

### Development Setup

```bash
# Clone the repository
git clone https://github.com/tahcohcat/ecolint.git
cd ecolint

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build the binary
go build -o ecolint cmd/ecolint/main.go

# Try it out!
./ecolint lint examples/
```

## 📈 Roadmap

- [X] 🔧 **Auto-fix mode**: Automatically fix common issues
- [X] 🐳 **Docker Images**: Pre-built container images
- [ ] 🌐 **Multi-language support**: Support for other env formats
- [ ] 📊 **Metrics**: Track environment file health over time
- [ ] 🔌 **Plugins**: Custom rule development API
- [ ] 📱 **IDE Integration**: VS Code, IntelliJ extensions
- [ ] 📚 **More Rules**: Expanded rule set for edge cases

## 🐛 Bug Reports & Feature Requests

Found a bug? Have an idea for a cool feature? We'd love to hear from you!

- 🐛 [Report a Bug](https://github.com/tahcohcat/ecolint/issues/new?template=bug_report.md)
- 💡 [Request a Feature](https://github.com/tahcohcat/ecolint/issues/new?template=feature_request.md)
- 💬 [Start a Discussion](https://github.com/tahcohcat/ecolint/discussions)

## 📜 License

MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by the amazing Go community
- Built with love using [Cobra CLI](https://github.com/spf13/cobra)
- Emoji game strong thanks to [Gitmoji](https://gitmoji.dev/)

## ⭐ Star History

[![Star History Chart](https://api.star-history.com/svg?repos=tahcohcat/ecolint&type=Date)](https://star-history.com/#tahcohcat/ecolint&Date)

---

<div align="center">
  <p><strong>Made with 💚 by developers, for developers</strong></p>
  <p>If ecolint helped you catch a bug, consider <a href="https://github.com/tahcohcat/ecolint">giving it a star ⭐</a></p>
</div>