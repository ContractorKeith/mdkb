package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type SemanticSettings struct {
	Enabled        bool   `toml:"enabled"`
	URL            string `toml:"url"`
	Model          string `toml:"model"`
	TimeoutSeconds int    `toml:"timeout_seconds"`
}

func Semantic() (SemanticSettings, error) {
	cfg := struct {
		Semantic SemanticSettings `toml:"semantic"`
	}{Semantic: SemanticSettings{false, "http://localhost:11434", "nomic-embed-text", 30}}
	home, err := os.UserHomeDir()
	if err != nil {
		return cfg.Semantic, err
	}
	_, err = toml.DecodeFile(filepath.Join(home, ".config/mdkb/config.toml"), &cfg)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return cfg.Semantic, err
	}
	switch os.Getenv("MDKB_SEMANTIC") {
	case "":
	case "0":
		cfg.Semantic.Enabled = false
	case "1":
		cfg.Semantic.Enabled = true
	default:
		return cfg.Semantic, fmt.Errorf("MDKB_SEMANTIC must be 0 or 1")
	}
	if v := os.Getenv("MDKB_OLLAMA_URL"); v != "" {
		cfg.Semantic.URL = v
	}
	if v := os.Getenv("MDKB_EMBED_MODEL"); v != "" {
		cfg.Semantic.Model = v
	}
	if cfg.Semantic.TimeoutSeconds < 1 || cfg.Semantic.URL == "" || cfg.Semantic.Model == "" {
		return cfg.Semantic, fmt.Errorf("semantic URL/model must be nonempty and timeout_seconds positive")
	}
	return cfg.Semantic, nil
}
