package main

import (
	"net/http"

	"go.uber.org/zap"
	"matthewpsimons.com/llm-orchestrator/internal/config"
	"matthewpsimons.com/llm-orchestrator/internal/logging"

	"matthewpsimons.com/llm-orchestrator/handlers"
)

func main() {
	config.Load()
	logging.Init()
	defer logging.Logger.Sync()

	log := logging.Logger
	handlers.InitSystemPrompt(log)

	http.HandleFunc("/api/v1/chat", handlers.HandleChat(logging.Logger))

	log.Info("Server starting", zap.String("addr", ":8080"))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed", zap.Error(err))
	}
}
