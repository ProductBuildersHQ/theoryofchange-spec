package prismx

import (
	"errors"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/grokify/prism-roadmap/goals/dmaic"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
	"github.com/ProductBuildersHQ/impact-model/pkg/measure"
)

// dmaicExample is the intermediary example with Shared Services marked for
// DMAIC and two years of weekly observations of its North Star, evaluated as
// of the last one.
func dmaicExample(t *testing.T) (*impact.Model, *impact.Status, []*impact.ObservationSet) {
	t.Helper()
	doc, err := impactio.LoadDocument("../../examples/intermediary/impact-model.yaml")
	if err != nil {
		t.Fatal(err)
	}
	p, ok := doc.Pathway("pw-shared")
	if !ok {
		t.Fatal("the example has no pw-shared pathway")
	}
	p.Projections = []impact.Projection{impact.ProjectionDMAIC}

	sets, err := impactio.LoadObservations("../../examples/intermediary/observations-2026-q3.yaml")
	if err != nil {
		t.Fatal(err)
	}
	weekly := &impact.ObservationSet{Kind: impact.KindObservationSet, SpecVersion: impact.SpecVersion, ModelID: doc.ID}
	start := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	for i := range 24 {
		v := 21.5 - 0.03*float64(i) + 0.2*float64(i%2)
		weekly.Observations = append(weekly.Observations, impact.Observation{
			IndicatorID: "ind-overhead-share", Date: start.AddDate(0, 0, 7*i).Format(time.DateOnly), Value: &v,
		})
	}
	sets = append(sets, weekly)

	g, err := graph.Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	asOf, err := time.Parse(time.DateOnly, "2026-09-30")
	if err != nil {
		t.Fatal(err)
	}
	st, err := measure.Evaluate(doc, g, sets, asOf, impact.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	return doc, st, sets
}

func metricByID(t *testing.T, out *dmaic.DMAICDocument, id string) dmaic.Metric {
	t.Helper()
	for _, c := range out.Categories {
		for _, m := range c.Metrics {
			if m.ID == id {
				return m
			}
		}
	}
	t.Fatalf("no metric %s", id)
	return dmaic.Metric{}
}

func TestDMAICExportValidates(t *testing.T) {
	doc, st, sets := dmaicExample(t)
	for name, in := range map[string]struct {
		st   *impact.Status
		sets []*impact.ObservationSet
	}{"with observations": {st, sets}, "without": {nil, nil}} {
		out, _, err := ToDMAIC(doc, in.st, in.sets)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if errs := dmaic.Errors(out.Validate(dmaic.DefaultValidationOptions())); len(errs) != 0 {
			t.Errorf("%s: the exported DMAIC document does not validate: %+v", name, errs)
		}
	}
}

func TestDMAICMapping(t *testing.T) {
	doc, st, sets := dmaicExample(t)
	out, omissions, err := ToDMAIC(doc, st, sets)
	if err != nil {
		t.Fatal(err)
	}

	// Only the marked pathway is a category.
	if len(out.Categories) != 1 || out.Categories[0].ID != "pw-shared" {
		t.Fatalf("categories = %+v, want only pw-shared", out.Categories)
	}
	if out.Metadata.Period != st.AsOf {
		t.Errorf("period = %q, want the status date %q", out.Metadata.Period, st.AsOf)
	}

	// Every indicator of the pathway is a metric or an omission with a reason.
	cat := out.Categories[0]
	placed := map[string]bool{}
	for _, m := range cat.Metrics {
		placed[m.ID] = true
	}
	for _, o := range omissions {
		if placed[o.IndicatorID] || o.Reason == "" {
			t.Errorf("omission %+v is also a metric, or has no reason", o)
		}
		placed[o.IndicatorID] = true
	}
	for _, ind := range doc.Indicators {
		if !ind.Rejected() && doc.IndicatorServes(ind, "pw-shared") && !placed[ind.ID] {
			t.Errorf("indicator %s of pw-shared is neither a metric nor an omission", ind.ID)
		}
	}

	ns := metricByID(t, out, "ind-overhead-share")
	ind, _ := doc.Indicator("ind-overhead-share")
	if ns.Baseline != *ind.Baseline.Value || ns.Unit != ind.Unit || ns.TrendDirection != dmaic.TrendLowerBetter {
		t.Errorf("baseline, unit, or trend not carried over: %+v", ns)
	}
	if len(ns.DataPoints) < impact.DMAICMinPoints || ns.ControlLimits == nil {
		t.Fatalf("with %d observations the metric has control limits: %+v", len(ns.DataPoints), ns.ControlLimits)
	}
	if !slices.IsSortedFunc(ns.DataPoints, func(a, b dmaic.DataPoint) int { return a.Timestamp.Compare(b.Timestamp) }) {
		t.Error("data points are not in date order")
	}
	if last := ns.DataPoints[len(ns.DataPoints)-1]; ns.Current != last.Value {
		t.Errorf("current = %v, want the latest observation %v", ns.Current, last.Value)
	}
	if ns.ProcessCapability != nil {
		t.Error("process capability needs specification limits the model does not have")
	}
	if ns.Phase == dmaic.PhaseMeasure {
		t.Errorf("with %d points the metric is past Measure", len(ns.DataPoints))
	}

	// Initiatives are the pathway's interventions, each naming its metrics.
	for _, in := range out.Initiatives {
		p, _ := doc.Pathway("pw-shared")
		if !slices.Contains(p.InterventionIDs, in.ID) {
			t.Errorf("initiative %s is not an intervention of pw-shared", in.ID)
		}
		for _, id := range in.MetricIDs {
			if !slices.Contains(metricByID(t, out, id).InitiativeIDs, in.ID) {
				t.Errorf("initiative %s lists metric %s, which does not list it back", in.ID, id)
			}
		}
	}
	if len(out.Initiatives) == 0 {
		t.Error("no initiatives")
	}

	// Without observations there are no data points or limits, and every
	// metric is still being measured.
	bare, _, err := ToDMAIC(doc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range bare.Categories {
		for _, m := range c.Metrics {
			if len(m.DataPoints) != 0 || m.ControlLimits != nil || m.Phase != dmaic.PhaseMeasure || m.Current != m.Baseline {
				t.Errorf("metric %s without observations: %+v", m.ID, m)
			}
		}
	}
}

func TestDMAICIndividualsChart(t *testing.T) {
	var points []dmaic.DataPoint
	for i := range 20 {
		points = append(points, dmaic.DataPoint{Value: 10 + 2*float64(i%2)})
	}
	// Mean 11; every moving range is 2, so sigma is 2 / 1.128.
	got := individualsChart(points)
	sigma := 2 / 1.128
	want := dmaic.ControlLimits{CenterLine: 11, UCL: 11 + 3*sigma, LCL: 11 - 3*sigma, Sigma: sigma}
	for name, pair := range map[string][2]float64{
		"center": {got.CenterLine, want.CenterLine}, "ucl": {got.UCL, want.UCL},
		"lcl": {got.LCL, want.LCL}, "sigma": {got.Sigma, want.Sigma},
	} {
		if math.Abs(pair[0]-pair[1]) > 0.001 {
			t.Errorf("%s = %v, want %v", name, pair[0], pair[1])
		}
	}
}

func TestDMAICPhase(t *testing.T) {
	enough := impact.DMAICMinPoints
	tests := []struct {
		state     impact.State
		points    int
		improving bool
		want      string
	}{
		{impact.StateAchieved, 3, false, dmaic.PhaseControl},
		{impact.StateOnTrack, enough - 1, true, dmaic.PhaseMeasure},
		{impact.StateOnTrack, enough, true, dmaic.PhaseImprove},
		{impact.StateOffTrack, enough, false, dmaic.PhaseAnalyze},
	}
	for _, tt := range tests {
		if got := metricPhase(tt.state, tt.points, tt.improving); got != tt.want {
			t.Errorf("metricPhase(%s, %d, %v) = %s, want %s", tt.state, tt.points, tt.improving, got, tt.want)
		}
	}
}

func TestDMAICRefusesWhatItCannotDraw(t *testing.T) {
	doc, _, _ := dmaicExample(t)
	unmarked := *doc
	unmarked.Pathways = slices.Clone(doc.Pathways)
	for i := range unmarked.Pathways {
		unmarked.Pathways[i].Projections = nil
	}
	_, _, err := ToDMAIC(&unmarked, nil, nil)
	if !errors.Is(err, ErrNotDMAICReady) || !strings.Contains(err.Error(), "projections: [dmaic]") {
		t.Errorf("unmarked: error = %v", err)
	}

	noNumbers := *doc
	noNumbers.Indicators = slices.Clone(doc.Indicators)
	for i := range noNumbers.Indicators {
		noNumbers.Indicators[i].Baseline = nil
	}
	_, omissions, err := ToDMAIC(&noNumbers, nil, nil)
	if !errors.Is(err, ErrNotDMAICReady) || !strings.Contains(err.Error(), "numeric baseline") {
		t.Errorf("no numbers: error = %v", err)
	}
	if len(omissions) == 0 {
		t.Error("a refusal for missing numbers still lists the omitted indicators")
	}
	if _, _, err := ToDMAIC(nil, nil, nil); err == nil {
		t.Error("a nil document was accepted")
	}
}
