package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type UpstashClient struct {
	URL   string
	Token string
}

func NewUpstashClient() *UpstashClient {
	return &UpstashClient{
		URL:   os.Getenv("UPSTASH_VECTOR_REST_URL"),
		Token: os.Getenv("UPSTASH_VECTOR_REST_TOKEN"),
	}
}

type VectorData struct {
	ID       string                 `json:"id"`
	Vector   []float32              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (c *UpstashClient) UpsertValues(ctx context.Context, vectors []VectorData) error {
	payload, err := json.Marshal(vectors)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/upsert", c.URL), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("upstash error: status %d", resp.StatusCode)
	}

	return nil
}

type QueryRequest struct {
	Vector       []float32              `json:"vector"`
	TopK         int                    `json:"topK"`
	IncludeMeta  bool                   `json:"includeMetadata"`
	Filter       string                 `json:"filter,omitempty"`
}

type QueryResponse struct {
	Result []struct {
		ID       string                 `json:"id"`
		Score    float32                `json:"score"`
		Metadata map[string]interface{} `json:"metadata"`
	} `json:"result"`
}

func (c *UpstashClient) Query(ctx context.Context, reqData QueryRequest) (*QueryResponse, error) {
	payload, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/query", c.URL), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("upstash query error: status %d", resp.StatusCode)
	}

	var qResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&qResp); err != nil {
		return nil, err
	}

	return &qResp, nil
}

func (c *UpstashClient) DeleteBySession(ctx context.Context, sessionID string) error {
	// Upstash vector doesn't have native delete-by-metadata yet.
	// But we can query the IDs and delete them.
	// Let's query up to 1000 items with this session_id.
	qReq := QueryRequest{
		Vector:      make([]float32, 768), // Dummy vector
		TopK:        1000,
		IncludeMeta: false,
		Filter:      fmt.Sprintf("session_id = '%s'", sessionID),
	}
	qResp, err := c.Query(ctx, qReq)
	if err != nil {
		return err
	}

	if len(qResp.Result) == 0 {
		return nil
	}

	var ids []string
	for _, r := range qResp.Result {
		ids = append(ids, r.ID)
	}

	payload, err := json.Marshal(ids)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/delete", c.URL), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("upstash delete error: status %d", resp.StatusCode)
	}

	return nil
}
