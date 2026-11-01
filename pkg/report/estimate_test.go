package report

import (
	"strings"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/measure"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// latestOf builds the view of one indicator observed once, with the given
// quality and note, and returns it.
func latestOf(t *testing.T, quality toc.DataQuality, note string) IndicatorView {
	t.Helper()
	doc := fullDoc(t)
	g, err := graph.Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	id := doc.Indicators[0].ID
	v := 12.0
	sets := []*toc.ObservationSet{{
		Kind: toc.KindObservationSet, SpecVersion: toc.SpecVersion, TocID: doc.ID,
		Observations: []toc.Observation{{IndicatorID: id, Date: "2026-06-30", Value: &v, Quality: quality, Note: note}},
	}}
	asOf, err := time.Parse(time.DateOnly, "2026-09-30")
	if err != nil {
		t.Fatal(err)
	}
	st, err := measure.Evaluate(doc, g, sets, asOf, toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	view := build(t, doc, st, Options{})
	for _, iv := range view.Indicators {
		if iv.ID == id {
			return iv
		}
	}
	t.Fatalf("no view of %s", id)
	return IndicatorView{}
}

// An estimate reads as an estimate, and a verified or unlabeled figure does
// not, so a rounded or partial count is not mistaken for an exact one.
func TestAnEstimatedLatestValueIsMarkedAsAnEstimate(t *testing.T) {
	est := latestOf(t, toc.QualityEstimated, `The source says "almost 12".`)
	if !strings.Contains(est.Latest, "estimate, on 2026-06-30") {
		t.Errorf("Latest = %q, want it marked as an estimate", est.Latest)
	}
	if est.LatestNote != `The source says "almost 12".` {
		t.Errorf("LatestNote = %q", est.LatestNote)
	}
	for _, q := range []toc.DataQuality{toc.QualityVerified, toc.QualityReported, ""} {
		got := latestOf(t, q, "")
		if strings.Contains(got.Latest, "estimate") || !strings.Contains(got.Latest, "(on 2026-06-30)") {
			t.Errorf("quality %q: Latest = %q, want a plain dated value", q, got.Latest)
		}
	}
}
