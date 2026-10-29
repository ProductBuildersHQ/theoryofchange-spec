package validate

import (
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func TestSelectionAndCollectionIssues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(d *toc.Document)
		code   string
		path   string
	}{
		{"criterion without name", func(d *toc.Document) { d.SelectionCriteria[0].Name = " " }, "VAL-SEL-001", "/selectionCriteria/0/name"},
		{"criterion subject invalid", func(d *toc.Document) { d.SelectionCriteria[0].AppliesTo = "metric" }, "VAL-ENUM-001", "/selectionCriteria/0/appliesTo"},
		{"collection mode invalid", func(d *toc.Document) { d.Indicators[0].Collection.Mode = "telepathy" }, "VAL-ENUM-001", "/indicators/0/collection/mode"},
		{"collection period invalid", func(d *toc.Document) { d.Indicators[0].Collection.Period = "monthly" }, "VAL-DUR-001", "/indicators/0/collection/period"},
		{"minutes without a period", func(d *toc.Document) { d.Indicators[0].Collection.Period = "" }, "VAL-SEL-003", "/indicators/0/collection/period"},
		{"questions without a period", func(d *toc.Document) {
			d.Indicators[1].Collection.Period = ""
			d.Indicators[1].Collection.MinutesPerPeriod = 0
		}, "VAL-SEL-003", "/indicators/1/collection/period"},
		{"selection status invalid", func(d *toc.Document) { d.Indicators[0].Selection.Status = "chosen" }, "VAL-ENUM-001", "/indicators/0/selection/status"},
		{"score too high", func(d *toc.Document) { d.Indicators[0].Selection.Scores[0].Score = 6 }, "VAL-SEL-002", "/indicators/0/selection/scores/0/score"},
		{"score too low", func(d *toc.Document) { d.Indicators[0].Selection.Scores[0].Score = 0 }, "VAL-SEL-002", "/indicators/0/selection/scores/0/score"},
		{"score criterion unknown", func(d *toc.Document) { d.Indicators[0].Selection.Scores[0].CriterionID = "crit-ghost" }, "VAL-REF-001", "/indicators/0/selection/scores/0/criterionId"},
		{"score criterion is not a criterion", func(d *toc.Document) { d.Indicators[0].Selection.Scores[0].CriterionID = "oc-idp-support" }, "VAL-REF-002", "/indicators/0/selection/scores/0/criterionId"},
		{"criterion scores outcomes", func(d *toc.Document) { d.SelectionCriteria[0].AppliesTo = toc.SubjectOutcome }, "VAL-SEL-004", "/indicators/0/selection/scores/0/criterionId"},
		{"criterion scored twice", func(d *toc.Document) { d.Indicators[0].Selection.Scores[1].CriterionID = "crit-decision-value" }, "VAL-SEL-005", "/indicators/0/selection/scores/1/criterionId"},
		{"budget range reversed", func(d *toc.Document) { d.Budget.TargetSelectedMetrics.Min = 20 }, "VAL-SEL-006", "/measurementBudget/targetSelectedMetrics"},
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
	if issues := Document(fullDoc(t)); len(issues) != 0 {
		t.Errorf("the full fixture has issues:%s", describe(issues))
	}
}
