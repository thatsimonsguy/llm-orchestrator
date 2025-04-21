package types

type ChatRequest struct {
	UserID string `json:"user_id"`
	Query  string `json:"query"`
}

type Chunk struct {
	Text  string  `json:"text"`
	Score float32 `json:"score"`
}

type ChatChunkResponse struct {
	Response string `json:"response"`
}
