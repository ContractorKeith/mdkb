package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// ResolveRoot selects flag > MDKB_ROOT > user config > ~/kb.
// Lower-priority configuration is not read when a root is already selected.
func ResolveRoot(flag string) (string, error) {
	root := flag
	if root == "" {
		root = os.Getenv("MDKB_ROOT")
	}
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		var cfg struct {
			Root string `toml:"root"`
		}
		path := filepath.Join(home, ".config", "mdkb", "config.toml")
		if _, err := toml.DecodeFile(path, &cfg); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("read config: %w", err)
		}
		root = cfg.Root
	}
	if root == "" {
		root = "~/kb"
	}
	return AbsolutePath(root)
}

// AbsolutePath expands a leading ~/ and resolves relative paths.
func AbsolutePath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return filepath.Abs(path)
}
