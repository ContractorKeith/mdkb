package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type WriteSettings struct {
	CommitOnWrite bool   `toml:"commit_on_write"`
	DailyDir      string `toml:"daily_dir"`
}

func Writes() (WriteSettings, error) {
	cfg := struct {
		Write WriteSettings `toml:"write"`
	}{Write: WriteSettings{true, "10-Daily"}}
	home, err := os.UserHomeDir()
	if err != nil {
		return cfg.Write, err
	}
	_, err = toml.DecodeFile(filepath.Join(home, ".config/mdkb/config.toml"), &cfg)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return cfg.Write, err
	}
	switch value := os.Getenv("MDKB_COMMIT_ON_WRITE"); value {
	case "":
	case "0":
		cfg.Write.CommitOnWrite = false
	case "1":
		cfg.Write.CommitOnWrite = true
	default:
		return cfg.Write, fmt.Errorf("MDKB_COMMIT_ON_WRITE must be 0 or 1")
	}
	return cfg.Write, nil
}
