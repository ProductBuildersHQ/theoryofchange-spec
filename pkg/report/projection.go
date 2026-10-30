package report

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// The projections below are views of the causal model, not second models.
// Flattening a network into a table or a chain loses the links and the
// reasons behind them, so each says it is a view and every row names the
// elements it was projected from.

const projectionNote = "_This is a view of the causal model. It flattens the links between outcomes; the pathway diagram shows them._\n\n"

// views returns the views to project: one per pathway, or the whole theory
// when it has none.
func views(whole *View, pathways []*View) []*View {
	if len(pathways) == 0 {
		return []*View{whole}
	}
	return pathways
}

func idRef(title, id string) string { return MDCell(title) + " (`" + id + "`)" }

// OutcomeMaps renders the outcomes framework for each pathway.
func OutcomeMaps(whole *View, pathways []*View) string {
	var b strings.Builder
	for _, v := range views(whole, pathways) {
		b.WriteString(OutcomeMap(v))
	}
	return b.String()
}

// OutcomeMap renders one view's outcomes framework: outcomes grouped by
// level from impact down, each with its sphere, category, kind of change,
// responsibility, delivering actors, the causes it addresses, and the
// indicators that measure it. Outcomes above the ceiling of accountability,
// and those that are monitored only, are marked.
func OutcomeMap(v *View) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Outcome map: %s\n\n%s", mdText(PathwayTitle(v)), projectionNote)
	if len(v.Outcomes) == 0 {
		b.WriteString("This view has no outcomes.\n\n")
		return b.String()
	}
	for _, lv := range toc.AllLevels() {
		var rows []string
		for _, o := range v.Outcomes {
			if o.Level != lv {
				continue
			}
			rows = append(rows, fmt.Sprintf("| %s%s | %s | %s | %s | %s | %s | %s | %s |", idRef(o.Title, o.ID), BasisMark(o.Basis),
				orDash(Label(string(o.Sphere))), orDash(MDCell(o.Category)), orDash(Label(string(o.ChangeType))),
				responsibilityText(o), orDash(MDCell(strings.Join(o.DeliveredBy, ", "))),
				orDash(strings.Join(o.AddressesCauses, ", ")), orDash(strings.Join(o.Indicators, ", "))))
		}
		if len(rows) == 0 {
			continue
		}
		fmt.Fprintf(&b, "### %s\n\n| Outcome | Sphere | Category | Change | Responsibility | Delivered by | Addresses | Measured by |\n", LevelLabel(lv))
		b.WriteString("|---------|--------|----------|--------|----------------|--------------|-----------|-------------|\n")
		b.WriteString(strings.Join(rows, "\n") + "\n\n")
	}
	return b.String()
}

func responsibilityText(o OutcomeView) string {
	text := Label(o.Responsibility)
	switch {
	case !o.Owned:
		return text + " · above the ceiling, monitored only"
	case !o.Accountable:
		return text + " · monitored only"
	}
	return text
}

// Logframes renders a logframe matrix for each pathway.
func Logframes(whole *View, pathways []*View) string {
	var b strings.Builder
	for _, v := range views(whole, pathways) {
		b.WriteString(Logframe(v))
	}
	return b.String()
}

// Logframe renders one view as a logical framework matrix. The rows are the
// goal (impact outcomes), the purpose (long-term and intermediate
// outcomes), the outputs (short-term and output outcomes), and the
// activities (interventions). The columns are the narrative, the
// indicators with their targets, how each indicator is verified, and the
// assumptions that bear on the row's links or interventions.
func Logframe(v *View) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Logframe: %s\n\n%s", mdText(PathwayTitle(v)), projectionNote)

	indicators := map[string]IndicatorView{}
	for _, i := range v.Indicators {
		indicators[i.ID] = i
	}
	byTarget := assumptionsByElement(v)

	b.WriteString("| Level | Narrative | Indicators and targets | Means of verification | Assumptions |\n|-------|-----------|------------------------|-----------------------|-------------|\n")
	row := func(level string, narrative string, inds []string, assumptions []string) {
		var ind, verify []string
		for _, id := range inds {
			iv := indicators[id]
			text := idRef(iv.Name, id)
			if len(iv.Targets) > 0 {
				text += ": " + strings.Join(iv.Targets, "; ")
			}
			ind = append(ind, text)
			verify = append(verify, verification(iv))
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", level, narrative, orDash(strings.Join(ind, "; ")),
			orDash(strings.Join(verify, "; ")), orDash(strings.Join(assumptions, "; ")))
	}
	rows := []struct {
		level  string
		levels []toc.Level
	}{
		{"Goal", []toc.Level{toc.LevelImpact}},
		{"Purpose", []toc.Level{toc.LevelLongTerm, toc.LevelIntermediate}},
		{"Outputs", []toc.Level{toc.LevelShortTerm, toc.LevelOutput}},
	}
	for _, r := range rows {
		for _, o := range v.Outcomes {
			if slices.Contains(r.levels, o.Level) {
				row(r.level, idRef(o.Title, o.ID)+BasisMark(o.Basis), o.Indicators, byTarget[o.ID])
			}
		}
	}
	for _, in := range v.Interventions {
		var inds []string
		for _, i := range v.Indicators {
			if i.Measures == in.ID {
				inds = append(inds, i.ID)
			}
		}
		row("Activities", idRef(in.Title, in.ID)+BasisMark(in.Basis), inds, byTarget[in.ID])
	}
	b.WriteString("\n")
	return b.String()
}

