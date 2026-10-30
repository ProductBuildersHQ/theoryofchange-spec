package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// StatusKinds lists the status report kinds that need observations, in the
// order the full report presents them.
var StatusKinds = []string{"exec", "results", "assumptions", "interventions", "evidence-gaps", "full"}

// StatusInput is what a status report is made from: the view (built with the
// status), the document it came from, the snapshot, and optionally what
// changed since the previous snapshot.
type StatusInput struct {
	View   *View
	Doc    *toc.Document
	Status *toc.Status
	Diff   *toc.StatusDiff
}

// stateRank orders states from worst to best for sorting.
func stateRank(s string) int {
	for i, st := range []toc.State{
		toc.StateOffTrack, toc.StateAtRisk, toc.StateStale, toc.StateNoData, toc.StateNotStarted,
		toc.StateBaselineOnly, toc.StateUnmeasured, toc.StateOnTrack, toc.StateAchieved,
	} {
		if string(st) == s {
			return i
		}
	}
	return 99
}

func pct(p *float64) string {
	if p == nil {
		return "—"
	}
	return fmt.Sprintf("%.0f%%", *p*100)
}

func (in StatusInput) titles() map[string]string {
	t := map[string]string{}
	for _, o := range in.View.Outcomes {
		t[o.ID] = o.Title
	}
	return t
}

func (in StatusInput) indicatorStatus() map[string]toc.IndicatorStatus {
	m := map[string]toc.IndicatorStatus{}
	for _, s := range in.Status.Indicators {
		m[s.IndicatorID] = s
	}
	return m
}

// cites explains an indicator's status from the numbers behind it.
func cites(s toc.IndicatorStatus, asOf string) string {
	switch {
	case s.Progress != nil && s.Expected != nil:
		return fmt.Sprintf("%s of the way to target; %s expected by %s", pct(s.Progress), pct(s.Expected), asOf)
	case s.Latest != nil:
		return "latest value reported; no targets to measure progress against"
	case s.State == toc.StateNotStarted:
		return "no observation yet; the first target is still ahead"
	case s.State == toc.StateNoData:
		return "no observation, and the first target date has passed"
	}
	return "no data"
}

func statusLabel(state string) string {
	if state == "" {
		return "—"
	}
	return Label(state)
}

func header(in StatusInput, title string) string {
	return fmt.Sprintf("# %s: %s\n\nAs of %s.\n\n", mdText(in.View.Document.Title), title, in.Status.AsOf)
}

// confidenceNote states how confidence and status were computed, so a reader
// can see the settings behind every number.
func confidenceNote(st *toc.Status) string {
	w := st.Options.Weights
	return fmt.Sprintf("_Method._ Status compares progress with what the dated targets call for, within a tolerance of %.0f%%; "+
		"an outcome takes the %s of its indicators. Confidence in a link is %.1f × evidence + %.1f × assumptions + %.1f × rationale, "+
		"from 0 to 1; a route to impact multiplies its links' confidence.\n\n",
		st.Options.Tolerance*100, st.Options.Rollup, w.Evidence, w.Assumptions, w.Rationale)
}

