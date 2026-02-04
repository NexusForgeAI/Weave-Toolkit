package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	ServerAddress  string            `json:"server_address"`
	LogLevel       string            `json:"log_level"`
	LogDir         string            `json:"log_dir"`
	MaxConnections int               `json:"max_connections"`
	ToolTimeout    time.Duration     `json:"tool_timeout"`
	MaxRequestSize int64             `json:"max_request_size"`
	ReadTimeout    time.Duration     `json:"read_timeout"`
	WriteTimeout   time.Duration     `json:"write_timeout"`
	IdleTimeout    time.Duration     `json:"idle_timeout"`
	APIKey         string            `json:"api_key"`
	CORSOrigin     string            `json:"cors_origin"`
	ToolConfig     ToolManagerConfig `json:"tool_config"`
}

// ToolManagerConfig 工具管理器配置
type ToolManagerConfig struct {
	Categories map[string]CategoryConfig `json:"categories"`
	Global     GlobalToolConfig          `json:"global"`
}

// CategoryConfig 分类配置
type CategoryConfig struct {
	Enabled   bool          `json:"enabled"`
	MaxTools  int           `json:"max_tools"`
	RateLimit int           `json:"rate_limit"`
	Timeout   time.Duration `json:"timeout"`
}

// GlobalToolConfig 全局工具配置
type GlobalToolConfig struct {
	MaxConcurrentCalls int           `json:"max_concurrent_calls"`
	DefaultTimeout     time.Duration `json:"default_timeout"`
	EnableMetrics      bool          `json:"enable_metrics"`
	EnableTracing      bool          `json:"enable_tracing"`
}

// Load 加载配置
func Load() (*Config, error) {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env file: %v", err)
	}

	cfg := &Config{
		ServerAddress:  os.Getenv("MCP_SERVER_ADDRESS"),
		LogLevel:       os.Getenv("MCP_LOG_LEVEL"),
		LogDir:         os.Getenv("MCP_LOG_DIR"),
		MaxConnections: parseInt(os.Getenv("MCP_MAX_CONNECTIONS")),
		ToolTimeout:    parseDuration(os.Getenv("MCP_TOOL_TIMEOUT")),
		MaxRequestSize: parseInt64(os.Getenv("MCP_MAX_REQUEST_SIZE")),
		ReadTimeout:    parseDuration(os.Getenv("MCP_READ_TIMEOUT")),
		WriteTimeout:   parseDuration(os.Getenv("MCP_WRITE_TIMEOUT")),
		IdleTimeout:    parseDuration(os.Getenv("MCP_IDLE_TIMEOUT")),
		APIKey:         os.Getenv("MCP_API_KEY"),
		CORSOrigin:     os.Getenv("MCP_CORS_ORIGIN"),
	}

	// 加载工具配置文件
	if err := cfg.loadToolConfigFile(); err != nil {
		return nil, fmt.Errorf("failed to load tool config: %v", err)
	}

	return cfg, nil
}

// loadToolConfigFile 加载工具配置文件
func (c *Config) loadToolConfigFile() error {
	configPath := os.Getenv("TOOL_CONFIG_PATH")
	if configPath == "" {
		configPath = "tool-config.json"
	}

	// 如果文件不存在，返回错误
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("tool config file not found: %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read tool config file: %v", err)
	}

	var toolConfig ToolManagerConfig
	if err := json.Unmarshal(data, &toolConfig); err != nil {
		return fmt.Errorf("failed to parse tool config file: %v", err)
	}

	c.ToolConfig = toolConfig

	return nil
}

// parseInt 解析字符串为整数
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return 0
}

// parseInt64 解析字符串为64位整数
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return v
	}
	return 0
}

// parseDuration 解析字符串为时间间隔
func parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	if v, err := time.ParseDuration(s); err == nil {
		return v
	}
	return 0
}
