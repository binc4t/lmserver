package centrifuge

import (
	"context"
)

// Client 接口定义，用于向后兼容
// 现在使用 Server 作为实现
type Client interface {
	Publish(ctx context.Context, channel string, data interface{}) error
}

// ClientWrapper 包装 Server 以实现 Client 接口
type ClientWrapper struct {
	server *Server
}

// NewClientWrapper 创建客户端包装器
func NewClientWrapper(server *Server) Client {
	return &ClientWrapper{server: server}
}

// Publish 发布消息
func (c *ClientWrapper) Publish(ctx context.Context, channel string, data interface{}) error {
	return c.server.Publish(ctx, channel, data)
}