// Exec renders the executive summary: the headline, status by level, where
// attention is needed, broken or untested critical assumptions, what changed,
// and decisions requested.
func Exec(in StatusInput) string {
	var b strings.Builder
	b.WriteString(header(in, "executive summary"))
	titles := in.titles()

	b.WriteString("## Headline\n\n")
	if len(in.Status.Overall) == 0 {
		b.WriteString("This theory has no impact outcome yet.\n\n")
	}
	for _, i := range in.Status.Overall {
		fmt.Fprintf(&b, "- **%s**: %s", mdText(titles[i.OutcomeID]), statusLabel(string(i.PathwayState)))
		if i.Confidence != nil {
			fmt.Fprintf(&b, " (confidence %.2f)", *i.Confidence)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	b.WriteString("## Status by level\n\n| Level | Outcomes | Pathway status |\n|-------|----------|----------------|\n")
	for _, lv := range toc.AllLevels() {
		counts := map[string]int{}
		n := 0
		for _, o := range in.View.Outcomes {
			if o.Level == lv && o.Status != nil {
				counts[o.Status.PathwayState]++
				n++
			}
		}
		if n > 0 {
			fmt.Fprintf(&b, "| %s | %d | %s |\n", LevelLabel(lv), n, stateCountText(counts))
		}
	}
	b.WriteString("\n")

	b.WriteString("## Where attention is needed\n\n")
	b.WriteString(attention(in, 5))
	b.WriteString(brokenAssumptions(in))
	b.WriteString(changes(in))
	b.WriteString(decisions(in))
	b.WriteString(confidenceNote(in.Status))
	return b.String()
}

func stateCountText(c map[string]int) string {
	states := make([]string, 0, len(c))
	for s := range c {
		states = append(states, s)
	}
	slices.SortFunc(states, func(a, b string) int { return stateRank(a) - stateRank(b) })
	parts := make([]string, len(states))
	for i, s := range states {
		parts[i] = fmt.Sprintf("%d %s", c[s], strings.ToLower(Label(s)))
	}
	return strings.Join(parts, " · ")
}

// attention lists the outcomes whose pathway is at risk or off track, worst
// first, each with the indicators that drive it and the preconditions that
// hold it back.
func attention(in StatusInput, limit int) string {
	type row struct {
		o     OutcomeView
		state string
	}
	var rows []row
	for _, o := range in.View.Outcomes {
		if o.Status == nil {
			continue
		}
		if s := o.Status.PathwayState; s == string(toc.StateOffTrack) || s == string(toc.StateAtRisk) {
			rows = append(rows, row{o, s})
		}
	}
	if len(rows) == 0 {
		return "Nothing is at risk or off track.\n\n"
	}
	slices.SortStableFunc(rows, func(a, b row) int {
		if d := stateRank(a.state) - stateRank(b.state); d != 0 {
			return d
		}
		return strings.Compare(a.o.ID, b.o.ID)
	})
	byInd := in.indicatorStatus()
	pathway := map[string]string{}
	for _, o := range in.View.Outcomes {
		if o.Status != nil {
			pathway[o.ID] = o.Status.PathwayState
		}
	}
	snapshot := map[string]toc.OutcomeStatus{}
	for _, o := range in.Status.Outcomes {
		snapshot[o.OutcomeID] = o
	}
	var b strings.Builder
	b.WriteString("| Outcome | Status | Because |\n|---------|--------|---------|\n")
	for i, r := range rows {
		if i == limit {
			fmt.Fprintf(&b, "\n%d more at risk or off track are listed in the results report.\n", len(rows)-limit)
			break
		}
		var why []string
		for _, id := range snapshot[r.o.ID].IndicatorIDs {
			if s, ok := byInd[id]; ok && (s.State == toc.StateOffTrack || s.State == toc.StateAtRisk) {
				why = append(why, fmt.Sprintf("indicator `%s` is %s (%s)", id, strings.ToLower(Label(string(s.State))), cites(s, in.Status.AsOf)))
			}
		}
		for _, p := range r.o.Preconditions {
			if ps := pathway[p]; ps == string(toc.StateOffTrack) || ps == string(toc.StateAtRisk) {
				why = append(why, fmt.Sprintf("held back by `%s` (%s)", p, strings.ToLower(Label(ps))))
			}
		}
		for _, a := range snapshot[r.o.ID].AssumptionBroken {
			why = append(why, fmt.Sprintf("assumption `%s` is broken or at risk, which caps what its pathway can be", a))
		}
		fmt.Fprintf(&b, "| %s (`%s`) | %s | %s |\n", MDCell(r.o.Title), r.o.ID, statusLabel(r.state), orDash(MDCell(strings.Join(why, "; "))))
	}
	b.WriteString("\n")
	return b.String()
}

// brokenAssumptions lists high-stakes assumptions that are invalidated, at
// risk, or still untested.
func brokenAssumptions(in StatusInput) string {
	var rows []AssumptionView
	for _, a := range in.View.Assumptions {
		high := a.Criticality == string(toc.CriticalityHigh) || a.Criticality == string(toc.CriticalityCritical)
		bad := a.Status == string(toc.AssumptionInvalidated) || a.Status == string(toc.AssumptionAtRisk) || a.Status == string(toc.AssumptionUntested) || a.Status == ""
		if high && bad {
			rows = append(rows, a)
		}
	}
	slices.SortStableFunc(rows, func(a, b AssumptionView) int {
		return assumptionRank(a.Status) - assumptionRank(b.Status)
	})
	var b strings.Builder
	b.WriteString("## Broken or untested critical assumptions\n\n")
	if len(rows) == 0 {
		b.WriteString("No high-stakes assumption is broken, at risk, or untested.\n\n")
		return b.String()
	}
	b.WriteString("| Assumption | Criticality | Status | How it will be tested |\n|------------|-------------|--------|-----------------------|\n")
	for _, a := range rows {
		fmt.Fprintf(&b, "| %s (`%s`) | %s | %s (%s) | %s |\n", MDCell(a.Statement), a.ID, Label(a.Criticality),
			statusLabel(a.Status), sourceText(a.StatusSource), orDash(MDCell(a.Test)))
	}
	b.WriteString("\n")
	return b.String()
}

func assumptionRank(s string) int {
	switch s {
	case string(toc.AssumptionInvalidated):
		return 0
	case string(toc.AssumptionAtRisk):
		return 1
	default:
		return 2
	}
}

func sourceText(s string) string {
	if s == string(toc.SourceAssessment) {
		return "assessed"
	}
	return "as written"
}

func changes(in StatusInput) string {
	if in.Diff == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "## Changes since %s\n\n", in.Diff.FromAsOf)
	if len(in.Diff.Changes) == 0 {
		b.WriteString("Nothing changed.\n\n")
		return b.String()
	}
	const shown = 20
	for i, c := range in.Diff.Changes {
		if i == shown {
			fmt.Fprintf(&b, "- …and %d more\n", len(in.Diff.Changes)-shown)
			break
		}
		fmt.Fprintf(&b, "- %s `%s` %s: %s → %s\n", c.ElementKind, c.ElementID, c.Field, dash(c.From), dash(c.To))
	}
	b.WriteString("\n")
	return b.String()
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// decisions lists the decisions still open, then a prompt for each broken
// assumption.
func decisions(in StatusInput) string {
	var lines []string
	for _, d := range in.Doc.Decisions {
		if d.Status == "" || d.Status == toc.DecisionOpen {
			line := mdText(d.Question)
			if d.DueDate != "" {
				line += " (by " + d.DueDate + ")"
			}
			lines = append(lines, line)
		}
	}
	for _, a := range in.View.Assumptions {
		if a.Status == string(toc.AssumptionInvalidated) || a.Status == string(toc.AssumptionAtRisk) {
			lines = append(lines, fmt.Sprintf("What should change now that the assumption \"%s\" (`%s`) is %s?", mdText(a.Statement), a.ID, strings.ToLower(Label(a.Status))))
		}
	}
	var b strings.Builder
	b.WriteString("## Decisions requested\n\n")
	if len(lines) == 0 {
		b.WriteString("None.\n\n")
		return b.String()
	}
	for _, l := range lines {
		b.WriteString("- " + l + "\n")
	}
	b.WriteString("\n")
	return b.String()
}

// Results renders the results framework: each outcome with its indicators,
// baseline, targets, latest value, and status, and the numbers behind the
// status.
func Results(in StatusInput) string {
	var b strings.Builder
	b.WriteString(header(in, "results"))
	byInd := in.indicatorStatus()
	indicators := map[string]IndicatorView{}
	for _, i := range in.View.Indicators {
		indicators[i.ID] = i
	}
	for _, lv := range toc.AllLevels() {
		var section strings.Builder
		for _, o := range in.View.Outcomes {
			if o.Level != lv {
				continue
			}
			fmt.Fprintf(&section, "### %s%s\n\n", mdText(o.Title), BasisMark(o.Basis))
			state := "—"
			if o.Status != nil {
				state = statusLabel(o.Status.PathwayState)
			}
			fmt.Fprintf(&section, "Pathway status: %s. Own indicators: %s.\n\n", state, orDash(strings.Join(o.Indicators, ", ")))
			if len(o.Indicators) == 0 {
				continue
			}
			section.WriteString("| Indicator | Baseline | Targets | Latest | Status | Because |\n|-----------|----------|---------|--------|--------|---------|\n")
			for _, id := range o.Indicators {
				iv, s := indicators[id], byInd[id]
				fmt.Fprintf(&section, "| %s | %s | %s | %s | %s | %s |\n", MDCell(iv.Name)+BasisMark(iv.Basis), orDash(MDCell(iv.Baseline)),
					orDash(MDCell(strings.Join(iv.Targets, "; "))), orDash(MDCell(iv.Latest)), statusLabel(string(s.State)), MDCell(cites(s, in.Status.AsOf)))
			}
			section.WriteString("\n")
		}
		if section.Len() > 0 {
			fmt.Fprintf(&b, "## %s\n\n%s", LevelLabel(lv), section.String())
		}
	}
	return b.String()
}

// Assumptions renders the assumption register with each assumption's current
// status and where that status came from.
func Assumptions(in StatusInput) string {
	var b strings.Builder
	b.WriteString(header(in, "assumptions"))
	evidence := map[string][]string{}
	for _, a := range in.Doc.Assumptions {
		evidence[a.ID] = a.EvidenceIDs
	}
	b.WriteString("| Assumption | Type | Criticality | Applies to | Test | Evidence | Status |\n|------------|------|-------------|------------|------|----------|--------|\n")
	for _, a := range in.View.Assumptions {
		fmt.Fprintf(&b, "| %s (`%s`)%s | %s | %s | %s | %s | %s | %s (%s) |\n", MDCell(a.Statement), a.ID, BasisMark(a.Basis),
			Label(a.Type), Label(a.Criticality), orDash(strings.Join(a.AppliesTo, ", ")), orDash(MDCell(a.Test)),
			orDash(strings.Join(evidence[a.ID], ", ")), statusLabel(a.Status), sourceText(a.StatusSource))
	}
	b.WriteString("\n")
	return b.String()
}

// Interventions renders each intervention with its owner, status, and the
// pathway status of what it is meant to bring about.
func Interventions(in StatusInput) string {
	var b strings.Builder
	b.WriteString(header(in, "interventions"))
	titles := in.titles()
	pathway := map[string]string{}
	for _, o := range in.View.Outcomes {
		if o.Status != nil {
			pathway[o.ID] = o.Status.PathwayState
		}
	}
	b.WriteString("| Intervention | Owner | Status | Targets | Pathway status of the targets |\n|--------------|-------|--------|---------|-------------------------------|\n")
	for _, iv := range in.View.Interventions {
		var targets, states []string
		for _, id := range iv.Targets {
			targets = append(targets, MDCell(titles[id]))
			states = append(states, statusLabel(pathway[id]))
		}
		fmt.Fprintf(&b, "| %s (`%s`)%s | %s | %s | %s | %s |\n", MDCell(iv.Title), iv.ID, BasisMark(iv.Basis), orDash(iv.Owner),
			statusLabel(iv.Status), orDash(strings.Join(targets, "; ")), orDash(strings.Join(states, "; ")))
	}
	b.WriteString("\n")
	return b.String()
}

// EvidenceGaps lists what is claimed without support: links with no
// evidence, high-stakes assumptions with none, and outcomes nothing measures.
func EvidenceGaps(in StatusInput) string {
	var b strings.Builder
	b.WriteString(header(in, "evidence gaps"))

	b.WriteString("## Links with no evidence\n\n")
	var links []string
	for _, l := range in.Doc.Links {
		if !l.Rejected() && len(l.EvidenceIDs) == 0 {
			links = append(links, fmt.Sprintf("`%s`: %s → %s", l.ID, in.titles()[l.From], in.titles()[l.To]))
		}
	}
	b.WriteString(bullets(links))

	b.WriteString("## High-stakes assumptions with no evidence\n\n")
	var asm []string
	for _, a := range in.Doc.Assumptions {
		if !a.Provenance.IsRejected() && (a.Criticality == toc.CriticalityHigh || a.Criticality == toc.CriticalityCritical) && len(a.EvidenceIDs) == 0 {
			asm = append(asm, fmt.Sprintf("`%s`: %s", a.ID, mdText(a.Statement)))
		}
	}
	b.WriteString(bullets(asm))

	b.WriteString("## Outcomes nothing measures\n\n")
	b.WriteString(bullets(outcomesWithoutIndicators(in.View, in.Doc)))
	return b.String()
}

// Full renders every status report in one document.
func Full(in StatusInput, mermaid string) string {
	var b strings.Builder
	b.WriteString(Exec(in))
	if mermaid != "" {
		b.WriteString("## Pathway\n\n```mermaid\n" + mermaid + "```\n\n")
	}
	for _, part := range []string{Results(in), Assumptions(in), Interventions(in), EvidenceGaps(in)} {
		// Each report opens with a title; inside the full report it is a section.
		b.WriteString(strings.Replace(part, "# ", "## ", 1))
	}
	return b.String()
}
