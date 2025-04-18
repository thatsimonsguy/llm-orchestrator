package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
	"matthewpsimons.com/llm-orchestrator/internal/config"
)

type EmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

func EmbedText(input string, logger *zap.Logger) ([]float32, error) {
	payload := map[string]string{
		"text": input,
	}

	data, _ := json.Marshal(payload)

	resp, err := http.Post(config.AppConfig.EmbedServiceURL+"/api/v1/embed", "application/json", bytes.NewBuffer(data))
	if err != nil {
		logger.Error("failed to call embed service", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("embed service returned non-200", zap.Int("status", resp.StatusCode), zap.ByteString("body", body))
		return nil, fmt.Errorf("embed service failed with status %d", resp.StatusCode)
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Error("failed to decode embed response", zap.Error(err))
		return nil, err
	}

	if len(result.Embedding) != 768 {
		logger.Warn("unexpected embedding size", zap.Int("dims", len(result.Embedding)))
		return nil, fmt.Errorf("invalid embedding vector")
	}

	return result.Embedding, nil
}
