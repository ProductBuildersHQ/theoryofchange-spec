package report

import (
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func TestMetricTreeReport(t *testing.T) {
	in := exampleInput(t)
	out, err := MetricTree(in, "")
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "metric-tree.md", out)
	if _, err := MetricTree(in, "pw-ghost"); err == nil {
		t.Error("an unknown pathway was accepted")
	}
}

func TestMetricTreeShowsFocusAndRoles(t *testing.T) {
	in := exampleInput(t)
	pw := &in.Doc.Pathways[0]
	var ns, gr string
	for _, ind := range in.Doc.Indicators {
		if in.Doc.IndicatorServes(ind, pw.ID) {
			if ns == "" {
				ns = ind.ID
			} else if gr == "" {
				gr = ind.ID
			}
		}
	}
	if gr == "" {
		t.Fatalf("pathway %s serves fewer than two indicators", pw.ID)
	}
	pw.Focus = &toc.Focus{NorthStarID: ns, GuardrailIDs: []string{gr}, Rationale: "WHY THIS ONE"}

	out, err := MetricTree(in, pw.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"- North Star: ", "- Guardrail: ", "Why: WHY THIS ONE", "(North Star)", "(guardrail)", "class i_" + mermaidSafe(ns) + " northstar"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "No North Star is chosen") {
		t.Errorf("a pathway with a North Star was reported as having none:\n%s", out)
	}
}

func TestMetricTreeOrdersTiersFromActivityToImpact(t *testing.T) {
	out, err := MetricTree(exampleInput(t), "")
	if err != nil {
		t.Fatal(err)
	}
	last := -1
	for _, tier := range []string{"| activity |", "| output |", "| outcome |", "| impact |"} {
		i := strings.Index(out, tier)
		if i < 0 {
			continue
		}
		if i < last {
			t.Errorf("tier row %q is out of order", tier)
		}
		last = i
	}
}
