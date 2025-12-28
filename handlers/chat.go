package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/binc4t/lmserver/centrifuge"
	openaiClient "github.com/binc4t/lmserver/openai"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sashabaranov/go-openai"
)

// ChatHandler 处理聊天请求
type ChatHandler struct {
	openaiClient     *openaiClient.Client
	centrifugeClient centrifuge.Client
}

// NewChatHandler 创建新的聊天处理器
func NewChatHandler(openaiClient *openaiClient.Client, centrifugeClient centrifuge.Client) *ChatHandler {
	return &ChatHandler{
		openaiClient:     openaiClient,
		centrifugeClient: centrifugeClient,
	}
}

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Message string `json:"message"`
}

// ChatResponse 聊天响应结构
type ChatResponse struct {
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
}

// HandleChat 处理聊天请求
func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	// 设置 CORS 头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	// 生成唯一的 channel ID
	channelID := fmt.Sprintf("chat:%s", uuid.New().String())

	// 返回 channel ID 给客户端
	response := ChatResponse{
		ChannelID: channelID,
		Message:   "Channel created, start listening for messages",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		return
	}

	// 在 goroutine 中处理 OpenAI 流式响应
	go h.processStreamingResponse(r.Context(), channelID, req.Message)
}

// processStreamingResponse 处理流式响应
func (h *ChatHandler) processStreamingResponse(ctx context.Context, channelID, userMessage string) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userMessage,
		},
	}

	// 发送开始消息
	if err := h.centrifugeClient.Publish(ctx, channelID, map[string]interface{}{
		"type":    "start",
		"message": "",
	}); err != nil {
		log.Printf("Failed to publish start message: %v", err)
		return
	}

	// 调用 OpenAI 流式 API
	err := h.openaiClient.StreamChat(ctx, messages, func(chunk string) error {
		// 通过 Centrifugo 发布每个响应块
		if err := h.centrifugeClient.Publish(ctx, channelID, map[string]interface{}{
			"type":    "chunk",
			"message": chunk,
		}); err != nil {
			log.Printf("Failed to publish chunk: %v", err)
			return err
		}
		return nil
	})

	if err != nil {
		log.Printf("Error processing stream: %v", err)
		// 发送错误消息
		if pubErr := h.centrifugeClient.Publish(ctx, channelID, map[string]interface{}{
			"type":    "error",
			"message": fmt.Sprintf("Error: %v", err),
		}); pubErr != nil {
			log.Printf("Failed to publish error message: %v", pubErr)
		}
		return
	}

	// 发送结束消息
	if err := h.centrifugeClient.Publish(ctx, channelID, map[string]interface{}{
		"type":    "end",
		"message": "",
	}); err != nil {
		log.Printf("Failed to publish end message: %v", err)
	}
}

// RegisterRoutes 注册路由
func (h *ChatHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/chat", h.HandleChat).Methods("POST", "OPTIONS")
}
