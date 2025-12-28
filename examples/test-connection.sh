#!/bin/bash

# 测试脚本：发送请求并连接到 Centrifugo 接收消息

API_URL="http://localhost:8080"
CENTRIFUGO_URL="ws://localhost:8080/connection/websocket"

echo "1. Sending chat request..."
RESPONSE=$(curl -s -X POST "$API_URL/api/chat" \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, how are you?"}')

echo "Response: $RESPONSE"

# 提取 channel_id
CHANNEL_ID=$(echo $RESPONSE | grep -o '"channel_id":"[^"]*' | cut -d'"' -f4)

if [ -z "$CHANNEL_ID" ]; then
    echo "Error: Could not extract channel_id"
    exit 1
fi

echo ""
echo "2. Channel ID: $CHANNEL_ID"
echo ""
echo "3. To receive messages, you can:"
echo "   - Use the HTML client: open examples/client-working.html in a browser"
echo "   - Use Centrifuge JS SDK in your own code"
echo "   - Connect via WebSocket client to: $CENTRIFUGO_URL"
echo "   - Subscribe to channel: $CHANNEL_ID"
echo ""
echo "Example JavaScript code:"
echo "---"
echo "const centrifuge = new Centrifuge('$CENTRIFUGO_URL');"
echo "const sub = centrifuge.newSubscription('$CHANNEL_ID');"
echo "sub.on('publication', (ctx) => {"
echo "  const data = JSON.parse(ctx.data);"
echo "  if (data.type === 'chunk') {"
echo "    console.log(data.message);"
echo "  }"
echo "});"
echo "sub.subscribe();"
echo "centrifuge.connect();"
echo "---"

