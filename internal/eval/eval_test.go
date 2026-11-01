package eval

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/search"
)

func TestMetrics(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := t.TempDir()
	for id, body := range map[string]string{"a": "needle needle", "b": "needle filler filler filler", "c": "unrelated"} {
		if err := os.WriteFile(filepath.Join(root, id+".md"), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	idx, err := index.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if err = idx.Sync(t.Context(), root, false); err != nil {
		t.Fatal(err)
	}
	result, err := Run(t.Context(), search.Search{Index: idx}, []Query{{"needle", []string{"b"}}, {"missing", []string{"a"}}}, "keyword")
	if err != nil || result.Queries[0].MRR != 0.5 || result.Hit3 != 0.5 || result.MRR != 0.25 {
		t.Fatal(result, err)
	}
}
