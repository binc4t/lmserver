package openai

import (
	"context"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

// Client 封装 OpenAI 客户端
type Client struct {
	client *openai.Client
}

// NewClient 创建新的 OpenAI 客户端
func NewClient(apiKey string) *Client {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	client := openai.NewClientWithConfig(config)

	return &Client{
		client: client,
	}
}

// StreamChat 流式调用 OpenAI Chat API
func (c *Client) StreamChat(ctx context.Context, messages []openai.ChatCompletionMessage, onChunk func(string) error) error {
	req := openai.ChatCompletionRequest{
		Model:     "ep-m-20251228184628-lfn92", // kimi-k2-250905
		Messages:  messages,
		Stream:    true,
		MaxTokens: 1000,
	}

	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create chat completion stream: %w", err)
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive stream response: %w", err)
		}

		if len(response.Choices) > 0 {
			delta := response.Choices[0].Delta.Content
			if delta != "" {
				if err := onChunk(delta); err != nil {
					return fmt.Errorf("failed to process chunk: %w", err)
				}
			}
		}
	}

	return nil
}
