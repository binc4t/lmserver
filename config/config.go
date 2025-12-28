package config

import (
	"fmt"
	"os"
)

// Config 应用配置
type Config struct {
	OpenAIAPIKey   string
	Port           string
	CentrifugePort string
}

// Load 从环境变量加载配置
func Load() (*Config, error) {
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	centrifugePort := os.Getenv("CENTRIFUGO_PORT")
	if centrifugePort == "" {
		centrifugePort = "8000"
	}

	return &Config{
		OpenAIAPIKey:   openaiAPIKey,
		Port:           port,
		CentrifugePort: centrifugePort,
	}, nil
}
