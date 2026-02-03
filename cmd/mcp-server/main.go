package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"Weave-Toolkit/config"
	"Weave-Toolkit/internal/mcp"

	"github.com/rs/zerolog"
)

func main() {
	// 初始化日志
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// 创建 MCP 服务器
	server, err := mcp.NewServer(cfg, &logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create MCP server")
	}

	// 优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务器
	go func() {
		if err := server.Start(ctx); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start MCP server")
		}
	}()

	logger.Info().Str("address", cfg.ServerAddress).Msg("MCP server started")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info().Msg("Shutting down MCP server...")
	cancel()
}
