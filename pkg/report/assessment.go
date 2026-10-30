package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// The assessment report renders from a theory alone. It tells a reader how
// much of a draft is the organization's own word and how much is ours, where
// the theory has gaps, and what to ask the organization. It is the document
// to hand over with a draft.

// tagged is one element in the provenance listings.
type tagged struct {
	kind        toc.ElementKind
	id, title   string
	basis       BasisView
	accountable bool
}

// taggedElements lists every element counted in View.BasisCounts, in view
// order.
func taggedElements(v *View) []tagged {
	var out []tagged
	titles := map[string]string{}
	for _, o := range v.Outcomes {
		titles[o.ID] = o.Title
	}
	for _, in := range v.Interventions {
		titles[in.ID] = in.Title
	}
	name := func(id string) string {
		if t, ok := titles[id]; ok {
			return t
		}
		return id
	}
	for _, x := range v.Pathways {
		out = append(out, tagged{toc.ElementPathway, x.ID, x.Title, x.Basis, true})
	}
	for _, x := range v.Outcomes {
		out = append(out, tagged{toc.ElementOutcome, x.ID, x.Title, x.Basis, x.Accountable})
	}
	for _, x := range v.Links {
		out = append(out, tagged{toc.ElementLink, x.ID, name(x.From) + " → " + name(x.To), x.Basis, true})
	}
	for _, x := range v.Groups {
		out = append(out, tagged{toc.ElementGroup, x.ID, strings.ToUpper(x.Logic) + " into " + name(x.Target), x.Basis, true})
	}
	for _, x := range v.Interventions {
		out = append(out, tagged{toc.ElementIntervention, x.ID, x.Title, x.Basis, true})
	}
	for _, x := range v.Assumptions {
		out = append(out, tagged{toc.ElementAssumption, x.ID, x.Statement, x.Basis, true})
	}
	for _, x := range v.Indicators {
		out = append(out, tagged{toc.ElementIndicator, x.ID, x.Name, x.Basis, true})
	}
	for _, x := range v.Metrics {
		out = append(out, tagged{toc.ElementMetric, x.ID, x.Name, x.Basis, true})
	}
	for _, x := range v.Actors {
		out = append(out, tagged{toc.ElementActor, x.ID, x.Name, x.Basis, true})
	}
	return out
}

func (t tagged) text() string {
	kind := Label(string(t.kind))
	if t.kind == toc.ElementGroup {
		kind = "Precondition group"
	}
	return fmt.Sprintf("%s: %s (`%s`)", kind, mdText(t.title), t.id)
}

