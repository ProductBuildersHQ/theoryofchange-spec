package measure

import (
	"fmt"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/lint"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

// large builds a layered theory of 200 outcomes, 400 links, and 300
// indicators with 5,000 observations: the size the specification budgets a
// second for.
func large() (*toc.Document, []*toc.ObservationSet) {
	const (
		outcomes   = 200
		indicators = 300
		perLayer   = 40 // five layers of 40
	)
	d := toc.NewDocument("large", "A large theory.")
	levels := []toc.Level{toc.LevelOutput, toc.LevelShortTerm, toc.LevelIntermediate, toc.LevelLongTerm, toc.LevelImpact}
	for i := 0; i < outcomes; i++ {
		d.Outcomes = append(d.Outcomes, toc.Outcome{ID: fmt.Sprintf("oc-%03d", i), Title: fmt.Sprintf("Outcome %d", i), Level: levels[i/perLayer]})
	}
	for i := 0; i < outcomes-perLayer; i++ {
		to := (i/perLayer+1)*perLayer + (i*7)%perLayer
		d.Links = append(d.Links,
			toc.CausalLink{ID: fmt.Sprintf("ln-a-%03d", i), From: fmt.Sprintf("oc-%03d", i), To: fmt.Sprintf("oc-%03d", to), Rationale: "Because."},
			toc.CausalLink{ID: fmt.Sprintf("ln-b-%03d", i), From: fmt.Sprintf("oc-%03d", i), To: fmt.Sprintf("oc-%03d", (i/perLayer+1)*perLayer+(i*13+5)%perLayer), Rationale: "Because."})
	}
	for i := 0; len(d.Links) < 400; i++ {
		d.Links = append(d.Links, toc.CausalLink{ID: fmt.Sprintf("ln-c-%03d", i), From: fmt.Sprintf("oc-%03d", i%perLayer), To: fmt.Sprintf("oc-%03d", perLayer+(i*3+1)%perLayer), Rationale: "Because."})
	}
	one, hundred := 1.0, 100.0
	var obs []toc.Observation
	for i := 0; i < indicators; i++ {
		id := fmt.Sprintf("ind-%03d", i)
		d.Indicators = append(d.Indicators, toc.Indicator{
			ID: id, Measures: fmt.Sprintf("oc-%03d", i%outcomes), Name: id, MeasurementType: toc.MeasurementQuantitative,
			Unit: "things", Direction: toc.DirectionIncrease, Frequency: "P1M", DataSource: "x",
			Baseline: &toc.Baseline{Value: &one, Date: "2026-01-01"},
			Targets:  []toc.Target{{Value: &hundred, Date: "2026-12-31"}},
		})
		for m := 0; m < 17 && len(obs) < 5000; m++ {
			v := one + float64(m*5)
			obs = append(obs, toc.Observation{IndicatorID: id, Date: time.Date(2026, time.Month(m%12+1), 28, 0, 0, 0, 0, time.UTC).Format(time.DateOnly), Value: &v})
		}
	}
	return d, []*toc.ObservationSet{{Kind: toc.KindObservationSet, TocID: "large", Observations: obs}}
}

func pipeline(d *toc.Document, sets []*toc.ObservationSet) error {
	if issues := validate.Document(d); validate.HasErrors(issues) {
		return fmt.Errorf("invalid: %v", issues[0])
	}
	g, err := graph.Build(d)
	if err != nil {
		return err
	}
	if _, err := lint.Run(d, g, lint.Options{Profile: lint.ProfileFull}); err != nil {
		return err
	}
	_, err = Evaluate(d, g, sets, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), toc.DefaultStatusOptions())
	return err
}

func BenchmarkValidateLintAndStatus(b *testing.B) {
	d, sets := large()
	b.ResetTimer()
	for b.Loop() {
		if err := pipeline(d, sets); err != nil {
			b.Fatal(err)
		}
	}
}

// The specification budgets one second for validate, lint, and status on a
// theory of this size. The pipeline takes a few dozen milliseconds on a
// laptop, so a busy machine passes with room to spare and a change that
// makes it an order of magnitude slower fails.
func TestLargeTheoryStaysWithinTheTimeBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("performance check")
	}
	d, sets := large()
	if len(d.Outcomes) != 200 || len(d.Links) != 400 || len(d.Indicators) != 300 || len(sets[0].Observations) != 5000 {
		t.Fatalf("fixture size: %d outcomes, %d links, %d indicators, %d observations", len(d.Outcomes), len(d.Links), len(d.Indicators), len(sets[0].Observations))
	}
	start := time.Now()
	if err := pipeline(d, sets); err != nil {
		t.Fatal(err)
	}
	if took := time.Since(start); took > time.Second {
		t.Errorf("validate, lint, and status took %v; the budget is 1s", took)
	} else {
		t.Logf("validate, lint, and status took %v", took)
	}
}
