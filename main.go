package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/binc4t/lmserver/centrifuge"
	"github.com/binc4t/lmserver/config"
	"github.com/binc4t/lmserver/handlers"
	"github.com/binc4t/lmserver/openai"
	centrifugeLib "github.com/centrifugal/centrifuge"
	"github.com/gorilla/mux"
)

// corsMiddleware 添加 CORS 头到响应
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

		// 处理 OPTIONS 预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 调用下一个处理器
		next.ServeHTTP(w, r)
	})
}

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化嵌入式 Centrifugo 服务器
	centrifugeServer, err := centrifuge.NewServer()
	if err != nil {
		log.Fatalf("Failed to create Centrifugo server: %v", err)
	}
	defer centrifugeServer.Shutdown(context.Background())

	// 创建 Centrifugo 客户端包装器
	centrifugeClient := centrifuge.NewClientWrapper(centrifugeServer)

	// 初始化 OpenAI 客户端
	openaiClient := openai.NewClient(cfg.OpenAIAPIKey)

	// 创建处理器
	chatHandler := handlers.NewChatHandler(openaiClient, centrifugeClient)

	// 设置路由
	r := mux.NewRouter()
	chatHandler.RegisterRoutes(r)

	// 添加健康检查端点
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// 集成 Centrifugo WebSocket/SSE 端点
	centrifugeNode := centrifugeServer.Node()

	// WebSocket handler - 允许所有来源（开发环境）
	wsHandler := centrifugeLib.NewWebsocketHandler(centrifugeNode, centrifugeLib.WebsocketConfig{
		CheckOrigin: func(r *http.Request) bool {
			// 开发环境允许所有来源
			return true
		},
	})
	r.Handle("/connection/websocket", wsHandler)

	// SSE handler - 添加 CORS 支持
	sseHandler := centrifugeLib.NewSSEHandler(centrifugeNode, centrifugeLib.SSEConfig{})
	r.Handle("/connection/sse", corsMiddleware(sseHandler))

	// Emulation handler - SSE 传输需要此端点来处理客户端到服务器的消息
	emulationHandler := centrifugeLib.NewEmulationHandler(centrifugeNode, centrifugeLib.EmulationConfig{})
	r.Handle("/emulation", corsMiddleware(emulationHandler))

	// 启动 HTTP 服务器
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("Centrifugo embedded (WebSocket: ws://localhost%s/connection/websocket, SSE: http://localhost%s/connection/sse)", addr, addr)

	// 优雅关闭
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
