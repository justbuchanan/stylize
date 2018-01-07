package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// This type defines the structure of the yml config file for stylize.
type Config struct {
	FormattersByExt map[string]string `yaml:"formatters"`
	ExcludeDirs     []string          `yaml:"exclude_dirs"`

	// TODO: do better
	ClangStyle string `yaml:"clang_style"`
	YapfStyle  string `yaml:"yapf_style"`
}

// Read the config file. Check returned error with IsNotExist() to differentiate
// failure modes.
func LoadConfig(file string) (*Config, error) {
	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(fileContent, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
