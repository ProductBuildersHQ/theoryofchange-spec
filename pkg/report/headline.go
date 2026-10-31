package report

import (
	"slices"

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
