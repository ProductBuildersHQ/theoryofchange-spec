package slides

import (
	"fmt"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

const (
	attentionPerSlide = 3
	changesShown      = 10
	questionsPerSlide = 4
	becauseWidth      = 150
	storiesPerSlide   = 2
)

// newDeck starts a deck with its Marp front matter.
func newDeck(title, theme string) *deck {
	if theme == "" {
		theme = "default"
	}
	d := &deck{}
	d.line("---")
	d.line("marp: true")
	d.line("theme: %s", theme)
	d.line("paginate: true")
	d.line("title: %q", title)
	d.line("---")
	return d
}

func (d *deck) titleSlide(name, kind, asOf, stamp string) {
	d.slide("lead")
	d.line("# %s", name)
	d.line("")
	d.line("**%s**", kind)
	d.line("")
	d.line("As of %s", asOf)
	if stamp != "" {
		d.line("")
		d.line("_%s_", stamp)
	}
}

// cut shortens text to fit a table cell on a slide.
func cut(s string, n int) string {
	s = oneLine(s)
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

func cell(s string) string { return report.MDCell(s) }

func label(s string) string {
	if s == "" {
		return "—"
	}
	return report.Label(s)
}

// Exec renders the executive deck: the headline, status by level, the
// pathway diagram colored by health, where attention is needed, broken
// assumptions, what changed, decisions requested, and how to read the
// numbers.
func Exec(in report.StatusInput, o Options) string {
	d := newDeck(in.View.Document.Title+": executive summary", o.Theme)
	d.titleSlide(in.View.Document.Title, "Executive summary", in.Status.AsOf, o.Stamp)
	d.notes("Lead with the headline. Say what the status is as of, and that every colour on the pathway comes from the status snapshot.")

	d.slide("")
	d.line("# Where we stand")
	d.line("")
	if len(in.Status.Overall) == 0 {
		d.line("This theory has no impact outcome yet.")
	}
	titles := map[string]string{}
	for _, oc := range in.View.Outcomes {
		titles[oc.ID] = oc.Title
	}
	for _, i := range in.Status.Overall {
		conf := ""
		if i.Confidence != nil {
			conf = fmt.Sprintf(" · confidence %.2f", *i.Confidence)
		}
		d.line("- **%s**: %s%s", oneLine(titles[i.OutcomeID]), label(string(i.PathwayState)), conf)
	}
	d.notes("Confidence is a rule-based score from evidence and assumptions, not a probability. It is explained on the last slide.")

	d.slide("")
	d.line("# Status by level")
	d.line("")
	d.line("| Level | Outcomes | Pathway status |")
	d.line("|-------|----------|----------------|")
	for _, lv := range toc.AllLevels() {
		counts := map[string]int{}
		n := 0
		for _, oc := range in.View.Outcomes {
			if oc.Level == lv && oc.Status != nil {
				counts[oc.Status.PathwayState]++
				n++
			}
		}
		if n > 0 {
			d.line("| %s | %d | %s |", report.LevelLabel(lv), n, report.StateCountText(counts))
		}
	}
	d.notes("Read from the bottom: early outcomes lead to later ones. A late outcome can be at risk because an early one is.")

	d.diagram("The pathway, coloured by status", report.Mermaid(in.View, report.MermaidOptions{}),
		"Green is on track or achieved, yellow is at risk or stale, red is off track, grey has no data yet.")

	items := report.Attention(in)
	for i, group := range chunk(items, attentionPerSlide) {
		d.slide("")
		d.line("# Where attention is needed%s", pageSuffix(i, len(items), attentionPerSlide))
		d.line("")
		d.line("| Outcome | Status | Because |")
		d.line("|---------|--------|---------|")
		for _, it := range group {
			d.line("| %s | %s | %s |", cell(cut(it.Title, 80)), label(it.State), cell(cut(it.Because, becauseWidth)))
		}
		d.notes("Start with the first row. Ask what would have to be true for it to recover, and who can make that happen.")
	}

	assumptions := report.HighStakesAssumptions(in)
	for i, group := range chunk(assumptions, attentionPerSlide) {
		d.slide("")
		d.line("# Assumptions we should watch%s", pageSuffix(i, len(assumptions), attentionPerSlide))
		d.line("")
		for _, a := range group {
			d.line("**%s**", oneLine(a.Statement))
			d.line("")
			d.line("- %s · %s", label(a.Criticality), label(a.Status))
			if a.Test != "" {
				d.line("- To test it: %s", cut(a.Test, 160))
			}
			d.line("")
		}
		d.notes("These are the beliefs the theory leans on that are broken, at risk, or still untested.")
	}

	if in.Diff != nil && len(in.Diff.Changes) > 0 {
		d.slide("")
		d.line("# What changed since %s", in.Diff.FromAsOf)
		d.line("")
		for i, c := range in.Diff.Changes {
			if i == changesShown {
				d.line("- …and %d more", len(in.Diff.Changes)-changesShown)
				break
			}
			d.line("- %s `%s` %s: %s → %s", c.ElementKind, c.ElementID, c.Field, dash(c.From), dash(c.To))
		}
		d.notes("Ask which of these were expected and which were a surprise.")
	}

	if prompts := report.DecisionPrompts(in); len(prompts) > 0 {
		d.slide("")
		d.line("# Decisions we need")
		d.line("")
		for _, p := range prompts {
			d.line("- %s", cut(p, 200))
		}
		d.notes("Leave time for these. Each one is something only the people in the room can decide.")
	}

	d.slide("")
	d.line("# How to read the numbers")
	d.line("")
	w := in.Status.Options.Weights
	d.line("- Status compares progress with what the dated targets call for, within a tolerance of %.0f%%.", in.Status.Options.Tolerance*100)
	d.line("- An outcome takes the %s of its indicators.", in.Status.Options.Rollup)
	d.line("- Confidence in a link is %.1f × evidence + %.1f × assumptions + %.1f × rationale, from 0 to 1.", w.Evidence, w.Assumptions, w.Rationale)
	d.notes("These are the settings the snapshot was computed with. The same inputs always give the same numbers.")
	return d.b.String()
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// LearningReview renders the learning-review deck: what changed, the
// surprises, assumptions that moved, themes and the stories that may be told,
// how far to trust the data, questions for the room, and what was decided.
func LearningReview(in report.LearningInput, o Options) string {
	title := "learning review"
	if in.Period != "" {
		title += " " + in.Period
	}
	d := newDeck(in.View.Document.Title+": "+title, o.Theme)
	d.titleSlide(in.View.Document.Title, strings.ToUpper(title[:1])+title[1:], in.Status.AsOf, o.Stamp)
	d.notes("A learning review is for finding out what we got wrong. Say so at the start.")

	d.slide("")
	d.line("# What changed")
	d.line("")
	if in.Diff == nil {
		d.line("No earlier snapshot was given.")
	} else {
		d.line("%d value(s) changed since %s.", len(in.Diff.Changes), in.Diff.FromAsOf)
		d.line("")
		for i, c := range in.Diff.Changes {
			if i == changesShown {
				d.line("- …and %d more", len(in.Diff.Changes)-changesShown)
				break
			}
			d.line("- %s `%s` %s: %s → %s", c.ElementKind, c.ElementID, c.Field, dash(c.From), dash(c.To))
		}
	}
	d.notes("Read the list once, then move on. The surprises are on the next slide.")

	ahead, behind := report.Surprises(in.StatusInput)
	d.surprises("Behind plan", behind)
	d.surprises("Ahead of plan", ahead)

	var moves []string
	if in.Diff != nil {
		for _, c := range in.Diff.Changes {
			if c.ElementKind == toc.ElementAssumption && c.Field == "status" {
				moves = append(moves, fmt.Sprintf("`%s`: %s → %s", c.ElementID, dash(c.From), dash(c.To)))
			}
		}
	}
	if len(moves) > 0 {
		d.slide("")
		d.line("# Assumptions that moved")
		d.line("")
		for _, m := range moves {
			d.line("- %s", m)
		}
		d.notes("An assumption that moved is the theory telling us something. Ask what it changes.")
	}

	d.themesAndStories(in)

	d.slide("")
	d.line("# How far to trust the data")
	d.line("")
	stale := 0
	for _, s := range in.Status.Indicators {
		if s.State == toc.StateStale || s.State == toc.StateNoData {
			stale++
		}
	}
	d.line("- %d indicator(s) have no current data.", stale)
	d.line("- Observations carry a stated quality: verified, reported, or estimated.")
	d.notes("Anything estimated or stale should be said out loud before it is used to decide something.")

	qs := report.ReviewQuestions(in, ahead, behind)
	for i, group := range chunk(qs, questionsPerSlide) {
		d.slide("")
		d.line("# For discussion%s", pageSuffix(i, len(qs), questionsPerSlide))
		d.line("")
		for _, q := range group {
			d.line("- %s", cut(q, 220))
		}
		d.notes("Pick two or three. Do not try to cover them all.")
	}

	d.recorded(in)
	return d.b.String()
}

func (d *deck) surprises(heading string, rows []report.Surprise) {
	if len(rows) == 0 {
		return
	}
	d.slide("")
	d.line("# %s", heading)
	d.line("")
	d.line("| Indicator | Progress | Expected | Gap |")
	d.line("|-----------|----------|----------|-----|")
	for _, r := range rows {
		d.line("| %s | %.0f%% | %.0f%% | %+.0f points |", cell(cut(r.Name, 70)), r.Progress*100, r.Expected*100, (r.Gap*100)+0)
	}
	d.notes("Ask why. A gap is as often a bad target as bad work.")
}

func (d *deck) themesAndStories(in report.LearningInput) {
	counts := report.ThemeCounts(in)
	asOf, err := time.Parse(time.DateOnly, in.Status.AsOf)
	if err != nil {
		return
	}
	shown, _ := report.Disclose(in.Sets, asOf, nil)
	if len(counts) == 0 && len(shown) == 0 {
		return
	}
	d.slide("")
	d.line("# What the stories are about")
	d.line("")
	d.line("| Theme | Stories |")
	d.line("|-------|---------|")
	for _, c := range counts {
		note := ""
		if c.New {
			note = " (new)"
		}
		d.line("| %s%s | %d |", cell(c.Label), note, c.Count)
	}
	d.notes("A new theme is worth a conversation: it may be something the theory did not expect.")
	for i, group := range chunk(shown, storiesPerSlide) {
		d.slide("")
		d.line("# Stories%s", pageSuffix(i, len(shown), storiesPerSlide))
		d.line("")
		for _, s := range group {
			d.line("**%s** (%s)", oneLine(s.Story.Title), s.Story.Date)
			d.line("")
			d.line("%s", cut(s.Story.Summary, 260))
			if s.ShowQuote {
				d.line("")
				d.line("> %s", oneLine(s.Story.Quote))
			}
			d.line("")
		}
		d.notes("Only stories with granted consent appear. Quotes appear only where the story may be quoted.")
	}
}

func (d *deck) recorded(in report.LearningInput) {
	var found []toc.LearningReview
	for _, set := range in.Sets {
		for _, r := range set.Reviews {
			if in.Period == "" || r.Period == in.Period {
				found = append(found, r)
			}
		}
	}
	for _, r := range found {
		if len(r.Decisions) == 0 && len(r.Actions) == 0 {
			continue
		}
		d.slide("")
		d.line("# What we decided")
		d.line("")
		for _, x := range r.Decisions {
			d.line("- **Decision.** %s", cut(x, 200))
		}
		for _, x := range r.Actions {
			d.line("- **Action.** %s", cut(x, 200))
		}
		d.notes("These were recorded for the period. Check they are still right.")
	}
}

// Story renders the audience story deck: what the audience needs, three to
// five headline numbers with their caveats, one or two stories they may be
// told, and a closing ask. A story appears only with granted consent, and a
// quote only where the story may be quoted.
func Story(in report.StatusInput, sets []*toc.ObservationSet, audienceID string, o Options) (string, error) {
	aud, ok := in.Doc.Audience(audienceID)
	if !ok {
		return "", fmt.Errorf("slides: unknown audience %q", audienceID)
	}
	asOf, err := time.Parse(time.DateOnly, in.Status.AsOf)
	if err != nil {
		return "", fmt.Errorf("slides: status date %q: %w", in.Status.AsOf, err)
	}
	d := newDeck(in.View.Document.Title+": "+aud.Name, o.Theme)
	d.titleSlide(in.View.Document.Title, "For "+aud.Name, in.Status.AsOf, o.Stamp)
	d.notes("Written for this audience. Everything here is what they asked to see.")

	d.slide("")
	d.line("# What you told us you need")
	d.line("")
	d.line("%s", oneLine(aud.Needs))
	d.notes("Say this back to them before showing a number.")

	byInd := map[string]toc.IndicatorStatus{}
	for _, s := range in.Status.Indicators {
		byInd[s.IndicatorID] = s
	}
	views := map[string]report.IndicatorView{}
	for _, v := range in.View.Indicators {
		views[v.ID] = v
	}
	var caveats []string
	// An audience's chosen headline list is shown whole, in its order and
	// wording. Without one, the first few indicators reported to it are shown.
	const unchosenLimit = 5
	rows, chosen := report.AudienceNumbers(in.Doc, audienceID)
	if !chosen && len(rows) > unchosenLimit {
		rows = rows[:unchosenLimit]
	}
	if len(rows) > 0 {
		d.slide("")
		d.line("# The numbers")
		d.line("")
		d.line("| Measure | Latest | Status |")
		d.line("|---------|--------|--------|")
		for _, n := range rows {
			ind := n.Indicator
			v := views[ind.ID]
			note := ""
			if ind.EffectiveClaim() == toc.ClaimContribution {
				note = " *"
			}
			d.line("| %s%s | %s | %s |", cell(cut(n.Label, 70)), note, cell(dashText(v.Latest)), label(string(byInd[ind.ID].State)))
			for _, c := range ind.Caveats {
				if !containsString(caveats, c) {
					caveats = append(caveats, c)
				}
			}
		}
		d.line("")
		d.line("\\* A contribution: the result also depends on others.")
		for _, c := range caveats {
			d.line("- %s", cut(c, 160))
		}
		d.notes("Use the numbers as given. Say 'contributed to' for anything marked.")
	}

	shown, _ := report.Disclose(sets, asOf, func(s toc.Story) bool { return containsString(s.AudienceIDs, audienceID) })
	for i, dis := range shown {
		if i == storiesPerSlide {
			break
		}
		d.slide("")
		d.line("# %s", oneLine(dis.Story.Title))
		d.line("")
		d.line("%s", cut(dis.Story.Summary, 320))
		if dis.ShowQuote {
			d.line("")
			d.line("> %s", oneLine(dis.Story.Quote))
		}
		d.notes("Shared with consent. Do not add a name.")
	}

	d.slide("lead")
	d.line("# What we ask of you")
	d.line("")
	d.line("_Fill in the request for this audience._")
	d.notes("End on a specific ask: a decision, an introduction, a commitment.")
	return d.b.String(), nil
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func dashText(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
