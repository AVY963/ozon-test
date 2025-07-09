package config

import (
	"fmt"
	"os"
	"ozon-posts/internal/repositories"
	"strconv"
)

type Config struct {
	Server   ServerConfig         `json:"server"`
	Database *repositories.Config `json:"database"`
	Log      LogConfig            `json:"log"`
}

type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("PORT", 8080),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		Database: repositories.LoadConfig(),
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
