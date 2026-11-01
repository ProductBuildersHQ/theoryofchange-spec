package slides

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/report"
)

var update = flag.Bool("update", false, "rewrite golden files")

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

func fullDoc(t *testing.T) *impact.Model {
	t.Helper()
	data, err := os.ReadFile("../impact/testdata/full.json")
	if err != nil {
		t.Fatal(err)
	}
	var d impact.Model
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		t.Fatal(err)
	}
	return &d
}

func views(t *testing.T, doc *impact.Model) (*report.View, []*report.View) {
	t.Helper()
	whole, err := report.BuildView(doc, nil, nil, report.Options{})
	if err != nil {
		t.Fatal(err)
	}
	pathways, err := report.PathwayViews(doc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return whole, pathways
}

func TestWorkshopGolden(t *testing.T) {
	whole, pathways := views(t, fullDoc(t))
	golden(t, "workshop.md", Workshop(whole, pathways, Options{}))
}

func TestWorkshopStructure(t *testing.T) {
	whole, pathways := views(t, fullDoc(t))
	out := Workshop(whole, pathways, Options{Theme: "gaia"})
	for _, want := range []string{
		"marp: true\ntheme: gaia\npaginate: true\n",
		"# The problem",
		"# The change we are working toward",
		"```mermaid",
		"# Logic model: ",
		"# How we will measure",
		"# What is established, and what is our reading",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("deck is missing %q", want)
		}
	}
	if got := SlideCount(out); got < 8 {
		t.Errorf("SlideCount = %d, want at least 8", got)
	}
}

func TestWorkshopIsDeterministic(t *testing.T) {
	whole, pathways := views(t, fullDoc(t))
	if a, b := Workshop(whole, pathways, Options{}), Workshop(whole, pathways, Options{}); a != b {
		t.Error("two renders of the same impact model differ")
	}
}

func TestWorkshopWithoutPathwaysRendersOneLogicModel(t *testing.T) {
	doc := fullDoc(t)
	doc.Pathways = nil
	for i := range doc.Indicators {
		doc.Indicators[i].PathwayIDs = nil
	}
	whole, pathways := views(t, doc)
	if n := strings.Count(Workshop(whole, pathways, Options{}), "# Logic model: "); n != 1 {
		t.Errorf("want one logic model, got %d", n)
	}
}

func TestChunkedSlidesAreNumbered(t *testing.T) {
	doc := fullDoc(t)
	whole, pathways := views(t, doc)
	for len(whole.Glossary) <= termsPerSlide {
		whole.Glossary = append(whole.Glossary, whole.Glossary[0])
	}
	out := Workshop(whole, pathways, Options{})
	if !strings.Contains(out, "# Words we are using (1 of 2)") || !strings.Contains(out, "(2 of 2)") {
		t.Error("a long glossary should split across numbered slides")
	}
}

func TestNotesCannotCloseTheComment(t *testing.T) {
	d := &deck{}
	d.notes("a --> b")
	if strings.Contains(d.b.String(), "a -->") {
		t.Error("speaker notes must not contain a comment terminator")
	}
}

// Each logic model table is followed by the layered diagram of the same logic
// model. The table stays.
func TestWorkshopShowsTheLayeredDiagramBesideTheTable(t *testing.T) {
	whole, pathways := views(t, fullDoc(t))
	out := Workshop(whole, pathways, Options{})
	tables := strings.Count(out, "# Logic model: ")
	layered := strings.Count(out, "# Logic model, step by step: ")
	if tables == 0 || layered != tables {
		t.Fatalf("%d logic model tables and %d layered diagrams", tables, layered)
	}
	i := strings.Index(out, "# Logic model: ")
	j := strings.Index(out, "# Logic model, step by step: ")
	if j < i {
		t.Error("the layered diagram comes before its table")
	}
	if !strings.Contains(out, "flowchart LR") || !strings.Contains(out, "subgraph n_inputs") {
		t.Errorf("the layered diagram is not drawn:\n%s", out)
	}
}
