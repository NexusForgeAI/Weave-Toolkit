package mcp

import "encoding/json"

// MCP 协议版本
const (
	ProtocolVersion = "2025-06-18"
)

// MCP 请求类型
const (
	MethodInitialize     = "initialize"
	MethodToolsList      = "tools/list"
	MethodToolsCall      = "tools/call"
	MethodResourcesList  = "resources/list"
	MethodResourcesRead  = "resources/read"
	MethodPromptsList    = "prompts/list"
	MethodPromptsGet     = "prompts/get"
	MethodRootsList      = "roots/list"
)

// MCP 流式响应相关常量
const (
	StreamEventToolCall = "tool/call"
	StreamEventContent  = "content"
	StreamEventDone     = "done"
	StreamEventError    = "error"
)

// 流式响应内容类型
const (
	ContentTypeText = "text"
	ContentTypeJSON = "json"
)

// InitializeRequest 初始化请求
type InitializeRequest struct {
	Method string `json:"method"`
	Params struct {
		ProtocolVersion string `json:"protocolVersion"`
		ClientInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"clientInfo"`
		Capabilities struct {
			Tools struct {
			} `json:"tools"`
		} `json:"capabilities"`
	} `json:"params"`
}

// InitializeResponse 初始化响应
type InitializeResponse struct {
	Result struct {
		ProtocolVersion string `json:"protocolVersion"`
		ServerInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
		Capabilities struct {
			Tools struct {
				ListChanged bool `json:"listChanged"`
			} `json:"tools"`
			Resources struct {
				ListChanged bool `json:"listChanged"`
				Subscribe   bool `json:"subscribe"`
			} `json:"resources"`
			Prompts struct {
				ListChanged bool `json:"listChanged"`
			} `json:"prompts"`
			Roots struct {
				ListChanged bool `json:"listChanged"`
			} `json:"roots"`
		} `json:"capabilities"`
	} `json:"result"`
}

// ToolsListRequest 工具列表请求
type ToolsListRequest struct {
	Method string `json:"method"`
}

// ToolsListResponse 工具列表响应
type ToolsListResponse struct {
	Result struct {
		Tools []ToolInfo `json:"tools"`
	} `json:"result"`
}

// ToolsCallRequest 工具调用请求
type ToolsCallRequest struct {
	Method string `json:"method"`
	Params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"params"`
}

// ToolsCallResponse 工具调用响应
type ToolsCallResponse struct {
	Result ToolCallResult `json:"result"`
}

// ToolInfo 工具信息
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
	Type string `json:"type"`
	Text string `json:"text"`
}

// ResourcesListRequest 资源列表请求
type ResourcesListRequest struct {
	Method string `json:"method"`
}

// ResourcesListResponse 资源列表响应
type ResourcesListResponse struct {
	Result struct {
		Resources []ResourceInfo `json:"resources"`
	} `json:"result"`
}

// ResourcesReadRequest 资源读取请求
type ResourcesReadRequest struct {
	Method string `json:"method"`
	Params struct {
		URI string `json:"uri"`
	} `json:"params"`
}

// ResourcesReadResponse 资源读取响应
type ResourcesReadResponse struct {
	Result struct {
		Contents []ResourceContent `json:"contents"`
	} `json:"result"`
}

// PromptsListRequest 提示词列表请求
type PromptsListRequest struct {
	Method string `json:"method"`
}

// PromptsListResponse 提示词列表响应
type PromptsListResponse struct {
	Result struct {
		Prompts []PromptInfo `json:"prompts"`
	} `json:"result"`
}

// PromptsGetRequest 提示词获取请求
type PromptsGetRequest struct {
	Method string `json:"method"`
	Params struct {
		Name string `json:"name"`
	} `json:"params"`
}

// PromptsGetResponse 提示词获取响应
type PromptsGetResponse struct {
	Result struct {
		Description string                   `json:"description"`
		Arguments   []PromptArgument         `json:"arguments"`
	} `json:"result"`
}

// RootsListRequest 根目录列表请求
type RootsListRequest struct {
	Method string `json:"method"`
}

// RootsListResponse 根目录列表响应
type RootsListResponse struct {
	Result struct {
		Roots []RootInfo `json:"roots"`
	} `json:"result"`
}

// ResourceInfo 资源信息
type ResourceInfo struct {
	URI      string `json:"uri"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	Description string `json:"description,omitempty"`
}

// ResourceContent 资源内容
type ResourceContent struct {
	URI     string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text    string `json:"text"`
}

// PromptInfo 提示词信息
type PromptInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PromptArgument 提示词参数
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// RootInfo 根目录信息
type RootInfo struct {
	URI string `json:"uri"`
}

// StreamToolCallRequest 流式工具调用请求
type StreamToolCallRequest struct {
	Method string `json:"method"`
	Params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
		Stream    bool            `json:"stream,omitempty"` // 是否启用流式响应
	} `json:"params"`
}

// StreamEvent 流式事件
type StreamEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// StreamContent 流式内容
type StreamContent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Index   int    `json:"index,omitempty"` // 内容索引，用于标识顺序
}

// StreamToolCallResult 流式工具调用结果
type StreamToolCallResult struct {
	Content  []StreamContent `json:"content"`
	IsStream bool            `json:"is_stream"`
}
