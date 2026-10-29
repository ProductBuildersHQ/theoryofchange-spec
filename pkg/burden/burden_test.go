package burden

import (
	"math"
	"reflect"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestCollectionsPerYear(t *testing.T) {
	tests := []struct {
		period string
		want   float64
	}{
		{"P1Y", 1}, {"P6M", 2}, {"P3M", 4}, {"P1M", 12}, {"P1W", 52}, {"P2W", 26},
		{"P7D", 52}, {"P1D", 364}, {"PT24H", 364}, {"P1Y6M", 2.0 / 3},
	}
	for _, tt := range tests {
		got, err := CollectionsPerYear(tt.period)
		if err != nil || !near(got, tt.want) {
			t.Errorf("CollectionsPerYear(%q) = %v, %v; want %v", tt.period, got, err, tt.want)
		}
	}
	for _, bad := range []string{"", "monthly", "P", "PT", "P0D", "1M"} {
		if _, err := CollectionsPerYear(bad); err == nil {
			t.Errorf("CollectionsPerYear(%q) accepted", bad)
		}
	}
}

func ind(id string, status toc.SelectionStatus, c *toc.Collection, pathways ...string) toc.Indicator {
	return toc.Indicator{
		ID: id, Measures: "oc", Name: id, PathwayIDs: pathways, Collection: c,
		Selection: &toc.Selection{Status: status},
	}
}

func doc() *toc.Document {
	return &toc.Document{
		Pathways: []toc.Pathway{{ID: "pw-a"}, {ID: "pw-b"}},
		Indicators: []toc.Indicator{
			// 30 min monthly = 360/yr = 6.923 min/week; 5 questions x 12 x half = 30
			ind("i-survey", toc.SelectionSelected, &toc.Collection{Mode: toc.ModeSurvey, MinutesPerPeriod: 30, Period: "P1M", QuestionsPerRespondent: 5, SampleRate: 0.5}, "pw-a"),
			// weekly 10 min = 10 min/week, serves two pathways
			ind("i-manual", toc.SelectionSelected, &toc.Collection{Mode: toc.ModeManual, MinutesPerPeriod: 10, Period: "P1W"}, "pw-a", "pw-b"),
			// yearly 120 min = 120/52 min/week; 6 questions per year; no pathway
			ind("i-form", toc.SelectionSelected, &toc.Collection{Mode: toc.ModeForm, MinutesPerPeriod: 120, Period: "P1Y", QuestionsPerRespondent: 6}),
			ind("i-uncosted", toc.SelectionSelected, nil, "pw-b"),
			ind("i-candidate", toc.SelectionCandidate, &toc.Collection{Mode: toc.ModeManual, MinutesPerPeriod: 999, Period: "P1W"}),
			ind("i-retired", toc.SelectionRetired, &toc.Collection{Mode: toc.ModeManual, MinutesPerPeriod: 999, Period: "P1W"}),
			{ID: "i-unselected", Measures: "oc", Name: "x", Collection: &toc.Collection{Mode: toc.ModeManual, MinutesPerPeriod: 999, Period: "P1W"}},
		},
	}
}

func TestCompute(t *testing.T) {
	r, err := Compute(doc())
	if err != nil {
		t.Fatal(err)
	}
	wantMin := 30*12.0/52 + 10 + 120.0/52
	if r.Overall.Indicators != 4 || !near(r.Overall.MinutesPerWeek, wantMin) || !near(r.Overall.QuestionsPerRespondentPerYear, 36) {
		t.Errorf("overall = %+v, want 4 indicators, %.4f min/week, 36 questions", r.Overall, wantMin)
	}
	if !reflect.DeepEqual(r.Uncosted, []string{"i-uncosted"}) {
		t.Errorf("uncosted = %v", r.Uncosted)
	}

	keys := func(gs []Group) []string {
		var out []string
		for _, g := range gs {
			out = append(out, g.Key)
		}
		return out
	}
	if got := keys(r.ByPathway); !reflect.DeepEqual(got, []string{"pw-a", "pw-b", ""}) {
		t.Errorf("pathway order = %v", got)
	}
	a, b, none := r.ByPathway[0], r.ByPathway[1], r.ByPathway[2]
	if a.Indicators != 2 || !near(a.MinutesPerWeek, 30*12.0/52+10) || !near(a.QuestionsPerRespondentPerYear, 30) {
		t.Errorf("pw-a = %+v", a)
	}
	if b.Indicators != 2 || !near(b.MinutesPerWeek, 10) || none.Indicators != 1 || !near(none.QuestionsPerRespondentPerYear, 6) {
		t.Errorf("pw-b = %+v, none = %+v", b, none)
	}
	if got := keys(r.ByMode); !reflect.DeepEqual(got, []string{"form", "survey", "manual", ""}) {
		t.Errorf("mode order = %v", got)
	}
	if r.Budget != nil {
		t.Error("a theory with no budget got a budget check")
	}
}

func TestBudgetComparison(t *testing.T) {
	d := doc()
	d.Budget = &toc.MeasurementBudget{
		StaffMinutesPerWeek: 15, MaxQuestionsPerRespondentPerYear: 40,
		TargetSelectedMetrics: &toc.Range{Min: 5, Max: 8},
	}
	r, _ := Compute(d)
	b := r.Budget
	if !b.StaffMinutesPerWeek.Over || !b.StaffMinutesPerWeek.Declared || b.StaffMinutesPerWeek.Limit != 15 {
		t.Errorf("staff = %+v", b.StaffMinutesPerWeek)
	}
	if b.QuestionsPerYear.Over || !b.QuestionsPerYear.Declared {
		t.Errorf("questions = %+v", b.QuestionsPerYear)
	}
	if sm := b.SelectedMetrics; sm.Within || sm.Actual != 4 || sm.Min != 5 {
		t.Errorf("selected = %+v", sm)
	}

	d.Budget = &toc.MeasurementBudget{StaffMinutesPerWeek: 60}
	r, _ = Compute(d)
	if r.Budget.StaffMinutesPerWeek.Over || r.Budget.QuestionsPerYear.Declared || !r.Budget.SelectedMetrics.Within {
		t.Errorf("budget = %+v", r.Budget)
	}
}

func TestBadPeriodIsReported(t *testing.T) {
	d := doc()
	d.Indicators[0].Collection.Period = "often"
	if _, err := Compute(d); err == nil {
		t.Fatal("a bad period was accepted")
	}
}

func TestRejectedIndicatorsAreNotCounted(t *testing.T) {
	d := doc()
	d.Indicators[1].Provenance = &toc.Provenance{Basis: toc.BasisRejected}
	r, _ := Compute(d)
	if r.Overall.Indicators != 3 {
		t.Errorf("overall = %+v", r.Overall)
	}
}
