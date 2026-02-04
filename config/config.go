package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config 服务器配置
type Config struct {
	ServerAddress  string
	LogLevel       string
	LogDir         string
	MaxConnections int
	ToolTimeout    time.Duration
	MaxRequestSize int64
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	APIKey         string
	CORSOrigin     string
}

// 加载配置
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env file: %v", err)
	}

	// 获取项目根目录
	projectRoot, err := getProjectRoot()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		ServerAddress:  getEnv("MCP_SERVER_ADDRESS", ":8080"),
		LogLevel:       getEnv("MCP_LOG_LEVEL", "info"),
		LogDir:         getEnv("MCP_LOG_DIR", filepath.Join(projectRoot, "log")),
		MaxConnections: getEnvInt("MCP_MAX_CONNECTIONS", 100),
		ToolTimeout:    getEnvDuration("MCP_TOOL_TIMEOUT", 30*time.Second),
		MaxRequestSize: getEnvInt64("MCP_MAX_REQUEST_SIZE", 1048576),
		ReadTimeout:    getEnvDuration("MCP_READ_TIMEOUT", 15*time.Second),
		WriteTimeout:   getEnvDuration("MCP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:    getEnvDuration("MCP_IDLE_TIMEOUT", 60*time.Second),
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

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return currentDir, nil
}
