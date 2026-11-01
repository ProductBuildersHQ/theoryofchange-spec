package measure

import (
	"math"
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/graph"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

func set(assessments ...impact.AssumptionAssessment) []*impact.ObservationSet {
	return []*impact.ObservationSet{{Assessments: assessments}}
}

func TestEffectiveAssumptions(t *testing.T) {
	doc := impact.NewModel("t", "p")
	doc.Assumptions = []impact.Assumption{
		{ID: "a1"}, {ID: "a2", Status: impact.AssumptionHolding}, {ID: "a3", Status: impact.AssumptionHolding},
		{ID: "a4", Provenance: &impact.Provenance{Basis: impact.BasisRejected}},
	}
	got := EffectiveAssumptions(doc, set(
		impact.AssumptionAssessment{AssumptionID: "a2", Date: "2026-03-01", Status: impact.AssumptionAtRisk},
		impact.AssumptionAssessment{AssumptionID: "a2", Date: "2026-06-01", Status: impact.AssumptionInvalidated},
		impact.AssumptionAssessment{AssumptionID: "a3", Date: "2027-01-01", Status: impact.AssumptionInvalidated}, // after asOf
	), day("2026-07-01"))
	want := []impact.AssumptionState{
		{AssumptionID: "a1", Status: impact.AssumptionUntested, Source: impact.SourceDocument},
		{AssumptionID: "a2", Status: impact.AssumptionInvalidated, Source: impact.SourceAssessment, AssessedAt: "2026-06-01"},
		{AssumptionID: "a3", Status: impact.AssumptionHolding, Source: impact.SourceDocument},
	}
	if !slices.Equal(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
}

func near6(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func twoOutcomes() *impact.Model {
	d := impact.NewModel("t", "p")
	d.Outcomes = []impact.Outcome{{ID: "low", Level: impact.LevelShortTerm}, {ID: "top", Level: impact.LevelImpact}}
	d.Evidence = []impact.Evidence{
		{ID: "ev-strong", Strength: impact.StrengthStrong, Direction: impact.EvidenceSupports},
		{ID: "ev-weak", Strength: impact.StrengthWeak},
		{ID: "ev-against", Strength: impact.StrengthStrong, Direction: impact.EvidenceContradicts},
	}
	return d
}

func TestLinkConfidenceIsRuleBased(t *testing.T) {
	w := impact.DefaultStatusOptions().Weights // 0.4 / 0.4 / 0.2
	tests := []struct {
		name    string
		link    impact.CausalLink
		assumed map[string]impact.AssumptionStatus
		want    float64
		state   impact.LinkState
	}{
		{"nothing at all: no assumptions score 0.5", impact.CausalLink{}, nil, 0.4 * 0.5, impact.LinkStateOK},
		{"rationale only", impact.CausalLink{Rationale: "r"}, nil, 0.4*0 + 0.4*0.5 + 0.2, impact.LinkStateOK},
		{"strong evidence, validated, rationale", impact.CausalLink{Rationale: "r", EvidenceIDs: []string{"ev-strong"}, AssumptionIDs: []string{"a"}}, map[string]impact.AssumptionStatus{"a": impact.AssumptionValidated}, 1, impact.LinkStateOK},
		{"best supporting evidence counts", impact.CausalLink{EvidenceIDs: []string{"ev-weak", "ev-strong"}}, nil, 0.4*1 + 0.4*0.5, impact.LinkStateOK},
		{"contradicting evidence takes half off", impact.CausalLink{EvidenceIDs: []string{"ev-strong", "ev-against"}}, nil, 0.4*0.5 + 0.4*0.5, impact.LinkStateOK},
		{"contradiction floors at zero", impact.CausalLink{EvidenceIDs: []string{"ev-against"}}, nil, 0.4 * 0.5, impact.LinkStateOK},
		{"lowest assumption wins", impact.CausalLink{AssumptionIDs: []string{"a", "b"}}, map[string]impact.AssumptionStatus{"a": impact.AssumptionValidated, "b": impact.AssumptionHolding}, 0.4 * 0.8, impact.LinkStateOK},
		{"at risk assumption", impact.CausalLink{AssumptionIDs: []string{"a"}}, map[string]impact.AssumptionStatus{"a": impact.AssumptionAtRisk}, 0.4 * 0.25, impact.LinkStateAtRisk},
		{"invalidated assumption breaks the link", impact.CausalLink{AssumptionIDs: []string{"a", "b"}}, map[string]impact.AssumptionStatus{"a": impact.AssumptionInvalidated, "b": impact.AssumptionValidated}, 0, impact.LinkStateBroken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := twoOutcomes()
			tt.link.ID, tt.link.From, tt.link.To = "ln", "low", "top"
			d.Links = []impact.CausalLink{tt.link}
			var states []impact.AssumptionState
			for id, s := range tt.assumed {
				states = append(states, impact.AssumptionState{AssumptionID: id, Status: s})
			}
			got := LinkStatuses(d, states, w)
			if len(got) != 1 || got[0].State != tt.state || !near6(got[0].Confidence, tt.want) {
				t.Errorf("got %+v, want %s %.4f", got, tt.state, tt.want)
			}
		})
	}
}

func TestAssumptionsThatNameALinkInAppliesToCount(t *testing.T) {
	d := twoOutcomes()
	d.Links = []impact.CausalLink{{ID: "ln", From: "low", To: "top"}}
	d.Assumptions = []impact.Assumption{{ID: "a", AppliesTo: []string{"ln"}}}
	got := LinkStatuses(d, []impact.AssumptionState{{AssumptionID: "a", Status: impact.AssumptionInvalidated}}, impact.DefaultStatusOptions().Weights)
	if got[0].State != impact.LinkStateBroken {
		t.Errorf("state = %s", got[0].State)
	}
}

func TestWeightsAreOptions(t *testing.T) {
	d := twoOutcomes()
	d.Links = []impact.CausalLink{{ID: "ln", From: "low", To: "top", Rationale: "r"}}
	got := LinkStatuses(d, nil, impact.ConfidenceWeights{Evidence: 0, Assumptions: 0, Rationale: 1})
	if !near6(got[0].Confidence, 1) {
		t.Errorf("confidence = %v", got[0].Confidence)
	}
}

// chain3: a -> b -> top, with an assumption on the first link.
func chain3() (*impact.Model, *graph.Graph) {
	d := impact.NewModel("t", "p")
	d.Outcomes = []impact.Outcome{{ID: "a", Title: "a", Level: impact.LevelShortTerm}, {ID: "b", Title: "b", Level: impact.LevelMediumTerm}, {ID: "top", Title: "top", Level: impact.LevelImpact}, {ID: "side", Title: "side", Level: impact.LevelShortTerm}}
	d.Links = []impact.CausalLink{
		{ID: "ln-ab", From: "a", To: "b", Rationale: "r", AssumptionIDs: []string{"asm"}},
		{ID: "ln-bt", From: "b", To: "top", Rationale: "r"},
		{ID: "ln-st", From: "side", To: "top", Rationale: "r"},
	}
	d.Assumptions = []impact.Assumption{{ID: "asm", Statement: "s", Criticality: impact.CriticalityHigh}}
	g, err := graph.Build(d)
	if err != nil {
		panic(err)
	}
	return d, g
}

func TestInvalidatedAssumptionFlagsEveryDownstreamOutcome(t *testing.T) {
	d, g := chain3()
	states := []impact.AssumptionState{{AssumptionID: "asm", Status: impact.AssumptionInvalidated}}
	p := Propagate(d, g, states)
	for _, id := range []string{"b", "top"} {
		if !slices.Equal(p.Broken[id], []string{"asm"}) || p.Cap[id] != impact.StateAtRisk {
			t.Errorf("%s: broken %v cap %s", id, p.Broken[id], p.Cap[id])
		}
	}
	for _, id := range []string{"a", "side"} {
		if _, flagged := p.Broken[id]; flagged {
			t.Errorf("%s is not downstream of the broken link", id)
		}
	}
	pathway := map[string]impact.State{"a": impact.StateAchieved, "b": impact.StateAchieved, "top": impact.StateAchieved, "side": impact.StateAchieved}
	p.Apply(pathway)
	if pathway["b"] != impact.StateAtRisk || pathway["top"] != impact.StateAtRisk || pathway["a"] != impact.StateAchieved || pathway["side"] != impact.StateAchieved {
		t.Errorf("after capping: %v", pathway)
	}
}

func TestAtRiskAssumptionsCapAtRisk(t *testing.T) {
	d, g := chain3()
	p := Propagate(d, g, []impact.AssumptionState{{AssumptionID: "asm", Status: impact.AssumptionAtRisk}})
	if p.Cap["top"] != impact.StateAtRisk {
		t.Errorf("cap = %s", p.Cap["top"])
	}
	for _, s := range []impact.AssumptionStatus{impact.AssumptionHolding, impact.AssumptionValidated, impact.AssumptionUntested} {
		if p := Propagate(d, g, []impact.AssumptionState{{AssumptionID: "asm", Status: s}}); len(p.Broken) != 0 {
			t.Errorf("%s flagged outcomes: %v", s, p.Broken)
		}
	}
}

func TestCriticalInvalidatedAssumptionCapsOffTrack(t *testing.T) {
	d, g := chain3()
	d.Assumptions[0].Criticality = impact.CriticalityCritical
	p := Propagate(d, g, []impact.AssumptionState{{AssumptionID: "asm", Status: impact.AssumptionInvalidated}})
	if p.Cap["b"] != impact.StateOffTrack || p.Cap["top"] != impact.StateOffTrack {
		t.Errorf("caps = %v", p.Cap)
	}
	// Critical but only at risk is still just at risk.
	p = Propagate(d, g, []impact.AssumptionState{{AssumptionID: "asm", Status: impact.AssumptionAtRisk}})
	if p.Cap["top"] != impact.StateAtRisk {
		t.Errorf("critical at-risk cap = %s", p.Cap["top"])
	}
}

func TestAssumptionsOnOutcomesAndInterventionsPropagate(t *testing.T) {
	d, _ := chain3()
	d.Links[0].AssumptionIDs = nil
	d.Assumptions = []impact.Assumption{{ID: "on-outcome", Statement: "s", AppliesTo: []string{"b"}}, {ID: "on-int", Statement: "s", AppliesTo: []string{"int"}}}
	d.Interventions = []impact.Intervention{{ID: "int", Title: "i", Targets: []string{"side"}}}
	g, err := graph.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	p := Propagate(d, g, []impact.AssumptionState{
		{AssumptionID: "on-outcome", Status: impact.AssumptionInvalidated}, {AssumptionID: "on-int", Status: impact.AssumptionAtRisk},
	})
	if !slices.Equal(p.Broken["b"], []string{"on-outcome"}) || !slices.Equal(p.Broken["top"], []string{"on-int", "on-outcome"}) || !slices.Equal(p.Broken["side"], []string{"on-int"}) {
		t.Errorf("broken = %v", p.Broken)
	}
}

func TestPathAndOutcomeConfidence(t *testing.T) {
	d, g := chain3()
	links := []impact.LinkStatus{{LinkID: "ln-ab", Confidence: 0.5}, {LinkID: "ln-bt", Confidence: 0.8}, {LinkID: "ln-st", Confidence: 0.3}}
	outcomes, impacts := Confidence(d, g, links)
	want := map[string]float64{"top": 1, "b": 0.8, "a": 0.4, "side": 0.3}
	for id, w := range want {
		if !near6(outcomes[id], w) {
			t.Errorf("confidence(%s) = %v, want %v", id, outcomes[id], w)
		}
	}
	// The impact's headline is its best complete chain: a->b->top is 0.4, side->top is 0.3.
	if !near6(impacts["top"], 0.4) {
		t.Errorf("impact confidence = %v, want 0.4", impacts["top"])
	}

	// Two links between the same outcomes: the stronger one counts.
	d.Links = append(d.Links, impact.CausalLink{ID: "ln-bt2", From: "b", To: "top"})
	g, err := graph.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	links = append(links, impact.LinkStatus{LinkID: "ln-bt2", Confidence: 0.9})
	outcomes, _ = Confidence(d, g, links)
	if !near6(outcomes["b"], 0.9) {
		t.Errorf("parallel links: %v", outcomes["b"])
	}
}

func TestOutcomeWithoutAPathHasNoConfidence(t *testing.T) {
	d := impact.NewModel("t", "p")
	d.Outcomes = []impact.Outcome{{ID: "top", Title: "t", Level: impact.LevelImpact}, {ID: "lost", Title: "l", Level: impact.LevelShortTerm}}
	g, err := graph.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	outcomes, impacts := Confidence(d, g, nil)
	if _, ok := outcomes["lost"]; ok {
		t.Error("a dead end has a confidence")
	}
	if outcomes["top"] != 1 {
		t.Errorf("impact outcome = %v", outcomes["top"])
	}
	if _, ok := impacts["top"]; ok {
		t.Error("an impact nothing leads to has a headline confidence")
	}
}
