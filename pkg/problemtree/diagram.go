package problemtree

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grokify/d2vision/generate"
)

// The zones of the diagram, from top to bottom, and the look of each.
const (
	zoneEffects = "effects"
	zoneProblem = "core"
	zoneCauses  = "causes"

	zoneClass     = "zone"
	rowClass      = "row"
	setAsideClass = "set_aside"

	rowGap  = 70 // between rows, leaving room for the arrows
	nodeGap = 30 // between boxes in a row
)

var zoneLabel = map[string]string{
	zoneEffects: "Effects (branches)",
	zoneProblem: "Core problem",
	zoneCauses:  "Causes (roots)",
}

var kindStyle = map[Kind]generate.StyleSpec{
	KindEffect:  {Fill: "#e4f3e5", Stroke: "#2e7d32"},
	KindProblem: {Fill: "#ffe0b2", Stroke: "#e08a1e", Bold: generate.BoolPtr(true), FontSize: generate.IntPtr(20)},
	KindCause:   {Fill: "#f3e6d8", Stroke: "#8d5a2b"},
}

// Spec describes the diagram for d2vision: three zones stacked from top to
// bottom, effects, the core problem, and causes, each holding one row per
// layer, and the edges between the boxes, which all point up. D2, Mermaid,
// and the JSON document are all written from it.
func (t *Tree) Spec() *generate.DiagramSpec {
	round := generate.IntPtr(12)
	spec := &generate.DiagramSpec{
		Direction:   "up",
		GridColumns: 1,
		VerticalGap: rowGap,
		Classes: map[string]generate.StyleSpec{
			zoneClass:     {Fill: "transparent", StrokeWidth: generate.IntPtr(0), FontSize: generate.IntPtr(20), Bold: generate.BoolPtr(true)},
			rowClass:      {Fill: "transparent", StrokeWidth: generate.IntPtr(0)},
			setAsideClass: {Fill: "#f3f3f3", Stroke: "#9e9e9e", Opacity: generate.Float64Ptr(0.7), BorderRadius: round},
		},
	}
	for k, style := range kindStyle {
		style.BorderRadius = round
		spec.Classes[string(k)] = style
	}

	path := map[string]string{} // node ID → its path in the diagram
	zones := map[string]*generate.ContainerSpec{}
	var order []string
	for _, r := range t.Rows {
		zone := zoneOf(r.Kind)
		z, ok := zones[zone]
		if !ok {
			z = &generate.ContainerSpec{ID: zone, Label: zoneLabel[zone], Class: zoneClass, GridColumns: 1, VerticalGap: rowGap}
			zones[zone] = z
			order = append(order, zone)
		}
		row := generate.ContainerSpec{ID: rowID(r.Index), Class: rowClass, GridRows: 1, HorizontalGap: nodeGap}
		for _, n := range r.Nodes {
			class := string(n.Kind)
			label := n.Label
			if n.SetAside {
				class = setAsideClass
				label += " (set aside)"
			}
			row.Nodes = append(row.Nodes, generate.NodeSpec{ID: nodeID(n.ID), Label: label, Class: class})
			path[n.ID] = zone + "." + row.ID + "." + nodeID(n.ID)
		}
		z.Containers = append(z.Containers, row)
	}
	for _, zone := range order {
		spec.Containers = append(spec.Containers, *zones[zone])
	}
	for _, e := range t.Edges {
		spec.Edges = append(spec.Edges, generate.EdgeSpec{From: path[e.From], To: path[e.To]})
	}
	return spec
}

// D2 writes the diagram as D2. Laid out with grids, each zone and row stays in
// place, so every arrow points up.
func (t *Tree) D2() string { return generate.NewGenerator().Generate(t.Spec()) }

// Mermaid writes the diagram as a Mermaid flowchart drawn bottom to top. Mermaid
// has no grid, so it places the rows itself.
func (t *Tree) Mermaid() (string, error) { return generate.NewGenerator().Mermaid(t.Spec()) }

// JSON writes the diagram spec, the document D2 and Mermaid are rendered from.
func (t *Tree) JSON() (string, error) {
	b, err := json.MarshalIndent(t.Spec(), "", "  ")
	if err != nil {
		return "", fmt.Errorf("problemtree: %w", err)
	}
	return string(b) + "\n", nil
}

func zoneOf(k Kind) string {
	switch k {
	case KindEffect:
		return zoneEffects
	case KindCause:
		return zoneCauses
	}
	return zoneProblem
}

// rowID names a row: r2 for the second row of effects, r0 for the core
// problem, r_1 for the first row of causes.
func rowID(i int) string {
	if i < 0 {
		return fmt.Sprintf("r_%d", -i)
	}
	return fmt.Sprintf("r%d", i)
}

// nodeID makes an element ID usable as a node name: a dot would read as a
// step into a container.
func nodeID(id string) string { return strings.ReplaceAll(id, ".", "_") }

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }
