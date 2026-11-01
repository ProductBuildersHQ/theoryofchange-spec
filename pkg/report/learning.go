package report

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// LearningInput adds to a status report what a learning review needs: the
// observation sets of the period, those of the period before (to see which
// themes are new), and the period's label.
type LearningInput struct {
	StatusInput
	Sets     []*impact.ObservationSet
	PrevSets []*impact.ObservationSet
	Period   string
}

const (
	surprisesShown = 5
	changesShown   = 15
	questionsShown = 12
)

// LearningReview drafts the review of a period: what changed, where reality
// departed most from the plan, which assumptions moved, what the stories are
// about, how far the data can be trusted, and questions to discuss. It then
// carries any review already recorded for the period. It decides nothing; it
// prepares the conversation.
func LearningReview(in LearningInput) string {
	var b strings.Builder
	title := "learning review"
	if in.Period != "" {
		title += " " + in.Period
	}
	b.WriteString(header(in.StatusInput, title))

	b.WriteString("## What changed\n\n")
	b.WriteString(changeSummary(in.Diff))

	ahead, behind := Surprises(in.StatusInput)
	b.WriteString("## Surprises\n\nThe indicators that moved furthest from what the targets called for by this date.\n\n")
	b.WriteString(surpriseTable("Ahead of plan", ahead, in.Status.AsOf))
	b.WriteString(surpriseTable("Behind plan", behind, in.Status.AsOf))

	b.WriteString("## Assumptions that moved\n\n")
	b.WriteString(assumptionMoves(in.Diff, in.View))

	b.WriteString("## What the stories are about\n\n")
	b.WriteString(themeSection(in))

	b.WriteString("## How far to trust the data\n\n")
	b.WriteString(dataQuality(in))

	b.WriteString("## Questions to discuss\n\n")
	for _, q := range ReviewQuestions(in, ahead, behind) {
		b.WriteString("- " + q + "\n")
	}
	b.WriteString("\n")

	b.WriteString(recordedReview(in.Sets, in.Period))
	return b.String()
}

func changeSummary(diff *impact.StatusDiff) string {
	if diff == nil {
		return "No earlier snapshot was given, so there is nothing to compare with.\n\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Since %s, %d value(s) changed.\n\n", diff.FromAsOf, len(diff.Changes))
	for i, c := range diff.Changes {
		if i == changesShown {
			fmt.Fprintf(&b, "- …and %d more\n", len(diff.Changes)-changesShown)
			break
		}
		fmt.Fprintf(&b, "- %s `%s` %s: %s → %s\n", c.ElementKind, c.ElementID, c.Field, dash(c.From), dash(c.To))
	}
	b.WriteString("\n")
	return b.String()
}

// Surprise is an indicator that departed from the plan.
type Surprise struct {
	Name, ID string
	Progress float64
	Expected float64
	Gap      float64 // progress minus expected
}

// Surprises returns the indicators furthest ahead of and behind plan, each
// list ordered by the size of the gap and then ID, and cut to the top few.
func Surprises(in StatusInput) (ahead, behind []Surprise) {
	names := map[string]string{}
	for _, i := range in.View.Indicators {
		names[i.ID] = i.Name
	}
	for _, s := range in.Status.Indicators {
		if s.Progress == nil || s.Expected == nil {
			continue
		}
		sp := Surprise{Name: names[s.IndicatorID], ID: s.IndicatorID, Progress: *s.Progress, Expected: *s.Expected, Gap: *s.Progress - *s.Expected}
		switch {
		case sp.Gap > 0.005:
			ahead = append(ahead, sp)
		case sp.Gap < -0.005:
			behind = append(behind, sp)
		}
	}
	slices.SortStableFunc(ahead, func(a, b Surprise) int { return cmpGap(b.Gap, a.Gap, a.ID, b.ID) })
	slices.SortStableFunc(behind, func(a, b Surprise) int { return cmpGap(a.Gap, b.Gap, a.ID, b.ID) })
	return trim(ahead), trim(behind)
}

func trim(s []Surprise) []Surprise {
	if len(s) > surprisesShown {
		return s[:surprisesShown]
	}
	return s
}

