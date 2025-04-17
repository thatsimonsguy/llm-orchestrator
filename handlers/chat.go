package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"matthewpsimons.com/llm-orchestrator/clients"
	"matthewpsimons.com/llm-orchestrator/types"
)

func HandleChat(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			logger.Warn("Invalid request method", zap.String("method", r.Method))
			http.Error(w, "Only POST supported", http.StatusMethodNotAllowed)
			return
		}

		var chatReq types.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
			logger.Warn("Failed to decode request body", zap.Error(err))
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		logger.Info("Received chat request",
			zap.String("user_id", chatReq.UserID),
			zap.String("query", chatReq.Query),
		)

		embedding, err := clients.EmbedText(chatReq.Query, logger)
		if err != nil {
			logger.Error("Embedding failed", zap.Error(err))
			http.Error(w, "Embedding failed", http.StatusInternalServerError)
			return
		}

		chunks, err := clients.SearchChunks(embedding, logger)
		if err != nil {
			logger.Error("Qdrant search failed", zap.Error(err))
			http.Error(w, "Vector search failed", http.StatusInternalServerError)
			return
		}

		logger.Info("Retrieved chunks from Qdrant", zap.Int("count", len(chunks)))

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(chunks); err != nil {
			logger.Error("Failed to write response", zap.Error(err))
			http.Error(w, "Failed to respond", http.StatusInternalServerError)
		}
	}
}
