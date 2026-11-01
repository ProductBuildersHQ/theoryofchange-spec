package logicmodel

import (
	"strings"

	"github.com/grokify/d2vision/generate"
)

// The look of the diagram. A stage heading has no box of its own, and each
// stage's items share a color.
const (
	stageClass = "stage"
	// columnGap leaves room between columns for the arrows.
	columnGap = 90
	// itemGap spaces the items in a column.
	itemGap = 24
)

var stageStyle = map[StageID]generate.StyleSpec{
	StageInputs:     {Fill: "#fff1dc", Stroke: "#e08a1e"},
	StageActivities: {Fill: "#e4f3e5", Stroke: "#2e7d32"},
	StageOutputs:    {Fill: "#ece4fb", Stroke: "#7e57c2"},
	StageShortTerm:  {Fill: "#e1f3fd", Stroke: "#2d9cdb"},
	StageMediumTerm: {Fill: "#dfe8f6", Stroke: "#3f5fa8"},
	StageLongTerm:   {Fill: "#d3dcee", Stroke: "#2c4a8a"},
	StageImpact:     {Fill: "#c5cfe8", Stroke: "#1f3a7a"},
}

// Spec describes the diagram for d2vision: one container for each stage, laid
// out as a row of columns, a node for each item, and the edges between them.
// D2 and Mermaid are both written from it.
func (g *Graph) Spec() *generate.DiagramSpec {
	round := generate.IntPtr(12)
	spec := &generate.DiagramSpec{
		Direction:     "right",
		GridColumns:   len(g.Stages),
		HorizontalGap: columnGap,
		Classes: map[string]generate.StyleSpec{
			stageClass: {Fill: "transparent", StrokeWidth: generate.IntPtr(0), FontSize: generate.IntPtr(22), Bold: generate.BoolPtr(true)},
		},
	}
	for _, st := range g.Stages {
		style := stageStyle[st.Stage]
		style.BorderRadius = round
		spec.Classes[string(st.Stage)] = style
	}
	for _, st := range g.Stages {
		c := generate.ContainerSpec{
			ID:          string(st.Stage),
			Label:       st.Stage.Name(),
			Class:       stageClass,
			GridColumns: 1,
			VerticalGap: itemGap,
		}
		for _, n := range st.Nodes {
			c.Nodes = append(c.Nodes, generate.NodeSpec{ID: diagramID(n.ID), Label: n.Label, Class: string(st.Stage)})
		}
		spec.Containers = append(spec.Containers, c)
	}
	stageOf := map[string]StageID{}
	for _, st := range g.Stages {
		for _, n := range st.Nodes {
			stageOf[n.ID] = st.Stage
		}
	}
	for _, e := range g.Edges {
		spec.Edges = append(spec.Edges, generate.EdgeSpec{
			From: string(stageOf[e.From]) + "." + diagramID(e.From),
			To:   string(stageOf[e.To]) + "." + diagramID(e.To),
		})
	}
	return spec
}

// D2 writes the diagram as D2. Laid out with a grid, the stages stay in their
// columns, in order.
func (g *Graph) D2() string {
	return generate.NewGenerator().Generate(g.Spec())
}

// Mermaid writes the diagram as a Mermaid flowchart. Mermaid has no grid, so it
// places the columns itself.
func (g *Graph) Mermaid() (string, error) {
	return generate.NewGenerator().Mermaid(g.Spec())
}

// diagramID makes an element ID usable as a node name: a dot would read as a
// step into a container.
func diagramID(id string) string { return strings.ReplaceAll(id, ".", "_") }
