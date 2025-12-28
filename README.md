架构设想（可能强行引入复杂度了，但是主要是为了熟悉各个组件。。。）
- Kong：API 网关，统一入口、JWT 鉴权、限流
- Centrifugo：维持长连接，推送响应流给客户端
- Temporal：编排多步任务，包括但不限于：
    1. Zep / Redis 处理记忆问题
    2. DB 处理用户连接、用户信息等
    3. Milvus 存储长期知识
    4. 网页搜索和解析（Jina Reader API 或其他MCP服务）
    5. Reranker 搜索结果重排
- OpenTelemetry：监控