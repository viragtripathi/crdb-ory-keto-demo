package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		URL string `yaml:"url"`
	} `yaml:"database"`

	Keto struct {
		BaseURL string `yaml:"base_url"`
	} `yaml:"keto"`

	Workload struct {
		TupleCount  int `yaml:"tuple_count"`
		Concurrency int `yaml:"concurrency"`
		ChecksPerSecond int `yaml:"checks_per_second"`
	} `yaml:"workload"`
}

var AppConfig Config

func LoadConfig() error {
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
