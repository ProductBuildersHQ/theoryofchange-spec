// Package graph builds the causal graph of a Theory of Change document and
// answers structural questions about it: transitive preconditions and
// consequences, paths to impact, cycles, levels, orphans, dead ends, and the
// ceiling of accountability.
//
// The graph has one node per outcome and per intervention. Edges are the
// document's causal links plus an implicit edge from each intervention to
// each outcome it targets. Feedback links are kept (see [Graph.Edges] and
// [Graph.FeedbackEdges]) but are excluded from every structural query, so an
// intentional loop never counts as a cycle. Precondition groups do not
// change reachability; they are exposed through [Graph.GroupsFor].
//
// All results are deterministic: sets of IDs come back in node order (the
// document's outcomes, then its interventions) unless a method says
// otherwise, and no map iteration order reaches any output. A Graph is
// immutable after Build and safe for concurrent use.
package graph

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

// Sentinel errors. Callers match them with errors.Is.
var (
	// ErrInvalidDocument is returned by Build when validation reports an
	// error-severity issue.
	ErrInvalidDocument = errors.New("graph: invalid document")
	// ErrCycle is returned by TopologicalOrder when the non-feedback graph
	// has a cycle.
	ErrCycle = errors.New("graph: cycle in causal graph")
)

// DefaultMaxPaths bounds PathsToImpact when max <= 0.
const DefaultMaxPaths = 1000

// maxIssuesInError is how many validation issues Build quotes in its error.
const maxIssuesInError = 3

// Node is an outcome or an intervention.
type Node struct {
	ID   string
	Kind toc.ElementKind // toc.ElementOutcome or toc.ElementIntervention
	// Level is set for outcomes only.
	Level toc.Level
}

// Edge is a directed causal edge.
type Edge struct {
	// LinkID is the causal link's ID; empty for implicit edges.
	LinkID   string
	From, To string
	Feedback bool
	// Implicit marks an intervention -> outcome edge derived from
	// Intervention.Targets.
	Implicit bool
}

// Graph is the causal graph of one document.
type Graph struct {
	nodes []Node
	index map[string]int
	edges []Edge // all edges, in order
	fb    []Edge // feedback edges, in order

	// Non-feedback adjacency by node index, de-duplicated, in edge order.
	succ [][]int
	pred [][]int
	// indegree counts non-feedback incoming edges (a self edge counts).
	hasIn []bool

	groups  []toc.PreconditionGroup
	ceiling *toc.Ceiling
	// ownedByID marks outcomes that are listed in, or upstream of,
	// Ceiling.OutcomeIDs.
	ownedByID []bool

	reachOnce sync.Once
	reach     []bool // node can reach an impact outcome (impact included)
}

