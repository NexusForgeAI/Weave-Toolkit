package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog"
)

// ToolManager MCP 工具管理器
type ToolManager struct {
	tools  map[string]Tool
	logger *zerolog.Logger
}

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error)
}

// ToolInfo 工具信息结构
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ToolCallResult 工具调用结果
type ToolCallResult struct {
	Content []ToolCallContent `json:"content"`
}

// ToolCallContent 工具调用内容
type ToolCallContent struct {
	Type string      `json:"type"`
	Text string      `json:"text"`
	Data interface{} `json:"data,omitempty"`
}

// NewToolManager 创建新的工具管理器
func NewToolManager(logger *zerolog.Logger) *ToolManager {
	return &ToolManager{
		tools:  make(map[string]Tool),
		logger: logger,
	}
}

// RegisterTool 注册工具
func (tm *ToolManager) RegisterTool(tool Tool) {
	tm.tools[tool.Name()] = tool
	tm.logger.Info().Str("tool", tool.Name()).Msg("Tool registered")
}

// RegisterAllTools 注册所有可用工具
func (tm *ToolManager) RegisterAllTools() {
	tm.RegisterTool(&CalculatorTool{})
	tm.RegisterTool(&FileTool{})
	// 添加更多工具
}

// GetTools 获取所有工具信息
func (tm *ToolManager) GetTools() []ToolInfo {
	var tools []ToolInfo
	for _, tool := range tm.tools {
		tools = append(tools, ToolInfo{
			Name:        tool.Name(),
			Description: tool.Description(),
		})
	}
	return tools
}

// CallTool 调用工具
func (tm *ToolManager) CallTool(ctx context.Context, name string, args json.RawMessage) (*ToolCallResult, error) {
	tool, exists := tm.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	tm.logger.Debug().Str("tool", name).Msg("Executing tool")

	result, err := tool.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	return &ToolCallResult{
		Content: []ToolCallContent{
			{
				Type: "text",
				Text: string(result),
			},
		},
	}, nil
}
