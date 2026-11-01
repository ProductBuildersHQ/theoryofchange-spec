package measure

import (
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

const (
	ach  = impact.StateAchieved
	on   = impact.StateOnTrack
	risk = impact.StateAtRisk
	off  = impact.StateOffTrack
	nd   = impact.StateNoData
	unm  = impact.StateUnmeasured
)

func TestWorse(t *testing.T) {
	cases := []struct{ a, b, want impact.State }{
		{ach, on, on}, {on, risk, risk}, {risk, off, off}, {impact.StateStale, nd, impact.StateStale},
		{nd, impact.StateNotStarted, nd}, {impact.StateNotStarted, impact.StateBaselineOnly, impact.StateNotStarted},
		{unm, on, on}, {off, unm, off}, {unm, unm, unm},
	}
	for _, c := range cases {
		if got := Worse(c.a, c.b); got != c.want {
			t.Errorf("Worse(%s, %s) = %s, want %s", c.a, c.b, got, c.want)
		}
	}
}

func st(id string, s impact.State, p, e float64) impact.IndicatorStatus {
	return impact.IndicatorStatus{IndicatorID: id, State: s, Progress: &p, Expected: &e}
}

func TestOutcomeStates(t *testing.T) {
	doc := &impact.Model{
		Outcomes: []impact.Outcome{{ID: "oc-a"}, {ID: "oc-b"}, {ID: "oc-c"}, {ID: "oc-d"}},
		Indicators: []impact.Indicator{
			{ID: "i1", Measures: "oc-a"}, {ID: "i2", Measures: "oc-a"},
			{ID: "i3", Measures: "oc-b", ProxyFor: []string{"oc-c"}},
			{ID: "cand", Measures: "oc-d", Selection: &impact.Selection{Status: impact.SelectionCandidate}},
			{ID: "gone", Measures: "oc-d", Provenance: &impact.Provenance{Basis: impact.BasisRejected}},
		},
	}
	statuses := map[string]impact.IndicatorStatus{
		"i1": st("i1", ach, 1.2, 1), "i2": st("i2", off, 0.1, 0.6), "i3": st("i3", on, 0.6, 0.6),
		"cand": st("cand", off, 0, 1), "gone": st("gone", off, 0, 1),
	}
	worst := OutcomeStates(doc, statuses, impact.RollupWorst, 0)
	if worst["oc-a"] != off || worst["oc-b"] != on || worst["oc-c"] != on {
		t.Errorf("worst: %v", worst)
	}
	if worst["oc-d"] != unm {
		t.Errorf("an outcome with only candidate and rejected indicators is %s, want unmeasured", worst["oc-d"])
	}
	// Mean progress (1.2+0.1)/2 = 0.65 against mean expected 0.8: within 2 tolerances.
	if mean := OutcomeStates(doc, statuses, impact.RollupMean, 0); mean["oc-a"] != risk {
		t.Errorf("mean: oc-a = %s", mean["oc-a"])
	}
	// With no progress to average, the mean method falls back to the worst state.
	bare := map[string]impact.IndicatorStatus{"i1": {IndicatorID: "i1", State: nd}, "i2": {IndicatorID: "i2", State: on}}
	if got := OutcomeStates(doc, bare, impact.RollupMean, 0)["oc-a"]; got != nd {
		t.Errorf("mean fallback = %s", got)
	}
}

// chain builds oc-top (impact) fed by the given sources, each a short-term
// outcome, with the given logic.
func chain(logic impact.Logic, sources ...string) *impact.Model {
	d := impact.NewModel("t", "p")
	d.Outcomes = append(d.Outcomes, impact.Outcome{ID: "oc-top", Title: "Top", Level: impact.LevelImpact, Logic: &logic})
	for _, s := range sources {
		d.Outcomes = append(d.Outcomes, impact.Outcome{ID: s, Title: s, Level: impact.LevelShortTerm})
		d.Links = append(d.Links, impact.CausalLink{ID: "ln-" + s, From: s, To: "oc-top"})
	}
	return d
}

func pathway(t *testing.T, d *impact.Model, own map[string]impact.State) map[string]impact.State {
	t.Helper()
	g, err := graph.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	got, err := PathwayStates(d, g, own)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestPathwayLogic(t *testing.T) {
	own := map[string]impact.State{"a": ach, "b": risk, "c": on}
	tests := []struct {
		name  string
		logic impact.Logic
		top   impact.State // the impact outcome's own state
		want  impact.State
	}{
		{"all takes the worst", impact.Logic{Type: impact.LogicAll}, ach, risk},
		{"any takes the best", impact.Logic{Type: impact.LogicAny}, ach, ach},
		{"at least 2 takes the second best", impact.Logic{Type: impact.LogicAtLeast, K: 2}, ach, on},
		{"at least 3 takes the third best", impact.Logic{Type: impact.LogicAtLeast, K: 3}, ach, risk},
		// The missing fourth counts as no data, which ranks above at risk, so the
		// fourth best of ach, on, no data, at risk is at risk.
		{"at least 4 of 3 pads with no data", impact.Logic{Type: impact.LogicAtLeast, K: 4}, ach, risk},
		{"own state worse than the signal wins", impact.Logic{Type: impact.LogicAny}, off, off},
		{"unmeasured own state defers to the signal", impact.Logic{Type: impact.LogicAll}, unm, risk},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := chain(tt.logic, "a", "b", "c")
			states := map[string]impact.State{"a": own["a"], "b": own["b"], "c": own["c"], "oc-top": tt.top}
			if got := pathway(t, d, states)["oc-top"]; got != tt.want {
				t.Errorf("oc-top = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestPathwayPropagatesUpward(t *testing.T) {
	d := impact.NewModel("t", "p")
	d.Outcomes = []impact.Outcome{
		{ID: "oc-top", Title: "T", Level: impact.LevelImpact}, {ID: "oc-mid", Title: "M", Level: impact.LevelMediumTerm}, {ID: "oc-low", Title: "L", Level: impact.LevelShortTerm},
	}
	d.Links = []impact.CausalLink{{ID: "l1", From: "oc-low", To: "oc-mid"}, {ID: "l2", From: "oc-mid", To: "oc-top"}}
	got := pathway(t, d, map[string]impact.State{"oc-low": off, "oc-mid": ach, "oc-top": ach})
	if got["oc-mid"] != off || got["oc-top"] != off {
		t.Errorf("a failing precondition must reach the impact: %v", got)
	}
	// Unmeasured outcomes on the way pass the signal through.
	got = pathway(t, d, map[string]impact.State{"oc-low": risk})
	if got["oc-mid"] != risk || got["oc-top"] != risk {
		t.Errorf("unmeasured middle did not defer: %v", got)
	}
	// Nothing measured anywhere stays unmeasured.
	if got = pathway(t, d, nil); got["oc-top"] != unm {
		t.Errorf("all-unmeasured = %v", got)
	}
}

func TestPathwayGroups(t *testing.T) {
	// oc-top needs: x, and "any of (y, z)". Group any-yz is a member of group all-x.
	d := chain(impact.Logic{Type: impact.LogicAll}, "x", "y", "z")
	d.Groups = []impact.PreconditionGroup{
		{ID: "grp-yz", Target: "oc-top", Members: []string{"y", "z"}, Logic: impact.Logic{Type: impact.LogicAny}},
	}
	got := pathway(t, d, map[string]impact.State{"x": on, "y": off, "z": ach, "oc-top": ach})
	// any(y=off, z=ach) = ach; all(x=on, ach) = on.
	if got["oc-top"] != on {
		t.Errorf("group then all: %s, want on_track", got["oc-top"])
	}

	// Nested: grp-outer (all of x and grp-inner), grp-inner (at least 1 of y, z).
	d = chain(impact.Logic{Type: impact.LogicAll}, "x", "y", "z")
	d.Groups = []impact.PreconditionGroup{
		{ID: "grp-outer", Target: "oc-top", Members: []string{"x", "grp-inner"}, Logic: impact.Logic{Type: impact.LogicAll}},
		{ID: "grp-inner", Target: "oc-top", Members: []string{"y", "z"}, Logic: impact.Logic{Type: impact.LogicAtLeast, K: 1}},
	}
	got = pathway(t, d, map[string]impact.State{"x": ach, "y": risk, "z": off, "oc-top": ach})
	// inner = best of (risk, off) = risk; outer = worst of (ach, risk) = risk; the group is the only item.
	if got["oc-top"] != risk {
		t.Errorf("nested groups: %s, want at_risk", got["oc-top"])
	}
}

func TestInhibitorsLimitAPathway(t *testing.T) {
	d := chain(impact.Logic{Type: impact.LogicAll}, "a")
	d.Outcomes = append(d.Outcomes, impact.Outcome{ID: "rival", Title: "Rival", Level: impact.LevelShortTerm})
	d.Links = append(d.Links, impact.CausalLink{ID: "ln-rival", From: "rival", To: "oc-top", Polarity: impact.PolarityNegative})

	strong := pathway(t, d, map[string]impact.State{"a": ach, "rival": ach, "oc-top": ach})
	if strong["oc-top"] != risk {
		t.Errorf("a thriving inhibitor: %s, want at_risk", strong["oc-top"])
	}
	weak := pathway(t, d, map[string]impact.State{"a": ach, "rival": off, "oc-top": ach})
	if weak["oc-top"] != ach {
		t.Errorf("a failing inhibitor must not matter: %s", weak["oc-top"])
	}
	// An inhibitor is not a precondition: it never counts toward "all".
	only := pathway(t, d, map[string]impact.State{"a": on, "rival": off, "oc-top": on})
	if only["oc-top"] != on {
		t.Errorf("inhibitor leaked into the logic: %s", only["oc-top"])
	}
}

func TestFeedbackLinksAreNotPreconditions(t *testing.T) {
	d := chain(impact.Logic{Type: impact.LogicAll}, "a")
	d.Links = append(d.Links, impact.CausalLink{ID: "ln-fb", From: "oc-top", To: "a", Feedback: true})
	got := pathway(t, d, map[string]impact.State{"a": ach, "oc-top": ach})
	if got["a"] != ach || got["oc-top"] != ach {
		t.Errorf("feedback changed the result: %v", got)
	}
}
