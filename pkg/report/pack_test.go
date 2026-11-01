package report

import (
	"strings"
	"testing"
)

func TestDefinitionsPackGolden(t *testing.T) {
	doc := fullDoc(t)
	whole := build(t, doc, nil, Options{})
	pathways, err := PathwayViews(doc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Definitions(whole, pathways, doc)
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "definitions.md", out)
}

func TestDefinitionsPackContainsEveryReport(t *testing.T) {
	doc := fullDoc(t)
	whole := build(t, doc, nil, Options{})
	pathways, err := PathwayViews(doc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Definitions(whole, pathways, doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## Causal pathways", "## Logic model:", "## Metric definitions", "## Glossary"} {
		if !strings.Contains(out, want) {
			t.Errorf("pack is missing %q", want)
		}
	}
	// The pack reuses the individual renderers, so it cannot drift from them.
	for _, part := range []string{LogicModels(whole, pathways), MetricDefinitions(whole), Glossary(whole)} {
		if !strings.Contains(out, part) {
			t.Errorf("pack does not embed a report verbatim:\n%s", part)
		}
	}
}

func TestCSVGolden(t *testing.T) {
	doc := fullDoc(t)
	whole := build(t, doc, nil, Options{})
	for _, table := range CSVTables {
		got, err := CSV(table, whole, doc)
		if err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		golden(t, table+".csv", got)
	}
}

func TestCSVIsRFC4180(t *testing.T) {
	whole := build(t, fullDoc(t), nil, Options{})
	whole.Glossary[0].Definition = "has, a comma and \"quotes\"\nand a newline"
	got, err := CSV("glossary", whole, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "\"has, a comma and \"\"quotes\"\"\r\nand a newline\"") {
		t.Errorf("field not quoted per RFC 4180:\n%q", got)
	}
	if !strings.HasPrefix(got, "id,term,definition,examples\r\n") {
		t.Errorf("want CRLF-terminated header, got %q", got)
	}
}

func TestCSVUnknownTable(t *testing.T) {
	if _, err := CSV("nope", &View{}, nil); err == nil || !strings.Contains(err.Error(), "indicators") {
		t.Errorf("want an error naming the valid tables, got %v", err)
	}
}
