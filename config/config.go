package config

import (
	"os"
	"strconv"
)

// Config 服务器配置
type Config struct {
	ServerAddress  string
	LogLevel       string
	MaxConnections int
}

// Load 加载配置，支持环境变量和默认值
func Load() (*Config, error) {
	cfg := &Config{
		ServerAddress:  getEnv("MCP_SERVER_ADDRESS", ":8080"),
		LogLevel:       getEnv("MCP_LOG_LEVEL", "info"),
		MaxConnections: getEnvInt("MCP_MAX_CONNECTIONS", 100),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
