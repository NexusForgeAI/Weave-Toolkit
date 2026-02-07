package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"Weave-Toolkit/config"
	"Weave-Toolkit/internal/logger"
)

// ToolCategory 工具分类
type ToolCategory string

const (
	CategoryMath    ToolCategory = "math"    // 计算工具
	CategoryAI      ToolCategory = "ai"      // AI工具
	CategorySystem  ToolCategory = "system"  // 系统工具
	CategoryUtility ToolCategory = "utility" // 实用工具
)

// ToolManager MCP 工具管理器
type ToolManager struct {
	categories map[ToolCategory]*CategoryManager
	mu         sync.RWMutex
	logger     *logger.Logger
}

// CategoryManager 分类管理器
type CategoryManager struct {
	Name    ToolCategory
	tools   map[string]Tool
	enabled bool
	config  CategoryConfig
}

// CategoryConfig 分类配置
type CategoryConfig struct {
	Enabled   bool          `json:"enabled"`
	MaxTools  int           `json:"max_tools"`
	RateLimit int           `json:"rate_limit"`
	Timeout   time.Duration `json:"timeout"`
}

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Category() ToolCategory
	Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error)
}

// StreamTool 流式工具接口
type StreamTool interface {
	Tool
	ExecuteStream(ctx context.Context, args json.RawMessage, callback func(content string, index int)) (json.RawMessage, error)
}

// StreamCallback 流式回调函数类型
type StreamCallback func(content string, index int)

// ToolInfo 工具信息结构
type ToolInfo struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Category    ToolCategory `json:"category"`
	Enabled     bool         `json:"enabled"`
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
func NewToolManager(logger *logger.Logger, toolConfig *config.ToolManagerConfig) *ToolManager {
	tm := &ToolManager{
		categories: make(map[ToolCategory]*CategoryManager),
		logger:     logger,
	}

	// 使用配置初始化分类
	tm.initCategoriesFromConfig(toolConfig)

	return tm
}

// initCategoriesFromConfig 从配置初始化分类
func (tm *ToolManager) initCategoriesFromConfig(toolConfig *config.ToolManagerConfig) {
	categoryMapping := map[string]ToolCategory{
		"math":    CategoryMath,
		"ai":      CategoryAI,
		"system":  CategorySystem,
		"utility": CategoryUtility,
	}

	for configCategory, configData := range toolConfig.Categories {
		category, exists := categoryMapping[configCategory]
		if !exists {
			// 未知分类跳过
			tm.logger.Warn().Str("category", configCategory).Msg("Unknown category in config, skipping")
			continue
		}

		tm.categories[category] = &CategoryManager{
			Name:    category,
			tools:   make(map[string]Tool),
			enabled: configData.Enabled,
			config: CategoryConfig{
				Enabled:   configData.Enabled,
				MaxTools:  configData.MaxTools,
				RateLimit: configData.RateLimit,
				Timeout:   configData.Timeout,
			},
		}
	}

	// 确保所有预定义分类都有管理器
	for _, category := range []ToolCategory{CategoryMath, CategoryAI, CategorySystem, CategoryUtility} {
		if _, exists := tm.categories[category]; !exists {
			tm.categories[category] = &CategoryManager{
				Name:    category,
				tools:   make(map[string]Tool),
				enabled: false, // 默认禁用
				config:  CategoryConfig{},
			}
		}
	}
}

// RegisterTool 注册工具到指定分类
func (tm *ToolManager) RegisterTool(tool Tool) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	category := tool.Category()
	categoryMgr, exists := tm.categories[category]
	if !exists {
		return fmt.Errorf("category not found: %s", category)
	}

	if !categoryMgr.enabled {
		return fmt.Errorf("category is disabled: %s", category)
	}

	if len(categoryMgr.tools) >= categoryMgr.config.MaxTools {
		return fmt.Errorf("category %s reached maximum tools limit: %d", category, categoryMgr.config.MaxTools)
	}

	categoryMgr.tools[tool.Name()] = tool
	tm.logger.Info().
		Str("tool", tool.Name()).
		Str("category", string(category)).
		Msg("Tool registered")

	return nil
}

// RegisterAllTools 注册所有可用工具
func (tm *ToolManager) RegisterAllTools() {
	tm.RegisterTool(&CalculatorTool{})
	tm.RegisterTool(&StreamTextProcessor{})
	// 添加更多工具
}

// EnableCategory 启用分类
func (tm *ToolManager) EnableCategory(category ToolCategory) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	categoryMgr, exists := tm.categories[category]
	if !exists {
		return fmt.Errorf("category not found: %s", category)
	}

	categoryMgr.enabled = true
	tm.logger.Info().Str("category", string(category)).Msg("Category enabled")
	return nil
}

// DisableCategory 禁用分类
func (tm *ToolManager) DisableCategory(category ToolCategory) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	categoryMgr, exists := tm.categories[category]
	if !exists {
		return fmt.Errorf("category not found: %s", category)
	}

	categoryMgr.enabled = false
	tm.logger.Info().Str("category", string(category)).Msg("Category disabled")
	return nil
}

