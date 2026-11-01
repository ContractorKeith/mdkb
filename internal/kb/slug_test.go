package kb

import (
	"strings"
	"testing"
)

func TestSlug(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{"First few words! more", "first-few-words-more"},
		{"", "note"}, {"!!!", "note"},
		{strings.Repeat("A", 80), strings.Repeat("a", 40)},
		{"  Héllo World  ", "héllo-world"},
	} {
		if got := Slug(tt.in); got != tt.want {
			t.Fatalf("%q: %q != %q", tt.in, got, tt.want)
		}
	}
}
