# 连接到 Centrifugo 接收流式响应

## 步骤

### 1. 发送聊天请求

```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, how are you?"}'
```

响应示例：
```json
{
  "channel_id": "chat:27531c1d-7d59-4256-a903-711c01cc0bab",
  "message": "Channel created, start listening for messages"
```

### 2. 连接到 Centrifugo 并订阅 channel

#### 方法 1: 使用 HTML 客户端（推荐）

打开 `examples/client-working.html` 在浏览器中，它会自动处理连接和订阅。

#### 方法 2: 使用 JavaScript (Centrifuge JS SDK)

```html
<!DOCTYPE html>
<html>
<head>
    <script src="https://unpkg.com/centrifuge@5.2.0/dist/centrifuge.js"></script>
</head>
<body>
    <script>
        // 连接到 Centrifugo - 使用明确的传输端点配置
        const centrifuge = new Centrifuge([
            {
                transport: 'websocket',
                endpoint: 'ws://localhost:8080/connection/websocket'
            },
            {
                transport: 'sse',
                endpoint: 'http://localhost:8080/connection/sse'
            }
        ]);
        
        // 订阅 channel（使用从 /api/chat 返回的 channel_id）
        const channelId = 'chat:27531c1d-7d59-4256-a903-711c01cc0bab';
        const sub = centrifuge.newSubscription(channelId);
        
        // 监听消息
        sub.on('publication', function(ctx) {
            const data = JSON.parse(ctx.data);
            if (data.type === 'chunk') {
                // 接收到的文本块
                console.log(data.message);
                document.body.innerHTML += data.message;
            } else if (data.type === 'end') {
                console.log('Response completed');
            } else if (data.type === 'error') {
                console.error('Error:', data.message);
            }
        });
        
        sub.on('subscribed', function() {
            console.log('Subscribed to channel:', channelId);
        });
        
        sub.on('error', function(ctx) {
            console.error('Subscription error:', ctx.error);
        });
        
        // 订阅并连接
        sub.subscribe();
        centrifuge.connect();
    </script>
</body>
</html>
```

#### 方法 3: 使用 Node.js

```javascript
const Centrifuge = require('centrifuge');

const centrifuge = new Centrifuge('ws://localhost:8080/connection/websocket');
const sub = centrifuge.newSubscription('chat:27531c1d-7d59-4256-a903-711c01cc0bab');

sub.on('publication', (ctx) => {
    const data = JSON.parse(ctx.data);
    if (data.type === 'chunk') {
        process.stdout.write(data.message);
    } else if (data.type === 'end') {
        console.log('\nResponse completed');
        process.exit(0);
    }
});

sub.subscribe();
centrifuge.connect();
```

## 消息格式

服务器会发送以下类型的消息：

- `start`: 开始响应
  ```json
  {"type": "start", "message": ""}
  ```

- `chunk`: 文本块
  ```json
  {"type": "chunk", "message": "Hello"}
  ```

- `end`: 响应结束
  ```json
  {"type": "end", "message": ""}
  ```

- `error`: 错误信息
  ```json
  {"type": "error", "message": "Error message"}
  ```

## 注意事项

1. **连接 URL**: 
   - WebSocket: `ws://localhost:8080/connection/websocket`
   - SSE: `http://localhost:8080/connection/sse`

2. **Channel ID**: 每次调用 `/api/chat` 都会返回一个新的唯一 channel ID

3. **连接时机**: 建议在收到 channel_id 后立即连接，因为服务器可能在连接建立前就开始发送消息

4. **重连**: Centrifugo JS SDK 会自动处理重连，但需要重新订阅 channel

