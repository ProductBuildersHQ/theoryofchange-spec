package report

import (
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
)

func TestAssessmentGoldenOverTheIntermediaryExample(t *testing.T) {
	doc, err := impactio.LoadDocument("../../examples/intermediary/impact-model.yaml")
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "assessment-intermediary.md", Assessment(build(t, doc, nil, Options{}), doc))
}

func TestAssessmentFullFixture(t *testing.T) {
	doc := fullDoc(t)
	out := Assessment(build(t, doc, nil, Options{}), doc)
	for _, want := range []string{
		"# Open agent authorization standard adoption: assessment",
		"## Where each element comes from",
		"### Hypothesis",
		"### Inferred",
		"## Gaps",
		"## Questions for the organization",
		"We inferred this from your public materials",
		"We are guessing here",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("assessment is missing %q:\n%s", want, out)
		}
	}
}

func TestAssessmentListsEqualTheCounts(t *testing.T) {
	doc := fullDoc(t)
	v := build(t, doc, nil, Options{})
	got := map[impact.Basis]int{}
	unmarked := 0
	for _, e := range taggedElements(v) {
		if e.basis.Basis == "" {
			unmarked++
			continue
		}
		got[e.basis.Basis]++
	}
	for basis, n := range v.BasisCounts {
		if got[basis] != n {
			t.Errorf("%s: listed %d, counted %d", basis, got[basis], n)
		}
	}
	if unmarked != v.Unmarked {
		t.Errorf("unmarked: listed %d, counted %d", unmarked, v.Unmarked)
	}
}

func TestAssessmentFlagsStatedWithoutCitation(t *testing.T) {
	doc := fullDoc(t)
	doc.Outcomes[0].Provenance = &impact.Provenance{Basis: impact.BasisStated}
	out := Assessment(build(t, doc, nil, Options{}), doc)
	if !strings.Contains(out, "stated element(s) cite no source") {
		t.Errorf("a stated element with no evidence should be flagged:\n%s", out)
	}
}

func TestAssessmentGapsAndDecisions(t *testing.T) {
	doc := fullDoc(t)
	doc.Decisions = []impact.Decision{{ID: "dec-scope", Question: "Which region first?", Options: []string{"North", "South"}, Status: impact.DecisionOpen}}
	doc.Indicators[0].DataAvailability = impact.DataUnknown
	out := Assessment(build(t, doc, nil, Options{}), doc)
	for _, want := range []string{
		"Which region first? (`dec-scope`). Options: North; South",
		"Decision needed: Which region first? (dec-scope)",
		"Where could the data for",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
}

func TestAssessmentIsDeterministic(t *testing.T) {
	doc := fullDoc(t)
	v := build(t, doc, nil, Options{})
	first, second := Assessment(v, doc), Assessment(v, doc)
	if first != second {
		t.Error("two renders differ")
	}
}
