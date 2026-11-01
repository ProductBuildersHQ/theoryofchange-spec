// Package problemtree draws the problem tree of an impact model: the causes
// (roots) below the core problem and the effects (branches) above it, with
// arrows that flow upward from the deepest causes to the core problem and on
// to the furthest effects.
//
// The tree is built first as rows of nodes and the edges between them, then
// described as a d2vision diagram spec, a JSON document from which D2 and
// Mermaid are rendered.
package problemtree

import (
	"math"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// Kind is what a node of the tree is.
type Kind string

// The kinds of node.
const (
	KindCause   Kind = "cause"
	KindProblem Kind = "problem"
	KindEffect  Kind = "effect"
)

// ProblemID is the node ID of the core problem.
const ProblemID = "problem"

// Node is one box of the tree.
type Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Kind  Kind   `json:"kind"`
	// Row is 0 for the core problem, negative for causes (deeper is lower),
	// and positive for effects (further is higher).
	Row int `json:"row"`
	// SetAside is true for a cause or effect marked out of scope.
	SetAside bool `json:"setAside,omitempty"`
}

// Edge points from a cause to what it leads to, or from the core problem or
// an effect to the effect that follows. Every edge points up a row or more.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Row is the nodes at one height, in file order.
type Row struct {
	Index int    `json:"index"`
	Kind  Kind   `json:"kind"`
	Nodes []Node `json:"nodes"`
}

// Tree is the laid-out problem tree.
type Tree struct {
	// Rows run from the top (the furthest effects) to the bottom (the deepest
	// causes).
	Rows  []Row  `json:"rows"`
	Edges []Edge `json:"edges"`
}

// Empty reports whether the tree holds only the core problem: no cause or
// effect has been recorded, or every one was rejected.
func (t *Tree) Empty() bool { return len(t.Rows) <= 1 }

// Build lays out the problem tree of doc. Rejected causes and effects are left
// out, and so are references to them. A cause with no parents is one row
// below the core problem, and any other cause is one row below its lowest
// parent; effects mirror this upward. Within a row, nodes keep file order.
func Build(doc *impact.Model) *Tree {
	causes := map[string]impact.Cause{}
	for _, c := range doc.Causes {
		if !c.Rejected() {
			causes[c.ID] = c
		}
	}
	effects := map[string]impact.Effect{}
	for _, e := range doc.Effects {
		if !e.Rejected() {
			effects[e.ID] = e
		}
	}

	causeRow := rows(func(id string) []string { return causes[id].Parents }, func(id string) bool { _, ok := causes[id]; return ok }, -1)
	effectRow := rows(func(id string) []string { return effects[id].Parents }, func(id string) bool { _, ok := effects[id]; return ok }, +1)

	byRow := map[int][]Node{0: {{ID: ProblemID, Label: oneLine(doc.Context.ProblemStatement), Kind: KindProblem}}}
	t := &Tree{}
	for _, c := range doc.Causes {
		if _, ok := causes[c.ID]; !ok {
			continue
		}
		r := causeRow(c.ID)
		byRow[r] = append(byRow[r], Node{ID: c.ID, Label: oneLine(c.Statement), Kind: KindCause, Row: r, SetAside: c.OutOfScope != nil})
		ups := live(c.Parents, causes)
		if len(ups) == 0 {
			ups = []string{ProblemID}
		}
		for _, p := range ups {
			t.Edges = append(t.Edges, Edge{From: c.ID, To: p})
		}
	}
	for _, e := range doc.Effects {
		if _, ok := effects[e.ID]; !ok {
			continue
		}
		r := effectRow(e.ID)
		byRow[r] = append(byRow[r], Node{ID: e.ID, Label: oneLine(e.Statement), Kind: KindEffect, Row: r, SetAside: e.OutOfScope != nil})
		downs := live(e.Parents, effects)
		if len(downs) == 0 {
			downs = []string{ProblemID}
		}
		for _, p := range downs {
			t.Edges = append(t.Edges, Edge{From: p, To: e.ID})
		}
	}

	top, bottom := 0, 0
	for r := range byRow {
		top, bottom = max(top, r), min(bottom, r)
	}
	for r := top; r >= bottom; r-- {
		nodes, ok := byRow[r]
		if !ok {
			continue
		}
		kind := KindProblem
		switch {
		case r > 0:
			kind = KindEffect
		case r < 0:
			kind = KindCause
		}
		t.Rows = append(t.Rows, Row{Index: r, Kind: kind, Nodes: nodes})
	}
	return t
}

// rows returns a function giving the row of a node: first (−1 or +1) when it
// has no live parents, otherwise one step further from the core problem than
// its furthest parent. A parent chain that loops, which validation reports,
// is cut where it repeats.
func rows(parents func(string) []string, alive func(string) bool, first int) func(string) int {
	memo := map[string]int{}
	visiting := map[string]bool{}
	var row func(id string) int
	row = func(id string) int {
		if r, ok := memo[id]; ok {
			return r
		}
		visiting[id] = true
		r := first
		far := 0
		for _, p := range parents(id) {
			if !alive(p) || visiting[p] {
				continue
			}
			pr := row(p)
			if math.Abs(float64(pr)) > math.Abs(float64(far)) {
				far = pr
			}
		}
		if far != 0 {
			r = far + first
		}
		visiting[id] = false
		memo[id] = r
		return r
	}
	return row
}

// live keeps the IDs that name a node of the tree.
func live[T any](ids []string, known map[string]T) []string {
	var out []string
	for _, id := range ids {
		if _, ok := known[id]; ok {
			out = append(out, id)
		}
	}
	return out
}
