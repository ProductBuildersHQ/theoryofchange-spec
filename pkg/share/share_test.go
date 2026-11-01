package share

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/report"
	"github.com/ProductBuildersHQ/impact-model/pkg/slides"
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

func TestStampHashMatchesCanonicalJSON(t *testing.T) {
	doc := fullDoc(t)
	want, err := impactio.ContentHash(doc)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewStamp(doc, "2026-09-30")
	if err != nil {
		t.Fatal(err)
	}
	if s.ContentHash != want || s.ModelID != doc.ID || s.AsOf != "2026-09-30" {
		t.Errorf("unexpected stamp %+v", s)
	}
}

func TestStampVersionFallsBack(t *testing.T) {
	doc := fullDoc(t)
	doc.Metadata.Version = ""
	doc.Revisions = nil
	if s, err := NewStamp(doc, ""); err != nil || s.Version != unversioned {
		t.Errorf("no version: %+v, %v", s, err)
	}
	doc.Revisions = []impact.Revision{{Version: "0.1", Date: "2026-01-01", Summary: "first"}, {Version: "0.2", Date: "2026-02-01", Summary: "second"}}
	if s, err := NewStamp(doc, ""); err != nil || s.Version != "0.2" {
		t.Errorf("latest revision: %+v, %v", s, err)
	}
	doc.Metadata.Version = "1.0"
	if s, err := NewStamp(doc, ""); err != nil || s.Version != "1.0" {
		t.Errorf("metadata version: %+v, %v", s, err)
	}
}

func TestStampChangesWhenTheDocumentChanges(t *testing.T) {
	doc := fullDoc(t)
	a, err := NewStamp(doc, "")
	if err != nil {
		t.Fatal(err)
	}
	doc.Outcomes[0].Title += " (edited)"
	b, err := NewStamp(doc, "")
	if err != nil {
		t.Fatal(err)
	}
	if a.ContentHash == b.ContentHash {
		t.Error("an edit did not change the content hash")
	}
}

func TestStampForms(t *testing.T) {
	s := Stamp{ModelID: "t1", Version: "0.3", ContentHash: strings.Repeat("ab", 32), AsOf: "2026-09-30"}
	if got, want := s.CSVRow(), "# model:t1 version:0.3 hash:"+s.ContentHash+"\r\n"; got != want {
		t.Errorf("CSVRow = %q, want %q", got, want)
	}
	if got := s.Line(); got != "Impact model t1 · version 0.3 · hash abababababab · as of 2026-09-30" {
		t.Errorf("Line = %q", got)
	}
	block := s.YAMLBlock(`A "quoted" title`)
	for _, want := range []string{"---\n", `title: "A \"quoted\" title"`, `version: "0.3"`, `date: "2026-09-30"`, `model-id: "t1"`, "content-hash: \"" + s.ContentHash + "\""} {
		if !strings.Contains(block, want) {
			t.Errorf("YAML block is missing %q:\n%s", want, block)
		}
	}
	s.AsOf = ""
	if strings.Contains(s.YAMLBlock("t"), "date:") || strings.Contains(s.Line(), "as of") {
		t.Error("an empty as-of date should be left out")
	}
}

func definitions(t *testing.T, doc *impact.Model) string {
	t.Helper()
	whole, err := report.BuildView(doc, nil, nil, report.Options{})
	if err != nil {
		t.Fatal(err)
	}
	pathways, err := report.PathwayViews(doc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defs, err := report.Definitions(whole, pathways, doc)
	if err != nil {
		t.Fatal(err)
	}
	return defs
}

func TestReportProfileGolden(t *testing.T) {
	doc := fullDoc(t)
	s, err := NewStamp(doc, "2026-09-30")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	p := Profile{Diagram: func(string) (string, error) {
		n++
		return fmt.Sprintf("diagrams/d%d.png", n), nil
	}}
	out, err := p.Report(s, doc.Metadata.Name, definitions(t, doc))
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "definitions-pandoc.md", out)
}

func TestReportProfileRules(t *testing.T) {
	doc := fullDoc(t)
	s := Stamp{ModelID: "t1", Version: "1", ContentHash: strings.Repeat("0", 64)}
	out, err := Profile{}.Report(s, "T", definitions(t, doc))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "_(inferred)_") || !strings.Contains(out, "[inferred]") {
		t.Error("basis badges should be plain text")
	}
	if !strings.HasPrefix(out, "---\ntitle: \"T\"") {
		t.Errorf("want a metadata block first:\n%.80s", out)
	}
	if strings.Count(out, "\n# ") != 0 {
		t.Error("the body's own title should be dropped in favor of the metadata title")
	}
	sections := strings.Count(out, "\n## ")
	if got := strings.Count(out, "\\newpage"); got != sections-1 {
		t.Errorf("want a page break before every section but the first: %d breaks, %d sections", got, sections)
	}
	if !strings.HasSuffix(out, "*"+s.Line()+"*\n") {
		t.Error("the stamp line should close the document")
	}
}

func TestDiagramErrorIsReturned(t *testing.T) {
	boom := errors.New("render failed")
	p := Profile{Diagram: func(string) (string, error) { return "", boom }}
	if _, err := p.Report(Stamp{}, "T", "```mermaid\nflowchart BT\n```\n"); !errors.Is(err, boom) {
		t.Errorf("err = %v, want it to wrap the renderer's error", err)
	}
}

func TestDeckProfileGolden(t *testing.T) {
	doc := fullDoc(t)
	s, err := NewStamp(doc, "")
	if err != nil {
		t.Fatal(err)
	}
	whole, err := report.BuildView(doc, nil, nil, report.Options{})
	if err != nil {
		t.Fatal(err)
	}
	pathways, err := report.PathwayViews(doc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	marp := slides.Workshop(whole, pathways, slides.Options{Stamp: s.Line()})
	out, err := Profile{Diagram: func(string) (string, error) { return "d.png", nil }}.Deck(s, doc.Metadata.Name, marp)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"marp: true", "<!-- _class", "```mermaid", "_(inferred)_"} {
		if strings.Contains(out, bad) {
			t.Errorf("deck still contains %q", bad)
		}
	}
	if !strings.Contains(out, "::: notes") || !strings.HasSuffix(out, "# Export details\n\n"+s.Line()+"\n") {
		t.Error("want speaker notes and a closing stamp slide")
	}
	golden(t, "workshop-pandoc.md", out)
}
