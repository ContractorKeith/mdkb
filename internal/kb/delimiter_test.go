package kb_test

import (
	"errors"
	"github.com/ContractorKeith/mdkb/internal/kb"
	"strings"
	"testing"
	"time"
)

func TestWriteIDDelimiters(t *testing.T) {
	root := t.TempDir()
	if err := kb.Init(root); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"has#heading", "has|alias", "has[bracket", "has]bracket", "has\nline", "has\tcontrol", "has\x7fcontrol", " leading", "trailing ", "dir/ leading", "dir /note"} {
		t.Run(strings.ReplaceAll(id, "/", "_"), func(t *testing.T) {
			if err := kb.ValidateID(id, true); !errors.Is(err, kb.ErrWriteRefused) {
				t.Errorf("accepted write id %q: %v", id, err)
			}
			if _, err := kb.Write(root, id, "body", "create"); !errors.Is(err, kb.ErrWriteRefused) {
				t.Errorf("write accepted %q", id)
			}
			if err := kb.PreviewWrite(root, id, "create"); !errors.Is(err, kb.ErrWriteRefused) {
				t.Errorf("preview accepted %q", id)
			}
		})
	}
	for _, title := range []string{"has#heading", "[[bracket]] | alias", "\n\t title "} {
		id, _, err := kb.CaptureContent("body", title, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))
		if err != nil || kb.ValidateID(id, true) != nil {
			t.Fatalf("unsafe capture id %q: %v", id, err)
		}
	}
}
