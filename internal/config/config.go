package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	RequiredVars []string `yaml:"required_vars"`
	Rules        Rules    `yaml:"rules"`
	Output       Output   `yaml:"output"`
}

type Rules struct {
	Duplicate   bool `yaml:"duplicate"`
	Missing     bool `yaml:"missing"`
	Security    bool `yaml:"security"`
	Convention  bool `yaml:"convention"`
	Syntax      bool `yaml:"syntax"`
	EmptyValues bool `yaml:"empty_values"`
}

type Output struct {
	Format string `yaml:"format"`
	Color  bool   `yaml:"color"`
}

func Load(configFile string) Config {
	// Default configuration
	cfg := Config{
		RequiredVars: []string{},
		Rules: Rules{
			Duplicate:   true,
			Missing:     true,
			Syntax:      true,
			EmptyValues: true,
		},
		Output: Output{
			Format: "pretty",
			Color:  true,
		},
	}

	// Try to find config file
	if configFile == "" {
		candidates := []string{
			".ecolint.yaml",
			".ecolint.yml",
			"ecolint.yaml",
			"ecolint.yml",
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				configFile = candidate
				break
			}
		}
	}

	if configFile == "" {
		return cfg // Return default config
	}

	// Load config file
	if data, err := ioutil.ReadFile(configFile); err == nil {
		yaml.Unmarshal(data, &cfg)
	}

	return cfg
}

func CreateSampleConfig(path string) error {
	sampleConfig := `# ecolint configuration file
# 🌱 cultivating clean environments

# Required environment variables that must be present
required_vars:
  - DATABASE_URL
  - API_KEY
  - PORT

# Rule configuration
rules:
  duplicate: true      # Check for duplicate variable definitions
  missing: true        # Check for missing required variables
  syntax: true         # Validate .env file syntax
  empty_values: true   # Warn about empty variable values

# Output configuration  
output:
  format: "pretty"     # Output format: pretty, json, github
  color: true          # Enable colored output
`

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(sampleConfig), 0644)
}
