// Package logicmodel derives the layered graph of a logic model from a report
// view: columns for the stages from inputs to impact, a box for each item in a
// stage, and an arrow wherever the theory says one thing leads to another.
//
// The graph holds no layout and no styling. It is the one description of the
// diagram that every output is made from.
package logicmodel

import (
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// StageID names a column of the logic model.
type StageID string

// The stages, from the start of the chain to its end.
const (
	StageInputs     StageID = "inputs"
	StageActivities StageID = "activities"
	StageOutputs    StageID = "outputs"
	StageShortTerm  StageID = "short_term"
	StageMediumTerm StageID = "medium_term"
	StageLongTerm   StageID = "long_term"
	StageImpact     StageID = "impact"
)

// Stages returns every stage, from the start of the chain to its end.
func Stages() []StageID {
	return []StageID{StageInputs, StageActivities, StageOutputs, StageShortTerm, StageMediumTerm, StageLongTerm, StageImpact}
}

// Name is the heading of the stage.
func (s StageID) Name() string {
	switch s {
	case StageInputs:
		return "Inputs"
	case StageActivities:
		return "Activities"
	case StageOutputs:
		return "Outputs"
	case StageShortTerm:
		return "Short-term outcomes"
	case StageMediumTerm:
		return "Medium-term outcomes"
	case StageLongTerm:
		return "Long-term outcomes"
	case StageImpact:
		return "Impact"
	}
	return string(s)
}

// Node is one item in a stage.
type Node struct {
	// ID is the element's ID, or a made-up one, starting with an underscore, for
	// an input or an output that is only text in the theory.
	ID string
	// Label is the text shown, with a mark when the element is inferred or a
	// hypothesis.
	Label string
}

// Edge says that one node leads to another.
type Edge struct {
	From, To string
}

// StageNodes is the nodes of one stage, in the order they are drawn.
type StageNodes struct {
	Stage StageID
	Nodes []Node
}

// Graph is the layered graph of a logic model.
type Graph struct {
	// Title is the pathway's title, or the theory's.
	Title string
	// Stages holds the stages that have nodes, in order.
	Stages []StageNodes
	// Edges are in a fixed order, and every edge joins two nodes of the graph.
	Edges []Edge
}

// Build derives the graph from a view.
//
// The inputs are the resources interventions state. Each activity is an
// intervention, and leads to the outputs it lists, which lead on to the
// outcomes it targets; with no outputs listed it leads to its targets
// directly. Outcomes are placed by level, with output-level outcomes in the
// outputs stage, and every causal link between two of them is an edge.
// Within a stage, nodes are ordered to keep arrows from crossing where they
// can, and otherwise stay in the order of the theory.
func Build(v *report.View) *Graph {
	b := &builder{stage: map[string]StageID{}, nodes: map[StageID][]Node{}, seen: map[Edge]bool{}}

	inputIDs := map[string]string{}
	outputIDs := map[string]string{}
	for _, o := range v.Outcomes {
		b.add(stageOf(o.Level), Node{ID: o.ID, Label: label(o.Title, o.Basis)})
	}
	for _, in := range v.Interventions {
		b.add(StageActivities, Node{ID: in.ID, Label: label(in.Title, in.Basis)})
	}
	for _, in := range v.Interventions {
		if in.Resources != "" {
			id, ok := inputIDs[in.Resources]
			if !ok {
				id = "_input" + strconv.Itoa(len(inputIDs)+1)
				inputIDs[in.Resources] = id
				b.add(StageInputs, Node{ID: id, Label: oneLine(in.Resources)})
			}
			b.edge(id, in.ID)
		}
		var produced []string
		for _, text := range in.Outputs {
			id, ok := outputIDs[text]
			if !ok {
				id = "_output" + strconv.Itoa(len(outputIDs)+1)
				outputIDs[text] = id
				b.add(StageOutputs, Node{ID: id, Label: oneLine(text)})
			}
			b.edge(in.ID, id)
			produced = append(produced, id)
		}
		for _, t := range in.Targets {
			if st, ok := b.stage[t]; ok && st == StageOutputs {
				b.edge(in.ID, t) // the target is itself an output
				continue
			}
			if len(produced) == 0 {
				b.edge(in.ID, t)
				continue
			}
			for _, id := range produced {
				b.edge(id, t)
			}
		}
	}
	for _, l := range v.Links {
		b.edge(l.From, l.To)
	}

	g := &Graph{Title: report.PathwayTitle(v)}
	order := b.order()
	for _, st := range Stages() {
		if nodes := order[st]; len(nodes) > 0 {
			g.Stages = append(g.Stages, StageNodes{Stage: st, Nodes: nodes})
		}
	}
	g.Edges = b.edges
	return g
}

// stageOf places an outcome by its level.
func stageOf(l toc.Level) StageID {
	switch l {
	case toc.LevelOutput:
		return StageOutputs
	case toc.LevelShortTerm:
		return StageShortTerm
	case toc.LevelMediumTerm:
		return StageMediumTerm
	case toc.LevelLongTerm:
		return StageLongTerm
	default:
		return StageImpact
	}
}

func label(title string, b report.BasisView) string {
	text := oneLine(title)
	if b.Marked {
		text += " (" + string(b.Basis) + ")"
	}
	return text
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

type builder struct {
	stage map[string]StageID
	nodes map[StageID][]Node
	edges []Edge
	seen  map[Edge]bool
}

func (b *builder) add(st StageID, n Node) {
	if _, dup := b.stage[n.ID]; dup {
		return
	}
	b.stage[n.ID] = st
	b.nodes[st] = append(b.nodes[st], n)
}

// edge adds an edge between two nodes of the graph, once. An edge to or from
// something that is not in the view, such as an outcome of another pathway, is
// dropped.
func (b *builder) edge(from, to string) {
	if _, ok := b.stage[from]; !ok {
		return
	}
	if _, ok := b.stage[to]; !ok {
		return
	}
	e := Edge{from, to}
	if from == to || b.seen[e] {
		return
	}
	b.seen[e] = true
	b.edges = append(b.edges, e)
}

// order sorts each stage after the first by where the nodes that lead to it
// sit, which keeps arrows from crossing where they can. Nodes nothing leads to
// keep the place they had.
func (b *builder) order() map[StageID][]Node {
	preds := map[string][]string{}
	for _, e := range b.edges {
		preds[e.To] = append(preds[e.To], e.From)
	}
	row := map[string]int{}
	out := map[StageID][]Node{}
	for _, st := range Stages() {
		nodes := slices.Clone(b.nodes[st])
		weight := map[string]float64{}
		for i, n := range nodes {
			weight[n.ID] = float64(i)
			var sum float64
			var count int
			for _, p := range preds[n.ID] {
				if r, ok := row[p]; ok && b.stage[p] != st {
					sum += float64(r)
					count++
				}
			}
			if count > 0 {
				weight[n.ID] = sum / float64(count)
			}
		}
		sort.SliceStable(nodes, func(i, j int) bool { return weight[nodes[i].ID] < weight[nodes[j].ID] })
		for i, n := range nodes {
			row[n.ID] = i
		}
		out[st] = nodes
	}
	return out
}
