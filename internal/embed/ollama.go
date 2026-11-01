// Package embed chunks notes and obtains local embedding vectors.
package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strings"
	"time"
)

type Embedder interface {
	Embed(context.Context, []string) ([][]float32, error)
	Model() string
}

type ErrUnavailable struct{ URL, Model, Reason string }

func (e *ErrUnavailable) Error() string {
	if e.URL == "" {
		return fmt.Sprintf("semantic unavailable (model %s): %s", e.Model, e.Reason)
	}
	return fmt.Sprintf("Ollama unavailable at %s (model %s): %s", e.URL, e.Model, strings.Join(strings.Fields(e.Reason), " "))
}

type Ollama struct {
	URL, ModelName string
	Client         *http.Client
}

func (o Ollama) Model() string { return o.ModelName }

// Embed sends at most sixteen inputs per request; failures never return partial vectors.
func (o Ollama) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	client := o.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	result := [][]float32{}
	dim := 0
	for start := 0; start < len(texts); start += 16 {
		raw, err := json.Marshal(struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}{o.Model(), texts[start:min(start+16, len(texts))]})
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(o.URL, "/")+"/api/embed", bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return nil, &ErrUnavailable{o.URL, o.Model(), err.Error()}
		}
		var body struct {
			Embeddings [][]float32 `json:"embeddings"`
			Error      string      `json:"error"`
		}
		err = json.NewDecoder(io.LimitReader(resp.Body, 64<<20)).Decode(&body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			reason := body.Error
			if reason == "" {
				reason = resp.Status
			}
			return nil, &ErrUnavailable{o.URL, o.Model(), reason}
		}
		if err != nil {
			var timeout net.Error
			if errors.As(err, &timeout) && timeout.Timeout() {
				return nil, &ErrUnavailable{o.URL, o.Model(), err.Error()}
			}
			return nil, fmt.Errorf("decode Ollama embeddings: %w", err)
		}
		if len(body.Embeddings) != min(16, len(texts)-start) {
			return nil, fmt.Errorf("Ollama returned wrong embedding count")
		}
		for _, v := range body.Embeddings {
			if len(v) == 0 || (dim != 0 && dim != len(v)) {
				return nil, fmt.Errorf("Ollama returned inconsistent embedding dimensions")
			}
			dim = len(v)
			for _, x := range v {
				if math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
					return nil, fmt.Errorf("Ollama returned non-finite vector")
				}
			}
		}
		result = append(result, body.Embeddings...)
	}
	return result, nil
}
