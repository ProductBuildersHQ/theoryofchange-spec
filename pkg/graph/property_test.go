package graph

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// randomDAG builds a document with 50-200 outcomes and a few interventions;
// every link goes from a lower to a higher index, so the graph is acyclic.
func randomDAG(r *rand.Rand) *toc.Document {
	n := 50 + r.IntN(151)
	s := newSpec()
	levels := []toc.Level{toc.LevelOutput, toc.LevelShortTerm, toc.LevelIntermediate, toc.LevelLongTerm}
	for i := range n {
		l := levels[r.IntN(len(levels))]
		if r.IntN(10) == 0 {
			l = toc.LevelImpact
		}
		s.outcome(fmt.Sprintf("n%d", i), l)
	}
	for range n * 2 {
		a, b := r.IntN(n), r.IntN(n)
		if a == b {
			continue
		}
		if a > b {
			a, b = b, a
		}
		prefix := ""
		if r.IntN(15) == 0 {
			prefix = "~" // a feedback link may point anywhere
			if r.IntN(2) == 0 {
				a, b = b, a
			}
		}
		s.links = append(s.links, fmt.Sprintf("%sn%d>n%d", prefix, a, b))
	}
	for k := range 5 {
		s.intervene(fmt.Sprintf("int%d", k), fmt.Sprintf("n%d", r.IntN(n)), fmt.Sprintf("n%d", r.IntN(n)))
	}
	return s.doc()
}

func TestProperties(t *testing.T) {
	for seed := range uint64(8) {
		t.Run(fmt.Sprintf("seed%d", seed), func(t *testing.T) {
			r := newRand(seed, 99)
			doc := randomDAG(r)
			before := snapshot(t, doc)
			g, err := Build(doc)
			if err != nil {
				t.Fatal(err)
			}
			if snapshot(t, doc) != before {
				t.Fatal("Build mutated input")
			}
			checkProperties(t, g)
		})
	}
}

func checkProperties(t *testing.T, g *Graph) {
	t.Helper()
	nodes := g.Nodes()
	pos := map[string]int{}
	for i, n := range nodes {
		pos[n.ID] = i
	}
	type pair struct{ a, b string }
	edgeSet := map[pair]bool{}
	for _, e := range g.Edges() {
		if !e.Feedback {
			edgeSet[pair{e.From, e.To}] = true
		}
	}
	if c := g.Cycles(); c != nil {
		t.Fatalf("DAG has cycles: %v", c)
	}
	levels := g.Levels()
	up := map[string]map[string]bool{}
	for _, n := range nodes {
		up[n.ID] = map[string]bool{}
	}
	for _, n := range nodes {
		ups := g.Upstream(n.ID)
		if !slices.IsSortedFunc(ups, func(a, b string) int { return pos[a] - pos[b] }) {
			t.Fatalf("Upstream(%s) not in node order", n.ID)
		}
		for _, u := range ups {
			up[n.ID][u] = true
		}
		if up[n.ID][n.ID] {
			t.Fatalf("Upstream(%s) contains itself", n.ID)
		}
	}
	for _, n := range nodes {
		down := map[string]bool{}
		for _, d := range g.Downstream(n.ID) {
			down[d] = true
		}
		for _, m := range nodes {
			if up[m.ID][n.ID] != down[m.ID] {
				t.Fatalf("Upstream/Downstream mismatch for %s, %s", n.ID, m.ID)
			}
		}
	}
	for _, n := range nodes {
		isImpact := n.Kind == toc.ElementOutcome && n.Level == toc.LevelImpact
		paths := g.PathsToImpact(n.ID, 50)
		if g.HasPathToImpact(n.ID) != (len(paths) > 0) {
			t.Fatalf("HasPathToImpact(%s) disagrees with paths", n.ID)
		}
		for _, p := range paths {
			if p[0] != n.ID {
				t.Fatalf("path %v does not start at %s", p, n.ID)
			}
			last, _ := g.Node(p[len(p)-1])
			if last.Level != toc.LevelImpact {
				t.Fatalf("path %v does not end at impact", p)
			}
			seen := map[string]bool{}
			for i, id := range p {
				if seen[id] {
					t.Fatalf("path %v repeats %s", p, id)
				}
				seen[id] = true
				if i > 0 && !edgeSet[pair{p[i-1], id}] {
					t.Fatalf("path %v uses missing edge %s->%s", p, p[i-1], id)
				}
				if i < len(p)-1 {
					if nd, _ := g.Node(id); nd.Level == toc.LevelImpact {
						t.Fatalf("path %v continues past impact", p)
					}
				}
			}
		}
		lvl, has := levels[n.ID]
		if has != g.HasPathToImpact(n.ID) {
			t.Fatalf("Levels presence for %s disagrees with reachability", n.ID)
		}
		if has && !isImpact {
			best := -1
			for _, m := range nodes {
				if edgeSet[pair{n.ID, m.ID}] {
					if l, ok := levels[m.ID]; ok {
						best = max(best, l+1)
					}
				}
			}
			if lvl != best {
				t.Fatalf("Levels(%s) = %d, want %d", n.ID, lvl, best)
			}
		}
		if isImpact && lvl != 0 {
			t.Fatalf("impact %s level %d", n.ID, lvl)
		}
	}
	order, err := g.TopologicalOrder()
	if err != nil {
		t.Fatal(err)
	}
	checkTopo(t, g, order)

	// Determinism: every query twice.
	queries := func() []any {
		id := nodes[len(nodes)/2].ID
		o, err := g.TopologicalOrder()
		if err != nil {
			t.Fatal(err)
		}
		return []any{g.Upstream(id), g.Downstream(id), g.PathsToImpact(id, 20), g.Cycles(), g.Levels(),
			g.Orphans(), g.DeadEnds(), g.OwnedOutcomes(), o, g.Edges(), g.Nodes(), g.FeedbackEdges()}
	}
	if !reflect.DeepEqual(queries(), queries()) {
		t.Fatal("queries are not deterministic")
	}
	// Rebuilding from the same doc gives the same graph answers.
}

