package report

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/burden"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/measure"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func TestCustomTemplate(t *testing.T) {
	in := exampleInput(t)
	out, err := RenderTemplate(`# {{.Document.Title}}
{{range .Outcomes}}{{if eq .Level "impact"}}- {{label (printf "%s" .Level)}}: {{.Title}}{{if .Status}} ({{label .Status.PathwayState}}){{end}}
{{end}}{{end}}Indicators: {{len .Indicators}}
`, in.View)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# Riverbend Nonprofit Network theory of change", "- Impact: Residents get timely", "(At risk)", fmt.Sprintf("Indicators: %d", len(in.View.Indicators))} {
		if !strings.Contains(out, want) {
			t.Errorf("output is missing %q:\n%s", want, out)
		}
	}
}

func TestTemplateErrors(t *testing.T) {
	in := exampleInput(t)
	for name, tmpl := range map[string]string{
		"syntax":           "{{.Document.Title",
		"unknown field":    "{{.Nope}}",
		"unknown function": `{{readFile "/etc/passwd"}}`,
		"env function":     `{{env "HOME"}}`,
		"exec function":    `{{exec "ls"}}`,
	} {
		if _, err := RenderTemplate(tmpl, in.View); !errors.Is(err, ErrTemplate) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
}

func TestTemplateLimits(t *testing.T) {
	in := exampleInput(t)
	if _, err := RenderTemplate(strings.Repeat("x", MaxTemplateBytes+1), in.View); !errors.Is(err, ErrTemplate) {
		t.Errorf("an oversized template was accepted: %v", err)
	}
	// Output grows by doubling through nested loops until it passes the cap.
	bomb := `{{range .Outcomes}}{{range .Preconditions}}{{range $.Indicators}}{{range $.Outcomes}}{{range $.Outcomes}}{{range $.Outcomes}}{{$.Document.Problem}}{{end}}{{end}}{{end}}{{end}}{{end}}{{end}}`
	if _, err := RenderTemplate(bomb, in.View); !errors.Is(err, ErrTemplate) || !strings.Contains(err.Error(), "too large") {
		t.Errorf("runaway output was not stopped: %v", err)
	}
}

func TestTemplatesAreDeterministic(t *testing.T) {
	in := exampleInput(t)
	a, _ := RenderTemplate(`{{json .Outcomes}}`, in.View)
	b, _ := RenderTemplate(`{{json .Outcomes}}`, in.View)
	if a == "" || a != b {
		t.Error("two renders differ")
	}
}

// secrets are the strings of a restricted evidence record. None may appear in
// any output of any report.
var secrets = []string{"SECRET-EXCERPT", "SECRET-SUMMARY", "SECRET-URL", "UNSET-EXCERPT", "UNSET-SUMMARY"}

// withRestrictedEvidence adds a restricted record and one with no sensitivity
// set (which counts as restricted), and cites them from a link, an
// assumption, an outcome, and a story-less element, so every report that
// touches evidence has the chance to leak them.
func withRestrictedEvidence(t *testing.T) StatusInput {
	t.Helper()
	in := exampleInput(t)
	doc := in.Doc
	doc.Evidence = append(doc.Evidence,
		toc.Evidence{ID: "ev-secret", Title: "Board minutes", Excerpt: "SECRET-EXCERPT", Summary: "SECRET-SUMMARY", URL: "https://example.org/SECRET-URL", Sensitivity: toc.SensitivityRestricted},
		toc.Evidence{ID: "ev-unset", Title: "Interview notes", Excerpt: "UNSET-EXCERPT", Summary: "UNSET-SUMMARY"},
	)
	doc.Links[0].EvidenceIDs = append(doc.Links[0].EvidenceIDs, "ev-secret", "ev-unset")
	doc.Assumptions[0].EvidenceIDs = append(doc.Assumptions[0].EvidenceIDs, "ev-secret")
	doc.Outcomes[0].Provenance = &toc.Provenance{Basis: toc.BasisStated, EvidenceIDs: []string{"ev-secret", "ev-unset"}}
	g, err := graph.Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	st, err := measure.Evaluate(doc, g, exampleSets(t), mustDate(t, "2026-09-30"), toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	v, err := BuildView(doc, g, st, Options{})
	if err != nil {
		t.Fatal(err)
	}
	return StatusInput{View: v, Doc: doc, Status: st, Diff: in.Diff}
}

func TestRestrictedEvidenceIsAbsentFromEveryOutput(t *testing.T) {
	in := withRestrictedEvidence(t)
	sets := exampleSets(t)
	cost, err := burden.Compute(in.Doc)
	if err != nil {
		t.Fatal(err)
	}
	pathways, err := PathwayViews(in.Doc, nil, in.Status)
	if err != nil {
		t.Fatal(err)
	}
	full, err := RenderTemplate(`{{json .}}`, in.View)
	if err != nil {
		t.Fatal(err)
	}
	story, err := Story(in, sets, "aud-funders")
	if err != nil {
		t.Fatal(err)
	}
	outputs := map[string]string{
		"view as json": full, "definitions": Definitions(in.View, pathways), "assessment": Assessment(in.View, in.Doc),
		"exec": Exec(in), "results": Results(in), "assumptions": Assumptions(in), "interventions": Interventions(in),
		"evidence gaps": EvidenceGaps(in), "full": Full(in, Mermaid(in.View, MermaidOptions{ShowIDs: true})),
		"story": story, "learning review": LearningReview(LearningInput{StatusInput: in, Sets: sets, Period: "2026-Q3"}),
		"portfolio": Portfolio(in, cost), "narrative": Narrative(in.View, in.Doc, in.Status), "actors": Actors(in.View, in.Doc, in.Status),
		"outcome map": OutcomeMaps(in.View, pathways), "logframe": Logframes(in.View, pathways), "diagrams": LogicModelDiagrams(in.View, pathways),
		"metric selection": MetricSelection(in.View, in.Doc), "collection plan": CollectionPlan(in.Doc, cost),
	}
	for _, table := range CSVTables {
		csv, err := CSV(table, in.View, in.Doc)
		if err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		outputs["csv "+table] = csv
	}
	for name, out := range outputs {
		for _, secret := range secrets {
			if strings.Contains(out, secret) {
				t.Errorf("%s leaks %q", name, secret)
			}
		}
	}
	// The record itself is still citable by ID and title.
	if !strings.Contains(full, "ev-secret") || !strings.Contains(full, "Board minutes") {
		t.Error("a restricted record should remain identifiable by ID and title")
	}
}
