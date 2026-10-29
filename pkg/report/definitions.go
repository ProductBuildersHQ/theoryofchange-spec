package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// The definition reports render from a theory alone, with no observations,
// so a workshop can review them before any data exists. They emit portable
// Markdown: headings, pipe tables, and plain text, nothing that GitHub or
// Pandoc cannot render.

// PathwayViews returns one View per pathway, each restricted to that
// pathway and its route to impact, in document order. A theory without
// pathways yields no views; render the whole-theory view instead.
func PathwayViews(doc *toc.Document, g *graph.Graph, st *toc.Status) ([]*View, error) {
	var views []*View
	for _, p := range doc.Pathways {
		if p.Rejected() {
			continue
		}
		v, err := BuildView(doc, g, st, Options{Pathway: p.ID})
		if err != nil {
			return nil, err
		}
		views = append(views, v)
	}
	return views, nil
}

// Stage is one column of a logic model.
type Stage struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

// LogicModelStages projects a view onto the logic-model columns. It is a
// projection of the causal graph, not a second model:
//
//   - Inputs are the resources of the view's interventions.
//   - Activities are the interventions.
//   - Outputs are the interventions' listed outputs and output-level outcomes.
//   - Outcomes are grouped by level.
//   - Impact is the impact-level outcomes.
//
// Item text is safe for a Markdown table cell, and inferred and hypothesis
// elements carry a visible marker. Stages with nothing in them have no items.
func LogicModelStages(v *View) []Stage {
	var inputs, activities, outputs []string
	for _, in := range v.Interventions {
		if in.Resources != "" {
			inputs = appendUnique(inputs, MDCell(in.Resources))
		}
		activities = append(activities, MDCell(in.Title)+BasisMark(in.Basis))
		for _, o := range in.Outputs {
			outputs = appendUnique(outputs, MDCell(o))
		}
	}
	byLevel := map[toc.Level][]string{}
	for _, o := range v.Outcomes {
		text := MDCell(o.Title) + BasisMark(o.Basis)
		if o.Level == toc.LevelOutput {
			outputs = appendUnique(outputs, text)
			continue
		}
		byLevel[o.Level] = append(byLevel[o.Level], text)
	}
	return []Stage{
		{"Inputs", inputs},
		{"Activities", activities},
		{"Outputs", outputs},
		{"Short-term outcomes", byLevel[toc.LevelShortTerm]},
		{"Intermediate outcomes", byLevel[toc.LevelIntermediate]},
		{"Long-term outcomes", byLevel[toc.LevelLongTerm]},
		{"Impact", byLevel[toc.LevelImpact]},
	}
}

// PathwayTitle is the title a logic model or diagram for the view should
// carry: the pathway's title when the view is restricted to one, otherwise
// the theory's.
func PathwayTitle(v *View) string {
	if v.Filter != "" && len(v.Pathways) == 1 {
		return v.Pathways[0].Title
	}
	return v.Document.Title
}

// LogicModel renders one view as a logic model table.
func LogicModel(v *View) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Logic model: %s\n\n", mdText(PathwayTitle(v)))
	if len(v.Pathways) == 1 && v.Pathways[0].Description != "" {
		fmt.Fprintf(&b, "%s\n\n", mdText(v.Pathways[0].Description))
	}
	b.WriteString("| Stage | Elements |\n|-------|----------|\n")
	for _, st := range LogicModelStages(v) {
		cell := "—"
		if len(st.Items) > 0 {
			cell = strings.Join(st.Items, "; ")
		}
		fmt.Fprintf(&b, "| %s | %s |\n", st.Name, cell)
	}
	b.WriteString("\n")
	return b.String()
}

// LogicModels renders a logic model for each pathway view, or a single one
// for the whole theory when it has no pathways.
func LogicModels(whole *View, pathways []*View) string {
	if len(pathways) == 0 {
		return LogicModel(whole)
	}
	var b strings.Builder
	for _, v := range pathways {
		b.WriteString(LogicModel(v))
	}
	return b.String()
}

// MetricDefinitions renders the shared metric library: for each metric its
// definition, exact formula, how it combines across pathways, and the
// indicators that instantiate it.
func MetricDefinitions(v *View) string {
	var b strings.Builder
	b.WriteString("## Metric definitions\n\n")
	if len(v.Metrics) == 0 {
		b.WriteString("This theory defines no shared metrics.\n\n")
		return b.String()
	}
	b.WriteString("| Metric | Definition | Formula | Unit | Direction | Combines across pathways | Layer | Used by |\n")
	b.WriteString("|--------|------------|---------|------|-----------|--------------------------|-------|---------|\n")
	names := map[string]string{}
	for _, ind := range v.Indicators {
		names[ind.ID] = ind.Name
	}
	for _, m := range v.Metrics {
		used := make([]string, len(m.Indicators))
		for i, id := range m.Indicators {
			used[i] = MDCell(names[id])
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s | %s |\n",
			MDCell(m.Name)+BasisMark(m.Basis), MDCell(m.Definition), MDCell(m.Formula), MDCell(m.Unit),
			Label(m.Direction), AggregationText(m), Label(m.Layer), orDash(strings.Join(used, "; ")))
	}
	b.WriteString("\n")
	return b.String()
}

// AggregationText explains in plain words how a metric combines.
func AggregationText(m MetricView) string {
	switch toc.Aggregation(m.Aggregation) {
	case toc.AggregationSum:
		return "Sum"
	case toc.AggregationMean:
		return "Mean"
	case toc.AggregationWeightedMean:
		if m.WeightedBy != "" {
			return "Mean weighted by " + MDCell(m.WeightedBy)
		}
		return "Weighted mean"
	case toc.AggregationMin:
		return "Minimum"
	case toc.AggregationMax:
		return "Maximum"
	case toc.AggregationLatest:
		return "Latest value"
	case toc.AggregationNone:
		return "Cannot be added up: observed directly at each level"
	default:
		return "—"
	}
}

// Glossary renders the shared vocabulary.
func Glossary(v *View) string {
	var b strings.Builder
	b.WriteString("## Glossary\n\n")
	if len(v.Glossary) == 0 {
		b.WriteString("This theory defines no glossary terms.\n\n")
		return b.String()
	}
	b.WriteString("| Term | Definition | Examples |\n|------|------------|----------|\n")
	terms := slices.Clone(v.Glossary)
	for _, t := range terms {
		examples := make([]string, len(t.Examples))
		for i, e := range t.Examples {
			examples[i] = MDCell(e)
		}
		fmt.Fprintf(&b, "| %s | %s | %s |\n", MDCell(t.Term), MDCell(t.Definition), orDash(strings.Join(examples, "; ")))
	}
	b.WriteString("\n")
	return b.String()
}

// ---- shared Markdown helpers ----------------------------------------------

// BasisMark appends a visible marker to inferred and hypothesis elements.
func BasisMark(b BasisView) string {
	if !b.Marked {
		return ""
	}
	return " _(" + string(b.Basis) + ")_"
}

// MDCell makes text safe inside a pipe-table cell.
func MDCell(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "|", `\|`)
	return strings.TrimSpace(s)
}

// mdText makes text safe in running prose by collapsing line breaks.
func mdText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func appendUnique(list []string, item string) []string {
	if slices.Contains(list, item) {
		return list
	}
	return append(list, item)
}
