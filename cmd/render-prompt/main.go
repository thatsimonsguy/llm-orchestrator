package main

import (
	"flag"
	"fmt"
	"log"

	"go.uber.org/zap"

	"matthewpsimons.com/llm-orchestrator/clients"
	"matthewpsimons.com/llm-orchestrator/internal/config"
	"matthewpsimons.com/llm-orchestrator/internal/logging"
	"matthewpsimons.com/llm-orchestrator/internal/promptbuilder"
)

func main() {
	query := flag.String("query", "", "User query to generate prompt for")
	flag.Parse()

	if *query == "" {
		log.Fatal("Missing --query input")
	}

	config.Load()
	logging.Init()
	defer logging.Logger.Sync()
	logger := logging.Logger

	instructions, err := promptbuilder.BuildSystemInstructions("internal/promptbuilder/prompt_instructions.txt")
	if err != nil {
		logger.Fatal("Failed to load instructions", zap.Error(err))
	}

	canonicalData, err := promptbuilder.LoadCanonicalData("data")
	if err != nil {
		logger.Fatal("Failed to load canonical data", zap.Error(err))
	}

	embedding, err := clients.EmbedText(*query, logger)
	if err != nil {
		logger.Fatal("Failed to embed query", zap.Error(err))
	}

	chunks, err := clients.SearchChunks(embedding, logger)
	if err != nil {
		logger.Fatal("Failed to fetch chunks", zap.Error(err))
	}

	var chunkTexts []string
	for _, chunk := range chunks {
		chunkTexts = append(chunkTexts, chunk.Text)
	}

	prompt := promptbuilder.BuildUserPrompt(*query, chunkTexts, canonicalData, instructions)

	fmt.Println(prompt)
}
