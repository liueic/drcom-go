package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Auth   AuthConfig   `mapstructure:"auth"`
	Daemon DaemonConfig `mapstructure:"daemon"`
	Alert  AlertConfig  `mapstructure:"alert"`
}

type AuthConfig struct {
	Host     string `mapstructure:"host"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type DaemonConfig struct {
	Interval int `mapstructure:"interval"` // Seconds
}

type AlertConfig struct {
	TrafficThreshold float64 `mapstructure:"traffic_threshold"` // GB
	WebhookURL       string  `mapstructure:"webhook_url"`
}

func InitConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	configDir := filepath.Join(home, ".config", "drcom-go")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
	}

	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetDefault("auth.host", "http://10.10.10.9:801")
	viper.SetDefault("daemon.interval", 60)
	viper.SetDefault("alert.traffic_threshold", 80.0)
	viper.SetDefault("alert.webhook_url", "")

	viper.AutomaticEnv() 

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
		} else {
			fmt.Printf("Error reading config file: %s\n", err)
		}
	}
}

func LoadConfig() (*Config, error) {
    file := viper.ConfigFileUsed()
    if file != "" {
        info, err := os.Stat(file)
        if err == nil {
            mode := info.Mode().Perm()
            if mode != 0600 {
                return nil, fmt.Errorf("配置文件 %s 权限不安全 (%o)。请运行 'chmod 600 %s'", file, mode, file)
            }
        }
    }

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
    viper.Set("auth.host", cfg.Auth.Host)
    viper.Set("auth.username", cfg.Auth.Username)
    viper.Set("auth.password", cfg.Auth.Password)
    viper.Set("daemon.interval", cfg.Daemon.Interval)
    viper.Set("alert.traffic_threshold", cfg.Alert.TrafficThreshold)
    viper.Set("alert.webhook_url", cfg.Alert.WebhookURL)
    
    return viper.SafeWriteConfig()
}