func cmpGap(x, y float64, idX, idY string) int {
	switch {
	case x < y:
		return -1
	case x > y:
		return 1
	}
	return strings.Compare(idX, idY)
}

func surpriseTable(heading string, rows []Surprise, asOf string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "### %s\n\n", heading)
	if len(rows) == 0 {
		b.WriteString("None.\n\n")
		return b.String()
	}
	fmt.Fprintf(&b, "| Indicator | Progress | Expected by %s | Gap |\n|-----------|----------|%s|-----|\n", asOf, strings.Repeat("-", len(asOf)+14))
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s (`%s`) | %.0f%% | %.0f%% | %+.0f points |\n", MDCell(r.Name), r.ID, r.Progress*100, r.Expected*100, math.Round(r.Gap*100))
	}
	b.WriteString("\n")
	return b.String()
}

func assumptionMoves(diff *impact.StatusDiff, v *View) string {
	statements := map[string]string{}
	for _, a := range v.Assumptions {
		statements[a.ID] = a.Statement
	}
	var lines []string
	if diff != nil {
		for _, c := range diff.Changes {
			if c.ElementKind == impact.ElementAssumption && c.Field == "status" {
				lines = append(lines, fmt.Sprintf("`%s` went from %s to %s: %s", c.ElementID, dash(c.From), dash(c.To), mdText(statements[c.ElementID])))
			}
		}
	}
	return bullets(lines)
}

// ThemeCount is how many of the period's stories carry a theme.
type ThemeCount struct {
	Label string
	Count int
	// New is true when the theme did not appear in the period before. It is
	// never set when no earlier period was given.
	New bool
}

// ThemeCounts counts the period's stories by codebook theme, in the
// codebook's order, counting every story whether or not it may be shown.
func ThemeCounts(in LearningInput) []ThemeCount {
	asOf, err := time.Parse(time.DateOnly, in.Status.AsOf)
	if err != nil {
		return nil
	}
	count := func(sets []*impact.ObservationSet) map[string]int {
		out := map[string]int{}
		seen := map[string]bool{}
		for _, set := range sets {
			for _, s := range set.Stories {
				if at, err := time.Parse(time.DateOnly, s.Date); err != nil || at.After(asOf) || seen[s.ID] {
					continue
				}
				seen[s.ID] = true
				for _, t := range s.ThemeIDs {
					out[t]++
				}
			}
		}
		return out
	}
	now, before := count(in.Sets), count(in.PrevSets)
	var out []ThemeCount
	for _, t := range in.Doc.Themes {
		if now[t.ID] > 0 {
			out = append(out, ThemeCount{Label: t.Label, Count: now[t.ID], New: len(in.PrevSets) > 0 && before[t.ID] == 0})
		}
	}
	return out
}

// themeSection counts the period's stories by theme, marks the themes that
// did not appear in the period before, and lists the stories that may be
// told.
func themeSection(in LearningInput) string {
	asOf, err := time.Parse(time.DateOnly, in.Status.AsOf)
	if err != nil {
		return "The status date could not be read.\n\n"
	}
	var b strings.Builder
	var rows []string
	for _, tc := range ThemeCounts(in) {
		note := ""
		if tc.New {
			note = " (new)"
		}
		rows = append(rows, fmt.Sprintf("| %s%s | %d |", MDCell(tc.Label), note, tc.Count))
	}
	if len(rows) == 0 {
		b.WriteString("No stories were coded this period.\n\n")
	} else {
		b.WriteString("| Theme | Stories |\n|-------|---------|\n" + strings.Join(rows, "\n") + "\n\n")
	}
	shown, withheld := Disclose(in.Sets, asOf, nil)
	if len(shown) > 0 {
		b.WriteString("Stories that may be told:\n\n")
		for _, d := range shown {
			fmt.Fprintf(&b, "- **%s** (%s): %s\n", mdText(d.Story.Title), d.Story.Date, mdText(d.Story.Summary))
			if d.ShowQuote {
				fmt.Fprintf(&b, "  > %s\n", mdText(d.Story.Quote))
			}
		}
		b.WriteString("\n")
	}
	if withheld > 0 {
		fmt.Fprintf(&b, "%d stor%s not shown because consent has not been given.\n\n", withheld, plural(withheld, "y is", "ies are"))
	}
	return b.String()
}

// dataQuality counts the period's observations by their stated quality and
// lists the indicators that have gone stale or have no data.
func dataQuality(in LearningInput) string {
	counts := map[impact.DataQuality]int{}
	unrated, total := 0, 0
	for _, set := range in.Sets {
		for _, o := range set.Observations {
			total++
			if o.Quality == "" {
				unrated++
				continue
			}
			counts[o.Quality]++
		}
	}
	var b strings.Builder
	if total == 0 {
		b.WriteString("There are no observations in this period.\n\n")
	} else {
		var parts []string
		for _, q := range impact.AllDataQualities() {
			if counts[q] > 0 {
				parts = append(parts, fmt.Sprintf("%d %s", counts[q], q))
			}
		}
		if unrated > 0 {
			parts = append(parts, fmt.Sprintf("%d with no quality stated", unrated))
		}
		fmt.Fprintf(&b, "%d observations: %s.\n\n", total, strings.Join(parts, ", "))
	}
	var gaps []string
	for _, s := range in.Status.Indicators {
		if s.State == impact.StateStale || s.State == impact.StateNoData {
			gaps = append(gaps, fmt.Sprintf("`%s` is %s", s.IndicatorID, strings.ToLower(Label(string(s.State)))))
		}
	}
	if len(gaps) > 0 {
		b.WriteString("Indicators without current data:\n\n")
		b.WriteString(bullets(gaps))
	}
	return b.String()
}

// ReviewQuestions drafts questions from the surprises, the assumptions in
// trouble, and the indicators without data. The wording is fixed, so the
// same inputs give the same questions.
func ReviewQuestions(in LearningInput, ahead, behind []Surprise) []string {
	var qs []string
	for _, s := range behind {
		qs = append(qs, fmt.Sprintf("Why is \"%s\" (`%s`) behind where we expected it to be, and is the target or the work the thing to change?", mdText(s.Name), s.ID))
	}
	for _, s := range ahead {
		qs = append(qs, fmt.Sprintf("What is making \"%s\" (`%s`) go better than planned, and can it be repeated elsewhere?", mdText(s.Name), s.ID))
	}
	for _, a := range in.View.Assumptions {
		if a.Status == string(impact.AssumptionInvalidated) || a.Status == string(impact.AssumptionAtRisk) {
			qs = append(qs, fmt.Sprintf("What have we learned about \"%s\" (`%s`), which is now %s, and what does it change?", mdText(a.Statement), a.ID, strings.ToLower(Label(a.Status))))
		}
	}
	for _, s := range in.Status.Indicators {
		if s.State == impact.StateStale || s.State == impact.StateNoData {
			qs = append(qs, fmt.Sprintf("What is stopping us collecting `%s`, and is it still worth collecting?", s.IndicatorID))
		}
	}
	if len(qs) > questionsShown {
		qs = append(qs[:questionsShown], fmt.Sprintf("…and %d more questions follow from the sections above.", len(qs)-questionsShown))
	}
	if len(qs) == 0 {
		qs = append(qs, "Nothing stands out this period. What would we want to know that these numbers cannot tell us?")
	}
	return qs
}

// recordedReview merges the learning review already recorded for the period,
// if any.
func recordedReview(sets []*impact.ObservationSet, period string) string {
	var found []impact.LearningReview
	for _, set := range sets {
		for _, r := range set.Reviews {
			if period == "" || r.Period == period {
				found = append(found, r)
			}
		}
	}
	if len(found) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## What was recorded\n\n")
	for _, r := range found {
		for _, sec := range []struct {
			name  string
			items []string
		}{{"Questions asked", r.Questions}, {"Findings", r.Findings}, {"Decisions", r.Decisions}, {"Actions", r.Actions}} {
			if len(sec.items) == 0 {
				continue
			}
			fmt.Fprintf(&b, "**%s**\n\n", sec.name)
			for _, it := range sec.items {
				b.WriteString("- " + mdText(it) + "\n")
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}
