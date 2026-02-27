package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	URL      string `mapstructure:"url"`
	Workers  int    `mapstructure:"workers"`
	RPS      int    `mapstructure:"rps"`
	Timeout  int    `mapstructure:"timeout"`
	LogLevel string `mapstructure:"log_level"`
}

var cfg *Config

func InitConfig() error {
	cfg = &Config{}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.portman")

	viper.SetDefault("workers", 10)
	viper.SetDefault("rps", 100)
	viper.SetDefault("timeout", 10)
	viper.SetDefault("log_level", "info")

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

	return nil
}

func GetConfig() *Config {
	return cfg
}
