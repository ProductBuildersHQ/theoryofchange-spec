package report

import (
	"strings"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/measure"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

// exampleInput evaluates the intermediary example at two dates and returns
// the later snapshot with the changes since the earlier.
func exampleInput(t *testing.T) StatusInput {
	t.Helper()
	doc, err := tocio.LoadDocument("../../examples/intermediary/theory.yaml")
	if err != nil {
		t.Fatal(err)
	}
	g, err := graph.Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	sets, err := tocio.LoadObservations("../../examples/intermediary/observations-2026-q3.yaml")
	if err != nil {
		t.Fatal(err)
	}
	at := func(s string) time.Time {
		d, err := time.Parse(time.DateOnly, s)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	prev, err := measure.Evaluate(doc, g, sets, at("2026-06-30"), toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	cur, err := measure.Evaluate(doc, g, sets, at("2026-09-30"), toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	v, err := BuildView(doc, g, cur, Options{})
	if err != nil {
		t.Fatal(err)
	}
	return StatusInput{View: v, Doc: doc, Status: cur, Diff: measure.Diff(prev, cur)}
}

func TestStatusReportGoldens(t *testing.T) {
	in := exampleInput(t)
	golden(t, "status-exec.md", Exec(in))
	golden(t, "status-results.md", Results(in))
	golden(t, "status-assumptions.md", Assumptions(in))
	golden(t, "status-interventions.md", Interventions(in))
	golden(t, "status-evidence-gaps.md", EvidenceGaps(in))
}

func TestFullReportContainsEveryStatusReport(t *testing.T) {
	in := exampleInput(t)
	full := Full(in, Mermaid(in.View, MermaidOptions{}))
	for _, want := range []string{"executive summary", "## Pathway", "```mermaid", "results", "assumptions", "interventions", "evidence gaps"} {
		if !strings.Contains(full, want) {
			t.Errorf("full report is missing %q", want)
		}
	}
	if !strings.HasPrefix(full, "# ") || strings.Contains(full, "\n# ") {
		t.Error("the full report should have exactly one title, at the top")
	}
}

func TestEveryStatusCitesItsInputs(t *testing.T) {
	in := exampleInput(t)
	exec, results := Exec(in), Results(in)
	// An outcome flagged in the summary says which indicator or precondition drives it.
	if !strings.Contains(exec, "of the way to target;") && !strings.Contains(exec, "held back by") {
		t.Errorf("the attention table gives no reason:\n%s", exec)
	}
	// Every indicator row in the results says why it has its status.
	for _, line := range strings.Split(results, "\n") {
		if strings.HasPrefix(line, "| ") && !strings.HasPrefix(line, "| Indicator") && !strings.HasPrefix(line, "|--") {
			cells := strings.Split(line, " | ")
			if last := strings.TrimSuffix(strings.TrimSpace(cells[len(cells)-1]), " |"); last == "" || last == "—" {
				t.Errorf("an indicator row has no reason: %s", line)
			}
		}
	}
	if !strings.Contains(exec, "Confidence in a link is 0.4 × evidence + 0.4 × assumptions + 0.2 × rationale") {
		t.Error("the method note must print the confidence weights")
	}
}

func TestExecWithoutADiffOrWithABrokenAssumption(t *testing.T) {
	in := exampleInput(t)
	in.Diff = nil
	if strings.Contains(Exec(in), "## Changes since") {
		t.Error("changes shown with no previous snapshot")
	}
	for i := range in.View.Assumptions {
		in.View.Assumptions[i].Status = string(toc.AssumptionInvalidated)
	}
	exec := Exec(in)
	if !strings.Contains(exec, "What should change now that the assumption") {
		t.Errorf("no prompt for a broken assumption:\n%s", exec)
	}
}

func TestExecIsDeterministic(t *testing.T) {
	first, second := Exec(exampleInput(t)), Exec(exampleInput(t))
	if first != second {
		t.Error("two renders differ")
	}
}
