package cors

import (
	"net/http"

	"matthewpsimons.com/llm-orchestrator/internal/config"
)

// ApplyCORS sets CORS headers based on origin
func ApplyCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	if origin == "http://localhost:3000" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Dev-Cors-Secret")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "https://matthewpsimons.com")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	}
}

// IsDevRequestAuthorized checks if a localhost request has the correct secret
func IsDevRequestAuthorized(cfg config.Config, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	corsSecretHeader := r.Header.Get("X-Dev-Cors-Secret")

	if origin == "http://localhost:3000" {
		return corsSecretHeader == cfg.DevCORSSecret
	}
	return true // allow production origin

}
