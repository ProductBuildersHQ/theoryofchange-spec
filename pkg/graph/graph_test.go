package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func loadFull(t *testing.T) *toc.Document {
	t.Helper()
	data, err := os.ReadFile("../toc/testdata/full.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc toc.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return &doc
}

func buildFull(t *testing.T) *Graph {
	t.Helper()
	g, err := Build(loadFull(t))
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return g
}

// link is a compact causal link: "from>to" with optional "~" prefix for feedback.
type spec struct {
	outcomes      map[string]toc.Level // id -> level, in order via order
	order         []string
	links         []string // "a>b" or "~a>b"
	interventions map[string][]string
	intOrder      []string
}

func newSpec() *spec {
	return &spec{outcomes: map[string]toc.Level{}, interventions: map[string][]string{}}
}

func (s *spec) outcome(id string, l toc.Level) *spec {
	s.outcomes[id] = l
	s.order = append(s.order, id)
	return s
}

func (s *spec) link(l ...string) *spec { s.links = append(s.links, l...); return s }

func (s *spec) intervene(id string, targets ...string) *spec {
	s.interventions[id] = targets
	s.intOrder = append(s.intOrder, id)
	return s
}

func (s *spec) doc() *toc.Document {
	d := toc.NewDocument("synthetic", "A synthetic problem.")
	for _, id := range s.order {
		d.Outcomes = append(d.Outcomes, toc.Outcome{ID: id, Title: id, Level: s.outcomes[id]})
	}
	for i, l := range s.links {
		fb := strings.HasPrefix(l, "~")
		l = strings.TrimPrefix(l, "~")
		from, to, _ := strings.Cut(l, ">")
		d.Links = append(d.Links, toc.CausalLink{ID: fmt.Sprintf("ln-%d", i), From: from, To: to, Feedback: fb})
	}
	for _, id := range s.intOrder {
		d.Interventions = append(d.Interventions, toc.Intervention{ID: id, Title: id, Targets: slices.Clone(s.interventions[id])})
	}
	return d
}

func (s *spec) build(t *testing.T) *Graph {
	t.Helper()
	g, err := Build(s.doc())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return g
}

func eq(t *testing.T, what string, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %v, want %v", what, got, want)
	}
}

func TestFullFixtureModel(t *testing.T) {
	g := buildFull(t)
	var ids []string
	for _, n := range g.Nodes() {
		ids = append(ids, n.ID)
	}
	eq(t, "nodes", ids, []string{
		"oc-safer-agents", "oc-enterprise-require", "oc-idp-support", "oc-platform-support",
		"oc-spec-stable", "oc-reference-impl", "int-interop", "int-reference",
	})
	n, ok := g.Node("int-interop")
	if !ok || n.Kind != toc.ElementIntervention || n.Level != "" {
		t.Fatalf("Node(int-interop) = %+v, %v", n, ok)
	}
	n, ok = g.Node("oc-safer-agents")
	if !ok || n.Kind != toc.ElementOutcome || n.Level != toc.LevelImpact {
		t.Fatalf("Node(oc-safer-agents) = %+v, %v", n, ok)
	}
	if _, ok := g.Node("nope"); ok {
		t.Fatal("unknown node found")
	}
	edges := g.Edges()
	if len(edges) != 9 {
		t.Fatalf("edges = %d, want 9: %+v", len(edges), edges)
	}
	eq(t, "implicit edges", edges[6:], []Edge{
		{From: "int-interop", To: "oc-spec-stable", Implicit: true},
		{From: "int-interop", To: "oc-idp-support", Implicit: true},
		{From: "int-reference", To: "oc-reference-impl", Implicit: true},
	})
	eq(t, "feedback", g.FeedbackEdges(), []Edge{{
		LinkID: "ln-require-platform-loop", From: "oc-enterprise-require", To: "oc-platform-support", Feedback: true,
	}})
	eq(t, "preconditions", g.Preconditions("oc-enterprise-require"), []string{"oc-idp-support", "oc-platform-support"})
	eq(t, "preconditions of platform excludes feedback", g.Preconditions("oc-platform-support"), []string{"oc-reference-impl"})
	eq(t, "preconditions unknown", g.Preconditions("nope"), []string(nil))
	eq(t, "preconditions none", g.Preconditions("int-interop"), []string(nil))

	groups := g.GroupsFor("oc-enterprise-require")
	if len(groups) != 1 || groups[0].ID != "grp-supply" {
		t.Fatalf("groups = %+v", groups)
	}
	groups[0].Members[0] = "mutated"
	if g.GroupsFor("oc-enterprise-require")[0].Members[0] != "oc-idp-support" {
		t.Fatal("GroupsFor leaked internal state")
	}
	eq(t, "no groups", g.GroupsFor("oc-safer-agents"), []toc.PreconditionGroup(nil))
}

