package measure

import (
	"bytes"
	"flag"
	"os"
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

var update = flag.Bool("update", false, "rewrite golden files")

func loadExample(t *testing.T) (*toc.Document, *graph.Graph, []*toc.ObservationSet) {
	t.Helper()
	doc, err := tocio.LoadDocument("../../examples/intermediary/theory.yaml")
	if err != nil {
		t.Fatal(err)
	}
	g, err := graph.Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	sets, err := tocio.LoadObservations("../../examples/intermediary/observations-2026-q3.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return doc, g, sets
}

func canonical(t *testing.T, v any) []byte {
	t.Helper()
	b, err := tocio.Canonical(v, tocio.FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestSnapshotIsByteIdenticalAndMatchesTheGolden(t *testing.T) {
	doc, g, sets := loadExample(t)
	asOf := day("2026-09-30")
	var runs [2][]byte
	for i := range runs {
		st, err := Evaluate(doc, g, sets, asOf, toc.DefaultStatusOptions())
		if err != nil {
			t.Fatal(err)
		}
		runs[i] = canonical(t, st)
	}
	if !bytes.Equal(runs[0], runs[1]) {
		t.Fatal("two evaluations of the same inputs differ")
	}
	path := "testdata/status-intermediary.json"
	if *update {
		if err := os.WriteFile(path, runs[0], 0o644); err != nil { //nolint:gosec // golden files are public
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v (run with -update)", err)
	}
	if !bytes.Equal(runs[0], want) {
		t.Errorf("snapshot differs from %s (run with -update to accept)", path)
	}
}

func TestSnapshotIsAValidStatusDocument(t *testing.T) {
	doc, g, sets := loadExample(t)
	st, err := Evaluate(doc, g, sets, day("2026-09-30"), toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tocio.ParseStatus(canonical(t, st), tocio.FormatJSON); err != nil {
		t.Fatalf("the snapshot does not conform to its schema: %v", err)
	}
	if st.Kind != toc.KindStatus || st.TocID != doc.ID || st.AsOf != "2026-09-30" {
		t.Errorf("header = %+v", st)
	}
}

func TestEvaluateWithoutObservations(t *testing.T) {
	doc, g, _ := loadExample(t)
	st, err := Evaluate(doc, g, nil, day("2026-06-30"), toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Outcomes) == 0 || len(st.Overall) == 0 || len(st.Links) == 0 {
		t.Fatalf("an empty observation set must still give a full snapshot: %d outcomes, %d impacts", len(st.Outcomes), len(st.Overall))
	}
	for _, i := range st.Indicators {
		switch i.State {
		case toc.StateNoData, toc.StateNotStarted, toc.StateBaselineOnly:
		default:
			t.Errorf("indicator %s is %s with no observations", i.IndicatorID, i.State)
		}
	}
}

func themeDoc() (*toc.Document, []*toc.ObservationSet) {
	d := toc.NewDocument("t", "p")
	d.Outcomes = []toc.Outcome{{ID: "oc-a", Title: "a", Level: toc.LevelImpact}, {ID: "oc-b", Title: "b", Level: toc.LevelImpact}}
	d.Themes = []toc.Theme{{ID: "thm-1", Label: "l", Definition: "d"}, {ID: "thm-2", Label: "l", Definition: "d"}}
	d.Pathways = []toc.Pathway{{ID: "pw-a", Title: "A", OutcomeIDs: []string{"oc-a"}}}
	d.Indicators = []toc.Indicator{
		{ID: "ind-any", Measures: "oc-a", Name: "n", MeasurementType: toc.MeasurementThemeCount, ThemeIDs: []string{"thm-1", "thm-2"}},
		{ID: "ind-pw", Measures: "oc-a", Name: "n", MeasurementType: toc.MeasurementThemeCount, ThemeIDs: []string{"thm-1"}, PathwayIDs: []string{"pw-a"}},
	}
	story := func(id, date string, themes, outcomes []string) toc.Story {
		return toc.Story{ID: id, Date: date, Title: id, Summary: "s", ThemeIDs: themes, OutcomeIDs: outcomes, Consent: toc.ConsentGranted, Sensitivity: toc.SensitivityPublic}
	}
	sets := []*toc.ObservationSet{
		{Stories: []toc.Story{
			story("s1", "2026-03-01", []string{"thm-1"}, []string{"oc-a"}),
			story("s2", "2026-04-01", []string{"thm-2"}, []string{"oc-b"}),
			story("s3", "2026-05-01", []string{"thm-1", "thm-2"}, []string{"oc-a"}),
			story("s4", "2026-06-01", []string{"thm-3"}, nil),
			story("s5", "2027-01-01", []string{"thm-1"}, []string{"oc-a"}), // after asOf
		}},
		{Stories: []toc.Story{story("s1", "2026-03-01", []string{"thm-1"}, []string{"oc-a"})}}, // the same story again
	}
	return d, sets
}

func TestThemeCounts(t *testing.T) {
	d, sets := themeDoc()
	asOf := day("2026-12-31")
	if n := ThemeCount(d, d.Indicators[0], sets, asOf); n != 3 {
		t.Errorf("any theme: %d, want 3 (s1, s2, s3; s1 repeated and counted once, s4 has another theme, s5 is later)", n)
	}
	if n := ThemeCount(d, d.Indicators[1], sets, asOf); n != 2 {
		t.Errorf("filtered by pathway: %d, want 2 (s1 and s3 are theme 1 about the pathway's outcome)", n)
	}
	if n := ThemeCount(d, d.Indicators[0], sets, day("2026-03-15")); n != 1 {
		t.Errorf("as of March: %d, want 1", n)
	}
}

func TestThemeCountsBecomeIndicatorValues(t *testing.T) {
	d, sets := themeDoc()
	d.Indicators[0].Baseline = &toc.Baseline{Value: f(0), Date: "2026-01-01"}
	d.Indicators[0].Targets = []toc.Target{{Value: f(3), Date: "2026-12-31"}}
	g, err := graph.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	st, err := Evaluate(d, g, sets, day("2026-12-31"), toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	got := st.Indicators[0]
	if got.State != toc.StateAchieved || got.Latest == nil || *got.Latest.Value != 3 || got.Latest.Source != "story theme count" {
		t.Errorf("theme count status = %+v", got)
	}
}

func TestRollupsBecomeIndicatorValues(t *testing.T) {
	d := rollupDoc(toc.AggregationSum)
	d.Outcomes = []toc.Outcome{{ID: "oc", Title: "t", Level: toc.LevelImpact}}
	d.Indicators[0].Baseline = &toc.Baseline{Value: f(0), Date: "2026-01-01"}
	d.Indicators[0].Targets = []toc.Target{{Value: f(60), Date: "2026-12-31"}}
	g, err := graph.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	sets := []*toc.ObservationSet{{Observations: values("a", "2026-12-01", 10.0, "b", "2026-12-01", 20.0, "c", "2026-12-01", 30.0)}}
	st, err := Evaluate(d, g, sets, day("2026-12-31"), toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	total := st.Indicators[0]
	if total.IndicatorID != "total" || total.Latest == nil || *total.Latest.Value != 60 || total.State != toc.StateAchieved {
		t.Errorf("rolled-up indicator = %+v", total)
	}
}

func TestDirectObservationsBeatRollups(t *testing.T) {
	d := rollupDoc(toc.AggregationSum)
	d.Outcomes = []toc.Outcome{{ID: "oc", Title: "t", Level: toc.LevelImpact}}
	g, err := graph.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	sets := []*toc.ObservationSet{{Observations: values("a", "2026-12-01", 1.0, "b", "2026-12-01", 2.0, "c", "2026-12-01", 3.0, "total", "2026-12-15", 100.0)}}
	st, _ := Evaluate(d, g, sets, day("2026-12-31"), toc.DefaultStatusOptions())
	if *st.Indicators[0].Latest.Value != 100 {
		t.Errorf("direct value lost: %v", *st.Indicators[0].Latest.Value)
	}
}

func TestEvaluateRespectsTheAsOfDate(t *testing.T) {
	doc, g, sets := loadExample(t)
	early, err := Evaluate(doc, g, sets, day("2026-02-01"), toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	late, _ := Evaluate(doc, g, sets, day("2026-09-30"), toc.DefaultStatusOptions())
	countObserved := func(st *toc.Status) int {
		n := 0
		for _, i := range st.Indicators {
			n += i.ObservationCount
		}
		return n
	}
	if countObserved(early) >= countObserved(late) {
		t.Errorf("observations after the date were used: %d early, %d late", countObserved(early), countObserved(late))
	}
}

func snapshot(asOf string, indicator, outcomePathway toc.State, conf float64, link toc.LinkState, assumption toc.AssumptionStatus) *toc.Status {
	return &toc.Status{
		TocID: "t", AsOf: asOf,
		Indicators:  []toc.IndicatorStatus{{IndicatorID: "ind", State: indicator}},
		Outcomes:    []toc.OutcomeStatus{{OutcomeID: "oc", State: indicator, PathwayState: outcomePathway, Confidence: &conf}},
		Links:       []toc.LinkStatus{{LinkID: "ln", State: link}},
		Assumptions: []toc.AssumptionState{{AssumptionID: "asm", Status: assumption}},
	}
}

func TestDiff(t *testing.T) {
	prev := snapshot("2026-06-30", toc.StateOnTrack, toc.StateOnTrack, 0.50, toc.LinkStateOK, toc.AssumptionHolding)
	cur := snapshot("2026-09-30", toc.StateAtRisk, toc.StateAtRisk, 0.50, toc.LinkStateAtRisk, toc.AssumptionAtRisk)
	d := Diff(prev, cur)
	if d.TocID != "t" || d.FromAsOf != "2026-06-30" || d.ToAsOf != "2026-09-30" {
		t.Errorf("header = %+v", d)
	}
	var got []string
	for _, c := range d.Changes {
		got = append(got, string(c.ElementKind)+":"+c.ElementID+":"+c.Field+":"+c.From+">"+c.To)
	}
	want := []string{
		"indicator:ind:state:on_track>at_risk", "outcome:oc:state:on_track>at_risk", "outcome:oc:pathwayState:on_track>at_risk",
		"link:ln:state:ok>at_risk", "assumption:asm:status:holding>at_risk",
	}
	if !slices.Equal(got, want) {
		t.Errorf("changes = %v\nwant     %v", got, want)
	}
	if same := Diff(prev, prev); len(same.Changes) != 0 {
		t.Errorf("a snapshot differs from itself: %v", same.Changes)
	}
}

func TestDiffReportsConfidenceAtTwoPlaces(t *testing.T) {
	prev := snapshot("a", toc.StateOnTrack, toc.StateOnTrack, 0.501, toc.LinkStateOK, toc.AssumptionHolding)
	cur := snapshot("b", toc.StateOnTrack, toc.StateOnTrack, 0.504, toc.LinkStateOK, toc.AssumptionHolding)
	if d := Diff(prev, cur); len(d.Changes) != 0 {
		t.Errorf("noise below two places is a change: %v", d.Changes)
	}
	cur.Outcomes[0].Confidence = f(0.62)
	d := Diff(prev, cur)
	if len(d.Changes) != 1 || d.Changes[0].Field != "confidence" || d.Changes[0].From != "0.50" || d.Changes[0].To != "0.62" {
		t.Errorf("changes = %v", d.Changes)
	}
}

func TestDiffReportsAddedAndRemovedElements(t *testing.T) {
	prev := snapshot("a", toc.StateOnTrack, toc.StateOnTrack, 0.5, toc.LinkStateOK, toc.AssumptionHolding)
	cur := snapshot("b", toc.StateOnTrack, toc.StateOnTrack, 0.5, toc.LinkStateOK, toc.AssumptionHolding)
	cur.Indicators = append(cur.Indicators, toc.IndicatorStatus{IndicatorID: "new", State: toc.StateNoData})
	prev.Links = append(prev.Links, toc.LinkStatus{LinkID: "gone", State: toc.LinkStateOK})
	d := Diff(prev, cur)
	if len(d.Changes) != 2 || d.Changes[0].Field != "presence" || d.Changes[0].To != "added" || d.Changes[1].To != "removed" {
		t.Errorf("changes = %+v", d.Changes)
	}
}
