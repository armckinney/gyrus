package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type OllamaEmbedder struct {
	URL    string
	Model  string
	Client *http.Client
}

func NewOllamaEmbedder(url, model string) *OllamaEmbedder {
	if url == "" {
		url = "http://127.0.0.1:11434/api/embeddings"
	} else {
		url = strings.Replace(url, "localhost", "127.0.0.1", 1)
		if !strings.HasSuffix(url, "/api/embeddings") {
			url = strings.TrimSuffix(url, "/") + "/api/embeddings"
		}
	}
	if model == "" {
		model = "nomic-embed-text"
	}
	return &OllamaEmbedder{
		URL:    url,
		Model:  model,
		Client: http.DefaultClient,
	}
}

func (o *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model":  o.Model,
		"prompt": text,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", o.URL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama error: status %d", resp.StatusCode)
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Embedding, nil
}
