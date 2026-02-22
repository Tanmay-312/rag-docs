package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	lib "pdf-ai-assistant/api/_lib"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

type ChatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" || req.Message == "" {
		http.Error(w, "session_id and message are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	gemini, err := lib.NewGeminiClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error init Gemini: %v", err), http.StatusInternalServerError)
		return
	}
	defer gemini.Close()

	upstash := lib.NewUpstashClient()

	// 1. Generate Query Vector
	vector, err := gemini.GenerateEmbedding(ctx, req.Message)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating embedding: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Perform Vector Search
	qReq := lib.QueryRequest{
		Vector:      vector,
		TopK:        3,
		IncludeMeta: true,
		Filter:      fmt.Sprintf("session_id = '%s'", req.SessionID),
	}
	qResp, err := upstash.Query(ctx, qReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying vector db: %v", err), http.StatusInternalServerError)
		return
	}

	var contextChunks []string
	var citations []string
	for _, result := range qResp.Result {
		if text, ok := result.Metadata["chunk_text"].(string); ok {
			contextChunks = append(contextChunks, text)
			// Truncate citation to show frontend what was found
			citations = append(citations, extractSnippet(text, 100))
		}
	}

	// 3. Prompt Gemini to answer with context
	prompt := fmt.Sprintf(`You are a helpful AI assistant. Answer the user's question based ONLY on the provided context retrieved from a document. 
If the answer is not in the context, say "I cannot answer this based on the provided document."

Context:
%s

Question:
%s`, strings.Join(contextChunks, "\n\n"), req.Message)

	model := gemini.Client.GenerativeModel("gemini-1.5-flash")
	model.SetTemperature(0.2) // lower temp for grounded RAG

	// 4. Stream response via Server-Sent Events style (Next.js AI SDK format or raw text)
	// We'll write an NDJSON stream for custom Typewriter effect
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	iter := model.GenerateContentStream(ctx, genai.Text(prompt))

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// First send citations so frontend can highlight them
	citJSON, _ := json.Marshal(map[string]interface{}{"type": "citations", "citations": citations})
	fmt.Fprintf(w, "data: %s\n\n", citJSON)
	flusher.Flush()

	for {
		resp, err := iter.Next()
		if err != nil {
			break
		}

		if len(resp.Candidates) > 0 {
			for _, part := range resp.Candidates[0].Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					chunkData, _ := json.Marshal(map[string]string{"type": "text", "text": string(txt)})
					fmt.Fprintf(w, "data: %s\n\n", chunkData)
					flusher.Flush()
				}
			}
		}
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func extractSnippet(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
