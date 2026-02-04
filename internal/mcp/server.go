package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"Weave-Toolkit/config"
	"Weave-Toolkit/internal/logger"
	"Weave-Toolkit/internal/tools"
)

// Server MCP 服务器
type Server struct {
	config       *config.Config
	logger       *logger.Logger
	httpSrv      *http.Server
	toolMgr      *tools.ToolManager
	activeOps    sync.WaitGroup // 等待正在执行的操作
	shuttingDown bool           // 关闭标志
	shutdownMu   sync.RWMutex   // 关闭状态锁
}

// NewServer 创建新的 MCP 服务器
func NewServer(cfg *config.Config, logger *logger.Logger) (*Server, error) {
	// 初始化工具管理器
	toolManager := tools.NewToolManager(logger, &cfg.ToolConfig)

	// 注册所有工具
	toolManager.RegisterAllTools()

	server := &Server{
		config:  cfg,
		logger:  logger,
		toolMgr: toolManager,
	}

	// 设置 HTTP 服务器
	server.setupHTTPServer()

	return server, nil
}

// Start 启动 MCP 服务器
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info().Str("address", s.config.ServerAddress).Msg("Starting MCP server")

	errChan := make(chan error, 1)

	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return s.Stop()
	}
}

// Stop 停止 MCP 服务器
func (s *Server) Stop() error {
	s.logger.Info().Msg("Stopping MCP server")

	// 设置关闭标志，拒绝新请求
	s.shutdownMu.Lock()
	s.shuttingDown = true
	s.shutdownMu.Unlock()

	s.logger.Info().Msg("Waiting for active operations to complete...")

	// 等待所有正在执行的操作完成（最多等待30秒）
	done := make(chan struct{})
	go func() {
		s.activeOps.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info().Msg("All active operations completed")
	case <-time.After(30 * time.Second):
		s.logger.Warn().Msg("Timeout waiting for active operations, forcing shutdown")
	}

	// 关闭 HTTP 服务器
	return s.httpSrv.Shutdown(context.Background())
}

// isShuttingDown 检查服务器是否正在关闭
func (s *Server) isShuttingDown() bool {
	s.shutdownMu.RLock()
	defer s.shutdownMu.RUnlock()
	return s.shuttingDown
}

func (s *Server) setupHTTPServer() {
	handler := http.NewServeMux()
	handler.HandleFunc("/mcp", s.handleMCPRequest)
	handler.HandleFunc("/health", s.handleHealthCheck)

	s.httpSrv = &http.Server{
		Addr:    s.config.ServerAddress,
		Handler: handler,
	}
}

func (s *Server) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	// 检查服务器是否正在关闭
	if s.isShuttingDown() {
		http.Error(w, "Service unavailable - server is shutting down", http.StatusServiceUnavailable)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 增加活跃操作计数
	s.activeOps.Add(1)
	defer s.activeOps.Done()

	// 读取请求体
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendErrorResponse(w, "Failed to read request body", -32700)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req map[string]interface{}
	if err := json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(&req); err != nil {
		s.sendErrorResponse(w, "Invalid JSON", -32700)
		return
	}

	method, ok := req["method"].(string)
	if !ok {
		s.sendErrorResponse(w, "Missing or invalid method", -32600)
		return
	}

	// 处理 MCP 请求
	result, err := s.handleMCPOperation(r.Context(), method, req)
	if err != nil {
		s.sendErrorResponse(w, err.Error(), -32603)
		return
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"result":  result,
		"id":      req["id"],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleMCPOperation(ctx context.Context, method string, req map[string]interface{}) (interface{}, error) {
	switch method {
	case MethodInitialize:
		return s.handleInitialize(req)
	case "notifications/initialized":
		return s.handleInitializedNotification(req)
	case MethodToolsList:
		return s.handleToolsList()
	case MethodToolsCall:
		return s.handleToolsCall(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func (s *Server) handleInitialize(req map[string]interface{}) (interface{}, error) {
	response := map[string]interface{}{
		"protocolVersion": ProtocolVersion,
		"serverInfo": map[string]interface{}{
			"name":    "Weave-Toolkit",
			"version": "1.0.0",
		},
		"capabilities": map[string]interface{}{
			"roots": map[string]interface{}{
				"listChanged": false,
			},
			"resources": map[string]interface{}{
				"listChanged": false,
			},
			"tools": map[string]interface{}{
				"listChanged": false,
			},
			"prompts": map[string]interface{}{
				"listChanged": false,
			},
		},
	}

	return response, nil
}

func (s *Server) handleInitializedNotification(req map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (s *Server) handleToolsList() (interface{}, error) {
	toolInfos := s.toolMgr.GetTools()

	// MCP 协议格式
	var tools []map[string]interface{}
	for _, tool := range toolInfos {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		})
	}

	return map[string]interface{}{
		"tools": tools,
	}, nil
}

func (s *Server) handleToolsCall(ctx context.Context, req map[string]interface{}) (interface{}, error) {
	params, ok := req["params"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params")
	}

	toolName, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid tool name")
	}

	arguments, err := json.Marshal(params["arguments"])
	if err != nil {
		return nil, fmt.Errorf("invalid arguments: %v", err)
	}

	result, err := s.toolMgr.CallTool(ctx, toolName, arguments)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Server) sendErrorResponse(w http.ResponseWriter, message string, code int) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
		"id": nil,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}
