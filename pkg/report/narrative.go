package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// Narrative renders a programme or strategy document scaffold, following the
// sections of the UN guidance: the causal analysis from problem to outcomes,
// priorities and why they were chosen, assumptions and risks, what partners
// deliver, who does what, how it will be monitored and evaluated, who has
// been consulted, and key messages. It needs no observations. When a status
// snapshot is given, outcomes show their status.
func Narrative(v *View, doc *toc.Document, st *toc.Status) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: programme narrative\n\n", mdText(v.Document.Title))
	if st != nil {
		fmt.Fprintf(&b, "Status as of %s.\n\n", st.AsOf)
	}
	b.WriteString(narrativeSummary(v))
	b.WriteString(causalAnalysis(v, doc))
	b.WriteString(priorities(doc))
	b.WriteString(risksAndAssumptions(v, doc))
	b.WriteString(partnerDelivered(v, st))
	b.WriteString(implementation(v))
	b.WriteString(monitoring(v, doc))
	b.WriteString(validationTrail(doc))
	b.WriteString(keyMessages(v))
	return b.String()
}

func narrativeSummary(v *View) string {
	var b strings.Builder
	b.WriteString("## Narrative summary\n\n")
	b.WriteString(mdText(v.Document.Problem) + "\n\n")
	if v.Document.TargetPopulation != "" {
		fmt.Fprintf(&b, "**Who is affected.** %s\n\n", mdText(v.Document.TargetPopulation))
	}
	if v.Document.Narrative != "" {
		b.WriteString(mdText(v.Document.Narrative) + "\n\n")
	}
	return b.String()
}

