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
		configDir = os.Getenv("TREEFARM_CONFIG_DIR")
	}
	if worktreeDir == "" {
		worktreeDir = os.Getenv("TREEFARM_WORKTREE_DIR")
	}
	if socketPath == "" {
		socketPath = os.Getenv("TREEFARM_SOCKET_PATH")
	}

	// Set up config directory
	if configDir == "" {
		configDir = GetConfigDir()
	}

	viper.SetConfigName("treefarm")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
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

// EnsureWorktreeDirForConfig ensures the worktree directory exists for a config
func (c *Config) EnsureWorktreeDir() error {
	if err := os.MkdirAll(c.WorktreeDir, 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}
	return nil
}
