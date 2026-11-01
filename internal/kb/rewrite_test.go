package kb

import "testing"

func TestRewriteLinks(t *testing.T) {
	before := []Note{{ID: "folder/plain", Aliases: []string{"Home"}}, {ID: "source"}}
	after := []Note{{ID: "renamed/plain", Aliases: []string{"Home"}}, {ID: "source"}}
	for _, tt := range []struct {
		in, want string
		count    int
	}{
		{"[[folder/plain]] [[folder/plain.md]]", "[[renamed/plain]] [[renamed/plain]]", 2},
		{"[[folder/plain#h|a]] ![[folder/plain|a]]", "[[renamed/plain#h|a]] ![[renamed/plain|a]]", 2},
		{"[[plain]] [[Home]]", "[[plain]] [[Home]]", 0},
		{"`[[folder/plain]]`\n~~~\n[[folder/plain]]\n~~~\n[[folder/plain]]", "`[[folder/plain]]`\n~~~\n[[folder/plain]]\n~~~\n[[renamed/plain]]", 1},
	} {
		got, count := RewriteLinks(tt.in, "folder/plain", "renamed/plain", before, after)
		if got != tt.want || count != tt.count {
			t.Fatalf("%q: %q (%d)", tt.in, got, count)
		}
	}
	after = append(after, Note{ID: "other/plain"})
	got, count := RewriteLinks("[[plain#h|a]]", "folder/plain", "renamed/plain", before, after)
	if got != "[[renamed/plain#h|a]]" || count != 1 {
		t.Fatal(got, count)
	}
	got, count = RewriteLinks("[[plain]]", "plain", "renamed/plain", []Note{{ID: "plain"}}, after)
	if got != "[[renamed/plain]]" || count != 1 {
		t.Fatal(got, count)
	}
}