func TestFullFixtureQueries(t *testing.T) {
	g := buildFull(t)
	tests := []struct {
		name string
		got  []string
		want []string
	}{
		{"upstream require", g.Upstream("oc-enterprise-require"),
			[]string{"oc-idp-support", "oc-platform-support", "oc-spec-stable", "oc-reference-impl", "int-interop", "int-reference"}},
		{"upstream spec", g.Upstream("oc-spec-stable"), []string{"int-interop"}},
		{"upstream interop", g.Upstream("int-interop"), nil},
		{"upstream unknown", g.Upstream("nope"), nil},
		{"downstream spec", g.Downstream("oc-spec-stable"), []string{"oc-safer-agents", "oc-enterprise-require", "oc-idp-support"}},
		{"downstream impact", g.Downstream("oc-safer-agents"), nil},
		{"downstream platform (feedback ignored)", g.Downstream("oc-platform-support"), []string{"oc-safer-agents", "oc-enterprise-require"}},
		{"downstream unknown", g.Downstream("nope"), nil},
		{"orphans", g.Orphans(), nil},
		{"dead ends", g.DeadEnds(), nil},
		{"owned", g.OwnedOutcomes(), []string{"oc-enterprise-require", "oc-idp-support", "oc-platform-support", "oc-spec-stable", "oc-reference-impl"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { eq(t, tt.name, tt.got, tt.want) })
	}
	eq(t, "cycles", g.Cycles(), [][]string(nil))
	eq(t, "paths spec", g.PathsToImpact("oc-spec-stable", 0),
		[][]string{{"oc-spec-stable", "oc-idp-support", "oc-enterprise-require", "oc-safer-agents"}})
	eq(t, "paths impact", g.PathsToImpact("oc-safer-agents", 0), [][]string{{"oc-safer-agents"}})
	eq(t, "paths interop", g.PathsToImpact("int-interop", 0), [][]string{
		{"int-interop", "oc-spec-stable", "oc-idp-support", "oc-enterprise-require", "oc-safer-agents"},
		{"int-interop", "oc-idp-support", "oc-enterprise-require", "oc-safer-agents"},
	})
	eq(t, "paths unknown", g.PathsToImpact("nope", 0), [][]string(nil))
	eq(t, "levels", g.Levels(), map[string]int{
		"oc-safer-agents": 0, "oc-enterprise-require": 1, "oc-idp-support": 2, "oc-platform-support": 2,
		"oc-spec-stable": 3, "oc-reference-impl": 3, "int-interop": 4, "int-reference": 4,
	})
	if l, ok := g.LevelOf("oc-reference-impl"); !ok || l != 3 {
		t.Fatalf("LevelOf = %d, %v", l, ok)
	}
	if _, ok := g.LevelOf("nope"); ok {
		t.Fatal("LevelOf unknown ok")
	}
	if !g.HasPathToImpact("int-reference") || g.HasPathToImpact("nope") {
		t.Fatal("HasPathToImpact mismatch")
	}
	for id, want := range map[string]bool{
		"oc-safer-agents": false, "oc-enterprise-require": true, "oc-idp-support": true,
		"oc-spec-stable": true, "int-interop": false, "nope": false,
	} {
		if got := g.BelowCeiling(id); got != want {
			t.Errorf("BelowCeiling(%s) = %v, want %v", id, got, want)
		}
	}
	order, err := g.TopologicalOrder()
	if err != nil {
		t.Fatal(err)
	}
	checkTopo(t, g, order)
}

func checkTopo(t *testing.T, g *Graph, order []string) {
	t.Helper()
	if len(order) != len(g.Nodes()) {
		t.Fatalf("order has %d ids, want %d", len(order), len(g.Nodes()))
	}
	pos := map[string]int{}
	for i, id := range order {
		pos[id] = i
	}
	for _, e := range g.Edges() {
		if e.Feedback {
			continue
		}
		if pos[e.From] >= pos[e.To] {
			t.Fatalf("edge %s->%s goes backward", e.From, e.To)
		}
	}
}

func TestCeilingVariants(t *testing.T) {
	// a(output) > b(short) > c(inter) > d(long) > e(impact); x(short) side branch to nothing.
	mk := func(c *toc.Ceiling) *Graph {
		s := newSpec().outcome("e", toc.LevelImpact).outcome("d", toc.LevelLongTerm).
			outcome("c", toc.LevelIntermediate).outcome("b", toc.LevelShortTerm).
			outcome("a", toc.LevelOutput).outcome("x", toc.LevelShortTerm).
			link("a>b", "b>c", "c>d", "d>e")
		doc := s.doc()
		doc.Ceiling = c
		g, err := Build(doc)
		if err != nil {
			t.Fatal(err)
		}
		return g
	}
	eq(t, "nil ceiling", mk(nil).OwnedOutcomes(), []string{"e", "d", "c", "b", "a", "x"})
	eq(t, "level", mk(&toc.Ceiling{Level: toc.LevelIntermediate}).OwnedOutcomes(), []string{"c", "b", "a", "x"})
	eq(t, "outcome ids", mk(&toc.Ceiling{OutcomeIDs: []string{"c"}}).OwnedOutcomes(), []string{"c", "b", "a"})
	eq(t, "both", mk(&toc.Ceiling{Level: toc.LevelShortTerm, OutcomeIDs: []string{"c"}}).OwnedOutcomes(), []string{"c", "b", "a", "x"})
	eq(t, "ids with nothing else", mk(&toc.Ceiling{OutcomeIDs: []string{"a"}}).OwnedOutcomes(), []string{"a"})
}

func TestCycleWithoutFeedback(t *testing.T) {
	s := newSpec().outcome("i", toc.LevelImpact).outcome("a", toc.LevelShortTerm).
		outcome("b", toc.LevelIntermediate).outcome("c", toc.LevelLongTerm).outcome("z", toc.LevelOutput).
		link("z>a", "a>b", "b>c", "c>a", "c>i")
	g := s.build(t)
	eq(t, "cycles", g.Cycles(), [][]string{{"a", "b", "c"}})
	// Upstream of a node on a cycle never contains itself.
	eq(t, "upstream a", g.Upstream("a"), []string{"b", "c", "z"})
	eq(t, "downstream a", g.Downstream("a"), []string{"i", "b", "c"})
	eq(t, "levels omit cycle and upstream", g.Levels(), map[string]int{"i": 0})
	eq(t, "paths a", g.PathsToImpact("a", 0), [][]string{{"a", "b", "c", "i"}})
	_, err := g.TopologicalOrder()
	if !errors.Is(err, ErrCycle) || !strings.Contains(err.Error(), "a -> b -> c") {
		t.Fatalf("TopologicalOrder err = %v", err)
	}

	s.links[3] = "~c>a"
	g = s.build(t)
	eq(t, "no cycles with feedback", g.Cycles(), [][]string(nil))
	eq(t, "feedback edge", len(g.FeedbackEdges()), 1)
	eq(t, "levels", g.Levels(), map[string]int{"i": 0, "c": 1, "b": 2, "a": 3, "z": 4})
	order, err := g.TopologicalOrder()
	if err != nil {
		t.Fatal(err)
	}
	checkTopo(t, g, order)
}

func TestCycleOrdering(t *testing.T) {
	s := newSpec().outcome("i", toc.LevelImpact).outcome("p", toc.LevelShortTerm).
		outcome("q", toc.LevelShortTerm).outcome("r", toc.LevelShortTerm).outcome("s", toc.LevelShortTerm).
		link("s>r", "r>s", "q>p", "p>q", "p>i")
	eq(t, "cycles", s.build(t).Cycles(), [][]string{{"p", "q"}, {"r", "s"}})
}

func TestDisconnectedNode(t *testing.T) {
	s := newSpec().outcome("i", toc.LevelImpact).outcome("a", toc.LevelShortTerm).outcome("lone", toc.LevelOutput).
		link("a>i").intervene("int-x", "a")
	g := s.build(t)
	eq(t, "orphans", g.Orphans(), []string{"lone"})
	eq(t, "dead ends", g.DeadEnds(), []string{"lone"})
	if g.HasPathToImpact("lone") {
		t.Fatal("lone reaches impact")
	}
	eq(t, "paths lone", g.PathsToImpact("lone", 0), [][]string(nil))
	if _, ok := g.LevelOf("lone"); ok {
		t.Fatal("lone has a level")
	}
	// Orphans exclude interventions; "a" is not an orphan thanks to the intervention.
	eq(t, "a not orphan", slices.Contains(g.Orphans(), "a"), false)
}

func TestNoImpactMakesEverythingDeadEnd(t *testing.T) {
	g := newSpec().outcome("a", toc.LevelShortTerm).outcome("b", toc.LevelLongTerm).link("a>b").build(t)
	eq(t, "dead ends", g.DeadEnds(), []string{"a", "b"})
	eq(t, "levels", g.Levels(), map[string]int{})
}

func TestImplicitEdgeDedup(t *testing.T) {
	s := newSpec().outcome("i", toc.LevelImpact).outcome("a", toc.LevelShortTerm).
		link("int-x>a", "a>i").intervene("int-x", "a", "a", "i")
	g := s.build(t)
	edges := g.Edges()
	// ln-0 explicit, ln-1 explicit, then implicit int-x->i only.
	eq(t, "edges", edges, []Edge{
		{LinkID: "ln-0", From: "int-x", To: "a"},
		{LinkID: "ln-1", From: "a", To: "i"},
		{From: "int-x", To: "i", Implicit: true},
	})
}

func TestParallelLinksDeduplicated(t *testing.T) {
	g := newSpec().outcome("i", toc.LevelImpact).outcome("a", toc.LevelShortTerm).link("a>i", "a>i").build(t)
	eq(t, "paths", g.PathsToImpact("a", 0), [][]string{{"a", "i"}})
	eq(t, "preconditions", g.Preconditions("i"), []string{"a"})
}

func TestMaxPathsTruncation(t *testing.T) {
	// Layered graph: s -> {m0..m3} -> i gives 4 paths.
	s := newSpec().outcome("i", toc.LevelImpact).outcome("s", toc.LevelOutput)
	for k := range 4 {
		m := fmt.Sprintf("m%d", k)
		s.outcome(m, toc.LevelIntermediate).link("s>"+m, m+">i")
	}
	g := s.build(t)
	all := g.PathsToImpact("s", 0)
	if len(all) != 4 {
		t.Fatalf("paths = %d", len(all))
	}
	two := g.PathsToImpact("s", 2)
	eq(t, "truncated", two, all[:2])
	eq(t, "one", len(g.PathsToImpact("s", 1)), 1)
}

func TestDiamondPathCounting(t *testing.T) {
	// Chain of 5 diamonds: 2^5 = 32 paths.
	s := newSpec().outcome("i", toc.LevelImpact)
	prev := "i"
	var links []string
	for k := range 5 {
		a, b, top := fmt.Sprintf("a%d", k), fmt.Sprintf("b%d", k), fmt.Sprintf("t%d", k)
		s.outcome(a, toc.LevelIntermediate).outcome(b, toc.LevelIntermediate).outcome(top, toc.LevelIntermediate)
		links = append(links, a+">"+prev, b+">"+prev, top+">"+a, top+">"+b)
		prev = top
	}
	g := s.link(links...).build(t)
	paths := g.PathsToImpact("t4", 0)
	if len(paths) != 32 {
		t.Fatalf("paths = %d, want 32", len(paths))
	}
	if len(g.PathsToImpact("t4", 10)) != 10 {
		t.Fatal("max not honoured")
	}
	if l, _ := g.LevelOf("t4"); l != 10 {
		t.Fatalf("level = %d", l)
	}
}

func TestPathsDoNotContinuePastImpact(t *testing.T) {
	g := newSpec().outcome("i1", toc.LevelImpact).outcome("i2", toc.LevelImpact).outcome("a", toc.LevelShortTerm).
		link("a>i1", "i1>i2").build(t)
	eq(t, "paths", g.PathsToImpact("a", 0), [][]string{{"a", "i1"}})
	eq(t, "levels", g.Levels(), map[string]int{"i1": 0, "i2": 0, "a": 1})
}

func TestBuildErrors(t *testing.T) {
	if _, err := Build(nil); !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("nil doc err = %v", err)
	}
	selfLoop := newSpec().outcome("a", toc.LevelShortTerm).link("a>a").doc()
	_, err := Build(selfLoop)
	if !errors.Is(err, ErrInvalidDocument) || !strings.Contains(err.Error(), "VAL-LINK-001") {
		t.Fatalf("self link err = %v", err)
	}
	dangling := newSpec().outcome("a", toc.LevelShortTerm).link("a>ghost").doc()
	if _, err := Build(dangling); !errors.Is(err, ErrInvalidDocument) || !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("dangling err = %v", err)
	}
	// Many issues: message is truncated.
	bad := newSpec().outcome("a", toc.LevelShortTerm).link("a>g1", "a>g2", "a>g3", "a>g4", "a>g5").doc()
	_, err = Build(bad)
	if !errors.Is(err, ErrInvalidDocument) || !strings.Contains(err.Error(), "more") {
		t.Fatalf("many err = %v", err)
	}
}

