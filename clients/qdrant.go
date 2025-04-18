package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"matthewpsimons.com/llm-orchestrator/internal/config"
	"matthewpsimons.com/llm-orchestrator/types"
)

type QdrantSearchRequest struct {
	Vector      []float32 `json:"vector"`
	Limit       int       `json:"limit"`
	WithPayload bool      `json:"with_payload"`
}

type QdrantPoint struct {
	Payload map[string]string `json:"payload"`
	Score   float32           `json:"score"`
}

type QdrantSearchResponse []QdrantPoint

func SearchChunks(vector []float32, logger *zap.Logger) ([]types.Chunk, error) {
	payload := QdrantSearchRequest{
		Vector:      vector,
		Limit:       5,
		WithPayload: true,
	}

	data, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/collections/%s/points/search", config.AppConfig.QdrantURL, config.AppConfig.CollectionName)
	logger.Info("Qdrant search config", zap.String("url", url), zap.String("collection", config.AppConfig.CollectionName))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		logger.Error("failed to call qdrant", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("qdrant search returned non-200", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("qdrant returned status %d", resp.StatusCode)
	}

	var decoded struct {
		Result QdrantSearchResponse `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		logger.Error("failed to decode qdrant response", zap.Error(err))
		return nil, err
	}

	var chunks []types.Chunk
	for _, point := range decoded.Result {
		chunks = append(chunks, types.Chunk{
			Text:  point.Payload["text"],
			Score: point.Score,
		})
	}

	return chunks, nil
}
