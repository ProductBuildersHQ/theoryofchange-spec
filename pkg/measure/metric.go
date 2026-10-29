package measure

import (
	"math"
	"slices"
	"sort"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// RollupResult is the value an indicator takes by combining other indicators
// of the same metric.
type RollupResult struct {
	IndicatorID string `json:"indicatorId"`
	// Value is the combined value; nil when a source has no value.
	Value *float64 `json:"value,omitempty"`
	// Missing lists the sources that had no value at the date, which is why
	// the roll-up is no data.
	Missing []string `json:"missing,omitempty"`
	// Direct is true when the indicator has an observation of its own at the
	// date. The direct value takes precedence; Value is still the computed
	// one, for comparison.
	Direct bool `json:"direct,omitempty"`
	// DirectValue is the indicator's own latest value when Direct.
	DirectValue *float64  `json:"directValue,omitempty"`
	State       toc.State `json:"state"`
}

// latestValue returns the latest numeric observation of an indicator at or
// before asOf.
func latestValue(id string, observations []toc.Observation, asOf time.Time) (float64, time.Time, bool) {
	var (
		best  time.Time
		val   float64
		found bool
	)
	for _, ob := range observations {
		if ob.IndicatorID != id || ob.Value == nil {
			continue
		}
		at, ok := parseDate(ob.Date)
		if !ok || at.After(asOf) {
			continue
		}
		if !found || !at.Before(best) {
			best, val, found = at, *ob.Value, true
		}
	}
	return val, best, found
}

// Rollup computes the value of an indicator that rolls up from other
// indicators, using the aggregation of its metric. Each source contributes
// its latest value at or before asOf. If any source has none, the result is
// no data and names the missing sources, so a partial total is never passed
// off as complete. A weighted mean weights each source by the latest value of
// the weight metric's indicator in the source's pathway.
//
// The result's state is no data when a value is missing, and otherwise
// baseline only, since a computed value carries no targets of its own; the
// indicator's own status, from EvaluateIndicator, uses the value as an
// observation. It reports false when the indicator does not roll up.
func Rollup(doc *toc.Document, ind toc.Indicator, observations []toc.Observation, asOf time.Time) (RollupResult, bool) {
	if len(ind.RollupFrom) == 0 || ind.MetricID == "" {
		return RollupResult{}, false
	}
	m, ok := doc.Metric(ind.MetricID)
	if !ok || m.Aggregation == toc.AggregationNone {
		return RollupResult{}, false
	}
	res := RollupResult{IndicatorID: ind.ID, State: toc.StateNoData}
	if v, _, found := latestValue(ind.ID, observations, asOf); found {
		res.Direct, res.DirectValue = true, &v
	}

	type source struct {
		id     string
		value  float64
		at     time.Time
		weight float64
	}
	var sources []source
	for _, id := range ind.RollupFrom {
		v, at, found := latestValue(id, observations, asOf)
		if !found {
			res.Missing = append(res.Missing, id)
			continue
		}
		src := source{id: id, value: v, at: at, weight: 1}
		if m.Aggregation == toc.AggregationWeightedMean {
			w, ok := weightFor(doc, id, m.WeightMetricID, observations, asOf)
			if !ok {
				res.Missing = append(res.Missing, id)
				continue
			}
			src.weight = w
		}
		sources = append(sources, src)
	}
	if len(res.Missing) > 0 {
		return res, true
	}

	vals := make([]float64, len(sources))
	for i, s := range sources {
		vals[i] = s.value
	}
	var out float64
	switch m.Aggregation {
	case toc.AggregationSum:
		for _, v := range vals {
			out += v
		}
	case toc.AggregationMean:
		out = mean(vals)
	case toc.AggregationWeightedMean:
		var sum, weights float64
		for _, s := range sources {
			sum += s.value * s.weight
			weights += s.weight
		}
		if weights == 0 {
			res.Missing = slices.Clone(ind.RollupFrom)
			return res, true
		}
		out = sum / weights
	case toc.AggregationMedian:
		out = median(vals)
	case toc.AggregationMode:
		out = mode(vals)
	case toc.AggregationMin:
		out = slices.Min(vals)
	case toc.AggregationMax:
		out = slices.Max(vals)
	case toc.AggregationLatest:
		latest := sources[0]
		for _, s := range sources[1:] {
			if !s.at.Before(latest.at) {
				latest = s
			}
		}
		out = latest.value
	case toc.AggregationNone:
		return RollupResult{}, false
	}
	res.Value, res.State = &out, toc.StateBaselineOnly
	return res, true
}

// weightFor finds the latest value of the weight metric in the same pathway
// as the source indicator.
func weightFor(doc *toc.Document, sourceID, weightMetric string, observations []toc.Observation, asOf time.Time) (float64, bool) {
	src, ok := doc.Indicator(sourceID)
	if !ok {
		return 0, false
	}
	for _, cand := range doc.Indicators {
		if cand.MetricID != weightMetric || cand.Provenance.IsRejected() {
			continue
		}
		if !sharesPathway(src.PathwayIDs, cand.PathwayIDs) {
			continue
		}
		if v, _, found := latestValue(cand.ID, observations, asOf); found {
			return v, true
		}
	}
	return 0, false
}

func sharesPathway(a, b []string) bool {
	return slices.ContainsFunc(a, func(id string) bool { return slices.Contains(b, id) })
}

func mean(v []float64) float64 {
	var sum float64
	for _, x := range v {
		sum += x
	}
	return sum / float64(len(v))
}

func median(v []float64) float64 {
	s := slices.Clone(v)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

// mode returns the most frequent value; of equal counts, the smallest.
func mode(v []float64) float64 {
	counts := map[float64]int{}
	for _, x := range v {
		counts[x]++
	}
	best, bestN := math.Inf(1), 0
	for x, n := range counts {
		if n > bestN || (n == bestN && x < best) {
			best, bestN = x, n
		}
	}
	return best
}
