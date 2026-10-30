package measure

import (
	"math"
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func set(assessments ...toc.AssumptionAssessment) []*toc.ObservationSet {
	return []*toc.ObservationSet{{Assessments: assessments}}
}

func TestEffectiveAssumptions(t *testing.T) {
	doc := toc.NewDocument("t", "p")
	doc.Assumptions = []toc.Assumption{
		{ID: "a1"}, {ID: "a2", Status: toc.AssumptionHolding}, {ID: "a3", Status: toc.AssumptionHolding},
		{ID: "a4", Provenance: &toc.Provenance{Basis: toc.BasisRejected}},
	}
	got := EffectiveAssumptions(doc, set(
		toc.AssumptionAssessment{AssumptionID: "a2", Date: "2026-03-01", Status: toc.AssumptionAtRisk},
		toc.AssumptionAssessment{AssumptionID: "a2", Date: "2026-06-01", Status: toc.AssumptionInvalidated},
		toc.AssumptionAssessment{AssumptionID: "a3", Date: "2027-01-01", Status: toc.AssumptionInvalidated}, // after asOf
	), day("2026-07-01"))
	want := []toc.AssumptionState{
		{AssumptionID: "a1", Status: toc.AssumptionUntested, Source: toc.SourceDocument},
		{AssumptionID: "a2", Status: toc.AssumptionInvalidated, Source: toc.SourceAssessment, AssessedAt: "2026-06-01"},
		{AssumptionID: "a3", Status: toc.AssumptionHolding, Source: toc.SourceDocument},
	}
	if !slices.Equal(got, want) {
		t.Errorf("got  %+v\nwant %+v", got, want)
	}
}

func near6(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func twoOutcomes() *toc.Document {
	d := toc.NewDocument("t", "p")
	d.Outcomes = []toc.Outcome{{ID: "low", Level: toc.LevelShortTerm}, {ID: "top", Level: toc.LevelImpact}}
	d.Evidence = []toc.Evidence{
		{ID: "ev-strong", Strength: toc.StrengthStrong, Direction: toc.EvidenceSupports},
		{ID: "ev-weak", Strength: toc.StrengthWeak},
		{ID: "ev-against", Strength: toc.StrengthStrong, Direction: toc.EvidenceContradicts},
	}
	return d
}

func TestLinkConfidenceIsRuleBased(t *testing.T) {
	w := toc.DefaultStatusOptions().Weights // 0.4 / 0.4 / 0.2
	tests := []struct {
		name    string
		link    toc.CausalLink
		assumed map[string]toc.AssumptionStatus
		want    float64
		state   toc.LinkState
	}{
		{"nothing at all: no assumptions score 0.5", toc.CausalLink{}, nil, 0.4 * 0.5, toc.LinkStateOK},
		{"rationale only", toc.CausalLink{Rationale: "r"}, nil, 0.4*0 + 0.4*0.5 + 0.2, toc.LinkStateOK},
		{"strong evidence, validated, rationale", toc.CausalLink{Rationale: "r", EvidenceIDs: []string{"ev-strong"}, AssumptionIDs: []string{"a"}}, map[string]toc.AssumptionStatus{"a": toc.AssumptionValidated}, 1, toc.LinkStateOK},
		{"best supporting evidence counts", toc.CausalLink{EvidenceIDs: []string{"ev-weak", "ev-strong"}}, nil, 0.4*1 + 0.4*0.5, toc.LinkStateOK},
		{"contradicting evidence takes half off", toc.CausalLink{EvidenceIDs: []string{"ev-strong", "ev-against"}}, nil, 0.4*0.5 + 0.4*0.5, toc.LinkStateOK},
		{"contradiction floors at zero", toc.CausalLink{EvidenceIDs: []string{"ev-against"}}, nil, 0.4 * 0.5, toc.LinkStateOK},
		{"lowest assumption wins", toc.CausalLink{AssumptionIDs: []string{"a", "b"}}, map[string]toc.AssumptionStatus{"a": toc.AssumptionValidated, "b": toc.AssumptionHolding}, 0.4 * 0.8, toc.LinkStateOK},
		{"at risk assumption", toc.CausalLink{AssumptionIDs: []string{"a"}}, map[string]toc.AssumptionStatus{"a": toc.AssumptionAtRisk}, 0.4 * 0.25, toc.LinkStateAtRisk},
		{"invalidated assumption breaks the link", toc.CausalLink{AssumptionIDs: []string{"a", "b"}}, map[string]toc.AssumptionStatus{"a": toc.AssumptionInvalidated, "b": toc.AssumptionValidated}, 0, toc.LinkStateBroken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := twoOutcomes()
			tt.link.ID, tt.link.From, tt.link.To = "ln", "low", "top"
			d.Links = []toc.CausalLink{tt.link}
			var states []toc.AssumptionState
			for id, s := range tt.assumed {
				states = append(states, toc.AssumptionState{AssumptionID: id, Status: s})
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
	d.Links = []toc.CausalLink{{ID: "ln", From: "low", To: "top"}}
	d.Assumptions = []toc.Assumption{{ID: "a", AppliesTo: []string{"ln"}}}
	got := LinkStatuses(d, []toc.AssumptionState{{AssumptionID: "a", Status: toc.AssumptionInvalidated}}, toc.DefaultStatusOptions().Weights)
	if got[0].State != toc.LinkStateBroken {
		t.Errorf("state = %s", got[0].State)
	}
}

func TestWeightsAreOptions(t *testing.T) {
	d := twoOutcomes()
	d.Links = []toc.CausalLink{{ID: "ln", From: "low", To: "top", Rationale: "r"}}
	got := LinkStatuses(d, nil, toc.ConfidenceWeights{Evidence: 0, Assumptions: 0, Rationale: 1})
	if !near6(got[0].Confidence, 1) {
		t.Errorf("confidence = %v", got[0].Confidence)
	}
}

// chain3: a -> b -> top, with an assumption on the first link.
func chain3() (*toc.Document, *graph.Graph) {
	d := toc.NewDocument("t", "p")
	d.Outcomes = []toc.Outcome{{ID: "a", Title: "a", Level: toc.LevelShortTerm}, {ID: "b", Title: "b", Level: toc.LevelIntermediate}, {ID: "top", Title: "top", Level: toc.LevelImpact}, {ID: "side", Title: "side", Level: toc.LevelShortTerm}}
	d.Links = []toc.CausalLink{
		{ID: "ln-ab", From: "a", To: "b", Rationale: "r", AssumptionIDs: []string{"asm"}},
		{ID: "ln-bt", From: "b", To: "top", Rationale: "r"},
		{ID: "ln-st", From: "side", To: "top", Rationale: "r"},
	}
	d.Assumptions = []toc.Assumption{{ID: "asm", Statement: "s", Criticality: toc.CriticalityHigh}}
	g, err := graph.Build(d)
	if err != nil {
		panic(err)
	}
	return d, g
}

func TestInvalidatedAssumptionFlagsEveryDownstreamOutcome(t *testing.T) {
	d, g := chain3()
	states := []toc.AssumptionState{{AssumptionID: "asm", Status: toc.AssumptionInvalidated}}
	p := Propagate(d, g, states)
	for _, id := range []string{"b", "top"} {
		if !slices.Equal(p.Broken[id], []string{"asm"}) || p.Cap[id] != toc.StateAtRisk {
			t.Errorf("%s: broken %v cap %s", id, p.Broken[id], p.Cap[id])
		}
	}
	for _, id := range []string{"a", "side"} {
		if _, flagged := p.Broken[id]; flagged {
			t.Errorf("%s is not downstream of the broken link", id)
		}
	}
	pathway := map[string]toc.State{"a": toc.StateAchieved, "b": toc.StateAchieved, "top": toc.StateAchieved, "side": toc.StateAchieved}
	p.Apply(pathway)
	if pathway["b"] != toc.StateAtRisk || pathway["top"] != toc.StateAtRisk || pathway["a"] != toc.StateAchieved || pathway["side"] != toc.StateAchieved {
		t.Errorf("after capping: %v", pathway)
	}
}

func TestAtRiskAssumptionsCapAtRisk(t *testing.T) {
	d, g := chain3()
	p := Propagate(d, g, []toc.AssumptionState{{AssumptionID: "asm", Status: toc.AssumptionAtRisk}})
	if p.Cap["top"] != toc.StateAtRisk {
		t.Errorf("cap = %s", p.Cap["top"])
	}
	for _, s := range []toc.AssumptionStatus{toc.AssumptionHolding, toc.AssumptionValidated, toc.AssumptionUntested} {
		if p := Propagate(d, g, []toc.AssumptionState{{AssumptionID: "asm", Status: s}}); len(p.Broken) != 0 {
			t.Errorf("%s flagged outcomes: %v", s, p.Broken)
		}
	}
}

func TestCriticalInvalidatedAssumptionCapsOffTrack(t *testing.T) {
	d, g := chain3()
	d.Assumptions[0].Criticality = toc.CriticalityCritical
	p := Propagate(d, g, []toc.AssumptionState{{AssumptionID: "asm", Status: toc.AssumptionInvalidated}})
	if p.Cap["b"] != toc.StateOffTrack || p.Cap["top"] != toc.StateOffTrack {
		t.Errorf("caps = %v", p.Cap)
	}
	// Critical but only at risk is still just at risk.
	p = Propagate(d, g, []toc.AssumptionState{{AssumptionID: "asm", Status: toc.AssumptionAtRisk}})
	if p.Cap["top"] != toc.StateAtRisk {
		t.Errorf("critical at-risk cap = %s", p.Cap["top"])
	}
}

func TestAssumptionsOnOutcomesAndInterventionsPropagate(t *testing.T) {
	d, _ := chain3()
	d.Links[0].AssumptionIDs = nil
	d.Assumptions = []toc.Assumption{{ID: "on-outcome", Statement: "s", AppliesTo: []string{"b"}}, {ID: "on-int", Statement: "s", AppliesTo: []string{"int"}}}
	d.Interventions = []toc.Intervention{{ID: "int", Title: "i", Targets: []string{"side"}}}
	g, err := graph.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	p := Propagate(d, g, []toc.AssumptionState{
		{AssumptionID: "on-outcome", Status: toc.AssumptionInvalidated}, {AssumptionID: "on-int", Status: toc.AssumptionAtRisk},
	})
	if !slices.Equal(p.Broken["b"], []string{"on-outcome"}) || !slices.Equal(p.Broken["top"], []string{"on-int", "on-outcome"}) || !slices.Equal(p.Broken["side"], []string{"on-int"}) {
		t.Errorf("broken = %v", p.Broken)
	}
}

func TestPathAndOutcomeConfidence(t *testing.T) {
	d, g := chain3()
	links := []toc.LinkStatus{{LinkID: "ln-ab", Confidence: 0.5}, {LinkID: "ln-bt", Confidence: 0.8}, {LinkID: "ln-st", Confidence: 0.3}}
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
	d.Links = append(d.Links, toc.CausalLink{ID: "ln-bt2", From: "b", To: "top"})
	g, err := graph.Build(d)
	if err != nil {
		t.Fatal(err)
	}
	links = append(links, toc.LinkStatus{LinkID: "ln-bt2", Confidence: 0.9})
	outcomes, _ = Confidence(d, g, links)
	if !near6(outcomes["b"], 0.9) {
		t.Errorf("parallel links: %v", outcomes["b"])
	}
}

func TestOutcomeWithoutAPathHasNoConfidence(t *testing.T) {
	d := toc.NewDocument("t", "p")
	d.Outcomes = []toc.Outcome{{ID: "top", Title: "t", Level: toc.LevelImpact}, {ID: "lost", Title: "l", Level: toc.LevelShortTerm}}
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
