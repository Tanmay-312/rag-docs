package lib

import (
	"context"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	Client *genai.Client
}

func NewGeminiClient(ctx context.Context) (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	// Use the stable v1 endpoint for 2026 production stability
	client, err := genai.NewClient(ctx,
		option.WithAPIKey(apiKey),
		option.WithEndpoint("https://generativelanguage.googleapis.com/v1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create GenAI client: %w", err)
	}

	return &GeminiClient{Client: client}, nil
}

func (g *GeminiClient) Close() {
	if g.Client != nil {
		g.Client.Close()
	}
}

// SanitizePII uses Gemini 2.5 Flash (the 2026 stable successor to 1.5)
func (g *GeminiClient) SanitizePII(ctx context.Context, text string) (string, error) {
	// gemini-2.5-flash is the current stable replacement for 1.5 flash
	model := g.Client.GenerativeModel("gemini-2.5-flash")
	model.SetTemperature(0)

	prompt := fmt.Sprintf(`You are a PII sanitization agent. Redact Personally Identifiable Information (PII) 
from the text (emails, phone numbers, API keys). Replace with [REDACTED EMAIL], [REDACTED PHONE], or [REDACTED KEY]. 
Return only the sanitized text.

Text to sanitize:
%s`, text)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("sanitization failed: %w", err)
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(txt), nil
		}
	}

	return text, nil
}

// GenerateEmbedding uses gemini-embedding-001 (Stable 2026 replacement)
func (g *GeminiClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// gemini-embedding-001 replaces text-embedding-004.
	// NOTE: This model uses 3072 dimensions by default.
	em := g.Client.EmbeddingModel("gemini-embedding-001")
	res, err := em.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	if len(res.Embedding.Values) > 0 {
		return res.Embedding.Values, nil
	}

	return nil, fmt.Errorf("no embedding values returned")
}
