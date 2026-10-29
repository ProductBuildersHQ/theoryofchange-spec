package service

import (
	"fmt"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// GraphQuery names a graph query.
type GraphQuery string

// Graph queries.
const (
	QuerySummary    GraphQuery = "summary"
	QueryUpstream   GraphQuery = "upstream"
	QueryDownstream GraphQuery = "downstream"
	QueryPaths      GraphQuery = "paths"
	QueryCycles     GraphQuery = "cycles"
	QueryLevels     GraphQuery = "levels"
	QueryOrphans    GraphQuery = "orphans"
	QueryDeadEnds   GraphQuery = "dead-ends"
	QueryGraph      GraphQuery = "graph"
)

// AllGraphQueries returns every query name.
func AllGraphQueries() []GraphQuery {
	return []GraphQuery{
		QuerySummary, QueryUpstream, QueryDownstream, QueryPaths, QueryCycles,
		QueryLevels, QueryOrphans, QueryDeadEnds, QueryGraph,
	}
}

// Valid reports whether q is a known query.
func (q GraphQuery) Valid() bool {
	for _, k := range AllGraphQueries() {
		if q == k {
			return true
		}
	}
	return false
}

// NeedsID reports whether the query requires a starting element.
func (q GraphQuery) NeedsID() bool {
	return q == QueryUpstream || q == QueryDownstream || q == QueryPaths
}

// Summary counts the parts of a graph and lists structural problems.
type Summary struct {
	Nodes          int        `json:"nodes"`
	Edges          int        `json:"edges"`
	FeedbackEdges  int        `json:"feedbackEdges"`
	ImpactOutcomes []string   `json:"impactOutcomes"`
	Cycles         [][]string `json:"cycles"`
	Orphans        []string   `json:"orphans"`
	DeadEnds       []string   `json:"deadEnds"`
}

// Export is the whole graph as data.
type Export struct {
	Nodes []graph.Node `json:"nodes"`
	Edges []graph.Edge `json:"edges"`
}

// LevelEntry is one node's longest distance to an impact outcome.
type LevelEntry struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
}

// BuildGraph builds the causal graph of a document.
func BuildGraph(doc *toc.Document) (*graph.Graph, error) {
	return graph.Build(doc)
}

// RunGraphQuery runs a graph query. Results are JSON-serializable and
// deterministic: ID lists in node order, levels sorted by node order.
func RunGraphQuery(g *graph.Graph, q GraphQuery, id string, maxPaths int) (any, error) {
	if !q.Valid() {
		return nil, fmt.Errorf("unknown graph query %q (want one of %s)", string(q), queryNames())
	}
	if q.NeedsID() {
		if id == "" {
			return nil, fmt.Errorf("query %q needs an element id", string(q))
		}
		if _, ok := g.Node(id); !ok {
			return nil, fmt.Errorf("unknown outcome or intervention %q", id)
		}
	}
	switch q {
	case QuerySummary:
		return summarize(g), nil
	case QueryUpstream:
		return nonNil(g.Upstream(id)), nil
	case QueryDownstream:
		return nonNil(g.Downstream(id)), nil
	case QueryPaths:
		paths := g.PathsToImpact(id, maxPaths)
		if paths == nil {
			paths = [][]string{}
		}
		return paths, nil
	case QueryCycles:
		cycles := g.Cycles()
		if cycles == nil {
			cycles = [][]string{}
		}
		return cycles, nil
	case QueryLevels:
		levels := g.Levels()
		out := []LevelEntry{}
		for _, n := range g.Nodes() {
			if l, ok := levels[n.ID]; ok {
				out = append(out, LevelEntry{ID: n.ID, Level: l})
			}
		}
		return out, nil
	case QueryOrphans:
		return nonNil(g.Orphans()), nil
	case QueryDeadEnds:
		return nonNil(g.DeadEnds()), nil
	default:
		return Export{Nodes: g.Nodes(), Edges: g.Edges()}, nil
	}
}

func queryNames() string {
	names := make([]string, 0, len(AllGraphQueries()))
	for _, q := range AllGraphQueries() {
		names = append(names, string(q))
	}
	return strings.Join(names, ", ")
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func summarize(g *graph.Graph) Summary {
	s := Summary{
		Nodes:          len(g.Nodes()),
		Edges:          len(g.Edges()),
		FeedbackEdges:  len(g.FeedbackEdges()),
		ImpactOutcomes: []string{},
		Cycles:         g.Cycles(),
		Orphans:        nonNil(g.Orphans()),
		DeadEnds:       nonNil(g.DeadEnds()),
	}
	if s.Cycles == nil {
		s.Cycles = [][]string{}
	}
	for _, n := range g.Nodes() {
		if n.Kind == toc.ElementOutcome && n.Level == toc.LevelImpact {
			s.ImpactOutcomes = append(s.ImpactOutcomes, n.ID)
		}
	}
	return s
}

// DOT renders the graph in Graphviz DOT. Feedback links are dashed and
// implicit intervention edges are dotted. Output is deterministic.
func DOT(g *graph.Graph) string {
	var b strings.Builder
	b.WriteString("digraph theoryofchange {\n  rankdir=BT;\n")
	for _, n := range g.Nodes() {
		shape, label := "box", n.ID
		if n.Kind == toc.ElementIntervention {
			shape = "ellipse"
		} else if n.Level != "" {
			label = fmt.Sprintf("%s\n(%s)", n.ID, n.Level)
		}
		fmt.Fprintf(&b, "  %s [shape=%s, label=%s];\n", dotQuote(n.ID), shape, dotQuote(label))
	}
	for _, e := range g.Edges() {
		style := ""
		switch {
		case e.Feedback:
			style = " [style=dashed]"
		case e.Implicit:
			style = " [style=dotted]"
		}
		fmt.Fprintf(&b, "  %s -> %s%s;\n", dotQuote(e.From), dotQuote(e.To), style)
	}
	b.WriteString("}\n")
	return b.String()
}

// dotQuote quotes a string for DOT: backslashes and quotes are escaped, and
// a newline becomes the DOT line-break escape.
func dotQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return `"` + s + `"`
}
