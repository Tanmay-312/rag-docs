package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	lib "pdf-ai-assistant/api/_lib"
)

type WipeRequest struct {
	SessionID string `json:"session_id"`
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

	var req WipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	upstash := lib.NewUpstashClient()

	err := upstash.DeleteBySession(ctx, req.SessionID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error wiping data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Nuclear Wipe Complete",
	})
}
