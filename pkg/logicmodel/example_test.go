package logicmodel_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
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

// The reports the CLI offers, over the intermediary example: one diagram for
// each of its three pathways, as D2 and as Mermaid.
func TestLayeredReportsOverTheExample(t *testing.T) {
	doc, err := impactio.LoadDocument("../../examples/intermediary/impact-model.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"logic-model-d2", "logic-model-layered"} {
		out, err := service.DefinitionReport(doc, kind, "")
		if err != nil {
			t.Fatal(kind, err)
		}
		if n := strings.Count(out, "## Layered logic model:"); n != 3 {
			t.Errorf("%s: %d diagrams, want 3", kind, n)
		}
		golden(t, kind+".md", out)
	}
}
