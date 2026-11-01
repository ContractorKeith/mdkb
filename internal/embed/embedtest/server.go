// Package embedtest provides a deterministic local Ollama double.
package embedtest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Server gives controlled paraphrases identical dimensions, without external I/O.
func Server(t testing.TB) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" || r.Method != "POST" {
			http.Error(w, "bad endpoint", 404)
			return
		}
		var in struct {
			Model string
			Input []string
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if in.Model == "missing" {
			http.Error(w, `{"error":"model missing not found"}`, 404)
			return
		}
		if len(in.Input) > 16 {
			http.Error(w, "too many inputs", 400)
			return
		}
		vectors := [][]float32{}
		for _, text := range in.Input {
			v := make([]float32, 64)
			for _, word := range strings.Fields(strings.ToLower(text)) {
				word = strings.Trim(word, "#.,:;!?[]")
				switch word {
				case "espresso", "coffee", "café", "shop":
					word = "cafe"
				case "isolated", "orphan":
					word = "alone"
				case "damaged", "broken", "recovery", "recoverable":
					word = "damage"
				}
				h := uint32(2166136261)
				for _, c := range word {
					h = (h ^ uint32(c)) * 16777619
				}
				v[h%64]++
			}
			vectors = append(vectors, v)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": vectors})
	}))
	t.Cleanup(s.Close)
	return s
}