func TestBuildDoesNotMutateOrAlias(t *testing.T) {
	doc := loadFull(t)
	before := snapshot(t, doc)
	g, err := Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	if after := snapshot(t, doc); after != before {
		t.Fatal("Build mutated the document")
	}
	want := g.OwnedOutcomes()
	wantGroup := g.GroupsFor("oc-enterprise-require")[0].Members[0]
	// Mutate the document after the build; results must not change.
	doc.Ceiling.OutcomeIDs = append(doc.Ceiling.OutcomeIDs, "oc-safer-agents")
	doc.Groups[0].Members[0] = "changed"
	doc.Links[0].To = "oc-safer-agents"
	doc.Outcomes[0].Level = toc.LevelOutput
	eq(t, "owned after mutation", g.OwnedOutcomes(), want)
	eq(t, "group after mutation", g.GroupsFor("oc-enterprise-require")[0].Members[0], wantGroup)
	eq(t, "edge after mutation", g.Edges()[0].To, "oc-idp-support")
}

func snapshot(t *testing.T, doc *toc.Document) string {
	t.Helper()
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestAccessorsReturnCopies(t *testing.T) {
	g := buildFull(t)
	g.Nodes()[0].ID = "x"
	g.Edges()[0].From = "x"
	if g.Nodes()[0].ID == "x" || g.Edges()[0].From == "x" {
		t.Fatal("accessors leaked internal slices")
	}
}

func TestConcurrentLazyQueries(t *testing.T) {
	g := buildFull(t)
	var wg sync.WaitGroup
	errs := make(chan string, 16)
	for range 16 {
		wg.Go(func() {
			if !g.HasPathToImpact("oc-spec-stable") {
				errs <- "HasPathToImpact"
			}
			if len(g.PathsToImpact("int-interop", 0)) != 2 {
				errs <- "PathsToImpact"
			}
			if len(g.DeadEnds()) != 0 {
				errs <- "DeadEnds"
			}
			if g.BelowCeiling("oc-safer-agents") {
				errs <- "BelowCeiling"
			}
		})
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}

func TestDeepGraphNoStackIssues(t *testing.T) {
	// A 10,000-node chain exercises Tarjan and DFS depth.
	const n = 10000
	s := newSpec().outcome("n0", toc.LevelImpact)
	var links []string
	for k := 1; k < n; k++ {
		s.outcome(fmt.Sprintf("n%d", k), toc.LevelIntermediate)
		links = append(links, fmt.Sprintf("n%d>n%d", k, k-1))
	}
	g := s.link(links...).build(t)
	if g.Cycles() != nil {
		t.Fatal("unexpected cycles")
	}
	if l, _ := g.LevelOf(fmt.Sprintf("n%d", n-1)); l != n-1 {
		t.Fatalf("level = %d", l)
	}
	if p := g.PathsToImpact(fmt.Sprintf("n%d", n-1), 0); len(p) != 1 || len(p[0]) != n {
		t.Fatal("deep path wrong")
	}
	// Close the chain into one big cycle.
	s.links = append(s.links, fmt.Sprintf("n0>n%d", n-1))
	s.outcomes["n0"] = toc.LevelOutput
	if c := s.build(t).Cycles(); len(c) != 1 || len(c[0]) != n {
		t.Fatal("big cycle not found")
	}
}
