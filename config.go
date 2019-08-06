package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// This type defines the structure of the yml config file for stylize.
type Config struct {
	FormattersByExt map[string]string `yaml:"formatters"`
	ExcludePatterns []string          `yaml:"exclude"`

	// Formatter arguments keyed by formatter name.
	// Example: {"clang": ["--style", "google"]}
	FormatterArgs map[string][]string `yaml:"formatter_args"`
}

// Read the config file. Check returned error with IsNotExist() to differentiate
// failure modes.
func LoadConfig(file string) (*Config, error) {
	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = yaml.Unmarshal(fileContent, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
