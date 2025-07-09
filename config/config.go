package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	FetchInterval time.Duration `mapstructure:"fetch_interval"`
	SocketPath    string        `mapstructure:"socket_path"`
}

func Load() (*Config, error) {
	viper.SetConfigName("treefarm")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.treefarm")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("fetch_interval", "15m")

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

	// Set socket path if not configured
	if cfg.SocketPath == "" {
		cfg.SocketPath = filepath.Join(GetWorktreeDir(), "daemon.sock")
	}

	return &cfg, nil
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