// Build constructs the graph. It returns an error wrapping
// ErrInvalidDocument if the document has error-severity validation issues.
// Build does not modify doc and keeps no reference to it.
func Build(doc *toc.Document) (*Graph, error) {
	if doc == nil {
		return nil, fmt.Errorf("%w: nil document", ErrInvalidDocument)
	}
	if issues := validate.Document(doc); validate.HasErrors(issues) {
		return nil, invalidError(issues)
	}
	// Rejected elements stay in the document for history but take no part
	// in the graph: neither they nor any edge touching them.
	g := &Graph{index: make(map[string]int, len(doc.Outcomes)+len(doc.Interventions))}
	for _, o := range doc.Outcomes {
		if !o.Rejected() {
			g.addNode(Node{ID: o.ID, Kind: toc.ElementOutcome, Level: o.Level})
		}
	}
	for _, in := range doc.Interventions {
		if !in.Rejected() {
			g.addNode(Node{ID: in.ID, Kind: toc.ElementIntervention})
		}
	}
	active := func(id string) bool { _, ok := g.index[id]; return ok }

	type pair struct{ from, to string }
	seen := make(map[pair]bool, len(doc.Links))
	for _, l := range doc.Links {
		if l.Rejected() || !active(l.From) || !active(l.To) {
			continue
		}
		g.edges = append(g.edges, Edge{LinkID: l.ID, From: l.From, To: l.To, Feedback: l.Feedback})
		seen[pair{l.From, l.To}] = true
	}
	for _, in := range doc.Interventions {
		if in.Rejected() {
			continue
		}
		for _, t := range in.Targets {
			p := pair{in.ID, t}
			if seen[p] || !active(t) {
				continue
			}
			seen[p] = true
			g.edges = append(g.edges, Edge{From: in.ID, To: t, Implicit: true})
		}
	}

	n := len(g.nodes)
	g.succ = make([][]int, n)
	g.pred = make([][]int, n)
	g.hasIn = make([]bool, n)
	for _, e := range g.edges {
		if e.Feedback {
			g.fb = append(g.fb, e)
			continue
		}
		f, t := g.index[e.From], g.index[e.To]
		g.hasIn[t] = true
		if !slices.Contains(g.succ[f], t) {
			g.succ[f] = append(g.succ[f], t)
			g.pred[t] = append(g.pred[t], f)
		}
	}

	activeGroup := map[string]bool{}
	for _, grp := range doc.Groups {
		if !grp.Rejected() && active(grp.Target) {
			activeGroup[grp.ID] = true
		}
	}
	g.groups = make([]toc.PreconditionGroup, 0, len(doc.Groups))
	for _, grp := range doc.Groups {
		if !activeGroup[grp.ID] {
			continue
		}
		members := make([]string, 0, len(grp.Members))
		for _, m := range grp.Members {
			if active(m) || activeGroup[m] {
				members = append(members, m)
			}
		}
		grp.Members = members
		g.groups = append(g.groups, grp)
	}
	if doc.Ceiling != nil {
		c := *doc.Ceiling
		c.OutcomeIDs = slices.DeleteFunc(slices.Clone(c.OutcomeIDs), func(id string) bool { return !active(id) })
		g.ceiling = &c
		g.ownedByID = g.closure(g.pred, c.OutcomeIDs)
	}
	return g, nil
}

func invalidError(issues []validate.Issue) error {
	var msgs []string
	total := 0
	for _, is := range issues {
		if is.Severity != validate.SeverityError {
			continue
		}
		total++
		if len(msgs) < maxIssuesInError {
			msgs = append(msgs, is.String())
		}
	}
	msg := strings.Join(msgs, "; ")
	if total > len(msgs) {
		msg += fmt.Sprintf("; and %d more", total-len(msgs))
	}
	return fmt.Errorf("%w: %s", ErrInvalidDocument, msg)
}

func (g *Graph) addNode(n Node) {
	g.index[n.ID] = len(g.nodes)
	g.nodes = append(g.nodes, n)
}

// Nodes returns every node: outcomes, then interventions, in document order.
func (g *Graph) Nodes() []Node { return slices.Clone(g.nodes) }

// Node returns the node with the given ID.
func (g *Graph) Node(id string) (Node, bool) {
	i, ok := g.index[id]
	if !ok {
		return Node{}, false
	}
	return g.nodes[i], true
}

// Edges returns every edge, feedback edges included: links in document
// order, then implicit intervention edges in intervention and target order.
// An implicit edge is omitted when an explicit link joins the same pair.
func (g *Graph) Edges() []Edge { return slices.Clone(g.edges) }

// FeedbackEdges returns the edges of feedback links, in edge order.
func (g *Graph) FeedbackEdges() []Edge { return slices.Clone(g.fb) }

// GroupsFor returns the precondition groups whose target is the given
// outcome, in document order. Groups never affect reachability.
func (g *Graph) GroupsFor(target string) []toc.PreconditionGroup {
	var out []toc.PreconditionGroup
	for _, grp := range g.groups {
		if grp.Target == target {
			grp.Members = slices.Clone(grp.Members)
			out = append(out, grp)
		}
	}
	return out
}

// Preconditions returns the direct sources of the node's non-feedback
// incoming edges, in edge order, without duplicates. Unknown ID returns nil.
func (g *Graph) Preconditions(id string) []string {
	i, ok := g.index[id]
	if !ok {
		return nil
	}
	return g.ids(g.pred[i])
}