// verification says how an indicator's value is established.
func verification(iv IndicatorView) string {
	var parts []string
	if iv.DataSource != "" {
		parts = append(parts, iv.DataSource)
	}
	if iv.Method != "" {
		parts = append(parts, iv.Method)
	}
	if iv.Frequency != "" {
		parts = append(parts, iv.Frequency)
	}
	if len(parts) == 0 {
		return "not stated"
	}
	return MDCell(strings.Join(parts, "; "))
}

// assumptionsByElement maps each outcome and intervention to the assumptions
// that bear on it: those that name it, and those that name a link into it.
func assumptionsByElement(v *View) map[string][]string {
	byID := map[string]AssumptionView{}
	for _, a := range v.Assumptions {
		byID[a.ID] = a
	}
	set := map[string]map[string]bool{}
	add := func(element, assumption string) {
		if set[element] == nil {
			set[element] = map[string]bool{}
		}
		set[element][assumption] = true
	}
	for _, a := range v.Assumptions {
		for _, id := range a.AppliesTo {
			add(id, a.ID)
		}
	}
	for _, l := range v.Links {
		for _, id := range l.Assumptions {
			add(l.To, id)
		}
		for _, a := range v.Assumptions {
			if slices.Contains(a.AppliesTo, l.ID) {
				add(l.To, a.ID)
			}
		}
	}
	out := map[string][]string{}
	for element, ids := range set {
		var sorted []string
		for id := range ids {
			sorted = append(sorted, id)
		}
		slices.Sort(sorted)
		for _, id := range sorted {
			if a, ok := byID[id]; ok {
				out[element] = append(out[element], idRef(a.Statement, id))
			}
		}
	}
	return out
}

// LogicModelDiagrams renders a left-to-right diagram of the logic model for
// each pathway.
func LogicModelDiagrams(whole *View, pathways []*View) string {
	var b strings.Builder
	for _, v := range views(whole, pathways) {
		fmt.Fprintf(&b, "## Logic model diagram: %s\n\n```mermaid\n%s```\n\n", mdText(PathwayTitle(v)), LogicModelMermaid(v))
	}
	return b.String()
}

// LogicModelMermaid draws the stages of a logic model, from inputs to impact,
// as a left-to-right flow. Stages with nothing in them are left out and the
// arrows join the stages that remain.
func LogicModelMermaid(v *View) string {
	var b strings.Builder
	b.WriteString("flowchart LR\n")
	type stage struct{ id, name string }
	var drawn []stage
	n := 0
	for i, st := range LogicModelStages(v) {
		if len(st.Items) == 0 {
			continue
		}
		id := "stage" + strconv.Itoa(i)
		fmt.Fprintf(&b, "  subgraph %s[\"%s\"]\n", id, mermaidText(st.Name))
		for _, item := range st.Items {
			n++
			fmt.Fprintf(&b, "    n%d[\"%s\"]\n", n, mermaidText(strings.ReplaceAll(item, "\\|", "|")))
		}
		b.WriteString("  end\n")
		drawn = append(drawn, stage{id, st.Name})
	}
	for i := 0; i+1 < len(drawn); i++ {
		fmt.Fprintf(&b, "  %s --> %s\n", drawn[i].id, drawn[i+1].id)
	}
	return b.String()
}
