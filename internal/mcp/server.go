package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"Weave-Toolkit/config"
	"Weave-Toolkit/internal/tools"

	"github.com/rs/zerolog"
)

// Server MCP 服务器
type Server struct {
	config  *config.Config
	logger  *zerolog.Logger
	httpSrv *http.Server
	toolMgr *tools.ToolManager
}

// NewServer 创建新的 MCP 服务器
func NewServer(cfg *config.Config, logger *zerolog.Logger) (*Server, error) {
	// 初始化工具管理器
	toolManager := tools.NewToolManager(logger)

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
	return s.httpSrv.Shutdown(context.Background())
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 处理 MCP 请求
	result, err := s.handleMCPOperation(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) handleMCPOperation(ctx context.Context, req MCPRequest) (interface{}, error) {
	switch req.Method {
	case "tools/list":
		return s.toolMgr.GetTools(), nil
	case "tools/call":
		var callReq struct {
			Name string          `json:"name"`
			Args json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &callReq); err != nil {
			return nil, fmt.Errorf("invalid call parameters: %v", err)
		}
		return s.toolMgr.CallTool(ctx, callReq.Name, callReq.Args)
	default:
		return nil, fmt.Errorf("unsupported method: %s", req.Method)
	}
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

// MCPRequest MCP 请求结构
type MCPRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}