func (g *Graph) ids(idx []int) []string {
	if len(idx) == 0 {
		return nil
	}
	out := make([]string, len(idx))
	for k, i := range idx {
		out[k] = g.nodes[i].ID
	}
	return out
}

// idsOf converts a membership mask to IDs in node order.
func (g *Graph) idsOf(mask []bool) []string {
	var out []string
	for i, in := range mask {
		if in {
			out = append(out, g.nodes[i].ID)
		}
	}
	return out
}

func (g *Graph) isImpact(i int) bool {
	n := g.nodes[i]
	return n.Kind == toc.ElementOutcome && n.Level == toc.LevelImpact
}

// closure marks every node reachable from the starts (inclusive) following
// adj. Unknown starts are ignored.
func (g *Graph) closure(adj [][]int, starts []string) []bool {
	mark := make([]bool, len(g.nodes))
	var stack []int
	for _, s := range starts {
		if i, ok := g.index[s]; ok && !mark[i] {
			mark[i] = true
			stack = append(stack, i)
		}
	}
	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, w := range adj[v] {
			if !mark[w] {
				mark[w] = true
				stack = append(stack, w)
			}
		}
	}
	return mark
}

func (g *Graph) reachable(adj [][]int, id string) []string {
	i, ok := g.index[id]
	if !ok {
		return nil
	}
	mark := g.closure(adj, []string{id})
	mark[i] = false
	return g.idsOf(mark)
}

// Upstream returns every node with a non-feedback path to id (its
// transitive preconditions), in node order. id itself is never included,
// even when it lies on a cycle. Unknown ID returns nil.
func (g *Graph) Upstream(id string) []string { return g.reachable(g.pred, id) }

// Downstream returns every node reachable from id by non-feedback edges
// (its transitive consequences), in node order. id itself is never
// included. Unknown ID returns nil.
func (g *Graph) Downstream(id string) []string { return g.reachable(g.succ, id) }

// canReachImpact lazily computes, once, which nodes have a non-feedback path
// to an impact outcome (impact outcomes included).
func (g *Graph) canReachImpact() []bool {
	g.reachOnce.Do(func() {
		var starts []string
		for i := range g.nodes {
			if g.isImpact(i) {
				starts = append(starts, g.nodes[i].ID)
			}
		}
		g.reach = g.closure(g.pred, starts)
	})
	return g.reach
}

// HasPathToImpact reports whether id is an impact outcome or has a
// non-feedback path to one. Unknown ID returns false. It runs in O(1) after
// a one-time O(V+E) computation.
func (g *Graph) HasPathToImpact(id string) bool {
	i, ok := g.index[id]
	return ok && g.canReachImpact()[i]
}

// PathsToImpact returns all simple paths (no repeated node) from id along
// non-feedback edges to an impact outcome, found by depth-first search in
// edge order. Each path starts at id and ends at an impact outcome; paths
// do not continue through an impact outcome. If id is itself an impact
// outcome the result is [[id]]. max <= 0 means DefaultMaxPaths; the search
// stops once max paths are found. Unknown ID, or no path, returns nil.
func (g *Graph) PathsToImpact(id string, max int) [][]string {
	start, ok := g.index[id]
	if !ok {
		return nil
	}
	if max <= 0 {
		max = DefaultMaxPaths
	}
	reach := g.canReachImpact()
	if !reach[start] {
		return nil
	}
	var (
		paths  [][]string
		path   []int
		onPath = make([]bool, len(g.nodes))
	)
	var dfs func(v int)
	dfs = func(v int) {
		path = append(path, v)
		onPath[v] = true
		if g.isImpact(v) {
			p := make([]string, len(path))
			for k, i := range path {
				p[k] = g.nodes[i].ID
			}
			paths = append(paths, p)
		} else {
			for _, w := range g.succ[v] {
				if len(paths) >= max {
					break
				}
				if reach[w] && !onPath[w] {
					dfs(w)
				}
			}
		}
		onPath[v] = false
		path = path[:len(path)-1]
	}
	dfs(start)
	return paths
}

