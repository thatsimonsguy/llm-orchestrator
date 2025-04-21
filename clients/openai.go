package clients

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"matthewpsimons.com/llm-orchestrator/internal/config"
)

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIChatRequest struct {
	Model         string          `json:"model"`
	Stream        bool            `json:"stream"`
	Messages      []OpenAIMessage `json:"messages"`
	StreamOptions map[string]any  `json:"stream_options,omitempty"` // usage stats if supported
}

type OpenAIStreamDelta struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content,omitempty"`
		} `json:"delta"`
	} `json:"choices"`
}

type OpenAIStreamUsage struct {
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func GenerateOpenAIResponseStream(cfg config.Config, userContent, systemInstructions string, logger *zap.Logger, onChunk func(string)) error {
	apiKey := cfg.OpenAIKey
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY not set in config")
	}

	model := cfg.OpenAIModel

	payload := OpenAIChatRequest{
		Model:  model,
		Stream: true,
		Messages: []OpenAIMessage{
			{Role: "system", Content: systemInstructions},
			{Role: "user", Content: userContent},
		},
		StreamOptions: map[string]any{
			"include_usage": true,
		},
	}

	data, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("failed to call OpenAI API", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("openai returned non-200", zap.Int("status", resp.StatusCode), zap.ByteString("body", body))
		return fmt.Errorf("openai returned status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")
			if line == "[DONE]" {
				break
			}

			// Try to decode regular delta message
			var delta OpenAIStreamDelta
			if err := json.Unmarshal([]byte(line), &delta); err == nil && len(delta.Choices) > 0 {
				chunk := delta.Choices[0].Delta.Content
				if chunk != "" {
					onChunk(chunk)
				}
				continue
			}

			// Try to decode final usage stats chunk
			var usage OpenAIStreamUsage
			if err := json.Unmarshal([]byte(line), &usage); err == nil && usage.Usage.TotalTokens > 0 {
				logger.Info("OpenAI usage stats",
					zap.Int("prompt_tokens", usage.Usage.PromptTokens),
					zap.Int("completion_tokens", usage.Usage.CompletionTokens),
					zap.Int("total_tokens", usage.Usage.TotalTokens),
				)
				continue
			}

			// If unrecognized
			logger.Warn("Unrecognized stream chunk", zap.String("raw", line))
		}
	}

	return scanner.Err()
}