// causalAnalysis lays the problem tree out by level, with the outcomes that
// address each cause, so the solution can be checked against the problem at
// every level.
func causalAnalysis(v *View, doc *toc.Document) string {
	var b strings.Builder
	b.WriteString("## Causal analysis\n\n")
	if len(doc.Causes) == 0 {
		b.WriteString("No problem tree has been drawn. The causes named in the context are:\n\n")
		b.WriteString(bullets(v.Document.RootCauses))
		return b.String()
	}
	titles := map[string]string{}
	for _, o := range v.Outcomes {
		titles[o.ID] = o.Title
	}
	groups := map[string]string{}
	for _, g := range doc.TargetGroups {
		groups[g.ID] = g.Name
	}
	addressedBy := map[string][]string{}
	for _, o := range doc.Outcomes {
		if o.Rejected() {
			continue
		}
		for _, c := range o.AddressesCauses {
			addressedBy[c] = append(addressedBy[c], mdText(titles[o.ID])+" (`"+o.ID+"`)")
		}
	}
	for _, lv := range []toc.CauseLevel{toc.CauseStructural, toc.CauseUnderlying, toc.CauseImmediate} {
		var rows []string
		for _, c := range doc.Causes {
			if c.Level != lv {
				continue
			}
			addressed := strings.Join(addressedBy[c.ID], "; ")
			switch {
			case addressed == "" && c.OutOfScope != nil:
				addressed = "Set aside: " + mdText(c.OutOfScope.Reason)
			case addressed == "":
				addressed = "**Not addressed**"
			}
			var who []string
			for _, id := range c.AffectedGroupIDs {
				who = append(who, groups[id])
			}
			rows = append(rows, fmt.Sprintf("| %s (`%s`) | %s | %s |", MDCell(c.Statement), c.ID, orDash(MDCell(strings.Join(who, ", "))), MDCell(addressed)))
		}
		if len(rows) > 0 {
			fmt.Fprintf(&b, "### %s causes\n\n| Cause | Who is affected | What addresses it |\n|-------|-----------------|-------------------|\n%s\n\n", Label(string(lv)), strings.Join(rows, "\n"))
		}
	}
	if len(doc.TargetGroups) > 0 {
		b.WriteString("### Priority groups\n\n")
		for _, g := range doc.TargetGroups {
			line := "- **" + mdText(g.Name) + "**"
			if g.Description != "" {
				line += ": " + mdText(g.Description)
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// priorities shows the weighted scores behind which outcomes and
// interventions were chosen, with the stated reasons.
func priorities(doc *toc.Document) string {
	var b strings.Builder
	b.WriteString("## Priorities and why\n\n")
	any := false
	for _, s := range []struct {
		heading string
		subject toc.SelectionSubject
	}{{"Outcomes", toc.SubjectOutcome}, {"Interventions", toc.SubjectIntervention}} {
		rows := priorityRows(doc, s.subject)
		if len(rows) == 0 {
			continue
		}
		any = true
		fmt.Fprintf(&b, "### %s\n\n", s.heading)
		scoreTable(&b, strings.TrimSuffix(s.heading, "s"), criteriaFor(doc, s.subject), rows, false)
	}
	if !any {
		b.WriteString("No priorities have been scored.\n\n")
	}
	return b.String()
}

func risksAndAssumptions(v *View, doc *toc.Document) string {
	var b strings.Builder
	b.WriteString("## Assumptions and risks\n\n")
	if len(v.Assumptions) == 0 {
		b.WriteString("No assumptions are stated.\n\n")
	} else {
		b.WriteString("| Assumption | Type | Criticality | Plausibility | Needs testing |\n|------------|------|-------------|--------------|---------------|\n")
		for _, a := range v.Assumptions {
			fmt.Fprintf(&b, "| %s (`%s`)%s | %s | %s | %s | %s |\n", MDCell(a.Statement), a.ID, BasisMark(a.Basis), Label(a.Type),
				Label(a.Criticality), orDash(Label(a.Plausibility)), yesNo(a.NeedsTesting))
		}
		b.WriteString("\n")
	}
	var risks []string
	for _, r := range doc.Risks {
		if r.Provenance.IsRejected() {
			continue
		}
		line := fmt.Sprintf("`%s` %s", r.ID, mdText(r.Description))
		if r.Mitigation != "" {
			line += " Mitigation: " + mdText(r.Mitigation)
		}
		risks = append(risks, line)
	}
	if len(risks) > 0 {
		b.WriteString("### Risks and how they are managed\n\n")
		b.WriteString(bullets(risks))
	}
	return b.String()
}

// partnerDelivered lists what the organization monitors but others deliver,
// and what nobody delivers.
func partnerDelivered(v *View, st *toc.Status) string {
	var b strings.Builder
	b.WriteString("## What partners deliver\n\n")
	b.WriteString(outsideAccountability(v, st))
	return b.String()
}

// outsideAccountability renders the outcomes that are monitored but not
// answered for, split into those partners deliver, those nobody delivers,
// and those above the ceiling of accountability.
func outsideAccountability(v *View, st *toc.Status) string {
	var partner, unaddressed, above []OutcomeView
	for _, o := range v.Outcomes {
		switch {
		case o.Accountable:
		case o.Responsibility == string(toc.ResponsibilityPartner):
			partner = append(partner, o)
		case o.Responsibility == string(toc.ResponsibilityUnaddressed):
			unaddressed = append(unaddressed, o)
		default:
			above = append(above, o)
		}
	}
	var b strings.Builder
	list := func(heading, note string, items []OutcomeView) {
		if len(items) == 0 {
			return
		}
		fmt.Fprintf(&b, "### %s\n\n%s\n\n| Outcome | Delivered by | Measured by | Status |\n|---------|--------------|-------------|--------|\n", heading, note)
		for _, o := range items {
			state := "—"
			if st != nil && o.Status != nil {
				state = statusLabel(o.Status.PathwayState)
			}
			fmt.Fprintf(&b, "| %s (`%s`) | %s | %s | %s |\n", MDCell(o.Title), o.ID, orDash(MDCell(strings.Join(o.DeliveredBy, ", "))),
				orDash(strings.Join(o.Indicators, ", ")), state)
		}
		b.WriteString("\n")
	}
	list("Delivered by partners", "These are part of the solution and are monitored, but others answer for them.", partner)
	list("Needed, and delivered by no one", "The theory depends on these. Nobody has taken them on.", unaddressed)
	list("Above the ceiling of accountability", "These are tracked and contributed to, but not owned.", above)
	if b.Len() == 0 {
		return "Everything in this theory is within what the organization answers for.\n\n"
	}
	return b.String()
}

func implementation(v *View) string {
	var b strings.Builder
	b.WriteString("## Who does what\n\n")
	if len(v.Interventions) == 0 {
		b.WriteString("No interventions are described.\n\n")
		return b.String()
	}
	b.WriteString("| Intervention | Owner | Status |\n|--------------|-------|--------|\n")
	for _, iv := range v.Interventions {
		fmt.Fprintf(&b, "| %s (`%s`) | %s | %s |\n", MDCell(iv.Title), iv.ID, orDash(iv.Owner), statusLabel(iv.Status))
	}
	b.WriteString("\nThe division of labour across all actors is in the actors report.\n\n")
	return b.String()
}

// monitoring sets out how the theory will be watched and tested: the
// assumptions that need testing, how often indicators are collected, and
// the questions an evaluation should answer.
func monitoring(v *View, doc *toc.Document) string {
	var b strings.Builder
	b.WriteString("## Monitoring and evaluation\n\n### Assumptions to test\n\n")
	var tests []string
	for _, a := range v.Assumptions {
		if a.NeedsTesting {
			tests = append(tests, fmt.Sprintf("`%s` %s Test: %s", a.ID, mdText(a.Statement), orDash(mdText(a.Test))))
		}
	}
	b.WriteString(bullets(tests))

	b.WriteString("### How often it is measured\n\n")
	cadence := map[string][]string{}
	for _, ind := range v.Indicators {
		cadence[ind.Frequency] = append(cadence[ind.Frequency], ind.ID)
	}
	var freqs []string
	for f := range cadence {
		freqs = append(freqs, f)
	}
	slices.Sort(freqs)
	var lines []string
	for _, f := range freqs {
		name := f
		if name == "" {
			name = "no frequency stated"
		}
		lines = append(lines, fmt.Sprintf("%s: %d indicator(s)", name, len(cadence[f])))
	}
	b.WriteString(bullets(lines))

	b.WriteString("### Questions an evaluation should answer\n\n")
	var qs []string
	for _, l := range doc.Links {
		if !l.Rejected() && len(l.EvidenceIDs) == 0 && len(qs) < 5 {
			qs = append(qs, fmt.Sprintf("Does it hold that \"%s\" leads to \"%s\"? (`%s` rests on no evidence yet.)", linkTitle(v, l.From), linkTitle(v, l.To), l.ID))
		}
	}
	b.WriteString(bullets(qs))
	return b.String()
}

func linkTitle(v *View, id string) string {
	for _, o := range v.Outcomes {
		if o.ID == id {
			return mdText(o.Title)
		}
	}
	return id
}

// validationTrail lists who has been consulted and what changed as a result.
func validationTrail(doc *toc.Document) string {
	var b strings.Builder
	b.WriteString("## Who has been consulted\n\n")
	if len(doc.Consultations) == 0 {
		b.WriteString("No consultation has been recorded.\n\n")
		return b.String()
	}
	b.WriteString("| Date | Step | Method | Who | What was learned |\n|------|------|--------|-----|------------------|\n")
	for _, c := range doc.Consultations {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", c.Date, Label(string(c.Step)), orDash(Label(string(c.Method))),
			orDash(MDCell(strings.Join(c.StakeholderGroups, ", "))), orDash(MDCell(strings.Join(c.Findings, " "))))
	}
	b.WriteString("\n")
	return b.String()
}

func keyMessages(v *View) string {
	var b strings.Builder
	b.WriteString("## Key messages\n\n")
	var goals []string
	accountable, monitored := 0, 0
	for _, o := range v.Outcomes {
		if o.Level == toc.LevelImpact {
			goals = append(goals, mdText(o.Title))
		}
		if o.Accountable {
			accountable++
		} else {
			monitored++
		}
	}
	if len(goals) > 0 {
		fmt.Fprintf(&b, "- The change this work is for: %s.\n", strings.Join(goals, "; "))
	}
	if v.Document.Ceiling != "" {
		fmt.Fprintf(&b, "- The organization answers for the theory up to: %s.\n", v.Document.Ceiling)
	}
	fmt.Fprintf(&b, "- %d outcome(s) are the organization's to deliver or share; %d are monitored but delivered by others or beyond the ceiling.\n", accountable, monitored)
	return b.String()
}
