package embed_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ContractorKeith/mdkb/internal/embed"
	"github.com/ContractorKeith/mdkb/internal/embed/embedtest"
)

func TestOllamaBatchAndUnavailable(t *testing.T) {
	server := embedtest.Server(t)
	o := embed.Ollama{URL: server.URL, ModelName: "test", Client: server.Client()}
	texts := make([]string, 35)
	for j := range texts {
		texts[j] = "coffee"
	}
	vectors, err := o.Embed(t.Context(), texts)
	if err != nil || len(vectors) != 35 {
		t.Fatalf("%d %v", len(vectors), err)
	}
	o.ModelName = "missing"
	_, err = o.Embed(t.Context(), texts)
	var unavailable *embed.ErrUnavailable
	if !errors.As(err, &unavailable) || !strings.Contains(err.Error(), server.URL) || !strings.Contains(err.Error(), "missing") {
		t.Fatal(err)
	}
	server.Close()
	o.ModelName = "test"
	_, err = o.Embed(t.Context(), texts)
	if !errors.As(err, &unavailable) {
		t.Fatal(err)
	}
}
func TestTimeoutAndMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(40 * time.Millisecond) }))
	defer server.Close()
	o := embed.Ollama{URL: server.URL, ModelName: "test", Client: &http.Client{Timeout: 10 * time.Millisecond}}
	_, err := o.Embed(t.Context(), []string{"x"})
	var unavailable *embed.ErrUnavailable
	if !errors.As(err, &unavailable) {
		t.Fatal(err)
	}
	for _, response := range []string{`{}`, `{"embeddings":[[]]}`, `{"embeddings":[[1],[1,2]]}`, `not json`} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(response)) }))
		o.URL = s.URL
		o.Client = s.Client()
		if _, err := o.Embed(t.Context(), []string{"x"}); err == nil {
			t.Fatal(response)
		}
		s.Close()
	}
}

// A timeout while receiving the body is as unavailable as one before headers.
func TestResponseBodyTimeout(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		time.Sleep(40 * time.Millisecond)
	}))
	defer s.Close()
	o := embed.Ollama{URL: s.URL, ModelName: "test", Client: &http.Client{Timeout: 10 * time.Millisecond}}
	_, err := o.Embed(t.Context(), []string{"x"})
	var unavailable *embed.ErrUnavailable
	if !errors.As(err, &unavailable) {
		t.Fatal(err)
	}
}
