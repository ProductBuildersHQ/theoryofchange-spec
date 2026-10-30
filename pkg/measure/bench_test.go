package measure

import (
	"fmt"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/internal/testfixture"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/lint"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

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
	d, sets := testfixture.Large()
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
	d, sets := testfixture.Large()
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
