package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// AudienceNumber is one number an audience is told.
type AudienceNumber struct {
	Indicator toc.Indicator
	// Label is the wording for this audience: the label its headline list gives,
	// or else the indicator's name.
	Label string
	// Note is why this audience is shown the number; only a headline list has one.
	Note string
}

// AudienceNumbers returns the indicators to tell an audience, in the order to
// tell them. When the audience has a headline list, that list is the answer,
// in its order and with its labels. Otherwise the answer is every indicator
// reported to the audience, in file order, and chosen is false.
//
// Every audience is told the same figure for an indicator: only the label and
// the note differ.
func AudienceNumbers(doc *toc.Document, audienceID string) (numbers []AudienceNumber, chosen bool) {
	a, ok := doc.Audience(audienceID)
	if !ok {
		return nil, false
	}
	if len(a.Headline) > 0 {
		for _, h := range a.Headline {
			ind, ok := doc.Indicator(h.IndicatorID)
			if !ok || ind.Rejected() {
				continue
			}
			label := h.Label
			if label == "" {
				label = ind.Name
			}
			numbers = append(numbers, AudienceNumber{Indicator: *ind, Label: label, Note: h.Note})
		}
		return numbers, true
	}
	for _, ind := range doc.Indicators {
		if !ind.Rejected() && slices.Contains(ind.AudienceIDs, audienceID) {
			numbers = append(numbers, AudienceNumber{Indicator: ind, Label: ind.Name})
		}
	}
	return numbers, false
}

// Headline renders the headline report: which numbers each audience is told,
// with the latest figure, its status, and the caveats that go with it. With an
// audience ID it covers that audience only; otherwise every audience with
// numbers. The same indicator carries the same figure for every audience, so
// the audiences differ only in what they lead with and how it is worded.
func Headline(in StatusInput, audienceID string) (string, error) {
	var auds []toc.Audience
	for _, a := range in.Doc.Audiences {
		if audienceID == "" || a.ID == audienceID {
			auds = append(auds, a)
		}
	}
	if audienceID != "" && len(auds) == 0 {
		return "", fmt.Errorf("report: unknown audience %q", audienceID)
	}
	byInd := in.indicatorStatus()
	views := map[string]IndicatorView{}
	for _, v := range in.View.Indicators {
		views[v.ID] = v
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# %s: headline metrics\n\nAs of %s.\n\n", mdText(in.View.Document.Title), in.Status.AsOf)

	type section struct {
		aud     toc.Audience
		numbers []AudienceNumber
		chosen  bool
	}
	var sections []section
	for _, a := range auds {
		ns, chosen := AudienceNumbers(in.Doc, a.ID)
		if len(ns) > 0 {
			sections = append(sections, section{a, ns, chosen})
		}
	}
	if len(sections) == 0 {
		b.WriteString("No indicator is reported to any audience.\n")
		return b.String(), nil
	}

	if len(sections) > 1 {
		b.WriteString("## Who is told what\n\n| Indicator |")
		for _, s := range sections {
			b.WriteString(" " + MDCell(s.aud.Name) + " |")
		}
		b.WriteString("\n|-----------|")
		for range sections {
			b.WriteString("---|")
		}
		b.WriteString("\n")
		var order []string
		seen := map[string]bool{}
		for _, s := range sections {
			for _, n := range s.numbers {
				if !seen[n.Indicator.ID] {
					seen[n.Indicator.ID] = true
					order = append(order, n.Indicator.ID)
				}
			}
		}
		for _, id := range order {
			ind, _ := in.Doc.Indicator(id)
			fmt.Fprintf(&b, "| %s |", MDCell(ind.Name))
			for _, s := range sections {
				cell := ""
				for i, n := range s.numbers {
					if n.Indicator.ID == id {
						cell = fmt.Sprintf("%d", i+1)
					}
				}
				b.WriteString(" " + cell + " |")
			}
			b.WriteString("\n")
		}
		b.WriteString("\nA number is the place the indicator takes in that audience's list.\n\n")
	}

	for _, s := range sections {
		fmt.Fprintf(&b, "## %s\n\n", mdText(s.aud.Name))
		if !s.chosen {
			b.WriteString("No headline list is chosen for this audience, so these are the indicators reported to it, in file order.\n\n")
		}
		b.WriteString("| Number | Latest | Status | How to describe it |\n|--------|--------|--------|--------------------|\n")
		var caveats, notes []string
		for _, n := range s.numbers {
			iv, st := views[n.Indicator.ID], byInd[n.Indicator.ID]
			claim := "attributable to the work"
			if n.Indicator.EffectiveClaim() == toc.ClaimContribution {
				claim = "a contribution, not attributable to the work alone"
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", MDCell(n.Label), orDash(MDCell(iv.Latest)), statusLabel(string(st.State)), claim)
			for _, c := range n.Indicator.Caveats {
				if !slices.Contains(caveats, c) {
					caveats = append(caveats, c)
				}
			}
			if n.Note != "" {
				notes = append(notes, fmt.Sprintf("%s: %s", n.Label, n.Note))
			}
			if iv.LatestNote != "" {
				notes = append(notes, fmt.Sprintf("%s: %s", n.Label, iv.LatestNote))
			}
		}
		b.WriteString("\n")
		writeList(&b, "Notes on these numbers", notes)
		writeList(&b, "Caveats that go with these numbers", caveats)
	}
	return b.String(), nil
}

func writeList(b *strings.Builder, heading string, items []string) {
	if len(items) == 0 {
		return
	}
	b.WriteString(heading + ":\n\n")
	for _, it := range items {
		b.WriteString("- " + mdText(it) + "\n")
	}
	b.WriteString("\n")
}
