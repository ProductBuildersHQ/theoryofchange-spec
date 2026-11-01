package logicmodel

import (
	"slices"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// view is a small theory: two activities share a resource, one lists an
// output, and the outcomes run from output level to impact.
func view() *report.View {
	return &report.View{
		Document: report.DocumentView{Title: "Whole theory"},
		Interventions: []report.InterventionView{
			{ID: "int-a", Title: "Run workshops", Resources: "Two staff", Outputs: []string{"Workshops held"}, Targets: []string{"oc-skills"}},
			{ID: "int-b", Title: "Coach leaders", Resources: "Two staff", Targets: []string{"oc-reach"}},
		},
		Outcomes: []report.OutcomeView{
			{ID: "oc-reach", Title: "Members reached", Level: toc.LevelOutput},
			{ID: "oc-skills", Title: "Skills gained", Level: toc.LevelShortTerm},
			{ID: "oc-practice", Title: "Practice applied", Level: toc.LevelMediumTerm, Basis: report.BasisView{Basis: toc.BasisInferred, Marked: true}},
			{ID: "oc-strong", Title: "Members are strong", Level: toc.LevelLongTerm},
			{ID: "oc-served", Title: "Residents served", Level: toc.LevelImpact},
		},
		Links: []report.LinkView{
			{ID: "l1", From: "oc-reach", To: "oc-skills"},
			{ID: "l2", From: "oc-skills", To: "oc-practice"},
			{ID: "l3", From: "oc-practice", To: "oc-strong"},
			{ID: "l4", From: "oc-strong", To: "oc-served"},
			{ID: "l5", From: "oc-strong", To: "oc-elsewhere"}, // an outcome of another pathway
			{ID: "l6", From: "oc-skills", To: "oc-skills"},    // a loop back to itself
		},
	}
}

func hasEdge(g *Graph, from, to string) bool { return slices.Contains(g.Edges, Edge{from, to}) }

func ids(ns []Node) []string {
	var out []string
	for _, n := range ns {
		out = append(out, n.ID)
	}
	return out
}

func TestStagesRunFromInputsToImpact(t *testing.T) {
	g := Build(view())
	var got []StageID
	for _, s := range g.Stages {
		got = append(got, s.Stage)
	}
	want := []StageID{StageInputs, StageActivities, StageOutputs, StageShortTerm, StageMediumTerm, StageLongTerm, StageImpact}
	if !slices.Equal(got, want) {
		t.Errorf("stages = %v, want %v", got, want)
	}
	if g.Title != "Whole theory" {
		t.Errorf("title = %q", g.Title)
	}
}

func TestEveryKindOfEdgeComesFromTheTheory(t *testing.T) {
	g := Build(view())
	for _, e := range []Edge{
		{"_input1", "int-a"}, {"_input1", "int-b"}, // one resource feeds two activities
		{"int-a", "_output1"},     // an activity leads to the output it lists
		{"_output1", "oc-skills"}, // which leads to the outcome the activity targets
		{"int-b", "oc-reach"},     // an activity leads to an output-level outcome it targets
		{"oc-reach", "oc-skills"}, // links join outcomes
		{"oc-skills", "oc-practice"}, {"oc-practice", "oc-strong"}, {"oc-strong", "oc-served"},
	} {
		if !hasEdge(g, e.From, e.To) {
			t.Errorf("missing edge %s -> %s; edges are %v", e.From, e.To, g.Edges)
		}
	}
	if hasEdge(g, "int-a", "oc-skills") {
		t.Error("an activity that lists outputs should lead to its targets through them, not around them")
	}
}

func TestEdgesToNowhereAndToSelfAreDropped(t *testing.T) {
	g := Build(view())
	for _, e := range g.Edges {
		if e.From == e.To || strings.Contains(e.To, "elsewhere") || strings.Contains(e.From, "elsewhere") {
			t.Errorf("edge %v should not be drawn", e)
		}
	}
	seen := map[Edge]bool{}
	for _, e := range g.Edges {
		if seen[e] {
			t.Errorf("edge %v is repeated", e)
		}
		seen[e] = true
	}
}

func TestSharedTextIsOneNode(t *testing.T) {
	g := Build(view())
	var inputs []Node
	for _, s := range g.Stages {
		if s.Stage == StageInputs {
			inputs = s.Nodes
		}
	}
	if len(inputs) != 1 || inputs[0].Label != "Two staff" {
		t.Errorf("two activities that state the same resource should share one input: %v", inputs)
	}
}

func TestMarkedElementsSayWhatTheyAre(t *testing.T) {
	g := Build(view())
	for _, s := range g.Stages {
		for _, n := range s.Nodes {
			if n.ID == "oc-practice" && n.Label != "Practice applied (inferred)" {
				t.Errorf("label = %q", n.Label)
			}
		}
	}
}

func TestStagesWithNothingInThemAreLeftOut(t *testing.T) {
	v := &report.View{Outcomes: []report.OutcomeView{{ID: "a", Title: "A", Level: toc.LevelShortTerm}, {ID: "b", Title: "B", Level: toc.LevelImpact}}}
	g := Build(v)
	if len(g.Stages) != 2 || g.Stages[0].Stage != StageShortTerm || g.Stages[1].Stage != StageImpact {
		t.Errorf("stages = %+v", g.Stages)
	}
}

// Nodes in a later stage are ordered by where the nodes that lead to them sit,
// so arrows do not cross when they need not.
func TestNodesAreOrderedToAvoidCrossings(t *testing.T) {
	v := &report.View{Outcomes: []report.OutcomeView{
		{ID: "s1", Title: "S1", Level: toc.LevelShortTerm},
		{ID: "s2", Title: "S2", Level: toc.LevelShortTerm},
		// Written in the opposite order to the nodes that lead to them.
		{ID: "m2", Title: "M2", Level: toc.LevelMediumTerm},
		{ID: "m1", Title: "M1", Level: toc.LevelMediumTerm},
	}, Links: []report.LinkView{{From: "s1", To: "m1"}, {From: "s2", To: "m2"}}}
	g := Build(v)
	var mediums []string
	for _, s := range g.Stages {
		if s.Stage == StageMediumTerm {
			mediums = ids(s.Nodes)
		}
	}
	if !slices.Equal(mediums, []string{"m1", "m2"}) {
		t.Errorf("order = %v, want [m1 m2]", mediums)
	}
}

func TestBuildIsDeterministic(t *testing.T) {
	a, b := Build(view()), Build(view())
	if len(a.Edges) != len(b.Edges) || !slices.Equal(a.Edges, b.Edges) {
		t.Error("the same view gave different edges")
	}
	for i := range a.Stages {
		if !slices.Equal(ids(a.Stages[i].Nodes), ids(b.Stages[i].Nodes)) {
			t.Error("the same view gave different node order")
		}
	}
}
