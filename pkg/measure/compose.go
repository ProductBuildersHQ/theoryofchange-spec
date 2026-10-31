package measure

import (
	"math"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// ComposeResult is the value of a composed indicator.
type ComposeResult struct {
	// Value is nil when a part is missing or the arithmetic has no answer.
	Value *float64
	// Missing names the parts with no value at the date.
	Missing []string
	// Quality is estimated when any part is, and reported otherwise.
	Quality toc.DataQuality
	// Direct is true when the indicator has an observation of its own, which
	// takes precedence over the composition.
	Direct bool
}

// Compose works out the value of an indicator that is composed of others, from
// each part's latest value at or before asOf. If any part has none, the
// result has no value and names the missing parts, so a partial figure is
// never passed off as complete. A ratio whose divisor is zero has no value.
// It reports false when the indicator is not composed.
func Compose(ind toc.Indicator, observations []toc.Observation, asOf time.Time) (ComposeResult, bool) {
	c := ind.ComposedOf
	if c == nil || len(c.IndicatorIDs) == 0 {
		return ComposeResult{}, false
	}
	var res ComposeResult
	if _, _, found := latestValue(ind.ID, observations, asOf); found {
		res.Direct = true
		return res, true
	}
	vals := make([]float64, 0, len(c.IndicatorIDs))
	res.Quality = toc.QualityReported
	for _, id := range c.IndicatorIDs {
		v, _, found := latestValue(id, observations, asOf)
		if !found {
			res.Missing = append(res.Missing, id)
			continue
		}
		vals = append(vals, v)
		if q, ok := latestQuality(id, observations, asOf); ok && q == toc.QualityEstimated {
			res.Quality = toc.QualityEstimated
		}
	}
	if len(res.Missing) > 0 {
		return res, true
	}
	out := vals[0]
	switch c.Op {
	case toc.CompositionRatio:
		if len(vals) != 2 || vals[1] == 0 {
			return res, true
		}
		out = vals[0] / vals[1]
	case toc.CompositionDifference:
		if len(vals) != 2 {
			return res, true
		}
		out = vals[0] - vals[1]
	case toc.CompositionProduct:
		for _, v := range vals[1:] {
			out *= v
		}
	case toc.CompositionSum:
		for _, v := range vals[1:] {
			out += v
		}
	default:
		return res, true
	}
	if math.IsNaN(out) || math.IsInf(out, 0) {
		return res, true
	}
	res.Value = &out
	return res, true
}

// latestQuality returns the quality of the latest numeric observation of an
// indicator at or before asOf.
func latestQuality(id string, observations []toc.Observation, asOf time.Time) (toc.DataQuality, bool) {
	var (
		best  time.Time
		q     toc.DataQuality
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
			best, q, found = at, ob.Quality, true
		}
	}
	return q, found
}