// Assessment renders the assessment report for a view of doc.
func Assessment(v *View, doc *toc.Document) string {
	els := taggedElements(v)
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: assessment\n\n", mdText(v.Document.Title))

	b.WriteString("## Summary\n\n")
	b.WriteString(assessmentSummary(v, els) + "\n\n")

	b.WriteString("## Where each element comes from\n\n")
	b.WriteString("| Basis | Meaning | Elements |\n|-------|---------|----------|\n")
	meaning := map[toc.Basis]string{
		toc.BasisStated:     "A cited source says it",
		toc.BasisInferred:   "Our reading of the sources",
		toc.BasisHypothesis: "A guess to test",
		toc.BasisConfirmed:  "The organization agreed to it",
	}
	for _, basis := range toc.AllBases() {
		if basis == toc.BasisRejected {
			fmt.Fprintf(&b, "| %s | Considered and set aside | %d |\n", Label(string(basis)), len(v.Rejected))
			continue
		}
		fmt.Fprintf(&b, "| %s | %s | %d |\n", Label(string(basis)), meaning[basis], v.BasisCounts[basis])
	}
	fmt.Fprintf(&b, "| No basis recorded | — | %d |\n\n", v.Unmarked)
	if n := statedWithoutCitation(els); n > 0 {
		fmt.Fprintf(&b, "%d stated element(s) cite no source; each needs one, or a different basis.\n\n", n)
	}
	for _, basis := range []toc.Basis{toc.BasisHypothesis, toc.BasisInferred, toc.BasisStated, toc.BasisConfirmed} {
		var lines []string
		for _, e := range els {
			if e.basis.Basis == basis {
				lines = append(lines, "- "+e.text())
			}
		}
		if len(lines) == 0 {
			continue
		}
		fmt.Fprintf(&b, "### %s\n\n%s\n\n", Label(string(basis)), strings.Join(lines, "\n"))
	}
	if len(v.Rejected) > 0 {
		b.WriteString("### Rejected\n\n")
		for _, r := range v.Rejected {
			fmt.Fprintf(&b, "- %s: %s (`%s`)\n", Label(string(r.Kind)), mdText(r.Title), r.ID)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Gaps\n\n")
	b.WriteString("### Outcomes within accountability with no indicator\n\n")
	b.WriteString(bullets(outcomesWithoutIndicators(v, doc)))
	b.WriteString("### Indicators by data availability\n\n")
	b.WriteString(dataAvailabilityTable(v))
	b.WriteString("### Assumptions needing testing\n\n")
	b.WriteString(bullets(assumptionsNeedingTesting(v)))
	b.WriteString("### Open decisions\n\n")
	b.WriteString(bullets(openDecisions(doc)))

	b.WriteString("## Questions for the organization\n\n")
	b.WriteString(numbered(questions(v, els, doc), "No questions were generated."))
	return b.String()
}

func assessmentSummary(v *View, els []tagged) string {
	owned := 0
	for _, o := range v.Outcomes {
		if o.Accountable {
			owned++
		}
	}
	marked := v.BasisCounts[toc.BasisInferred] + v.BasisCounts[toc.BasisHypothesis]
	s := fmt.Sprintf("This draft has %d outcomes (%d within the organization's accountability), %d interventions, %d assumptions, and %d indicators",
		len(v.Outcomes), owned, len(v.Interventions), len(v.Assumptions), len(v.Indicators))
	if n := len(v.Pathways); n > 0 {
		s += fmt.Sprintf(", organized into %d pathway(s)", n)
	}
	s += fmt.Sprintf(". Of its %d elements, %d are stated by cited sources, %d are our inference or hypothesis, %d are confirmed by the organization, and %d have no basis recorded.",
		len(els), v.BasisCounts[toc.BasisStated], marked, v.BasisCounts[toc.BasisConfirmed], v.Unmarked)
	if marked > 0 {
		s += " The inferred and hypothesis elements are the first to check with the organization."
	}
	return s
}

func statedWithoutCitation(els []tagged) int {
	n := 0
	for _, e := range els {
		if e.basis.Basis == toc.BasisStated && len(e.basis.EvidenceIDs) == 0 {
			n++
		}
	}
	return n
}

// measuredByProxy returns the outcomes some active indicator stands in for.
func measuredByProxy(doc *toc.Document) map[string]bool {
	out := map[string]bool{}
	for _, ind := range doc.Indicators {
		if ind.Provenance.IsRejected() {
			continue
		}
		for _, id := range ind.ProxyFor {
			out[id] = true
		}
	}
	return out
}

// outcomesWithoutIndicators lists owned outcomes that nothing measures: no
// indicator of their own, and no indicator standing in for them.
func outcomesWithoutIndicators(v *View, doc *toc.Document) []string {
	var out []string
	proxied := measuredByProxy(doc)
	for _, o := range v.Outcomes {
		if o.Accountable && len(o.Indicators) == 0 && !proxied[o.ID] {
			out = append(out, fmt.Sprintf("%s (`%s`)", mdText(o.Title), o.ID))
		}
	}
	return out
}

func dataAvailabilityTable(v *View) string {
	byAvail := map[string][]string{}
	for _, ind := range v.Indicators {
		key := ind.DataAvailability
		if key == "" {
			key = "not recorded"
		}
		byAvail[key] = append(byAvail[key], MDCell(ind.Name))
	}
	if len(v.Indicators) == 0 {
		return "No indicators.\n\n"
	}
	var b strings.Builder
	b.WriteString("| Data availability | Indicators |\n|-------------------|------------|\n")
	for _, a := range toc.AllDataAvailabilities() {
		if names := byAvail[string(a)]; len(names) > 0 {
			fmt.Fprintf(&b, "| %s | %s |\n", Label(string(a)), strings.Join(names, "; "))
		}
	}
	if names := byAvail["not recorded"]; len(names) > 0 {
		fmt.Fprintf(&b, "| Not recorded | %s |\n", strings.Join(names, "; "))
	}
	b.WriteString("\n")
	return b.String()
}

func needsTesting(a AssumptionView) bool {
	if a.NeedsTesting {
		return true
	}
	high := a.Criticality == string(toc.CriticalityHigh) || a.Criticality == string(toc.CriticalityCritical)
	return high && (a.Status == "" || a.Status == "untested")
}

func assumptionsNeedingTesting(v *View) []string {
	var out []string
	for _, a := range v.Assumptions {
		if needsTesting(a) {
			s := fmt.Sprintf("%s (`%s`)", mdText(a.Statement), a.ID)
			if a.Test != "" {
				s += ": " + mdText(a.Test)
			}
			out = append(out, s)
		}
	}
	return out
}

func openDecisions(doc *toc.Document) []string {
	var out []string
	for _, d := range doc.Decisions {
		if d.Status == toc.DecisionDecided {
			continue
		}
		s := fmt.Sprintf("%s (`%s`)", mdText(d.Question), d.ID)
		if len(d.Options) > 0 {
			s += ". Options: " + strings.Join(d.Options, "; ")
		}
		out = append(out, s)
	}
	return out
}

// questions generates the questions to put to the organization: one for each
// inferred or hypothesis element at an accountable level, then one for each
// gap. The wording is fixed so the same document always yields the same
// questions.
func questions(v *View, els []tagged, doc *toc.Document) []string {
	var qs []string
	for _, e := range els {
		if !e.basis.Marked || !e.accountable {
			continue
		}
		verb := "We inferred this from your public materials"
		if e.basis.Basis == toc.BasisHypothesis {
			verb = "We are guessing here"
		}
		qs = append(qs, fmt.Sprintf("%s — %s. Is it right, and what would you change? (%s)", e.text(), verb, e.id))
	}
	proxied := measuredByProxy(doc)
	for _, o := range v.Outcomes {
		if o.Accountable && len(o.Indicators) == 0 && !proxied[o.ID] {
			qs = append(qs, fmt.Sprintf("How would you know that \"%s\" (%s) has happened?", mdText(o.Title), o.ID))
		}
	}
	for _, ind := range v.Indicators {
		if ind.DataAvailability == string(toc.DataUnknown) {
			qs = append(qs, fmt.Sprintf("Where could the data for \"%s\" (%s) come from, and who holds it?", mdText(ind.Name), ind.ID))
		}
	}
	for _, a := range v.Assumptions {
		if needsTesting(a) {
			qs = append(qs, fmt.Sprintf("What have you seen that supports or contradicts \"%s\" (%s)?", mdText(a.Statement), a.ID))
		}
	}
	for _, d := range doc.Decisions {
		if d.Status != toc.DecisionDecided {
			qs = append(qs, fmt.Sprintf("Decision needed: %s (%s)", mdText(d.Question), d.ID))
		}
	}
	return slices.Clip(qs)
}

func bullets(items []string) string {
	const none = "None."
	if len(items) == 0 {
		return none + "\n\n"
	}
	var b strings.Builder
	for _, it := range items {
		b.WriteString("- " + it + "\n")
	}
	b.WriteString("\n")
	return b.String()
}

func numbered(items []string, none string) string {
	if len(items) == 0 {
		return none + "\n"
	}
	var b strings.Builder
	for i, it := range items {
		fmt.Fprintf(&b, "%d. %s\n", i+1, it)
	}
	return b.String()
}
