package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ReconciliationInterval time.Duration `mapstructure:"reconciliation_interval"`
	SocketPath             string        `mapstructure:"socket_path"`
	// Custom paths (can be set via CLI flags)
	ConfigDir   string `mapstructure:"-"`
	WorktreeDir string `mapstructure:"-"`
}

func Load() (*Config, error) {
	return LoadWithCustomPaths("", "", "")
}

func LoadWithCustomPaths(configDir, worktreeDir, socketPath string) (*Config, error) {
	// Check environment variables first
	if configDir == "" {
		configDir = os.Getenv("GITPOOL_CONFIG_DIR")
	}
	if worktreeDir == "" {
		worktreeDir = os.Getenv("GITPOOL_WORKTREE_DIR")
	}
	if socketPath == "" {
		socketPath = os.Getenv("GITPOOL_SOCKET_PATH")
	}

	// Set up config directory
	if configDir == "" {
		configDir = GetConfigDir()
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Set defaults
	viper.SetDefault("reconciliation_interval", "1m")

	// Read config if exists
	_ = viper.ReadInConfig() // Ignore error - config file is optional

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set custom paths
	cfg.ConfigDir = configDir
	if worktreeDir != "" {
		cfg.WorktreeDir = worktreeDir
	} else {
		cfg.WorktreeDir = GetWorktreeDir()
	}

	// Set socket path if not configured
	if socketPath != "" {
		cfg.SocketPath = socketPath
	} else if cfg.SocketPath == "" {
		cfg.SocketPath = filepath.Join(cfg.WorktreeDir, "daemon.sock")
	}

	return &cfg, nil
}


// GetWorktreeDir returns the hardcoded worktree directory
func GetWorktreeDir() string {
	return filepath.Join(os.Getenv("HOME"), ".gitpool", "worktrees")
}

// GetConfigDir returns the config directory
func GetConfigDir() string {
	return filepath.Join(os.Getenv("HOME"), ".gitpool")
}

// EnsureWorktreeDir ensures the worktree directory exists
func EnsureWorktreeDir() error {
	if err := os.MkdirAll(GetWorktreeDir(), 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}
	return nil
}

// EnsureWorktreeDirForConfig ensures the worktree directory exists for a config
func (c *Config) EnsureWorktreeDir() error {
	if err := os.MkdirAll(c.WorktreeDir, 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}
	return nil
}
