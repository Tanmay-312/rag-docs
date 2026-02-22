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

	// FORCE STABLE V1: Most 404s are caused by the SDK defaulting to v1beta
	// which no longer supports retired model names.
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

// SanitizePII uses the 2026 stable workhorse: Gemini 2.0 Flash
func (g *GeminiClient) SanitizePII(ctx context.Context, text string) (string, error) {
	// gemini-2.0-flash is the 2026 successor for all "flash" tasks.
	model := g.Client.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0)

	prompt := fmt.Sprintf(`You are a PII sanitization agent. Redact Personally Identifiable Information (PII) 
from the text (emails, phone numbers, API keys). Replace with [REDACTED EMAIL], [REDACTED PHONE], or [REDACTED KEY]. 
Return ONLY the sanitized text with no conversational filler.

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

// GenerateEmbedding uses the modern unified text-embedding-004 on the stable v1 endpoint
func (g *GeminiClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// text-embedding-004 is supported on v1.
	// Ensure you aren't using older IDs like 'embedding-001' or 'models/embedding'
	em := g.Client.EmbeddingModel("text-embedding-004")
	res, err := em.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	if len(res.Embedding.Values) > 0 {
		return res.Embedding.Values, nil
	}

	return nil, fmt.Errorf("no embedding values returned")
}
