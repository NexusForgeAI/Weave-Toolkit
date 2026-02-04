package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

// Logger 日志管理器
type Logger struct {
	zerolog.Logger
	file *os.File
}

// NewLogger 创建新的日志管理器
func NewLogger(logDir string, level string) (*Logger, error) {
	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// 创建日志文件
	logFile := filepath.Join(logDir, fmt.Sprintf("mcp-%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// 设置日志级别
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// 创建多输出日志器
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	multiWriter := io.MultiWriter(consoleWriter, file)

	logger := zerolog.New(multiWriter).
		Level(logLevel).
		With().
		Timestamp().
		Logger()

	return &Logger{
		Logger: logger,
		file:   file,
	}, nil
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// LogToolCall 记录工具调用日志
func (l *Logger) LogToolCall(toolName string, args interface{}, result interface{}, err error, duration time.Duration) {
	event := l.Info().
		Str("tool", toolName).
		Dur("duration", duration).
		Interface("args", args)

	if err != nil {
		event = l.Error().
			Str("tool", toolName).
			Dur("duration", duration).
			Interface("args", args).
			Err(err)
	} else {
		event = event.Interface("result", result)
	}

	event.Msg("Tool call completed")
}