// Cycles returns the non-feedback cycles: the strongly connected components
// that have more than one node, or one node with an edge to itself. Each
// component lists its IDs in node order; components are ordered by the
// position of their first node. No cycles returns nil.
func (g *Graph) Cycles() [][]string {
	comps := g.sccs()
	var out [][]string
	for _, c := range comps {
		out = append(out, g.ids(c))
	}
	return out
}

// sccs runs an iterative Tarjan and returns cyclic components, each sorted
// by node index and the list sorted by first node.
func (g *Graph) sccs() [][]int {
	n := len(g.nodes)
	const unvisited = -1
	idx := make([]int, n)
	low := make([]int, n)
	on := make([]bool, n)
	for i := range idx {
		idx[i] = unvisited
	}
	var (
		stack, comp []int
		counter     int
		out         [][]int
	)
	type frame struct{ v, next int }
	for root := range n {
		if idx[root] != unvisited {
			continue
		}
		call := []frame{{root, 0}}
		idx[root], low[root] = counter, counter
		counter++
		stack = append(stack, root)
		on[root] = true
		for len(call) > 0 {
			f := &call[len(call)-1]
			v := f.v
			if f.next < len(g.succ[v]) {
				w := g.succ[v][f.next]
				f.next++
				switch {
				case idx[w] == unvisited:
					idx[w], low[w] = counter, counter
					counter++
					stack = append(stack, w)
					on[w] = true
					call = append(call, frame{w, 0})
				case on[w]:
					low[v] = min(low[v], idx[w])
				}
				continue
			}
			call = call[:len(call)-1]
			if len(call) > 0 {
				p := call[len(call)-1].v
				low[p] = min(low[p], low[v])
			}
			if low[v] != idx[v] {
				continue
			}
			comp = comp[:0]
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				on[w] = false
				comp = append(comp, w)
				if w == v {
					break
				}
			}
			if len(comp) > 1 || slices.Contains(g.succ[v], v) {
				c := slices.Clone(comp)
				slices.Sort(c)
				out = append(out, c)
			}
		}
	}
	slices.SortFunc(out, func(a, b []int) int { return a[0] - b[0] })
	return out
}

// Levels maps each node to the length, in edges, of the LONGEST
// non-feedback path from it to an impact outcome. Impact outcomes are
// always 0 (a path stops at the first impact outcome). Every other node is
// omitted if it has no path to an impact outcome, or if it lies on a
// non-feedback cycle or can reach one anywhere downstream, since its
// longest distance is then unbounded.
func (g *Graph) Levels() map[string]int {
	n := len(g.nodes)
	lvl := make([]int, n)
	outdeg := make([]int, n)
	var queue []int
	for i := range n {
		lvl[i] = -1
		outdeg[i] = len(g.succ[i])
		if outdeg[i] == 0 {
			queue = append(queue, i)
		}
	}
	// Kahn from the sinks: a node is processed only when all of its
	// successors are, so unprocessed nodes are on or upstream of a cycle.
	out := make(map[string]int)
	for head := 0; head < len(queue); head++ {
		v := queue[head]
		if g.isImpact(v) {
			lvl[v] = 0
		} else {
			for _, w := range g.succ[v] {
				if lvl[w] >= 0 {
					lvl[v] = max(lvl[v], lvl[w]+1)
				}
			}
		}
		if lvl[v] >= 0 {
			out[g.nodes[v].ID] = lvl[v]
		}
		for _, p := range g.pred[v] {
			if outdeg[p]--; outdeg[p] == 0 {
				queue = append(queue, p)
			}
		}
	}
	for i := range n {
		if g.isImpact(i) {
			out[g.nodes[i].ID] = 0
		}
	}
	return out
}

// LevelOf returns the node's entry in Levels.
func (g *Graph) LevelOf(id string) (int, bool) {
	l, ok := g.Levels()[id]
	return l, ok
}

// Orphans returns the outcomes (not interventions) with no incoming
// non-feedback edge: no causal link into them and no intervention targeting
// them. Node order.
func (g *Graph) Orphans() []string {
	var out []string
	for i, n := range g.nodes {
		if n.Kind == toc.ElementOutcome && !g.hasIn[i] {
			out = append(out, n.ID)
		}
	}
	return out
}

