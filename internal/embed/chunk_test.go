package embed

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestChunks(t *testing.T) {
	for _, tt := range []struct {
		name, title, body string
		want              []string
	}{
		{"empty", "Title", "", []string{"Title\n"}},
		{"small", "Title", "a  b\n\nc d", []string{"Title\na b c d"}},
		{"unicode", "Café", "☕ résumé", []string{"Café\n☕ résumé"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := Chunks(tt.title, tt.body); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("%q", got)
			}
		})
	}
	words := []string{}
	for j := 0; j < 450; j++ {
		words = append(words, fmt.Sprint(j))
	}
	chunks := Chunks("Title", strings.Join(words, " "))
	if len(chunks) != 3 {
		t.Fatal(len(chunks))
	}
	if chunks[1] != strings.Join(words[160:360], " ") || chunks[2] != strings.Join(words[320:], " ") {
		t.Fatal("overlap")
	}
	body := strings.Join(words[:175], " ") + "\n\n" + strings.Join(words[175:], " ")
	chunks = Chunks("Title", body)
	if len(strings.Fields(chunks[0])) != 176 || !strings.HasPrefix(chunks[1], "135 ") {
		t.Fatal("paragraph boundary", chunks)
	}
}
