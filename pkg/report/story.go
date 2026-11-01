package report

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// Disclosure is a story as a report may show it.
type Disclosure struct {
	Story impact.Story
	// ShowQuote is true only when consent is granted and the story's
	// sensitivity may be rendered verbatim.
	ShowQuote bool
}

// Disclose decides which of the stories a report may show. A story is shown
// only with granted consent: one whose consent is not granted or unknown is
// withheld, its summary and quote both, and counted in Withheld. A shown
// story's summary is a paraphrase and is always shown; its quote appears
// only if its sensitivity may be rendered verbatim, so a restricted story
// appears without the words of the person. Stories dated after asOf are not
// considered. The order is by date, then ID, so output is stable.
func Disclose(sets []*impact.ObservationSet, asOf time.Time, keep func(impact.Story) bool) (shown []Disclosure, withheld int) {
	seen := map[string]bool{}
	var all []impact.Story
	for _, set := range sets {
		for _, s := range set.Stories {
			if seen[s.ID] {
				continue
			}
			if at, err := time.Parse(time.DateOnly, s.Date); err != nil || at.After(asOf) {
				continue
			}
			if keep != nil && !keep(s) {
				continue
			}
			seen[s.ID] = true
			all = append(all, s)
		}
	}
	slices.SortStableFunc(all, func(a, b impact.Story) int {
		if c := strings.Compare(a.Date, b.Date); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})
	for _, s := range all {
		if s.Consent != impact.ConsentGranted {
			withheld++
			continue
		}
		shown = append(shown, Disclosure{Story: s, ShowQuote: s.Quote != "" && s.Sensitivity.Renderable()})
	}
	return shown, withheld
}

// Story renders the audience story report: what the audience needs to see,
// the numbers and outcomes reported to it, the stories it may be told, and
// how the stories fall under the codebook's themes. It is a scaffold for
// someone to turn into prose; every number comes from the snapshot.
func Story(in StatusInput, sets []*impact.ObservationSet, audienceID string) (string, error) {
	aud, ok := in.Doc.Audience(audienceID)
	if !ok {
		return "", fmt.Errorf("report: unknown audience %q", audienceID)
	}
	asOf, err := time.Parse(time.DateOnly, in.Status.AsOf)
	if err != nil {
		return "", fmt.Errorf("report: status date %q: %w", in.Status.AsOf, err)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s: story for %s\n\nAs of %s.\n\n", mdText(in.View.Document.Title), mdText(aud.Name), in.Status.AsOf)

	b.WriteString("## What this audience needs\n\n")
	b.WriteString(orDash(mdText(aud.Needs)) + "\n\n")
	if len(aud.Channels) > 0 {
		fmt.Fprintf(&b, "Reached through: %s.\n\n", strings.Join(aud.Channels, ", "))
	}

	b.WriteString(audienceNumbers(in, audienceID))
	b.WriteString(audienceOutcomes(in, audienceID))

	shown, withheld := Disclose(sets, asOf, func(s impact.Story) bool { return slices.Contains(s.AudienceIDs, audienceID) })
	b.WriteString("## Stories\n\n")
	if len(shown) == 0 {
		b.WriteString("No story that this audience may be told is available yet.\n\n")
	}
	for _, d := range shown {
		s := d.Story
		fmt.Fprintf(&b, "### %s\n\n_%s_\n\n%s\n\n", mdText(s.Title), s.Date, mdText(s.Summary))
		if d.ShowQuote {
			fmt.Fprintf(&b, "> %s\n\n", mdText(s.Quote))
		}
		if s.Significance != "" {
			fmt.Fprintf(&b, "Why it was chosen: %s\n\n", mdText(s.Significance))
		}
	}
	if withheld > 0 {
		fmt.Fprintf(&b, "%d further stor%s for this audience %s not shown because consent has not been given.\n\n", withheld, plural(withheld, "y", "ies"), plural(withheld, "is", "are"))
	}

	b.WriteString(themeCounts(in.Doc, shown))
	b.WriteString("## For the writer\n\n")
	b.WriteString("- Use the numbers as given; each is the latest value in the status snapshot above.\n")
	b.WriteString("- Say \"contributed to\" for any figure marked as a contribution.\n")
	b.WriteString("- Quote only the words shown here, and do not add a name.\n")
	return b.String(), nil
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// audienceNumbers lists the numbers the audience is told, with their latest
// value and status, and the caveats that must travel with them. They are the
// audience's chosen headline list, in its order and wording, or else every
// indicator reported to it.
func audienceNumbers(in StatusInput, audienceID string) string {
	var b strings.Builder
	b.WriteString("## Numbers for this audience\n\n")
	byInd := in.indicatorStatus()
	views := map[string]IndicatorView{}
	for _, v := range in.View.Indicators {
		views[v.ID] = v
	}
	var rows []string
	var caveats []string
	numbers, _ := AudienceNumbers(in.Doc, audienceID)
	for _, n := range numbers {
		ind := n.Indicator
		iv, s := views[ind.ID], byInd[ind.ID]
		claim := "attributable to the work"
		if ind.EffectiveClaim() == impact.ClaimContribution {
			claim = "a contribution, not attributable to the work alone"
		}
		rows = append(rows, fmt.Sprintf("| %s | %s | %s | %s |", MDCell(n.Label), orDash(MDCell(iv.Latest)), statusLabel(string(s.State)), claim))
		for _, c := range ind.Caveats {
			if !slices.Contains(caveats, c) {
				caveats = append(caveats, c)
			}
		}
	}
	if len(rows) == 0 {
		b.WriteString("No indicator is reported to this audience.\n\n")
		return b.String()
	}
	b.WriteString("| Indicator | Latest | Status | How to describe it |\n|-----------|--------|--------|--------------------|\n")
	for _, r := range rows {
		b.WriteString(r + "\n")
	}
	b.WriteString("\n")
	if len(caveats) > 0 {
		b.WriteString("Caveats that go with these numbers:\n\n")
		for _, c := range caveats {
			b.WriteString("- " + mdText(c) + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func audienceOutcomes(in StatusInput, audienceID string) string {
	var b strings.Builder
	var rows []string
	for _, o := range in.Doc.Outcomes {
		if o.Rejected() || !slices.Contains(o.AudienceIDs, audienceID) {
			continue
		}
		state := "—"
		for _, ov := range in.View.Outcomes {
			if ov.ID == o.ID && ov.Status != nil {
				state = statusLabel(ov.Status.PathwayState)
			}
		}
		rows = append(rows, fmt.Sprintf("| %s | %s |", MDCell(o.Title), state))
	}
	if len(rows) == 0 {
		return ""
	}
	b.WriteString("## Outcomes for this audience\n\n| Outcome | Status |\n|---------|--------|\n")
	for _, r := range rows {
		b.WriteString(r + "\n")
	}
	b.WriteString("\n")
	return b.String()
}

// themeCounts counts the shown stories under each codebook theme, in the
// codebook's order.
func themeCounts(doc *impact.Model, shown []Disclosure) string {
	counts := map[string]int{}
	for _, d := range shown {
		for _, id := range d.Story.ThemeIDs {
			counts[id]++
		}
	}
	var rows []string
	for _, t := range doc.Themes {
		if counts[t.ID] > 0 {
			rows = append(rows, fmt.Sprintf("| %s | %d |", MDCell(t.Label), counts[t.ID]))
		}
	}
	if len(rows) == 0 {
		return ""
	}
	return "## Themes in these stories\n\n| Theme | Stories |\n|-------|---------|\n" + strings.Join(rows, "\n") + "\n\n"
}
