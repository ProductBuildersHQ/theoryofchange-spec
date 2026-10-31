package report

import (
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// numbersDoc reports three indicators to one audience, and one to nobody.
func numbersDoc() *toc.Document {
	return &toc.Document{
		Audiences: []toc.Audience{
			{ID: "aud-a", Name: "A"},
			{ID: "aud-b", Name: "B", Headline: []toc.HeadlineMetric{
				{IndicatorID: "ind-3", Label: "Third, first", Note: "why B cares"},
				{IndicatorID: "ind-gone"},
				{IndicatorID: "ind-missing"},
				{IndicatorID: "ind-1"},
			}},
		},
		Indicators: []toc.Indicator{
			{ID: "ind-1", Name: "One", AudienceIDs: []string{"aud-a", "aud-b"}},
			{ID: "ind-2", Name: "Two", AudienceIDs: []string{"aud-a"}},
			{ID: "ind-3", Name: "Three", AudienceIDs: []string{"aud-a", "aud-b"}},
			{ID: "ind-gone", Name: "Gone", AudienceIDs: []string{"aud-a"}, Provenance: &toc.Provenance{Basis: toc.BasisRejected}},
			{ID: "ind-none", Name: "None"},
		},
	}
}

func labels(ns []AudienceNumber) string {
	var out []string
	for _, n := range ns {
		out = append(out, n.Label)
	}
	return strings.Join(out, ",")
}

func TestAudienceNumbersFollowTheHeadlineList(t *testing.T) {
	ns, chosen := AudienceNumbers(numbersDoc(), "aud-b")
	if !chosen {
		t.Error("an audience with a headline list should be chosen")
	}
	// The list's order and wording win; a rejected or missing entry is skipped.
	if got := labels(ns); got != "Third, first,One" {
		t.Errorf("labels = %q", got)
	}
	if ns[0].Note != "why B cares" || ns[1].Note != "" {
		t.Errorf("notes = %q, %q", ns[0].Note, ns[1].Note)
	}
}

func TestAudienceNumbersFallBackToFileOrder(t *testing.T) {
	ns, chosen := AudienceNumbers(numbersDoc(), "aud-a")
	if chosen {
		t.Error("an audience without a headline list is not chosen")
	}
	if got := labels(ns); got != "One,Two,Three" {
		t.Errorf("labels = %q, want every live indicator in file order", got)
	}
}

func TestAudienceNumbersOfUnknownAudience(t *testing.T) {
	if ns, chosen := AudienceNumbers(numbersDoc(), "aud-ghost"); ns != nil || chosen {
		t.Errorf("got %v, %v", ns, chosen)
	}
}

// The story report tells an audience its chosen numbers, in the wording and
// order the headline list gives, and not every indicator reported to it.
func TestStoryReportUsesTheHeadlineList(t *testing.T) {
	in := exampleInput(t)
	var aud *toc.Audience
	for i := range in.Doc.Audiences {
		if in.Doc.Audiences[i].ID == "aud-funders" {
			aud = &in.Doc.Audiences[i]
		}
	}
	if aud == nil {
		t.Fatal("example has no aud-funders")
	}
	var reported []string
	for _, ind := range in.Doc.Indicators {
		for _, a := range ind.AudienceIDs {
			if a == "aud-funders" {
				reported = append(reported, ind.ID)
			}
		}
	}
	if len(reported) < 2 {
		t.Fatalf("need at least two indicators reported to funders, have %v", reported)
	}
	last := reported[len(reported)-1]
	aud.Headline = []toc.HeadlineMetric{{IndicatorID: last, Label: "THE ONE NUMBER FUNDERS GET"}}

	out, err := Story(in, exampleSets(t), "aud-funders")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "| THE ONE NUMBER FUNDERS GET |") {
		t.Errorf("the headline label is not used:\n%s", out)
	}
	if strings.Contains(out, "Attendee hours across all three program areas") {
		t.Errorf("an indicator outside the headline list was still told:\n%s", out)
	}
}
