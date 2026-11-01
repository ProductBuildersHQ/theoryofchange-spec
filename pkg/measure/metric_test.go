package measure

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

func rollupDoc(agg impact.Aggregation) *impact.Model {
	d := impact.NewModel("t", "p")
	d.Metrics = []impact.Metric{
		{ID: "met", Name: "M", Definition: "D", Formula: "F", Aggregation: agg, WeightMetricID: weightIf(agg)},
		{ID: "met-w", Name: "W", Definition: "D", Formula: "F", Aggregation: impact.AggregationSum},
	}
	d.Pathways = []impact.Pathway{{ID: "pw-1", Title: "1"}, {ID: "pw-2", Title: "2"}, {ID: "pw-3", Title: "3"}}
	d.Indicators = []impact.Indicator{
		{ID: "total", Measures: "oc", Name: "t", MetricID: "met", RollupFrom: []string{"a", "b", "c"}},
		{ID: "a", Measures: "oc", Name: "a", MetricID: "met", PathwayIDs: []string{"pw-1"}},
		{ID: "b", Measures: "oc", Name: "b", MetricID: "met", PathwayIDs: []string{"pw-2"}},
		{ID: "c", Measures: "oc", Name: "c", MetricID: "met", PathwayIDs: []string{"pw-3"}},
		{ID: "wa", Measures: "oc", Name: "wa", MetricID: "met-w", PathwayIDs: []string{"pw-1"}},
		{ID: "wb", Measures: "oc", Name: "wb", MetricID: "met-w", PathwayIDs: []string{"pw-2"}},
		{ID: "wc", Measures: "oc", Name: "wc", MetricID: "met-w", PathwayIDs: []string{"pw-3"}},
	}
	return d
}

func weightIf(agg impact.Aggregation) string {
	if agg == impact.AggregationWeightedMean {
		return "met-w"
	}
	return ""
}

func values(pairs ...any) []impact.Observation {
	var out []impact.Observation
	for i := 0; i+2 < len(pairs); i += 3 {
		v := pairs[i+2].(float64)
		out = append(out, impact.Observation{IndicatorID: pairs[i].(string), Date: pairs[i+1].(string), Value: &v})
	}
	return out
}

func TestRollupAggregations(t *testing.T) {
	asOf := day("2026-12-31")
	obs := values(
		"a", "2026-12-01", 10.0, "b", "2026-12-01", 20.0, "c", "2026-12-01", 20.0,
		"wa", "2026-12-01", 1.0, "wb", "2026-12-01", 1.0, "wc", "2026-12-01", 2.0,
	)
	tests := []struct {
		agg  impact.Aggregation
		want float64
	}{
		{impact.AggregationSum, 50},
		{impact.AggregationMean, 50.0 / 3},
		{impact.AggregationWeightedMean, (10*1 + 20*1 + 20*2) / 4.0},
		{impact.AggregationMedian, 20},
		{impact.AggregationMode, 20},
		{impact.AggregationMin, 10},
		{impact.AggregationMax, 20},
		{impact.AggregationLatest, 20},
	}
	for _, tt := range tests {
		t.Run(string(tt.agg), func(t *testing.T) {
			d := rollupDoc(tt.agg)
			res, ok := Rollup(d, d.Indicators[0], obs, asOf)
			if !ok || res.Value == nil || math.Abs(*res.Value-tt.want) > 1e-9 {
				t.Fatalf("got %+v ok=%v, want %v", res, ok, tt.want)
			}
			if res.State != impact.StateBaselineOnly || len(res.Missing) != 0 {
				t.Errorf("state %s missing %v", res.State, res.Missing)
			}
		})
	}
}

func TestRollupMedianOfAnEvenCount(t *testing.T) {
	d := rollupDoc(impact.AggregationMedian)
	d.Indicators[0].RollupFrom = []string{"a", "b"}
	res, _ := Rollup(d, d.Indicators[0], values("a", "2026-12-01", 10.0, "b", "2026-12-01", 20.0), day("2026-12-31"))
	if res.Value == nil || *res.Value != 15 {
		t.Errorf("median = %v", res.Value)
	}
}

func TestRollupLatestAndTheDateCutoff(t *testing.T) {
	d := rollupDoc(impact.AggregationLatest)
	obs := values("a", "2026-12-01", 10.0, "b", "2026-12-20", 77.0, "c", "2026-12-10", 5.0, "b", "2027-02-01", 999.0)
	res, _ := Rollup(d, d.Indicators[0], obs, day("2026-12-31"))
	if res.Value == nil || *res.Value != 77 {
		t.Errorf("latest = %v (a later observation must be ignored)", res.Value)
	}
}

func TestRollupReportsMissingSources(t *testing.T) {
	d := rollupDoc(impact.AggregationSum)
	res, ok := Rollup(d, d.Indicators[0], values("a", "2026-12-01", 10.0), day("2026-12-31"))
	if !ok || res.Value != nil || res.State != impact.StateNoData || !slices.Equal(res.Missing, []string{"b", "c"}) {
		t.Errorf("result = %+v", res)
	}
	// A weighted mean missing a weight is missing that source.
	w := rollupDoc(impact.AggregationWeightedMean)
	res, _ = Rollup(w, w.Indicators[0], values("a", "2026-12-01", 1.0, "b", "2026-12-01", 1.0, "c", "2026-12-01", 1.0, "wa", "2026-12-01", 1.0, "wb", "2026-12-01", 1.0), day("2026-12-31"))
	if res.Value != nil || !slices.Equal(res.Missing, []string{"c"}) {
		t.Errorf("missing weight: %+v", res)
	}
}

func TestRollupDirectObservationsAreReportedAlongside(t *testing.T) {
	d := rollupDoc(impact.AggregationSum)
	obs := values("a", "2026-12-01", 1.0, "b", "2026-12-01", 2.0, "c", "2026-12-01", 3.0, "total", "2026-12-15", 100.0)
	res, _ := Rollup(d, d.Indicators[0], obs, day("2026-12-31"))
	if !res.Direct || res.DirectValue == nil || *res.DirectValue != 100 || res.Value == nil || *res.Value != 6 {
		t.Errorf("direct value must take precedence and the computed one stay visible: %+v", res)
	}
}

func TestRollupDoesNotApply(t *testing.T) {
	d := rollupDoc(impact.AggregationNone)
	if _, ok := Rollup(d, d.Indicators[0], nil, time.Now()); ok {
		t.Error("a metric that cannot be combined rolled up")
	}
	s := rollupDoc(impact.AggregationSum)
	if _, ok := Rollup(s, s.Indicators[1], nil, time.Now()); ok {
		t.Error("an indicator with no rollupFrom rolled up")
	}
}
