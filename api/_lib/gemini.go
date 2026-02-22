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

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &GeminiClient{Client: client}, nil
}

func (g *GeminiClient) Close() {
	if g.Client != nil {
		g.Client.Close()
	}
}

// SanitizePII uses Gemini 1.5 Flash to redact PII
func (g *GeminiClient) SanitizePII(ctx context.Context, text string) (string, error) {
	model := g.Client.GenerativeModel("gemini-flash-latest")
	model.SetTemperature(0)

	prompt := fmt.Sprintf(`You are a PII sanitization agent. Your job is to redact any Personally Identifiable Information (PII) from the following text, including emails, phone numbers, and API keys. 
Replace them with [REDACTED EMAIL], [REDACTED PHONE], or [REDACTED KEY]. 
Do NOT change any other text, just return the exact same text with the PII redacted. 
Text to sanitize:

%s`, text)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		if txt, ok := part.(genai.Text); ok {
			return string(txt), nil
		}
	}

	return text, nil
}

// GenerateEmbedding generates an embedding for a text chunk
func (g *GeminiClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	em := g.Client.EmbeddingModel("text-embedding-004")
	res, err := em.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, err
	}

	if len(res.Embedding.Values) > 0 {
		return res.Embedding.Values, nil
	}

	return nil, fmt.Errorf("no embedding values returned")
}
