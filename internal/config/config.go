package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	EmbedServiceURL   string
	QdrantURL         string
	MistralServiceURL string
	MistralModel      string
	MistralStream     bool
	CollectionName    string

	UseOpenAI   bool
	OpenAIKey   string
	OpenAIModel string
}

var AppConfig Config

func Load() {
	// Load .env file if present
	err := godotenv.Load()
	if err != nil {
		log.Println("[INFO] No .env file found, using environment variables")
	}

	AppConfig = Config{
		Port:              getEnv("PORT", "8080"),
		EmbedServiceURL:   getEnv("EMBED_SERVICE_URL", "http://embedding-service.llm.svc.cluster.local"),
		QdrantURL:         getEnv("QDRANT_URL", "http://qdrant.llm.svc.cluster.local:6333"),
		MistralServiceURL: getEnv("MISTRAL_SERVICE_URL", "http://ollama.llm.svc.cluster.local:11434"),
		MistralModel:      getEnv("MISTRAL_MODEL", "mistral"),
		MistralStream:     getEnv("MISTRAL_STREAM", "true") == "true",
		CollectionName:    getEnv("QDRANT_COLLECTION", "matt-chunks"),

		UseOpenAI:   getEnv("USE_OPENAI", "false") == "true",
		OpenAIKey:   getEnv("OPENAI_API_KEY", ""), // required if using openai
		OpenAIModel: getEnv("OPENAI_MODEL", "gpt-4o"),
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Printf("[WARN] %s not set, defaulting to %s\n", key, fallback)
		return fallback
	}
	return val
}