// UpdateCategoryConfig 更新分类配置
func (tm *ToolManager) UpdateCategoryConfig(category ToolCategory, config CategoryConfig) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	categoryMgr, exists := tm.categories[category]
	if !exists {
		return fmt.Errorf("category not found: %s", category)
	}

	categoryMgr.config = config
	tm.logger.Info().
		Str("category", string(category)).
		Interface("config", config).
		Msg("Category config updated")

	return nil
}

// GetTools 获取所有工具信息
func (tm *ToolManager) GetTools() []ToolInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var tools []ToolInfo
	for _, categoryMgr := range tm.categories {
		if !categoryMgr.enabled {
			continue
		}

		for _, tool := range categoryMgr.tools {
			tools = append(tools, ToolInfo{
				Name:        tool.Name(),
				Description: tool.Description(),
				Category:    tool.Category(),
				Enabled:     true,
			})
		}
	}

	return tools
}

// GetToolsByCategory 按分类获取工具信息
func (tm *ToolManager) GetToolsByCategory(category ToolCategory) []ToolInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	categoryMgr, exists := tm.categories[category]
	if !exists || !categoryMgr.enabled {
		return []ToolInfo{}
	}

	var tools []ToolInfo
	for _, tool := range categoryMgr.tools {
		tools = append(tools, ToolInfo{
			Name:        tool.Name(),
			Description: tool.Description(),
			Category:    category,
			Enabled:     true,
		})
	}

	return tools
}

// GetCategories 获取所有分类信息
func (tm *ToolManager) GetCategories() map[ToolCategory]CategoryConfig {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	categories := make(map[ToolCategory]CategoryConfig)
	for category, categoryMgr := range tm.categories {
		categories[category] = categoryMgr.config
	}

	return categories
}

// CallTool 调用工具
func (tm *ToolManager) CallTool(ctx context.Context, name string, args json.RawMessage) (*ToolCallResult, error) {
	startTime := time.Now()

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// 在所有分类中查找工具
	var tool Tool
	var category ToolCategory

	for cat, categoryMgr := range tm.categories {
		if !categoryMgr.enabled {
			continue
		}

		if t, exists := categoryMgr.tools[name]; exists {
			tool = t
			category = cat
			break
		}
	}

	if tool == nil {
		tm.logger.Error().Str("tool", name).Msg("Tool not found")
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	// 应用分类级别的超时设置
	categoryMgr := tm.categories[category]
	if categoryMgr.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, categoryMgr.config.Timeout)
		defer cancel()
	}

	// 记录工具调用开始
	tm.logger.Info().
		Str("tool", name).
		Str("category", string(category)).
		RawJSON("args", args).
		Msg("Tool call started")

	result, err := tool.Execute(ctx, args)
	duration := time.Since(startTime)

	// 记录工具调用结果
	if err != nil {
		tm.logger.Error().
			Str("tool", name).
			Str("category", string(category)).
			Dur("duration", duration).
			Err(err).
			Msg("Tool call failed")
	} else {
		tm.logger.Info().
			Str("tool", name).
			Str("category", string(category)).
			Dur("duration", duration).
			Msg("Tool call completed successfully")
	}

	if err != nil {
		return nil, err
	}

	// MCP 兼容格式
	return &ToolCallResult{
		Content: []ToolCallContent{
			{
				Type: "text",
				Text: string(result),
			},
		},
	}, nil
}

// CallToolStream 流式调用工具
func (tm *ToolManager) CallToolStream(ctx context.Context, name string, args json.RawMessage, callback StreamCallback) (*ToolCallResult, error) {
	startTime := time.Now()

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// 在所有分类中查找工具
	var tool Tool
	var category ToolCategory

	for cat, categoryMgr := range tm.categories {
		if !categoryMgr.enabled {
			continue
		}

		if t, exists := categoryMgr.tools[name]; exists {
			tool = t
			category = cat
			break
		}
	}

	if tool == nil {
		tm.logger.Error().Str("tool", name).Msg("Tool not found")
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	// 检查工具是否支持流式调用
	streamTool, supportsStream := tool.(StreamTool)
	if !supportsStream {
		tm.logger.Warn().Str("tool", name).Msg("Tool does not support streaming, returning regular execution result")
		// 若不支持流式，返回普通调用结果
		return tm.CallTool(ctx, name, args)
	}

	// 应用分类级别的超时设置
	categoryMgr := tm.categories[category]
	if categoryMgr.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, categoryMgr.config.Timeout)
		defer cancel()
	}

	// 记录流式工具调用开始
	tm.logger.Info().
		Str("tool", name).
		Str("category", string(category)).
		RawJSON("args", args).
		Msg("Stream tool call started")

	result, err := streamTool.ExecuteStream(ctx, args, callback)
	duration := time.Since(startTime)

	// 记录流式工具调用结果
	if err != nil {
		tm.logger.Error().
			Str("tool", name).
			Str("category", string(category)).
			Dur("duration", duration).
			Err(err).
			Msg("Stream tool call failed")
	} else {
		tm.logger.Info().
			Str("tool", name).
			Str("category", string(category)).
			Dur("duration", duration).
			Msg("Stream tool call completed successfully")
	}

	if err != nil {
		return nil, err
	}

	// MCP 兼容格式
	return &ToolCallResult{
		Content: []ToolCallContent{
			{
				Type: "text",
				Text: string(result),
			},
		},
	}, nil
}
