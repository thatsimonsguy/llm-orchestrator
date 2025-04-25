package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	EmbedServiceURL string

	QdrantURL      string
	CollectionName string

	MistralServiceURL string
	MistralModel      string
	MistralStream     bool

	UseOpenAI   bool
	OpenAIKey   string
	OpenAIModel string

	DevCORSSecret string
}

var AppConfig Config

func Load() {
	// Load .env file if present
	err := godotenv.Load()
	if err != nil {
		log.Println("[INFO] No .env file found, using environment variables")
	}

	AppConfig = Config{
		Port: getEnv("PORT", "8080"),

		EmbedServiceURL: getEnv("EMBED_SERVICE_URL", "http://embedding-service.llm.svc.cluster.local"),

		QdrantURL:      getEnv("QDRANT_URL", "http://qdrant.llm.svc.cluster.local:6333"),
		CollectionName: getEnv("QDRANT_COLLECTION", "matt-chunks"),

		MistralServiceURL: getEnv("MISTRAL_SERVICE_URL", "http://ollama.llm.svc.cluster.local:11434"),
		MistralModel:      getEnv("MISTRAL_MODEL", "mistral"),
		MistralStream:     getEnv("MISTRAL_STREAM", "true") == "true",

		UseOpenAI:   getEnv("USE_OPENAI", "false") == "true",
		OpenAIKey:   getEnv("OPENAI_API_KEY", ""),
		OpenAIModel: getEnv("OPENAI_MODEL", "gpt-4o"),

		DevCORSSecret: getEnv("DEV_CORS_SECRET", ""),
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
