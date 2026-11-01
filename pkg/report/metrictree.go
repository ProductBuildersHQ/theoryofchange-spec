package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// MetricTree renders the metric tree report: for each pathway, the one number
// it steers by and the guardrails that watch it, then every indicator of the
// pathway laid out from what was done to what changed, as a table and as a
// diagram. With a pathway ID it covers that pathway only.
func MetricTree(in StatusInput, pathwayID string) (string, error) {
	var pws []impact.Pathway
	for _, p := range in.Doc.Pathways {
		if !p.Rejected() && (pathwayID == "" || p.ID == pathwayID) {
			pws = append(pws, p)
		}
	}
	if pathwayID != "" && len(pws) == 0 {
		return "", fmt.Errorf("report: unknown pathway %q", pathwayID)
	}
	views := map[string]IndicatorView{}
	for _, v := range in.View.Indicators {
		views[v.ID] = v
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# %s: metric tree\n\nAs of %s.\n\n", mdText(in.View.Document.Title), in.Status.AsOf)
	if len(pws) == 0 {
		b.WriteString("The impact model has no pathways.\n")
		return b.String(), nil
	}
	for _, p := range pws {
		fmt.Fprintf(&b, "## %s\n\n", mdText(p.Title))
		writeFocus(&b, in.Doc, p, views)

		byTier := map[impact.Tier][]impact.Indicator{}
		for _, ind := range in.Doc.Indicators {
			if !ind.Rejected() && in.Doc.IndicatorServes(ind, p.ID) {
				t := in.Doc.IndicatorTier(ind)
				byTier[t] = append(byTier[t], ind)
			}
		}
		if len(byTier) == 0 {
			b.WriteString("No indicator serves this pathway.\n\n")
			continue
		}
		b.WriteString("| Tier | Indicator | Latest | Target | Status |\n|------|-----------|--------|--------|--------|\n")
		for _, t := range impact.AllTiers() {
			for _, ind := range byTier[t] {
				v := views[ind.ID]
				fmt.Fprintf(&b, "| %s | %s%s | %s | %s | %s |\n", t, MDCell(ind.Name), roleMark(p, ind.ID),
					orDash(MDCell(v.Latest)), orDash(MDCell(strings.Join(v.Targets, "; "))), statusLabel(stateOf(v)))
			}
		}
		b.WriteString("\n")
		b.WriteString(treeDiagram(p, byTier))
	}
	return b.String(), nil
}

func stateOf(v IndicatorView) string {
	if v.Status == nil {
		return ""
	}
	return v.Status.State
}

// roleMark tags the indicator that is the pathway's North Star, one metric
// that matters, or a guardrail.
func roleMark(p impact.Pathway, indicatorID string) string {
	if p.Focus == nil {
		return ""
	}
	var roles []string
	if p.Focus.NorthStarID == indicatorID {
		roles = append(roles, "North Star")
	}
	if p.Focus.OneMetricID == indicatorID {
		roles = append(roles, "one metric")
	}
	if slices.Contains(p.Focus.GuardrailIDs, indicatorID) {
		roles = append(roles, "guardrail")
	}
	if len(roles) == 0 {
		return ""
	}
	return " (" + strings.Join(roles, ", ") + ")"
}

func writeFocus(b *strings.Builder, doc *impact.Model, p impact.Pathway, views map[string]IndicatorView) {
	f := p.Focus
	if f == nil || (f.NorthStarID == "" && f.OneMetricID == "" && len(f.GuardrailIDs) == 0) {
		b.WriteString("No North Star is chosen for this pathway.\n\n")
		return
	}
	line := func(role, id string) {
		if id == "" {
			return
		}
		name, latest := id, ""
		if ind, ok := doc.Indicator(id); ok {
			name = ind.Name
		}
		if v, ok := views[id]; ok && v.Latest != "" {
			latest = ": " + v.Latest
		}
		fmt.Fprintf(b, "- %s: %s%s\n", role, mdText(name), mdText(latest))
	}
	line("North Star", f.NorthStarID)
	line("One metric that matters", f.OneMetricID)
	for _, id := range f.GuardrailIDs {
		line("Guardrail", id)
	}
	if f.Rationale != "" {
		fmt.Fprintf(b, "\nWhy: %s\n", mdText(f.Rationale))
	}
	b.WriteString("\n")
}

// treeDiagram draws the pathway's indicators in their tiers, left to right.
func treeDiagram(p impact.Pathway, byTier map[impact.Tier][]impact.Indicator) string {
	var b strings.Builder
	b.WriteString("```mermaid\nflowchart LR\n")
	var prev string
	var starred, guards []string
	for _, t := range impact.AllTiers() {
		inds := byTier[t]
		if len(inds) == 0 {
			continue
		}
		sg := "tier_" + string(t)
		fmt.Fprintf(&b, "  subgraph %s[\"%s\"]\n", sg, t)
		for _, ind := range inds {
			n := "i_" + mermaidSafe(ind.ID)
			fmt.Fprintf(&b, "    %s[\"%s\"]\n", n, mermaidText(ind.Name))
			if p.Focus != nil {
				if p.Focus.NorthStarID == ind.ID {
					starred = append(starred, n)
				}
				if slices.Contains(p.Focus.GuardrailIDs, ind.ID) {
					guards = append(guards, n)
				}
			}
		}
		b.WriteString("  end\n")
		if prev != "" {
			fmt.Fprintf(&b, "  %s --> %s\n", prev, sg)
		}
		prev = sg
	}
	b.WriteString("  classDef northstar fill:#fde68a,stroke:#b45309,stroke-width:2px\n  classDef guardrail fill:#e5e7eb,stroke:#6b7280,stroke-dasharray:4 3\n")
	if len(starred) > 0 {
		fmt.Fprintf(&b, "  class %s northstar\n", strings.Join(starred, ","))
	}
	if len(guards) > 0 {
		fmt.Fprintf(&b, "  class %s guardrail\n", strings.Join(guards, ","))
	}
	b.WriteString("```\n\n")
	return b.String()
}

func mermaidSafe(id string) string {
	return strings.Map(func(c rune) rune {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			return c
		}
		return '_'
	}, id)
}
