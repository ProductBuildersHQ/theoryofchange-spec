package validate

import (
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

func TestHeadlineAndFocusIssues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(d *impact.Model)
		code   string
		path   string
	}{
		{"headline names an unknown indicator", func(d *impact.Model) {
			d.Audiences[0].Headline = []impact.HeadlineMetric{{IndicatorID: "ind-ghost"}}
		}, "VAL-REF-001", "/audiences/0/headline/0/indicatorId"},
		{"headline names an outcome", func(d *impact.Model) {
			d.Audiences[0].Headline = []impact.HeadlineMetric{{IndicatorID: d.Outcomes[0].ID}}
		}, "VAL-REF-002", "/audiences/0/headline/0/indicatorId"},
		{"headline entry names nothing", func(d *impact.Model) {
			d.Audiences[0].Headline = []impact.HeadlineMetric{{IndicatorID: " "}}
		}, "VAL-AUD-003", "/audiences/0/headline/0/indicatorId"},
		{"headline lists an indicator twice", func(d *impact.Model) {
			id := d.Indicators[0].ID
			d.Audiences[0].Headline = []impact.HeadlineMetric{{IndicatorID: id}, {IndicatorID: id, Label: "again"}}
		}, "VAL-AUD-002", "/audiences/0/headline/1/indicatorId"},
		{"north star is unknown", func(d *impact.Model) {
			d.Pathways[0].Focus = &impact.Focus{NorthStarID: "ind-ghost"}
		}, "VAL-REF-001", "/pathways/0/focus/northStarId"},
		{"one metric is an outcome", func(d *impact.Model) {
			d.Pathways[0].Focus = &impact.Focus{OneMetricID: d.Outcomes[0].ID}
		}, "VAL-REF-002", "/pathways/0/focus/oneMetricId"},
		{"guardrail is unknown", func(d *impact.Model) {
			d.Pathways[0].Focus = &impact.Focus{GuardrailIDs: []string{"ind-ghost"}}
		}, "VAL-REF-001", "/pathways/0/focus/guardrailIds/0"},
		{"guardrail is listed twice", func(d *impact.Model) {
			id := d.Indicators[0].ID
			d.Pathways[0].Focus = &impact.Focus{GuardrailIDs: []string{id, id}}
		}, "VAL-PTH-006", "/pathways/0/focus/guardrailIds/1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := fullDoc(t)
			tt.mutate(doc)
			for _, i := range Document(doc) {
				if i.Code == tt.code && i.Path == tt.path {
					return
				}
			}
			t.Fatalf("expected %s at %s; got:%s", tt.code, tt.path, describe(Document(doc)))
		})
	}
}

// A headline list, a focus, a product pathway, and a mission and vision are
// all valid together, and a document that uses none of them is unaffected.
func TestHeadlineFocusMissionAndProductAreValid(t *testing.T) {
	doc := fullDoc(t)
	before := len(Document(doc))
	ids := []string{doc.Indicators[0].ID, doc.Indicators[1].ID, doc.Indicators[2].ID}
	doc.Audiences[0].Headline = []impact.HeadlineMetric{
		{IndicatorID: ids[0], Label: "People we reached", Note: "The number the community asks for"},
		{IndicatorID: ids[1]},
	}
	doc.Pathways[0].Kind = impact.PathwayProduct
	doc.Pathways[0].Focus = &impact.Focus{NorthStarID: ids[0], OneMetricID: ids[1], GuardrailIDs: []string{ids[2]}, Rationale: "Why these"}
	doc.Context.Mission = "Why we exist."
	doc.Context.Vision = "The future we work toward."
	if got := len(Document(doc)); got != before {
		t.Errorf("valid new fields changed the issue count from %d to %d:%s", before, got, describe(Document(doc)))
	}
}
