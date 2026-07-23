package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OpenAIEmbedder struct {
	APIKey string
	Model  string
	URL    string
	Client *http.Client
}

func NewOpenAIEmbedder(apiKey string) *OpenAIEmbedder {
	return &OpenAIEmbedder{
		APIKey: apiKey,
		Model:  "text-embedding-3-small",
		URL:    "https://api.openai.com/v1/embeddings",
		Client: http.DefaultClient,
	}
}

func (o *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"input": text,
		"model": o.Model,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", o.URL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai error: status %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return result.Data[0].Embedding, nil
}
