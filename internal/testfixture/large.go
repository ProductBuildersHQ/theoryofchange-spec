// Package testfixture builds impact models for tests and benchmarks that need a
// size the hand-written fixtures do not reach.
package testfixture

import (
	"fmt"
	"time"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// Large builds a layered impact model of 200 outcomes, 400 links, and 300
// indicators with 5,000 observations: the size the specification budgets a
// second for.
func Large() (*impact.Model, []*impact.ObservationSet) {
	const (
		outcomes   = 200
		indicators = 300
		perLayer   = 40 // five layers of 40
	)
	d := impact.NewModel("large", "A large impact model.")
	levels := []impact.Level{impact.LevelOutput, impact.LevelShortTerm, impact.LevelMediumTerm, impact.LevelLongTerm, impact.LevelImpact}
	for i := 0; i < outcomes; i++ {
		d.Outcomes = append(d.Outcomes, impact.Outcome{ID: fmt.Sprintf("oc-%03d", i), Title: fmt.Sprintf("Outcome %d", i), Level: levels[i/perLayer]})
	}
	for i := 0; i < outcomes-perLayer; i++ {
		to := (i/perLayer+1)*perLayer + (i*7)%perLayer
		d.Links = append(d.Links,
			impact.CausalLink{ID: fmt.Sprintf("ln-a-%03d", i), From: fmt.Sprintf("oc-%03d", i), To: fmt.Sprintf("oc-%03d", to), Rationale: "Because."},
			impact.CausalLink{ID: fmt.Sprintf("ln-b-%03d", i), From: fmt.Sprintf("oc-%03d", i), To: fmt.Sprintf("oc-%03d", (i/perLayer+1)*perLayer+(i*13+5)%perLayer), Rationale: "Because."})
	}
	for i := 0; len(d.Links) < 400; i++ {
		d.Links = append(d.Links, impact.CausalLink{ID: fmt.Sprintf("ln-c-%03d", i), From: fmt.Sprintf("oc-%03d", i%perLayer), To: fmt.Sprintf("oc-%03d", perLayer+(i*3+1)%perLayer), Rationale: "Because."})
	}
	one, hundred := 1.0, 100.0
	var obs []impact.Observation
	for i := 0; i < indicators; i++ {
		id := fmt.Sprintf("ind-%03d", i)
		d.Indicators = append(d.Indicators, impact.Indicator{
			ID: id, Measures: fmt.Sprintf("oc-%03d", i%outcomes), Name: id, MeasurementType: impact.MeasurementQuantitative,
			Unit: "things", Direction: impact.DirectionIncrease, Frequency: "P1M", DataSource: "x",
			Baseline: &impact.Baseline{Value: &one, Date: "2026-01-01"},
			Targets:  []impact.Target{{Value: &hundred, Date: "2026-12-31"}},
		})
		for m := 0; m < 17 && len(obs) < 5000; m++ {
			v := one + float64(m*5)
			obs = append(obs, impact.Observation{IndicatorID: id, Date: time.Date(2026, time.Month(m%12+1), 28, 0, 0, 0, 0, time.UTC).Format(time.DateOnly), Value: &v})
		}
	}
	return d, []*impact.ObservationSet{{Kind: impact.KindObservationSet, ModelID: "large", Observations: obs}}
}
