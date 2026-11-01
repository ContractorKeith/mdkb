package app

import (
	"github.com/ContractorKeith/mdkb/internal/config"
	"github.com/ContractorKeith/mdkb/internal/kb"
)

// commitWrite reads user settings at the application layer after publication.
func commitWrite(root string, paths []string, message string) (bool, error) {
	cfg, err := config.Writes()
	if err != nil {
		return false, err
	}
	return kb.CommitWriteStatus(root, paths, message, cfg.CommitOnWrite)
}
