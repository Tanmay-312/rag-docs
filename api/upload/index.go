package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	lib "pdf-ai-assistant/api/_lib"
	"sync"
	"time"

	"github.com/google/uuid"
)

type WorkerResult struct {
	ChunkIndex int
	Vector     []float32
	Text       string
	Error      error
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

	err := r.ParseMultipartForm(10 << 20) // 10MB limit
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sessionID := r.FormValue("session_id")
	if sessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "error reading file", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()

	text, err := lib.ReadPDF(fileData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing PDF: %v", err), http.StatusInternalServerError)
		return
	}

	chunks := lib.ChunkText(text, 300, 50)
	if len(chunks) == 0 {
		http.Error(w, "No text found in PDF", http.StatusBadRequest)
		return
	}

	gemini, err := lib.NewGeminiClient(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error init Gemini: %v", err), http.StatusInternalServerError)
		return
	}
	defer gemini.Close()

	upstash := lib.NewUpstashClient()

	// Worker Pool setup
	numWorkers := 5
	jobs := make(chan struct {
		Index int
		Text  string
	}, len(chunks))
	results := make(chan WorkerResult, len(chunks))

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				// Sanitize PII via Gemini 1.5 Flash
				sanitizedText, err := gemini.SanitizePII(ctx, job.Text)
				if err != nil {
					results <- WorkerResult{ChunkIndex: job.Index, Error: err}
					continue
				}

				// Generate Embedding
				vector, err := gemini.GenerateEmbedding(ctx, sanitizedText)
				if err != nil {
					results <- WorkerResult{ChunkIndex: job.Index, Error: err}
					continue
				}

				results <- WorkerResult{
					ChunkIndex: job.Index,
					Vector:     vector,
					Text:       sanitizedText,
				}
			}
		}()
	}

	// Queue chunks
	for i, chunk := range chunks {
		jobs <- struct {
			Index int
			Text  string
		}{Index: i, Text: chunk}
	}
	close(jobs)

	// Wait since Vercel serverless has a max execution time anyway and stops when response is sent
	wg.Wait()
	close(results)

	var vectors []lib.VectorData
	for res := range results {
		if res.Error != nil {
			fmt.Printf("Error processing chunk %d: %v\n", res.ChunkIndex, res.Error)
			continue
		}

		vectors = append(vectors, lib.VectorData{
			ID:     uuid.New().String(),
			Vector: res.Vector,
			Metadata: map[string]interface{}{
				"session_id": sessionID,
				"chunk_text": res.Text,
				"timestamp":  time.Now().Unix(),
			},
		})
	}

	if len(vectors) > 0 {
		err = upstash.UpsertValues(ctx, vectors)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error saving to Upstash: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"chunks":  len(vectors),
	})
}
