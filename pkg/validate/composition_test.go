package validate

import (
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// Each way a composition can be wrong is reported at the place that is wrong.
// In the full fixture, ind-idp-count-eu is composed as a ratio of
// ind-adoption-stories to ind-idp-count-na, and ind-idp-total rolls up.
func TestCompositionIssues(t *testing.T) {
	comp := func(d *toc.Document, id string) *toc.Composition {
		for i := range d.Indicators {
			if d.Indicators[i].ID == id {
				return d.Indicators[i].ComposedOf
			}
		}
		t.Fatalf("no indicator %s", id)
		return nil
	}
	set := func(d *toc.Document, id string, c *toc.Composition) {
		for i := range d.Indicators {
			if d.Indicators[i].ID == id {
				d.Indicators[i].ComposedOf = c
			}
		}
	}
	const eu = "ind-idp-count-eu"
	tests := []struct {
		name   string
		mutate func(d *toc.Document)
		code   string
		path   string
	}{
		{"no operation", func(d *toc.Document) { comp(d, eu).Op = "" }, "VAL-MET-007", "/indicators/4/composedOf/op"},
		{"unknown operation", func(d *toc.Document) { comp(d, eu).Op = "average" }, "VAL-ENUM-001", "/indicators/4/composedOf/op"},
		{"a ratio of three", func(d *toc.Document) { c := comp(d, eu); c.IndicatorIDs = append(c.IndicatorIDs, "ind-idp-total") }, "VAL-MET-007", "/indicators/4/composedOf/indicatorIds"},
		{"a difference of one", func(d *toc.Document) {
			c := comp(d, eu)
			c.Op, c.IndicatorIDs = toc.CompositionDifference, c.IndicatorIDs[:1]
		}, "VAL-MET-007", "/indicators/4/composedOf/indicatorIds"},
		{"a sum of one", func(d *toc.Document) { c := comp(d, eu); c.Op, c.IndicatorIDs = toc.CompositionSum, c.IndicatorIDs[:1] }, "VAL-MET-007", "/indicators/4/composedOf/indicatorIds"},
		{"a part that does not exist", func(d *toc.Document) { comp(d, eu).IndicatorIDs[0] = "ind-ghost" }, "VAL-REF-001", "/indicators/4/composedOf/indicatorIds/0"},
		{"composed of itself", func(d *toc.Document) { comp(d, eu).IndicatorIDs[1] = eu }, "VAL-MET-008", "/indicators/4/composedOf/indicatorIds/1"},
		{"composed in a circle", func(d *toc.Document) {
			set(d, "ind-idp-count-na", &toc.Composition{Op: toc.CompositionSum, IndicatorIDs: []string{eu, "ind-adoption-stories"}})
		}, "VAL-MET-008", "/indicators/4/composedOf/indicatorIds/1"},
		{"rolled up and composed", func(d *toc.Document) {
			set(d, "ind-idp-total", &toc.Composition{Op: toc.CompositionSum, IndicatorIDs: []string{eu, "ind-idp-count-na"}})
		}, "VAL-MET-009", "/indicators/5/composedOf"},
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

func TestACleanCompositionIsValid(t *testing.T) {
	for _, i := range Document(fullDoc(t)) {
		if i.Severity == SeverityError {
			t.Errorf("the full fixture is not valid: %s %s: %s", i.Code, i.Path, i.Message)
		}
	}
}
