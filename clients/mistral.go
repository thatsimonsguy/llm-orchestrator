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

type ChatRequest struct {
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GenerateResponse sends the prompt to Mistral and either returns a full string or streams it if enabled
// Note: Mistral seems to ignore system messages, so all instructions and context are loaded in to userContext
func GenerateResponse(userContent string, systemInstructions string, logger *zap.Logger) (string, error) {
	payload := ChatRequest{
		Model:  config.AppConfig.MistralModel,
		Stream: config.AppConfig.MistralStream,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemInstructions,
			},
			{
				Role:    "user",
				Content: userContent,
			},
		},
	}

	data, _ := json.Marshal(payload)

	resp, err := http.Post(config.AppConfig.MistralServiceURL+"/v1/chat/completions", "application/json", bytes.NewBuffer(data))
	if err != nil {
		logger.Error("failed to call mistral model", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("mistral returned non-200", zap.Int("status", resp.StatusCode), zap.ByteString("body", body))
		return "", fmt.Errorf("mistral failed with status %d", resp.StatusCode)
	}

	if !config.AppConfig.MistralStream {
		var result struct {
			Choices []struct {
				Message Message `json:"message"`
			} `json:"choices"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			logger.Error("failed to decode mistral response", zap.Error(err))
			return "", err
		}
		if len(result.Choices) == 0 {
			return "", fmt.Errorf("mistral returned no choices")
		}
		return result.Choices[0].Message.Content, nil
	}

	// Handle streaming response
	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")
			if line == "[DONE]" {
				break
			}

			var delta struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(line), &delta); err != nil {
				logger.Warn("failed to unmarshal streaming chunk", zap.String("line", line), zap.Error(err))
				continue
			}
			if len(delta.Choices) > 0 {
				fullContent.WriteString(delta.Choices[0].Delta.Content)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Error("error reading mistral streaming response", zap.Error(err))
		return "", err
	}

	return fullContent.String(), nil
}

func GenerateResponseStream(userContent string, systemInstructions string, logger *zap.Logger, onChunk func(string)) error {
	payload := ChatRequest{
		Model:  config.AppConfig.MistralModel,
		Stream: true,
		Messages: []Message{
			{Role: "system", Content: systemInstructions},
			{Role: "user", Content: userContent},
		},
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(config.AppConfig.MistralServiceURL+"/v1/chat/completions", "application/json", bytes.NewBuffer(data))
	if err != nil {
		logger.Error("failed to call mistral model", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("mistral returned non-200", zap.Int("status", resp.StatusCode), zap.ByteString("body", body))
		return fmt.Errorf("mistral failed with status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")
			if line == "[DONE]" {
				break
			}

			var delta struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(line), &delta); err != nil {
				logger.Warn("failed to parse chunk", zap.String("line", line), zap.Error(err))
				continue
			}

			if len(delta.Choices) > 0 {
				chunk := delta.Choices[0].Delta.Content
				onChunk(chunk)
			}
		}
	}
	return scanner.Err()
}
