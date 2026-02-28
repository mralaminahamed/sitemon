package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	URL          string            `mapstructure:"url"`
	Workers      int               `mapstructure:"workers"`
	RPS          int               `mapstructure:"rps"`
	Timeout      int               `mapstructure:"timeout"`
	LogLevel     string            `mapstructure:"log_level"`
	LogFile      string            `mapstructure:"log_file"`
	OutputFormat string            `mapstructure:"output_format"`
	Headers      map[string]string `mapstructure:"headers"`
	Method       string            `mapstructure:"method"`
	MaxRetries   int               `mapstructure:"max_retries"`
	RetryWaitMs  int               `mapstructure:"retry_wait_ms"`
	TLSInsecure  bool              `mapstructure:"tls_insecure"`
}

var cfg *Config

func InitConfig() error {
	cfg = &Config{}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.sitemon")
	viper.AddConfigPath("$HOME/.config/sitemon")

	viper.SetDefault("workers", 10)
	viper.SetDefault("rps", 100)
	viper.SetDefault("timeout", 10)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_file", "")
	viper.SetDefault("output_format", "text")
	viper.SetDefault("method", "GET")
	viper.SetDefault("max_retries", 3)
	viper.SetDefault("retry_wait_ms", 500)
	viper.SetDefault("tls_insecure", false)

	viper.BindEnv("url", "SITEMON_URL", "PORTMAN_URL")
	viper.BindEnv("workers", "SITEMON_WORKERS", "PORTMAN_WORKERS")
	viper.BindEnv("rps", "SITEMON_RPS", "PORTMAN_RPS")
	viper.BindEnv("timeout", "SITEMON_TIMEOUT", "PORTMAN_TIMEOUT")
	viper.BindEnv("log_level", "SITEMON_LOG_LEVEL", "PORTMAN_LOG_LEVEL")
	viper.BindEnv("log_file", "SITEMON_LOG_FILE", "PORTMAN_LOG_FILE")
	viper.BindEnv("output_format", "SITEMON_OUTPUT_FORMAT", "PORTMAN_OUTPUT_FORMAT")
	viper.BindEnv("method", "SITEMON_METHOD", "PORTMAN_METHOD")
	viper.BindEnv("tls_insecure", "SITEMON_TLS_INSECURE", "PORTMAN_TLS_INSECURE")

	if err := viper.ReadInConfig(); err != nil {
		var pathErr *os.PathError
		if errors.As(err, &pathErr) {
			fmt.Println("No config file found, using defaults")
		} else {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.LogFile != "" {
		logPath := expandPath(cfg.LogFile)
		if err := ensureDir(filepath.Dir(logPath)); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	return nil
}

func GetConfig() *Config {
	return cfg
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return home + path[1:]
	}
	return path
}

func ensureDir(dir string) error {
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}
