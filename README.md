# PaiChat-Golang+Ollama LLM Service

<div align="center">  

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)  

</div>  

---

## 📂 项目结构  

```plaintext
PaiChat/
├── cmd/                 # 主程序入口
│   └── main.go          # 程序启动入口
├── internal/            # 内部核心模块
│   ├── chat/            # 聊天核心逻辑
│   │   ├── chat.go      # 聊天会话实现
│   │   ├── pool.go      # LLM连接池实现
│   │   └── service.go   # 聊天服务接口
│   ├── config/          # 配置管理
│   │   └── config.go    # 配置加载实现
│   ├── models/          # 数据模型
│   │   └── models.go    # API数据结构定义
│   └── server/          # Web服务
│       └── server.go    # HTTP服务器实现
├── static/              # 静态资源
│   ├── css/             
│   ├── js/              
│   └── index.html       
├── go.mod               # Go模块定义
├── go.sum               # 依赖校验文件
└── config.json          # 配置文件(可选)
```

---

## 🌟 功能特性
- **多模型对话** / Multi-model Chat  
  `DeepSeek` `llama2` `mistral` `gemma` 等Ollama官方模型支持 
- **双模交互** / Dual Interface  
  CLI命令行 + Web图形界面
- **智能管理** / Smart Management  
  LLM实例连接池 | 历史记录检索 | 上下文保持 | 完善日志

---

## Ollama + AstrBot 多平台LLM开发应用


---

## 🚀 快速开始
### 环境要求
- Go 1.20+
- Ollama服务(本地或远程)
- Ollama部署
```bash
# Download:  https://ollama.com/download
# Windows 
OllamaSetup.exe /DIR=D:\xxx  //修改默认安装目录
ollama -v     //查看版本
ollama serve  //启动服务
ollama list   //本地模型
ollama pull deepseek-r1   //拉取模型(若中途中断，尝试多次拉取)
ollama run deepseek-r1    //运行模型
ollama ps                 //查看运行中模型
ollama rm model_name      //删除模型
ollama show model_name    //查看模型详细信息
ollama stop model_name    //停止运行中模型

# Linux
# Ollama提供了一键安装脚本,可以快速在Linux系统上安装Ollama
curl -fsSL https://ollama.com/install.sh | sh
ollama serve   
```
可直接终端运行：
![image](https://github.com/user-attachments/assets/8adf4ce1-7203-48f1-bc8f-562ae40ee7de)

### 启动方式 
```bash
git clone https://github.com/liaotxcn/PaiChat.git  # 克隆仓库
```
### 命令行模式
```bash
go run cmd/main.go
```
### 交互支持：
- help 查看帮助
- exit 退出程序
- clear 清空记录
- history 查看历史对话

### Web服务(默认自启动)，访问 https://localhost:8080

### 研发中，持续更新...
