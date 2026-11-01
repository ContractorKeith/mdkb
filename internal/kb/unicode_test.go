package kb_test

import (
	"github.com/ContractorKeith/mdkb/internal/kb"
	"path/filepath"
	"testing"
)

func TestUnicodeIDsAgree(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	if err := kb.Init(root); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "Cafe\u0301.md"), "---\naliases: [Re\u0301sume\u0301]\n---\n# Coffee\n")
	notes, err := kb.Walk(root)
	if err != nil || len(notes) != 1 {
		t.Fatalf("walk: %+v %v", notes, err)
	}
	if notes[0].ID != "Café" || notes[0].Aliases[0] != "Résumé" {
		t.Errorf("unnormalized note: %+v", notes[0])
	}
	for _, id := range []string{"Café", "Cafe\u0301"} {
		note, _, err := kb.Read(root, id)
		if err != nil || note.ID != "Café" {
			t.Errorf("read %q: %+v %v", id, note, err)
		}
	}
	for _, target := range []string{"Café", "Cafe\u0301", "Résumé", "Re\u0301sume\u0301"} {
		id, reason := kb.NewResolver(notes).ResolveLink(target)
		if id != "Café" || reason != "" {
			t.Errorf("resolve %q: %q %q", target, id, reason)
		}
	}
	refs := kb.ExtractLinks("[[Cafe\u0301]]")
	if len(refs) != 1 || refs[0].Target != "Café" {
		t.Errorf("unnormalized link: %+v", refs)
	}
	result, err := kb.Write(root, "Cafe\u0301", "append", "append")
	if err != nil || result.ID != "Café" {
		t.Errorf("write: %+v %v", result, err)
	}
	if err := kb.PreviewWrite(root, "Café", "create"); err == nil {
		t.Error("create preview missed NFD collision")
	}
}
