package report

import (
	"math"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/burden"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func TestWeightedScore(t *testing.T) {
	doc := &toc.Document{SelectionCriteria: []toc.SelectionCriterion{
		{ID: "a", Weight: 3, AppliesTo: toc.SubjectIndicator},
		{ID: "cost", Weight: 2, Inverted: true, AppliesTo: toc.SubjectIndicator},
		{ID: "plain", AppliesTo: toc.SubjectIndicator}, // no weight counts as 1
	}}
	tests := []struct {
		name   string
		scores []toc.CriterionScore
		want   float64
		ok     bool
	}{
		{"one criterion", []toc.CriterionScore{{CriterionID: "a", Score: 4}}, 4, true},
		{"weights", []toc.CriterionScore{{CriterionID: "a", Score: 5}, {CriterionID: "plain", Score: 1}}, 4, true}, // (15+1)/4
		{"inverted flips", []toc.CriterionScore{{CriterionID: "cost", Score: 1}}, 5, true},
		{"mixed", []toc.CriterionScore{{CriterionID: "a", Score: 4}, {CriterionID: "cost", Score: 2}}, 4, true}, // (12+8)/5
		{"rounds to two places", []toc.CriterionScore{{CriterionID: "a", Score: 4}, {CriterionID: "plain", Score: 2}}, 3.5, true},
		{"unknown criterion ignored", []toc.CriterionScore{{CriterionID: "ghost", Score: 5}}, 0, false},
		{"nothing", nil, 0, false},
	}
	for _, tt := range tests {
		got, ok := WeightedScore(doc, tt.scores)
		if ok != tt.ok || math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("%s: WeightedScore = %v, %v; want %v, %v", tt.name, got, ok, tt.want, tt.ok)
		}
	}
}

func TestMetricSelectionGolden(t *testing.T) {
	doc := fullDoc(t)
	golden(t, "metric-selection.md", MetricSelection(build(t, doc, nil, Options{}), doc))
}

func TestMetricSelectionOrdersBestFirst(t *testing.T) {
	doc := fullDoc(t)
	out := MetricSelection(build(t, doc, nil, Options{}), doc)
	// ind-idp-count scores 5 and 2 (inverted) on weights 3 and 2: (15+8)/5 = 4.6;
	// ind-idp-roadmaps scores 3 on one criterion.
	if a, b := strings.Index(out, "ind-idp-count"), strings.Index(out, "ind-idp-roadmaps"); a < 0 || b < 0 || a > b {
		t.Errorf("the higher score should come first:\n%s", out)
	}
	if !strings.Contains(out, "4.60") || !strings.Contains(out, "Cost to collect (lower is better)") {
		t.Errorf("missing the weighted score or the inverted marker:\n%s", out)
	}
}

func TestMetricSelectionEmpty(t *testing.T) {
	doc := &toc.Document{}
	if out := MetricSelection(&View{}, doc); !strings.Contains(out, "No indicator has a selection record") {
		t.Errorf("empty output = %q", out)
	}
}

func TestCollectionPlanGolden(t *testing.T) {
	doc := fullDoc(t)
	rep, err := burden.Compute(doc)
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "collection-plan.md", CollectionPlan(doc, rep))
}

func TestCollectionPlanTotalsEqualBurden(t *testing.T) {
	doc := fullDoc(t)
	rep, err := burden.Compute(doc)
	if err != nil {
		t.Fatal(err)
	}
	out := CollectionPlan(doc, rep)
	want := rounded(rep.Overall.MinutesPerWeek)
	if !strings.Contains(out, " take "+want+" staff minutes a week") {
		t.Errorf("totals do not match the burden report (%s):\n%s", want, out)
	}
	csv, err := CSV("collection-plan", nil, doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range rep.Lines {
		if !strings.Contains(csv, l.ID+",") {
			t.Errorf("csv is missing %s", l.ID)
		}
	}
	if _, err := CSV("collection-plan", nil, nil); err == nil {
		t.Error("collection-plan without a document was accepted")
	}
}
