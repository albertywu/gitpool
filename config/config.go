package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type RepoConfig struct {
	FetchInterval time.Duration `mapstructure:"fetch_interval"`
}

type Config struct {
	ReconciliationInterval time.Duration          `mapstructure:"reconciliation_interval"`
	SocketPath             string                 `mapstructure:"socket_path"`
	Repos                  map[string]*RepoConfig `mapstructure:"repos"`
}

func Load() (*Config, error) {
	viper.SetConfigName("treefarm")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.treefarm")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("reconciliation_interval", "1m")
	viper.SetDefault("repos", map[string]*RepoConfig{})

	// Read config if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Initialize repos map if nil
	if cfg.Repos == nil {
		cfg.Repos = make(map[string]*RepoConfig)
	}

	// Set socket path if not configured
	if cfg.SocketPath == "" {
		cfg.SocketPath = filepath.Join(GetWorktreeDir(), "daemon.sock")
	}

	return &cfg, nil
}

// GetRepoFetchInterval returns the fetch interval for a repository, or 1h if not configured
func (c *Config) GetRepoFetchInterval(repoName string) time.Duration {
	if repoConfig, exists := c.Repos[repoName]; exists {
		return repoConfig.FetchInterval
	}
	return time.Hour // Default to 1 hour
}

// GetWorktreeDir returns the hardcoded worktree directory
func GetWorktreeDir() string {
	return filepath.Join(os.Getenv("HOME"), ".treefarm", "worktrees")
}

// GetConfigDir returns the config directory
func GetConfigDir() string {
	return filepath.Join(os.Getenv("HOME"), ".treefarm")
}

// EnsureWorktreeDir ensures the worktree directory exists
func EnsureWorktreeDir() error {
	if err := os.MkdirAll(GetWorktreeDir(), 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}
	return nil
}
