package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/burden"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// Portfolio renders one row of evidence per pathway: where it stands, how
// well supported its routes are, what it costs to measure, how much of its
// data comes free, who it reports to, and which decisions are open. It is
// the evidence for a grow, adjust, or sunset conversation and makes no
// recommendation.
func Portfolio(in StatusInput, cost *burden.Report) string {
	var b strings.Builder
	b.WriteString(header(in, "portfolio"))
	if len(in.View.Pathways) == 0 {
		b.WriteString("This theory has no pathways, so there is nothing to compare.\n\n")
		return b.String()
	}
	costOf := map[string]burden.Group{}
	for _, g := range cost.ByPathway {
		costOf[g.Key] = g
	}
	outcomes := map[string]OutcomeView{}
	for _, o := range in.View.Outcomes {
		outcomes[o.ID] = o
	}

	b.WriteString("| Pathway | Status | Outcomes by status | Confidence | Measurement cost | Free data | Reports to | Open decisions |\n")
	b.WriteString("|---------|--------|--------------------|------------|------------------|-----------|------------|----------------|\n")
	for _, p := range in.View.Pathways {
		state, spread, conf := pathwayStanding(p, outcomes)
		g := costOf[p.ID]
		fmt.Fprintf(&b, "| %s (`%s`) | %s | %s | %s | %s | %s | %s | %s |\n", MDCell(p.Title), p.ID, state, orDash(spread), conf,
			costText(g), freeDataShare(in.Doc, p.ID), orDash(MDCell(strings.Join(audiencesOf(in.Doc, p.ID), ", "))), orDash(MDCell(strings.Join(openDecisionsFor(in.Doc, p), "; "))))
	}
	b.WriteString("\n")
	b.WriteString("_How to read this._ A pathway's status is the worst pathway status among its outcomes. Confidence is the mean confidence of those outcomes' routes to impact. " +
		"Free data is the share of the pathway's selected indicators collected passively or derived from other indicators. The table compares pathways; it does not rank them.\n\n")
	b.WriteString(confidenceNote(in.Status))
	return b.String()
}

func pathwayStanding(p PathwayView, outcomes map[string]OutcomeView) (state, spread, confidence string) {
	counts := map[string]int{}
	worst := ""
	var sum float64
	n := 0
	for _, id := range p.OutcomeIDs {
		o, ok := outcomes[id]
		if !ok || o.Status == nil {
			continue
		}
		s := o.Status.PathwayState
		counts[s]++
		if worst == "" || stateRank(s) < stateRank(worst) {
			worst = s
		}
		if o.Status.Confidence != nil {
			sum += *o.Status.Confidence
			n++
		}
	}
	state = statusLabel(worst)
	confidence = "—"
	if n > 0 {
		confidence = fmt.Sprintf("%.2f", sum/float64(n))
	}
	return state, stateCountText(counts), confidence
}

func costText(g burden.Group) string {
	if g.Indicators == 0 {
		return "—"
	}
	return fmt.Sprintf("%s min/week · %s questions/year · %d indicator(s)", rounded(g.MinutesPerWeek), rounded(g.QuestionsPerRespondentPerYear), g.Indicators)
}

// pathwayIndicators returns the active indicators that serve a pathway.
func pathwayIndicators(doc *toc.Document, pathwayID string) []toc.Indicator {
	p, ok := doc.Pathway(pathwayID)
	if !ok {
		return nil
	}
	var out []toc.Indicator
	for _, ind := range doc.Indicators {
		if ind.Provenance.IsRejected() {
			continue
		}
		if slices.Contains(ind.PathwayIDs, p.ID) || slices.Contains(p.OutcomeIDs, ind.Measures) || slices.Contains(p.InterventionIDs, ind.Measures) {
			out = append(out, ind)
		}
	}
	return out
}

func freeDataShare(doc *toc.Document, pathwayID string) string {
	selected, free := 0, 0
	for _, ind := range pathwayIndicators(doc, pathwayID) {
		if ind.Selection == nil || ind.Selection.Status != toc.SelectionSelected {
			continue
		}
		selected++
		if c := ind.Collection; c != nil && (c.Mode == toc.ModePassive || c.Mode == toc.ModeDerived) {
			free++
		}
	}
	if selected == 0 {
		return "—"
	}
	return fmt.Sprintf("%d of %d (%.0f%%)", free, selected, float64(free)/float64(selected)*100)
}

func audiencesOf(doc *toc.Document, pathwayID string) []string {
	names := map[string]string{}
	for _, a := range doc.Audiences {
		names[a.ID] = a.Name
	}
	seen := map[string]bool{}
	var out []string
	add := func(ids []string) {
		for _, id := range ids {
			if n, ok := names[id]; ok && !seen[id] {
				seen[id] = true
				out = append(out, n)
			}
		}
	}
	for _, ind := range pathwayIndicators(doc, pathwayID) {
		add(ind.AudienceIDs)
	}
	if p, ok := doc.Pathway(pathwayID); ok {
		for _, id := range p.OutcomeIDs {
			if o, ok := doc.Outcome(id); ok {
				add(o.AudienceIDs)
			}
		}
	}
	slices.Sort(out)
	return out
}

// openDecisionsFor lists the open decisions that relate to the pathway, its
// outcomes, or its interventions.
func openDecisionsFor(doc *toc.Document, p PathwayView) []string {
	related := append(slices.Clone(p.OutcomeIDs), p.InterventionIDs...)
	related = append(related, p.ID)
	var out []string
	for _, d := range doc.Decisions {
		if (d.Status == "" || d.Status == toc.DecisionOpen) && slices.ContainsFunc(d.RelatedIDs, func(id string) bool { return slices.Contains(related, id) }) {
			out = append(out, mdText(d.Question))
		}
	}
	return out
}
