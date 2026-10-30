package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// Actors renders the division of labour: for each actor, what it delivers,
// what it changes or co-owns, and what it runs; then the outcomes that need
// several parties, the outcomes nobody delivers, and the outcomes partners
// deliver that the organization monitors.
func Actors(v *View, doc *toc.Document, st *toc.Status) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: who does what\n\n", mdText(v.Document.Title))
	titles := map[string]string{}
	for _, o := range v.Outcomes {
		titles[o.ID] = mdText(o.Title)
	}
	label := func(ids []string) string {
		out := make([]string, len(ids))
		for i, id := range ids {
			out[i] = titles[id]
		}
		return strings.Join(out, "; ")
	}

	b.WriteString("## Division of labour\n\n| Actor | Roles | Delivers | Changes or co-owns | Runs |\n|-------|-------|----------|--------------------|------|\n")
	for _, a := range v.Actors {
		var delivers, changes, runs []string
		for _, o := range doc.Outcomes {
			if o.Rejected() {
				continue
			}
			if slices.Contains(o.DeliveredBy, a.ID) {
				delivers = append(delivers, o.ID)
			}
			if slices.Contains(o.ActorIDs, a.ID) {
				changes = append(changes, o.ID)
			}
		}
		for _, in := range doc.Interventions {
			if !in.Provenance.IsRejected() && in.OwnerID == a.ID {
				runs = append(runs, mdText(in.Title))
			}
		}
		fmt.Fprintf(&b, "| %s (`%s`) | %s | %s | %s | %s |\n", MDCell(a.Name), a.ID, orDash(Label(strings.Join(a.Roles, ", "))),
			orDash(MDCell(label(delivers))), orDash(MDCell(label(changes))), orDash(MDCell(strings.Join(runs, "; "))))
	}
	b.WriteString("\n")

	b.WriteString("## Outcomes that need collaboration\n\n")
	var collab []string
	for _, o := range v.Outcomes {
		if o.Responsibility == string(toc.ResponsibilityShared) || o.Responsibility == string(toc.ResponsibilityPartner) || len(o.DeliveredBy) > 1 {
			collab = append(collab, fmt.Sprintf("%s (`%s`), %s: %s", mdText(o.Title), o.ID, strings.ToLower(o.Responsibility), orDash(strings.Join(o.DeliveredBy, ", "))))
		}
	}
	b.WriteString(bullets(collab))

	b.WriteString("## Outcomes with no one delivering them\n\n")
	b.WriteString(bullets(undelivered(v, doc)))

	b.WriteString("## Outcomes partners deliver, which are monitored\n\n")
	b.WriteString(partnerOnly(v, st))
	return b.String()
}

// undelivered lists the outcomes the organization answers for that no actor
// is named as delivering and no owned intervention targets.
func undelivered(v *View, doc *toc.Document) []string {
	targeted := map[string]bool{}
	for _, in := range doc.Interventions {
		if in.Provenance.IsRejected() || in.OwnerID == "" {
			continue
		}
		for _, t := range in.Targets {
			targeted[t] = true
		}
	}
	var out []string
	for _, o := range v.Outcomes {
		if o.Accountable && len(o.DeliveredBy) == 0 && !targeted[o.ID] && o.Level != toc.LevelImpact {
			out = append(out, fmt.Sprintf("%s (`%s`)", mdText(o.Title), o.ID))
		}
	}
	return out
}

func partnerOnly(v *View, st *toc.Status) string {
	var b strings.Builder
	found := false
	b.WriteString("| Outcome | Delivered by | Measured by | Status |\n|---------|--------------|-------------|--------|\n")
	for _, o := range v.Outcomes {
		if o.Responsibility != string(toc.ResponsibilityPartner) {
			continue
		}
		found = true
		state := "—"
		if st != nil && o.Status != nil {
			state = statusLabel(o.Status.PathwayState)
		}
		fmt.Fprintf(&b, "| %s (`%s`) | %s | %s | %s |\n", MDCell(o.Title), o.ID, orDash(MDCell(strings.Join(o.DeliveredBy, ", "))), orDash(strings.Join(o.Indicators, ", ")), state)
	}
	if !found {
		return "None.\n\n"
	}
	b.WriteString("\n")
	return b.String()
}