// DeadEnds returns every node except impact outcomes that has no
// non-feedback path to an impact outcome, in node order. A document with no
// impact outcome makes every node a dead end; lint rule TOC-STR-001
// reports that case.
func (g *Graph) DeadEnds() []string {
	reach := g.canReachImpact()
	var out []string
	for i, n := range g.nodes {
		if !reach[i] {
			out = append(out, n.ID)
		}
	}
	return out
}

// BelowCeiling reports whether an outcome is within the ceiling of
// accountability (owned). Unknown IDs and interventions return false.
//
//   - No ceiling: every outcome is owned.
//   - Ceiling.Level set: an outcome is owned when its level's rank is at
//     most the ceiling's (impact ranks highest, so a long_term ceiling
//     owns long_term and everything below it).
//   - Ceiling.OutcomeIDs set: an outcome is owned when it is listed or is
//     upstream (a transitive non-feedback precondition) of a listed outcome.
//   - When both are set an outcome is owned if either rule owns it.
//   - A ceiling with neither set owns every outcome.
func (g *Graph) BelowCeiling(id string) bool {
	i, ok := g.index[id]
	if !ok || g.nodes[i].Kind != toc.ElementOutcome {
		return false
	}
	c := g.ceiling
	if c == nil || (c.Level == "" && len(c.OutcomeIDs) == 0) {
		return true
	}
	if c.Level != "" && g.nodes[i].Level.Rank() <= c.Level.Rank() {
		return true
	}
	return g.ownedByID[i]
}

// OwnedOutcomes returns the outcomes for which BelowCeiling is true, in
// node order.
func (g *Graph) OwnedOutcomes() []string {
	var out []string
	for _, n := range g.nodes {
		if g.BelowCeiling(n.ID) {
			out = append(out, n.ID)
		}
	}
	return out
}

// TopologicalOrder returns all node IDs so that every non-feedback edge goes
// from an earlier to a later node (Kahn's algorithm). Among nodes that are
// ready at the same time the one earlier in node order comes first. If the
// graph has a cycle it returns an error wrapping ErrCycle that names the
// first cycle.
func (g *Graph) TopologicalOrder() ([]string, error) {
	n := len(g.nodes)
	indeg := make([]int, n)
	for i := range n {
		indeg[i] = len(g.pred[i])
	}
	ready := &intHeap{}
	for i := range n {
		if indeg[i] == 0 {
			ready.push(i)
		}
	}
	order := make([]string, 0, n)
	for ready.len() > 0 {
		v := ready.pop()
		order = append(order, g.nodes[v].ID)
		for _, w := range g.succ[v] {
			if indeg[w]--; indeg[w] == 0 {
				ready.push(w)
			}
		}
	}
	if len(order) < n {
		cyc := g.Cycles()
		if len(cyc) == 0 {
			return nil, fmt.Errorf("%w: unresolved dependencies", ErrCycle)
		}
		return nil, fmt.Errorf("%w: %s", ErrCycle, strings.Join(cyc[0], " -> "))
	}
	return order, nil
}

// intHeap is a minimal binary min-heap of ints.
type intHeap struct{ a []int }

func (h *intHeap) len() int { return len(h.a) }

func (h *intHeap) push(x int) {
	h.a = append(h.a, x)
	i := len(h.a) - 1
	for i > 0 {
		p := (i - 1) / 2
		if h.a[p] <= h.a[i] {
			break
		}
		h.a[p], h.a[i] = h.a[i], h.a[p]
		i = p
	}
}

func (h *intHeap) pop() int {
	top := h.a[0]
	last := len(h.a) - 1
	h.a[0] = h.a[last]
	h.a = h.a[:last]
	i := 0
	for {
		l, r, s := 2*i+1, 2*i+2, i
		if l < last && h.a[l] < h.a[s] {
			s = l
		}
		if r < last && h.a[r] < h.a[s] {
			s = r
		}
		if s == i {
			break
		}
		h.a[s], h.a[i] = h.a[i], h.a[s]
		i = s
	}
	return top
}
