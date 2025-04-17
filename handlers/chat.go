package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"matthewpsimons.com/llm-orchestrator/internal/logging"
	"matthewpsimons.com/llm-orchestrator/types"
)

var log = logging.Logger

func HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Warn("Invalid request method", zap.String("method", r.Method))
		http.Error(w, "Only POST supported", http.StatusMethodNotAllowed)
		return
	}

	var chatReq types.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		log.Warn("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Info("Received chat request",
		zap.String("user_id", chatReq.UserID),
		zap.String("query", chatReq.Query),
	)

	// Placeholder response
	response := types.ChatChunkResponse{
		Response: "This is a stub. Real response coming soon.",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("Failed to write response", zap.Error(err))
		http.Error(w, "Failed to respond", http.StatusInternalServerError)
	}
}
