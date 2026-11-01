package problemtree_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/service"
)

var update = flag.Bool("update", false, "rewrite the golden files")

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with -update to create it)", path, err)
	}
	if string(want) != got {
		t.Errorf("%s differs from the golden file (run with -update to accept):\n--- got\n%s\n--- want\n%s", name, got, want)
	}
}

// The three reports the CLI offers, over the eating-health illustration.
func TestProblemTreeReports(t *testing.T) {
	doc, err := impactio.LoadDocument("testdata/eating.json")
	if err != nil {
		t.Fatal(err)
	}
	for kind, file := range map[string]string{
		"problem-tree": "problem-tree.md", "problem-tree-d2": "problem-tree-d2.md", "problem-tree-json": "problem-tree.json",
	} {
		out, err := service.DefinitionReport(doc, kind, "")
		if err != nil {
			t.Fatal(kind, err)
		}
		golden(t, file, out)
	}
}
