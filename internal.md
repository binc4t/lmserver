# LMServer - LLM 流式响应服务

基于 Go 和 Centrifugo 实现的 LLM 流式响应 MVP，接收用户请求并转发给大模型，通过 SSE 将响应流式发送给客户端。

## MVP 功能

- 接收用户 HTTP 请求
- 将请求转发给 OpenAI API（流式模式）
- 通过 Centrifugo 将响应流式推送给客户端（SSE）

## 快速开始

### 前置要求

1. Go 1.19+
2. OpenAI API Key

**注意**：Centrifugo 以嵌入式模式运行，不需要单独的 Centrifugo 服务器进程。

### 安装和运行

1. 克隆项目并安装依赖：
```bash
go mod download
```

2. 配置环境变量：
```bash
export OPENAI_API_KEY=your_openai_api_key
export PORT=8080
export CENTRIFUGO_PORT=8000  # 可选，Centrifugo 连接端口（默认 8000）
```

3. 启动服务：
```bash
go run main.go
```

服务启动后，Centrifugo 会自动嵌入在同一个进程中，客户端可以通过以下端点连接：
- WebSocket: `ws://localhost:8080/connection/websocket`
- SSE: `http://localhost:8080/connection/sse`

### API 使用

1. 发送聊天请求：
```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, how are you?"}'
```

响应示例：
```json
{
  "channel_id": "chat:uuid-here",
  "message": "Channel created, start listening for messages"
}
```

2. 客户端通过 WebSocket 或 SSE 连接到服务（`/connection/websocket` 或 `/connection/sse`），订阅返回的 `channel_id`，接收流式响应。

### 客户端示例

查看 `examples/client.html` 获取完整的 HTML/JavaScript 客户端示例。

## 项目结构

```
lmserver/
├── main.go              # 主服务入口
├── config/              # 配置管理
│   └── config.go
├── handlers/            # HTTP 处理器
│   └── chat.go
├── centrifuge/          # Centrifugo 集成
│   └── client.go
├── openai/              # OpenAI 集成
│   └── client.go
├── examples/            # 客户端示例
│   ├── client.html
│   └── client-simple.html
└── README.md
```

## 环境变量

- `OPENAI_API_KEY`: OpenAI API 密钥（必需）
- `PORT`: 服务端口（默认: 8080）
- `CENTRIFUGO_PORT`: Centrifugo 连接端口（可选，默认: 8000，仅用于日志显示）

## 开发

```bash
# 运行测试
go test ./...

# 构建
go build -o lmserver

# 运行
./lmserver
```

## 架构说明

### 数据流

1. 客户端发送 POST 请求到 `/api/chat`，包含消息内容
2. Go 服务生成唯一的 channel ID
3. 服务返回 channel ID 和连接信息
4. 客户端通过 SSE 连接到 Centrifugo，订阅该 channel
5. Go 服务调用 OpenAI API（流式模式）
6. 每收到一个响应块，通过 Centrifugo 发布到 channel
7. 客户端通过 SSE 实时接收响应块

### 核心组件说明

- **Centrifugo 服务器** (`centrifuge/server.go`): 嵌入式 Centrifugo 服务器，在同一进程中运行
- **Centrifugo 客户端包装** (`centrifuge/client.go`): 封装发布消息的接口
- **OpenAI 客户端** (`openai/client.go`): 流式调用 OpenAI API
- **HTTP 处理器** (`handlers/chat.go`): 处理聊天请求，协调 OpenAI 和 Centrifugo
- **配置管理** (`config/config.go`): 从环境变量加载配置

### 架构优势

使用嵌入式模式的优势：
- **简化部署**：不需要单独的 Centrifugo 进程
- **更高性能**：无需网络调用，直接内存通信
- **统一管理**：所有服务在一个进程中，便于监控和日志

