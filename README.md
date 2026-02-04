# Weave-Toolkit

<div align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/Tool-MCP-FF6F00?style=for-the-badge&logo=tool&logoColor=white" alt="Tool MCP">
</div>

基于 Golang 高性能 MCP (Model Context Protocol) 工具服务器，可为 [Weave](https://github.com/liaotxcn/Weave) 平台提供可扩展的工具服务

---

## 🚀 特性

- **MCP 协议支持** 
- **工具分类管理** 
- **配置驱动** 
- **高性能** 
- **结构化日志** 

---

## 📦 快速开始

### 环境要求

- Go 1.24+
- 支持 MCP 协议的客户端

### 安装运行

```bash
# 克隆项目
git clone https://github.com/NexusForgeAI/Weave-Toolkit.git
cd Weave-Toolkit

# 配置环境变量(.env)、工具配置(tool-config.json)

# 启动服务
go run ./cmd/mcp-server
```

---

## 🔧 工具集成

### 添加新工具

1. 在 `internal/tools/` 目录创建新工具文件
2. 实现 `Tool` 接口：

```go
type Tool interface {
    Name() string
    Description() string 
    Category() ToolCategory
    Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error)
}
```

3. 在 `manager.go` 中注册工具

## 🌐 接口

### MCP 协议端点

- `POST /mcp` - MCP 协议主端点
- `GET /health` - 健康检查端点

### 协议方法

- `initialize` - 初始化连接
- `tools/list` - 获取可用工具列表
- `tools/call` - 调用具体工具

### 项目结构

```
Weave-Toolkit/
├── cmd/mcp-server/     # 启动入口
├── config/             # 配置管理
├── internal/           # 核心实现
│   ├── logger/         # 日志系统
│   ├── mcp/            # MCP 协议
│   └── tools/          # 工具管理
├── .env                # 环境配置
└── tool-config.json    # 工具配置
```

### 构建部署

```bash
# 构建
go build -o mcp-server ./cmd/mcp-server

# 运行
./mcp-server

# Docker 运行
docker build -t mcp-server:1.0.0 .
docker run -p 8888:8888 -v $(pwd)/.env:/app/.env -v $(pwd)/tool-config.json:/app/tool-config.json mcp-server:1.0.0
```

## 🤝 贡献指南

欢迎对项目进行贡献！感谢！

1. **Fork 仓库**并克隆到本地
2. **创建分支**进行开发（`git checkout -b feature/your-feature`）
3. **提交代码**并确保通过测试
4. **创建 Pull Request** 描述您的更改
5. 等待**代码审查**并根据反馈进行修改

---

### <div align="center"> <strong>✨ 持续更新完善中... ✨</strong> </div>
