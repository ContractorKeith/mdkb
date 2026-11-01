package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSemanticSettings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, key := range []string{"MDKB_SEMANTIC", "MDKB_OLLAMA_URL", "MDKB_EMBED_MODEL"} {
		t.Setenv(key, "")
	}
	cfg, err := Semantic()
	if err != nil || cfg.Enabled || cfg.Model != "nomic-embed-text" || cfg.URL != "http://localhost:11434" || cfg.TimeoutSeconds != 30 {
		t.Fatal(cfg, err)
	}
	dir := filepath.Join(home, ".config/mdkb")
	if err = os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "config.toml"), []byte("[semantic]\nenabled=true\nurl='http://configured'\nmodel='custom'\ntimeout_seconds=5\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err = Semantic()
	if err != nil || !cfg.Enabled || cfg.Model != "custom" || cfg.TimeoutSeconds != 5 {
		t.Fatal(cfg, err)
	}
	t.Setenv("MDKB_SEMANTIC", "0")
	t.Setenv("MDKB_OLLAMA_URL", "http://override")
	t.Setenv("MDKB_EMBED_MODEL", "override")
	cfg, err = Semantic()
	if err != nil || cfg.Enabled || cfg.URL != "http://override" || cfg.Model != "override" {
		t.Fatal(cfg, err)
	}
	t.Setenv("MDKB_SEMANTIC", "yes")
	if _, err = Semantic(); err == nil {
		t.Fatal("accepted invalid boolean")
	}
}
