package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/config"
)

func TestRootPrecedence(t *testing.T) {
	for _, tt := range []struct{ name, flag, env, file, want string }{
		{"default", "", "", "", "kb"},
		{"empty config", "", "", "# no root", "kb"},
		{"file", "", "", "root = '~/from-file'", "from-file"},
		{"env over file", "", "~/from-env", "root = '~/from-file'", "from-env"},
		{"flag over all", "~/from-flag", "~/from-env", "root = '~/from-file'", "from-flag"},
		{"env without file", "", "~/from-env", "", "from-env"},
		{"flag without env", "~/from-flag", "", "", "from-flag"},
		{"flag skips invalid file", "~/from-flag", "", "not valid toml [", "from-flag"},
		{"env skips invalid file", "", "~/from-env", "not valid toml [", "from-env"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("MDKB_ROOT", tt.env)
			if tt.file != "" {
				dir := filepath.Join(home, ".config", "mdkb")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(tt.file), 0644); err != nil {
					t.Fatal(err)
				}
			}
			got, err := config.ResolveRoot(tt.flag)
			if err != nil {
				t.Fatal(err)
			}
			if want := filepath.Join(home, tt.want); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestInvalidConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MDKB_ROOT", "")
	dir := filepath.Join(home, ".config", "mdkb")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("root = ["), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := config.ResolveRoot(""); err == nil {
		t.Fatal("expected invalid TOML error")
	}
}

func TestAbsoluteAndRelativeRoots(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MDKB_ROOT", "")
	for _, path := range []string{home, ".", "notes/../kb", "~"} {
		got, err := config.ResolveRoot(path)
		want := home
		if path != "~" {
			want, _ = filepath.Abs(path)
		}
		if err != nil || got != want {
			t.Fatalf("%q: got %q, %v; want %q", path, got, err, want)
		}
	}
}