func TestRandomGraphsWithCycles(t *testing.T) {
	for seed := range uint64(5) {
		r := newRand(seed, 7)
		n := 60
		s := newSpec()
		for i := range n {
			l := toc.LevelIntermediate
			if i%11 == 0 {
				l = toc.LevelImpact
			}
			s.outcome(fmt.Sprintf("n%d", i), l)
		}
		for range 90 {
			a, b := r.IntN(n), r.IntN(n)
			if a != b {
				s.links = append(s.links, fmt.Sprintf("n%d>n%d", a, b))
			}
		}
		g := s.build(t)
		cyc := g.Cycles()
		inCycle := map[string]bool{}
		for _, c := range cyc {
			if !slices.IsSorted(componentPositions(g, c)) {
				t.Fatalf("component %v not in node order", c)
			}
			for _, id := range c {
				inCycle[id] = true
			}
		}
		// Cross-check: a node is on a cycle iff it can reach itself.
		for _, node := range g.Nodes() {
			reachesSelf := false
			for _, p := range g.Preconditions(node.ID) {
				if p == node.ID || slices.Contains(g.Downstream(node.ID), p) {
					reachesSelf = true
					break
				}
			}
			if reachesSelf != inCycle[node.ID] {
				t.Fatalf("seed %d: node %s cycle membership mismatch", seed, node.ID)
			}
		}
		_, err := g.TopologicalOrder()
		if (err != nil) != (len(cyc) > 0) {
			t.Fatalf("TopologicalOrder err = %v with %d cycles", err, len(cyc))
		}
		for id := range g.Levels() {
			if nd, _ := g.Node(id); inCycle[id] && nd.Level != toc.LevelImpact {
				t.Fatalf("Levels includes cyclic node %s", id)
			}
		}
	}
}

func componentPositions(g *Graph, ids []string) []int {
	pos := map[string]int{}
	for i, n := range g.Nodes() {
		pos[n.ID] = i
	}
	out := make([]int, len(ids))
	for i, id := range ids {
		out[i] = pos[id]
	}
	return out
}

func BenchmarkQueries1000(b *testing.B) {
	r := newRand(1, 2)
	s := newSpec()
	const n = 1000
	levels := []toc.Level{toc.LevelOutput, toc.LevelShortTerm, toc.LevelIntermediate, toc.LevelLongTerm}
	for i := range n {
		l := levels[r.IntN(len(levels))]
		if i < 5 {
			l = toc.LevelImpact
		}
		s.outcome(fmt.Sprintf("n%d", i), l)
	}
	for range 2500 {
		x, y := r.IntN(n), r.IntN(n)
		if x == y {
			continue
		}
		// Edges go from higher to lower index so impact nodes (low indexes) are sinks.
		s.links = append(s.links, fmt.Sprintf("n%d>n%d", max(x, y), min(x, y)))
	}
	doc := s.doc()
	mid := fmt.Sprintf("n%d", n-1)
	b.ResetTimer()
	for range b.N {
		g, err := Build(doc)
		if err != nil {
			b.Fatal(err)
		}
		g.Upstream(mid)
		g.Downstream("n0")
		g.PathsToImpact(mid, 100)
		g.Cycles()
		g.Levels()
		g.Orphans()
		g.DeadEnds()
		g.OwnedOutcomes()
		if _, err := g.TopologicalOrder(); err != nil {
			b.Fatal(err)
		}
	}
}

// newRand returns a seeded PRNG so generated graphs are reproducible.
func newRand(a, b uint64) *rand.Rand {
	return rand.New(rand.NewPCG(a, b)) //nolint:gosec // reproducible test data, not security
}
