package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"matthewpsimons.com/llm-orchestrator/clients"
	"matthewpsimons.com/llm-orchestrator/internal/promptbuilder"
	"matthewpsimons.com/llm-orchestrator/types"
	"matthewpsimons.com/llm-orchestrator/internal/config"
)

var (
	systemInstructions string
	canonicalData      map[string]string
)

func InitSystemPrompt(logger *zap.Logger) {
	var err error
	systemInstructions, err = promptbuilder.BuildSystemInstructions("internal/promptbuilder/prompt_instructions.txt")
	if err != nil {
		logger.Fatal("Failed to load system instructions", zap.Error(err))
	}

	canonicalData, err = promptbuilder.LoadCanonicalData("data")
	if err != nil {
		logger.Fatal("Failed to load canonical data", zap.Error(err))
	}
}

func HandleChat(cfg config.Config, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://matthewpsimons.com")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

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

		var chunkTexts []string
		for _, chunk := range chunks {
			chunkTexts = append(chunkTexts, chunk.Text)
		}

		userPrompt := promptbuilder.BuildUserPrompt(chatReq.Query, chunkTexts, canonicalData, systemInstructions)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		if cfg.UseOpenAI {
			logger.Info("Using OpenAI model")
			err = clients.GenerateOpenAIResponseStream(cfg, userPrompt, "", logger, func(chunk string) {
				fmt.Fprintf(w, "data: %s\n\n", chunk)
				flusher.Flush()
			})
		} else {
			logger.Info("Using Mistral model")
			err = clients.GenerateResponseStream(userPrompt, "", logger, func(chunk string) {
				fmt.Fprintf(w, "data: %s\n\n", chunk)
				flusher.Flush()
			})
		}

		if err != nil {
			logger.Error("streaming response failed", zap.Error(err))
		}
	}
}
