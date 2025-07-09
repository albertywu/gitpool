package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	WorkDir       string        `mapstructure:"work_dir"`
	FetchInterval time.Duration `mapstructure:"fetch_interval"`
	SocketPath    string        `mapstructure:"socket_path"`
}

func Load() (*Config, error) {
	viper.SetConfigName("treefarm")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.treefarm")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("work_dir", filepath.Join(os.Getenv("HOME"), ".treefarm", "workdir"))
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
		cfg.SocketPath = filepath.Join(cfg.WorkDir, "daemon.sock")
	}

	return &cfg, nil
}

func (c *Config) EnsureWorkDir() error {
	if err := os.MkdirAll(c.WorkDir, 0755); err != nil {
		return fmt.Errorf("failed to create work directory: %w", err)
	}
	return nil
}
