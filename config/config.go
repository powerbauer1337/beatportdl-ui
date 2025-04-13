package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// AppConfig holds the application configuration
type AppConfig struct {
	MaxGlobalWorkers   int `json:"maxGlobalWorkers" yaml:"maxGlobalWorkers"`
	MaxDownloadWorkers int `json:"maxDownloadWorkers" yaml:"maxDownloadWorkers"`
}

// DefaultConfig returns a new AppConfig with default values
func DefaultConfig() *AppConfig {
	return &AppConfig{
		MaxGlobalWorkers:   5,
		MaxDownloadWorkers: 3,
	}
}

// Parse loads the configuration from the specified YAML file
func Parse(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfig loads the configuration from the specified JSON file
func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the configuration to the specified file
func (c *AppConfig) SaveConfig(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Save saves the configuration to the specified YAML file
func (c *AppConfig) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
