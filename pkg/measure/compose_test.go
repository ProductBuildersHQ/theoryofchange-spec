package measure

import (
	"math"
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func composed(op toc.CompositionOp, parts ...string) toc.Indicator {
	return toc.Indicator{ID: "x", Measures: "oc", Name: "x", ComposedOf: &toc.Composition{Op: op, IndicatorIDs: parts}}
}

func TestComposeArithmetic(t *testing.T) {
	asOf := day("2026-12-31")
	obs := values("cost", "2026-12-01", 900.0, "n", "2026-12-01", 12.0, "m", "2026-12-01", 3.0)
	tests := []struct {
		name string
		ind  toc.Indicator
		want float64
	}{
		{"ratio is the first over the second", composed(toc.CompositionRatio, "cost", "n"), 75},
		{"ratio is not commutative", composed(toc.CompositionRatio, "n", "cost"), 12.0 / 900},
		{"difference is the first less the second", composed(toc.CompositionDifference, "cost", "n"), 888},
		{"product multiplies every part", composed(toc.CompositionProduct, "cost", "n", "m"), 32400},
		{"sum adds every part", composed(toc.CompositionSum, "cost", "n", "m"), 915},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, ok := Compose(tt.ind, obs, asOf)
			if !ok || res.Value == nil || math.Abs(*res.Value-tt.want) > 1e-9 {
				t.Fatalf("Compose = %+v, %v; want %v", res, ok, tt.want)
			}
		})
	}
}

// A figure is never worked out from part of its parts, from a division by
// zero, or over an indicator that already has its own observation.
func TestComposeWithholdsWhatCannotBeWorkedOut(t *testing.T) {
	asOf := day("2026-12-31")

	res, _ := Compose(composed(toc.CompositionRatio, "cost", "n"), values("cost", "2026-12-01", 900.0), asOf)
	if res.Value != nil || !slices.Equal(res.Missing, []string{"n"}) {
		t.Errorf("a missing part: %+v", res)
	}

	res, _ = Compose(composed(toc.CompositionRatio, "cost", "n"), values("cost", "2026-12-01", 900.0, "n", "2026-12-01", 0.0), asOf)
	if res.Value != nil {
		t.Errorf("a division by zero was given the value %v", *res.Value)
	}

	res, _ = Compose(composed(toc.CompositionRatio, "cost", "n"), values("x", "2026-12-02", 5.0, "cost", "2026-12-01", 900.0, "n", "2026-12-01", 12.0), asOf)
	if !res.Direct || res.Value != nil {
		t.Errorf("an indicator with its own observation: %+v", res)
	}

	res, _ = Compose(composed(toc.CompositionRatio, "cost", "n"), values("cost", "2027-02-01", 900.0, "n", "2027-02-01", 12.0), asOf)
	if res.Value != nil {
		t.Errorf("parts observed after the date were used: %+v", res)
	}

	if _, ok := Compose(toc.Indicator{ID: "plain"}, nil, asOf); ok {
		t.Error("an indicator that is not composed was composed")
	}
}

func TestComposeIsAnEstimateWhenAnyPartIs(t *testing.T) {
	asOf := day("2026-12-31")
	est := func(q toc.DataQuality) []toc.Observation {
		obs := values("cost", "2026-12-01", 900.0, "n", "2026-12-01", 12.0)
		obs[1].Quality = q
		return obs
	}
	if res, _ := Compose(composed(toc.CompositionRatio, "cost", "n"), est(toc.QualityEstimated), asOf); res.Quality != toc.QualityEstimated {
		t.Errorf("quality = %q, want estimated", res.Quality)
	}
	if res, _ := Compose(composed(toc.CompositionRatio, "cost", "n"), est(toc.QualityVerified), asOf); res.Quality != toc.QualityReported {
		t.Errorf("quality = %q, want reported", res.Quality)
	}
}

// A composed indicator takes its place in the evaluated status like any other
// with an observation, and can itself be a part of another composition.
func TestComposedValueReachesTheStatus(t *testing.T) {
	d := toc.NewDocument("t", "p")
	d.Indicators = []toc.Indicator{
		{ID: "per", Measures: "oc", Name: "per", ComposedOf: &toc.Composition{Op: toc.CompositionRatio, IndicatorIDs: []string{"cost", "n"}}},
		{ID: "twice", Measures: "oc", Name: "twice", ComposedOf: &toc.Composition{Op: toc.CompositionProduct, IndicatorIDs: []string{"per", "two"}}},
		{ID: "cost", Measures: "oc", Name: "cost"},
		{ID: "n", Measures: "oc", Name: "n"},
		{ID: "two", Measures: "oc", Name: "two"},
	}
	sets := []*toc.ObservationSet{{Observations: values("cost", "2026-12-01", 900.0, "n", "2026-12-01", 12.0, "two", "2026-12-01", 2.0)}}
	obs := derivedObservations(d, sets, day("2026-12-31"))
	got := map[string]float64{}
	for _, o := range obs {
		if o.Value != nil {
			got[o.IndicatorID] = *o.Value
		}
	}
	if got["per"] != 75 || got["twice"] != 150 {
		t.Errorf("derived values = %v, want per 75 and twice 150", got)
	}
}
