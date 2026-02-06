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

// ConnectionPool 连接池
type ConnectionPool struct {
	pool    chan *MCPConnection
	maxSize int
	active  int
	mu      sync.RWMutex
	logger  *logger.Logger
}

// MCPConnection MCP 连接
type MCPConnection struct {
	ID         string
	ClientInfo *ClientInfo
	CreatedAt  time.Time
	LastActive time.Time
	Session    map[string]interface{}
}

// ClientInfo 客户端信息
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Server MCP 服务器
type Server struct {
	config       *config.Config
	logger       *logger.Logger
	httpSrv      *http.Server
	toolMgr      *tools.ToolManager
	connPool     *ConnectionPool // 连接池
	activeOps    sync.WaitGroup  // 等待正在执行的操作
	shuttingDown bool            // 关闭标志
	shutdownMu   sync.RWMutex    // 关闭状态锁
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

	// 初始化连接池
	maxConnections := 100 // 默认最大连接数
	if cfg.MaxConnections > 0 {
		maxConnections = cfg.MaxConnections
	}
	server.connPool = NewConnectionPool(maxConnections, logger)

	return server, nil
}

// NewConnectionPool 创建新的连接池
func NewConnectionPool(maxSize int, logger *logger.Logger) *ConnectionPool {
	return &ConnectionPool{
		pool:    make(chan *MCPConnection, maxSize),
		maxSize: maxSize,
		logger:  logger,
	}
}

// Acquire 获取连接
func (cp *ConnectionPool) Acquire(clientInfo *ClientInfo) (*MCPConnection, error) {
	select {
	case conn := <-cp.pool:
		// 从池中获取连接
		conn.LastActive = time.Now()
		cp.mu.Lock()
		cp.active++
		cp.mu.Unlock()
		return conn, nil
	default:
		// 池为空，创建新连接
		if cp.active >= cp.maxSize {
			return nil, fmt.Errorf("connection pool exhausted, max size: %d", cp.maxSize)
		}

		conn := &MCPConnection{
			ID:         generateConnectionID(),
			ClientInfo: clientInfo,
			CreatedAt:  time.Now(),
			LastActive: time.Now(),
			Session:    make(map[string]interface{}),
		}

		cp.mu.Lock()
		cp.active++
		cp.mu.Unlock()

		cp.logger.Debug().
			Str("connection_id", conn.ID).
			Str("client", clientInfo.Name).
			Msg("Created new connection")

		return conn, nil
	}
}

// Release 释放连接
func (cp *ConnectionPool) Release(conn *MCPConnection) {
	conn.LastActive = time.Now()

	select {
	case cp.pool <- conn:
		// 成功放回池中
		cp.mu.Lock()
		cp.active--
		cp.mu.Unlock()

		cp.logger.Debug().
			Str("connection_id", conn.ID).
			Msg("Connection released to pool")
	default:
		// 池已满，关闭连接
		cp.mu.Lock()
		cp.active--
		cp.mu.Unlock()

		cp.logger.Debug().
			Str("connection_id", conn.ID).
			Msg("Connection closed (pool full)")
	}
}

// Stats 获取连接池统计信息
func (cp *ConnectionPool) Stats() map[string]interface{} {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	return map[string]interface{}{
		"max_size":  cp.maxSize,
		"active":    cp.active,
		"pool_size": len(cp.pool),
		"available": cap(cp.pool) - len(cp.pool),
	}
}

