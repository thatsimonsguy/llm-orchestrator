package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"matthewpsimons.com/llm-orchestrator/internal/config"
)

func setupTestHandler(devCorsSecret string) http.Handler {
	cfg := config.Config{
		DevCORSSecret: devCorsSecret,
	}
	logger, _ := zap.NewDevelopment()
	return HandleChat(cfg, logger)
}

func TestHandleChat_CORSAndAuthorization(t *testing.T) {
	handler := setupTestHandler("my-secret")

	tests := []struct {
		name         string
		method       string
		origin       string
		secretHeader string
		body         string
		expectedCode int
	}{
		{
			name:         "OPTIONS from localhost gets OK",
			method:       http.MethodOptions,
			origin:       "http://localhost:3000",
			expectedCode: http.StatusOK,
		},
		{
			name:         "OPTIONS from matthewpsimons.com gets OK",
			method:       http.MethodOptions,
			origin:       "https://matthewpsimons.com",
			expectedCode: http.StatusOK,
		},
		{
			name:         "POST from localhost without the secret gets 401",
			method:       http.MethodPost,
			origin:       "http://localhost:3000",
			body:         `{"user_id":"test","query":"hello"}`,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "POST from matthewpsimons.com does not get 401",
			method:       http.MethodPost,
			origin:       "https://matthewpsimons.com",
			body:         `{"user_id":"test","query":"hello"}`,
			expectedCode: http.StatusInternalServerError, // will fail later because EmbedText will error, but NOT 401
		},
		{
			name:         "POST from matthewpsimons.com with invalid JSON gets server error",
			method:       http.MethodPost,
			origin:       "https://matthewpsimons.com",
			body:         `{invalid}`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/chat", strings.NewReader(tt.body))
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Content-Type", "application/json")
			if tt.secretHeader != "" {
				req.Header.Set("X-Dev-Cors-Secret", tt.secretHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectedCode, rr.Code)
			}
		})
	}
}
