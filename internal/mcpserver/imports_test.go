package mcpserver_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"
)

// SQL is expected transitively through internal/index, never as a direct import.
func TestImportBoundary(t *testing.T) {
	raw, err := exec.Command("go", "list", "-deps", "-json", ".").Output()
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	found := false
	for decoder.More() {
		var pkg struct {
			ImportPath string
			Imports    []string
		}
		if err := decoder.Decode(&pkg); err != nil {
			t.Fatal(err)
		}
		if pkg.ImportPath == "github.com/ContractorKeith/mdkb/internal/kb" {
			for _, path := range pkg.Imports {
				if path == "github.com/ContractorKeith/mdkb/internal/config" {
					t.Fatal("kb must receive settings from app")
				}
			}
		}
		if pkg.ImportPath != "github.com/ContractorKeith/mdkb/internal/mcpserver" {
			continue
		}
		found = true
		for _, path := range pkg.Imports {
			if path == "database/sql" || path == "os" || path == "os/exec" || path == "io/ioutil" {
				t.Fatalf("mcpserver must delegate storage access: %s", path)
			}
		}
	}
	if !found {
		t.Fatal("mcpserver absent from import graph")
	}
}
