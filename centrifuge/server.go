package centrifuge

import (
	"context"
	"encoding/json"
	"fmt"

	centrifugeLib "github.com/centrifugal/centrifuge"
)

// Server 封装嵌入式 Centrifugo 服务器
type Server struct {
	node *centrifugeLib.Node
}

// NewServer 创建新的嵌入式 Centrifugo 服务器
func NewServer() (*Server, error) {
	cfg := centrifugeLib.Config{}

	// 创建节点
	node, err := centrifugeLib.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create centrifuge node: %w", err)
	}

	// 配置连接处理 - 允许匿名连接（MVP 简化版本）
	node.OnConnecting(func(ctx context.Context, event centrifugeLib.ConnectEvent) (centrifugeLib.ConnectReply, error) {
		// 允许所有连接，返回匿名用户凭证
		return centrifugeLib.ConnectReply{
			Credentials: &centrifugeLib.Credentials{
				UserID: "anonymous",
			},
		}, nil
	})

	// 启动节点
	if err := node.Run(); err != nil {
		return nil, fmt.Errorf("failed to run centrifuge node: %w", err)
	}

	return &Server{
		node: node,
	}, nil
}

// Publish 发布消息到指定的 channel
func (s *Server) Publish(ctx context.Context, channel string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	_, err = s.node.Publish(channel, payload)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Node 返回 Centrifuge 节点，用于集成到 HTTP 服务器
func (s *Server) Node() *centrifugeLib.Node {
	return s.node
}

// Shutdown 关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	return s.node.Shutdown(ctx)
}
