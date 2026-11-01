package problemtree

import (
	"encoding/json"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

func model(causes []impact.Cause, effects []impact.Effect) *impact.Model {
	d := impact.NewModel("tree", "Poor eating health")
	d.Causes, d.Effects = causes, effects
	return d
}

func rowOf(t *testing.T, tr *Tree, id string) int {
	t.Helper()
	for _, r := range tr.Rows {
		for _, n := range r.Nodes {
			if n.ID == id {
				return r.Index
			}
		}
	}
	t.Fatalf("no node %s", id)
	return 0
}

func TestLayering(t *testing.T) {
	tr := Build(model(
		[]impact.Cause{
			{ID: "c-direct", Statement: "Too busy", Level: impact.CauseImmediate},
			{ID: "c-deep", Statement: "Fast food", Level: impact.CauseUnderlying, Parents: []string{"c-direct"}},
			// Two parents at different depths: it sits below the lower one.
			{ID: "c-both", Statement: "Long commutes", Level: impact.CauseStructural, Parents: []string{"c-direct", "c-deep"}},
		},
		[]impact.Effect{
			{ID: "e-direct", Statement: "Diabetes"},
			{ID: "e-far", Statement: "Health costs", Parents: []string{"e-direct"}},
			{ID: "e-both", Statement: "Lost work", Parents: []string{"e-direct", "e-far"}},
		}))
	for id, want := range map[string]int{
		ProblemID: 0, "c-direct": -1, "c-deep": -2, "c-both": -3,
		"e-direct": 1, "e-far": 2, "e-both": 3,
	} {
		if got := rowOf(t, tr, id); got != want {
			t.Errorf("%s is in row %d, want %d", id, got, want)
		}
	}
	// Rows run from the top down, and every edge points up.
	for i := 1; i < len(tr.Rows); i++ {
		if tr.Rows[i].Index >= tr.Rows[i-1].Index {
			t.Fatalf("rows are not top to bottom: %d then %d", tr.Rows[i-1].Index, tr.Rows[i].Index)
		}
	}
	for _, e := range tr.Edges {
		if rowOf(t, tr, e.From) >= rowOf(t, tr, e.To) {
			t.Errorf("edge %s -> %s does not point up", e.From, e.To)
		}
	}
}

func TestRejectedSetAsideAndLoops(t *testing.T) {
	rejected := &impact.Provenance{Basis: impact.BasisRejected}
	tr := Build(model(
		[]impact.Cause{
			{ID: "c-kept", Statement: "Stress", Level: impact.CauseImmediate, Parents: []string{"c-gone"}},
			{ID: "c-gone", Statement: "Rejected", Level: impact.CauseImmediate, Provenance: rejected},
			{ID: "c-aside", Statement: "Poverty", Level: impact.CauseStructural, OutOfScope: &impact.OutOfScope{Reason: "Beyond us."}},
			// A loop, which validation reports, is cut rather than followed forever.
			{ID: "c-a", Statement: "A", Level: impact.CauseImmediate, Parents: []string{"c-b"}},
			{ID: "c-b", Statement: "B", Level: impact.CauseImmediate, Parents: []string{"c-a"}},
		}, nil))
	for _, r := range tr.Rows {
		for _, n := range r.Nodes {
			if n.ID == "c-gone" {
				t.Error("a rejected cause is drawn")
			}
			if n.ID == "c-aside" && !n.SetAside {
				t.Error("a cause set aside is not marked")
			}
		}
	}
	// A cause whose only parent was rejected leads straight to the problem.
	if got := rowOf(t, tr, "c-kept"); got != -1 {
		t.Errorf("c-kept is in row %d, want -1", got)
	}
	for _, e := range tr.Edges {
		if e.From == "c-gone" || e.To == "c-gone" {
			t.Errorf("an edge touches the rejected cause: %+v", e)
		}
	}
}

func TestSpecIsJSONAndDeterministic(t *testing.T) {
	d := model([]impact.Cause{{ID: "c-1", Statement: "Stress", Level: impact.CauseImmediate}},
		[]impact.Effect{{ID: "e-1", Statement: "Diabetes"}})
	first, err := Build(d).JSON()
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(first), &v); err != nil {
		t.Fatalf("the spec is not JSON: %v", err)
	}
	if v["direction"] != "up" {
		t.Errorf("direction = %v, want up", v["direction"])
	}
	firstD2 := Build(d).D2()
	for range 5 {
		again, err := Build(d).JSON()
		if err != nil {
			t.Fatal(err)
		}
		if again != first {
			t.Fatal("two specs differ")
		}
		if Build(d).D2() != firstD2 {
			t.Fatal("two D2 renderings differ")
		}
	}
}
