package report

import (
	"testing"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/internal/testfixture"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/measure"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// execPipeline builds the graph, evaluates the status, builds the view, and
// renders the executive report.
func execPipeline(d *toc.Document, sets []*toc.ObservationSet) (string, error) {
	g, err := graph.Build(d)
	if err != nil {
		return "", err
	}
	st, err := measure.Evaluate(d, g, sets, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), toc.DefaultStatusOptions())
	if err != nil {
		return "", err
	}
	v, err := BuildView(d, g, st, Options{})
	if err != nil {
		return "", err
	}
	return Exec(StatusInput{View: v, Doc: d, Status: st}), nil
}

func BenchmarkExecReport(b *testing.B) {
	d, sets := testfixture.Large()
	b.ResetTimer()
	for b.Loop() {
		if _, err := execPipeline(d, sets); err != nil {
			b.Fatal(err)
		}
	}
}

// The executive report for a theory of 200 outcomes, 400 links, and 300
// indicators must take under a second, from graph to rendered text.
func TestExecReportForALargeTheoryTakesUnderASecond(t *testing.T) {
	if testing.Short() {
		t.Skip("performance check")
	}
	d, sets := testfixture.Large()
	start := time.Now()
	out, err := execPipeline(d, sets)
	if err != nil {
		t.Fatal(err)
	}
	took := time.Since(start)
	if took > time.Second {
		t.Errorf("the executive report took %v; the budget is 1s", took)
	}
	t.Logf("the executive report took %v (%d bytes)", took, len(out))
}