// generateConnectionID 生成连接ID
func generateConnectionID() string {
	return fmt.Sprintf("conn_%d_%s", time.Now().UnixNano(), randomString(8))
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
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

	// 获取客户端信息并创建连接
	clientInfo := extractClientInfo(req)
	conn, err := s.connPool.Acquire(clientInfo)
	if err != nil {
		s.sendErrorResponse(w, fmt.Sprintf("Connection limit exceeded: %v", err), -32000)
		return
	}
	defer s.connPool.Release(conn)

	// 在处理请求前更新连接活跃时间
	conn.LastActive = time.Now()

	// 处理 MCP 请求
	result, err := s.handleMCPOperation(r.Context(), method, req, conn)
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

func (s *Server) handleMCPOperation(ctx context.Context, method string, req map[string]interface{}, conn *MCPConnection) (interface{}, error) {
	switch method {
	case MethodInitialize:
		return s.handleInitialize(req)
	case "notifications/initialized":
		return s.handleInitializedNotification(req)
	case MethodToolsList:
		return s.handleToolsList()
	case MethodToolsCall:
		return s.handleToolsCall(ctx, req, conn)
	case MethodResourcesList:
		return s.handleResourcesList()
	case MethodResourcesRead:
		return s.handleResourcesRead(ctx, req, conn)
	case MethodPromptsList:
		return s.handlePromptsList()
	case MethodPromptsGet:
		return s.handlePromptsGet(ctx, req, conn)
	case MethodRootsList:
		return s.handleRootsList()
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

func (s *Server) handleToolsCall(ctx context.Context, req map[string]interface{}, conn *MCPConnection) (interface{}, error) {
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

// handleResourcesList 处理资源列表请求
func (s *Server) handleResourcesList() (interface{}, error) {
	// 返回空资源列表（可根据需要扩展）
	return map[string]interface{}{
		"resources": []interface{}{},
	}, nil
}

// handleResourcesRead 处理资源读取请求
func (s *Server) handleResourcesRead(ctx context.Context, req map[string]interface{}, conn *MCPConnection) (interface{}, error) {
	params, ok := req["params"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params")
	}

	uri, ok := params["uri"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid uri")
	}

	// 可以支持文件系统、数据库、HTTP资源等
	content, err := s.readResource(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource: %v", err)
	}

	return map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"uri":      uri,
				"mimeType": "text/plain",
				"text":     content,
			},
		},
	}, nil
}

// handlePromptsList 处理提示词列表请求
func (s *Server) handlePromptsList() (interface{}, error) {
	// 返回空提示词列表（可根据需要扩展）
	return map[string]interface{}{
		"prompts": []interface{}{},
	}, nil
}

// handlePromptsGet 处理提示词获取请求
func (s *Server) handlePromptsGet(ctx context.Context, req map[string]interface{}, conn *MCPConnection) (interface{}, error) {
	params, ok := req["params"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params")
	}

	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid prompt name")
	}

	// 提示词获取（可根据需要扩展）
	prompt, err := s.getPrompt(name)
	if err != nil {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}

	return prompt, nil
}

// handleRootsList 处理根目录列表请求
func (s *Server) handleRootsList() (interface{}, error) {
	// 返回空根目录列表（可根据需要扩展）
	return map[string]interface{}{
		"roots": []interface{}{},
	}, nil
}

// readResource 读取资源内容
func (s *Server) readResource(uri string) (string, error) {
	// 可以扩展支持文件系统、HTTP资源等
	if uri == "file:///example.txt" {
		return "This is an example resource content.", nil
	}

	return "", fmt.Errorf("resource not found: %s", uri)
}

// getPrompt 获取提示词
func (s *Server) getPrompt(name string) (map[string]interface{}, error) {
	// 可以扩展支持数据库、配置文件等
	prompts := map[string]map[string]interface{}{
		"example_prompt": {
			"description": "An example prompt for demonstration",
			"arguments": []map[string]interface{}{
				{
					"name":        "topic",
					"description": "The topic to write about",
					"required":    true,
				},
			},
		},
	}

	if prompt, exists := prompts[name]; exists {
		return prompt, nil
	}

	return nil, fmt.Errorf("prompt not found")
}

// extractClientInfo 从请求中提取客户端信息
func extractClientInfo(req map[string]interface{}) *ClientInfo {
	clientInfo := &ClientInfo{
		Name:    "unknown",
		Version: "1.0.0",
	}

	// 尝试从初始化请求中获取客户端信息
	if params, ok := req["params"].(map[string]interface{}); ok {
		if client, ok := params["clientInfo"].(map[string]interface{}); ok {
			if name, ok := client["name"].(string); ok {
				clientInfo.Name = name
			}
			if version, ok := client["version"].(string); ok {
				clientInfo.Version = version
			}
		}
	}

	return clientInfo
}

