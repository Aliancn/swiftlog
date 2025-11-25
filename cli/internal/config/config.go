package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds CLI configuration
type Config struct {
	Token          string `mapstructure:"token"`
	ServerAddr     string `mapstructure:"server_addr"`
	DefaultProject string `mapstructure:"default_project"`
	LastProject    string `mapstructure:"last_project"`
	TLS            bool   `mapstructure:"tls"` // Enable TLS/secure connection
}

// ProjectConfig holds project-level configuration
type ProjectConfig struct {
	Project string `json:"project" yaml:"project"`
}

var (
	configDir  string
	configFile string
)

func init() {
	// Determine config directory (~/.swiftlog/)
	home, err := os.UserHomeDir()
	if err != nil {
		configDir = ".swiftlog"
	} else {
		configDir = filepath.Join(home, ".swiftlog")
	}
	configFile = filepath.Join(configDir, "config.json")
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("json")

	// Set defaults
	viper.SetDefault("server_addr", "localhost:50051")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found, use defaults
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to disk
func Save(cfg *Config) error {
	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.Set("token", cfg.Token)
	viper.Set("server_addr", cfg.ServerAddr)
	viper.Set("default_project", cfg.DefaultProject)
	viper.Set("last_project", cfg.LastProject)
	viper.Set("tls", cfg.TLS)

	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return configFile
}

// IsConfigured checks if the CLI is configured with a token
func IsConfigured() bool {
	cfg, err := Load()
	if err != nil {
		return false
	}
	return cfg.Token != ""
}

// UpdateLastUsed updates the last used project
func UpdateLastUsed(project, _ string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	cfg.LastProject = project

	return Save(cfg)
}

// LoadProjectConfig loads project-level configuration from current directory
func LoadProjectConfig() (*ProjectConfig, error) {
	// Try to load .swiftlog.json or .swiftlog.yaml from current directory
	configFiles := []string{".swiftlog.json", ".swiftlog.yaml", ".swiftlog.yml"}

	for _, filename := range configFiles {
		if _, err := os.Stat(filename); err == nil {
			// File exists, try to load it
			v := viper.New()
			v.SetConfigFile(filename)

			if err := v.ReadInConfig(); err != nil {
				continue
			}

			var cfg ProjectConfig
			if err := v.Unmarshal(&cfg); err != nil {
				continue
			}

			return &cfg, nil
		}
	}

	// No project config found
	return nil, nil
}
