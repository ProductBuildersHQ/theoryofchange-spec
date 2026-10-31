package lint

import (
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// A common metric counts for a pathway when an indicator of it measures an
// outcome the pathway lists, with no pathwayIds on the indicator. The rules
// that ask which pathways a metric is measured in share one answer with the
// rest of the module (Document.IndicatorServes).
func TestCommonMetricNeedsNoExplicitPathwayList(t *testing.T) {
	d := clean()
	d.Pathways = []toc.Pathway{
		{ID: "pw-1", Title: "One", OutcomeIDs: []string{"oc-s1"}},
		{ID: "pw-2", Title: "Two", OutcomeIDs: []string{"oc-s2"}},
	}
	d.Metrics = append(d.Metrics, toc.Metric{ID: "met-1", Name: "M", Definition: "D", Formula: "F", Layer: toc.LayerCommon})
	addIndicator(d, "ind-1", "oc-s1", func(i *toc.Indicator) { i.MetricID = "met-1" })

	if !slices.Contains(fires(t, d, "TOC-MET-001"), "met-1") {
		t.Fatalf("the common metric is measured in only one of two pathways, but TOC-MET-001 reported %v", fires(t, d, "TOC-MET-001"))
	}
	addIndicator(d, "ind-2", "oc-s2", func(i *toc.Indicator) { i.MetricID = "met-1" })
	if got := fires(t, d, "TOC-MET-001"); slices.Contains(got, "met-1") {
		t.Errorf("both pathways measure the metric through the outcomes they list, but TOC-MET-001 reported %v", got)
	}
}

// A weighted mean needs its weight measured wherever it is used, and
// membership is derived the same way.
func TestWeightedMeanWeightIsFoundThroughTheMeasuredOutcome(t *testing.T) {
	d := clean()
	d.Pathways = []toc.Pathway{{ID: "pw-1", Title: "One", OutcomeIDs: []string{"oc-s1", "oc-s2"}}}
	d.Metrics = append(d.Metrics,
		toc.Metric{ID: "met-wt", Name: "Weight", Definition: "D", Formula: "F"},
		toc.Metric{ID: "met-w", Name: "W", Definition: "D", Formula: "F", Aggregation: toc.AggregationWeightedMean, WeightMetricID: "met-wt"})
	addIndicator(d, "ind-w", "oc-s1", func(i *toc.Indicator) { i.MetricID = "met-w" })
	if !slices.Contains(fires(t, d, "TOC-MET-003"), "met-w") {
		t.Fatal("the pathway uses the weighted mean without its weight, but TOC-MET-003 was silent")
	}
	addIndicator(d, "ind-wt", "oc-s2", func(i *toc.Indicator) { i.MetricID = "met-wt" })
	if slices.Contains(fires(t, d, "TOC-MET-003"), "met-w") {
		t.Error("the weight is measured through an outcome the pathway lists, but TOC-MET-003 still reported it")
	}
}
