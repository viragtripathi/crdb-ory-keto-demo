package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Keto struct {
		WriteAPI string `yaml:"write_api"`
		ReadAPI  string `yaml:"read_api"`
	} `yaml:"keto"`

	Workload struct {
		Concurrency     int `yaml:"concurrency"`
		ChecksPerSecond int `yaml:"checks_per_second"`
		ReadRatio       int `yaml:"read_ratio"`
        DurationSec     int `yaml:"duration_sec"`
	} `yaml:"workload"`
}

var AppConfig Config

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