// setupHTTPServer 设置 HTTP 服务器
func (s *Server) setupHTTPServer() {
	mux := http.NewServeMux()

	// MCP 协议端点
	mux.HandleFunc("/mcp", s.handleMCPRequest)
	mux.HandleFunc("/mcp/stream", s.handleMCPStreamRequest)

	// 新增健康检查端点
	mux.HandleFunc("/health", s.handleHealthCheck)
	mux.HandleFunc("/stats", s.handleStats)

	s.httpSrv = &http.Server{
		Addr:         s.config.ServerAddress,
		Handler:      mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}
}

// handleStats 统计信息端点
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.connPool.Stats()

	response := map[string]interface{}{
		"server_info": map[string]interface{}{
			"name":    "Weave-Toolkit",
			"version": "1.0.0",
		},
		"connections": stats,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMCPStreamRequest 处理流式 MCP 请求
func (s *Server) handleMCPStreamRequest(w http.ResponseWriter, r *http.Request) {
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

	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// 读取请求体
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendStreamError(w, "Failed to read request body")
		return
	}

	var req map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		s.sendStreamError(w, "Invalid JSON")
		return
	}

	method, ok := req["method"].(string)
	if !ok {
		s.sendStreamError(w, "Missing or invalid method")
		return
	}

	if method != MethodToolsCall {
		s.sendStreamError(w, "Streaming only supported for tools/call method")
		return
	}

	// 处理流式工具调用
	s.handleStreamToolsCall(r.Context(), w, req)
}

// handleStreamToolsCall 处理流式工具调用
func (s *Server) handleStreamToolsCall(ctx context.Context, w http.ResponseWriter, req map[string]interface{}) {
	params, ok := req["params"].(map[string]interface{})
	if !ok {
		s.sendStreamError(w, "Invalid params")
		return
	}

	toolName, ok := params["name"].(string)
	if !ok {
		s.sendStreamError(w, "Missing or invalid tool name")
		return
	}

	arguments, err := json.Marshal(params["arguments"])
	if err != nil {
		s.sendStreamError(w, fmt.Sprintf("Invalid arguments: %v", err))
		return
	}

	// 发送开始事件
	s.sendStreamEvent(w, StreamEventToolCall, map[string]interface{}{
		"tool":   toolName,
		"status": "started",
	})

	// 调用工具并获取流式结果
	result, err := s.toolMgr.CallToolStream(ctx, toolName, arguments, func(content string, index int) {
		// 发送内容事件
		s.sendStreamEvent(w, StreamEventContent, map[string]interface{}{
			"type":    ContentTypeText,
			"content": content,
			"index":   index,
		})
	})

	if err != nil {
		// 发送错误事件
		s.sendStreamEvent(w, StreamEventError, map[string]interface{}{
			"message": err.Error(),
		})
		return
	}

	// 发送完成事件
	s.sendStreamEvent(w, StreamEventDone, map[string]interface{}{
		"result": result,
	})
}

// sendStreamEvent 发送流式事件
func (s *Server) sendStreamEvent(w http.ResponseWriter, event string, data interface{}) {
	eventData, err := json.Marshal(data)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to marshal stream event data")
		return
	}

	eventStr := fmt.Sprintf("event: %s\ndata: %s\n\n", event, string(eventData))

	if _, err := w.Write([]byte(eventStr)); err != nil {
		s.logger.Error().Err(err).Msg("Failed to write stream event")
		return
	}

	// 刷新缓冲区
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// sendStreamError 发送流式错误
func (s *Server) sendStreamError(w http.ResponseWriter, message string) {
	s.sendStreamEvent(w, StreamEventError, map[string]interface{}{
		"message": message,
	})
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
