package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"Weave-Toolkit/config"
	"Weave-Toolkit/internal/logger"
	"Weave-Toolkit/internal/mcp"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// 初始化日志
	logMgr, err := logger.NewLogger(cfg.LogDir, cfg.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logMgr.Close()

	logMgr.Info().Str("log_dir", cfg.LogDir).Msg("Logger initialized")

	// 创建 MCP 服务器
	server, err := mcp.NewServer(cfg, logMgr)
	if err != nil {
		logMgr.Fatal().Err(err).Msg("Failed to create MCP server")
	}

	// 优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务器
	go func() {
		if err := server.Start(ctx); err != nil {
			logMgr.Fatal().Err(err).Msg("Failed to start MCP server")
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logMgr.Info().Msg("Shutting down MCP server...")
	cancel()
}
