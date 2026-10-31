package slides

import (
	"strings"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/measure"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

func statusInput(t *testing.T) (report.StatusInput, []*toc.ObservationSet) {
	t.Helper()
	doc, err := tocio.LoadDocument("../../examples/intermediary/theory.yaml")
	if err != nil {
		t.Fatal(err)
	}
	g, err := graph.Build(doc)
	if err != nil {
		t.Fatal(err)
	}
	sets, err := tocio.LoadObservations("../../examples/intermediary/observations-2026-q3.yaml")
	if err != nil {
		t.Fatal(err)
	}
	at := func(s string) time.Time {
		d, err := time.Parse(time.DateOnly, s)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	prev, err := measure.Evaluate(doc, g, sets, at("2026-06-30"), toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	cur, err := measure.Evaluate(doc, g, sets, at("2026-09-30"), toc.DefaultStatusOptions())
	if err != nil {
		t.Fatal(err)
	}
	v, err := report.BuildView(doc, g, cur, report.Options{})
	if err != nil {
		t.Fatal(err)
	}
	return report.StatusInput{View: v, Doc: doc, Status: cur, Diff: measure.Diff(prev, cur)}, sets
}

func TestStatusDeckGoldens(t *testing.T) {
	in, sets := statusInput(t)
	golden(t, "exec.md", Exec(in, Options{Stamp: "Theory of Change riverbend-network · version 0.2 · hash abc123"}))
	golden(t, "learning-review.md", LearningReview(report.LearningInput{StatusInput: in, Sets: sets, Period: "2026-Q3"}, Options{}))
	story, err := Story(in, sets, "aud-funders", Options{})
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "story.md", story)
}

func TestDecksAreWellFormedMarp(t *testing.T) {
	in, sets := statusInput(t)
	story, _ := Story(in, sets, "aud-board", Options{})
	for name, md := range map[string]string{
		"exec": Exec(in, Options{}), "learning review": LearningReview(report.LearningInput{StatusInput: in, Sets: sets, Period: "2026-Q3"}, Options{}), "story": story,
	} {
		if !strings.HasPrefix(md, "---\nmarp: true\n") {
			t.Errorf("%s: no Marp front matter", name)
		}
		if strings.Contains(md, "---\n\n---") || strings.Contains(md, "\n---\n\n\n---\n") {
			t.Errorf("%s: an empty slide", name)
		}
		if n := SlideCount(md); n < 4 || n > 40 {
			t.Errorf("%s: %d slides", name, n)
		}
		if strings.Count(md, "<!--\n") == 0 {
			t.Errorf("%s: no speaker notes", name)
		}
	}
}

func TestExecDeckContent(t *testing.T) {
	in, _ := statusInput(t)
	md := Exec(in, Options{})
	for _, want := range []string{"# Where we stand", "# Status by level", "```mermaid", "# Where attention is needed", "# What changed since 2026-06-30", "# How to read the numbers", "0.4 × evidence"} {
		if !strings.Contains(md, want) {
			t.Errorf("missing %q", want)
		}
	}
	in.Diff = nil
	if strings.Contains(Exec(in, Options{}), "# What changed") {
		t.Error("a change slide with nothing to compare")
	}
}

func TestStoryDeckHonoursConsent(t *testing.T) {
	in, _ := statusInput(t)
	story := func(id, consent, sensitivity, quote string) toc.Story {
		return toc.Story{ID: id, Date: "2026-03-01", Title: "Title " + id, Summary: "Summary " + id, Quote: quote,
			AudienceIDs: []string{"aud-funders"}, Consent: toc.Consent(consent), Sensitivity: toc.Sensitivity(sensitivity)}
	}
	sets := []*toc.ObservationSet{{Stories: []toc.Story{
		story("ok", "granted", "public", "QUOTE-OK"), story("restricted", "granted", "restricted", "QUOTE-RESTRICTED"),
		story("refused", "not_granted", "public", "QUOTE-REFUSED"), story("unknown", "unknown", "public", "QUOTE-UNKNOWN"),
	}}}
	md, err := Story(in, sets, "aud-funders", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, "QUOTE-OK") || !strings.Contains(md, "Summary ok") {
		t.Error("a consented, public story should be told")
	}
	for _, leak := range []string{"QUOTE-RESTRICTED", "QUOTE-REFUSED", "QUOTE-UNKNOWN", "Summary refused", "Summary unknown", "Title refused"} {
		if strings.Contains(md, leak) {
			t.Errorf("%q reached the deck", leak)
		}
	}
	if _, err := Story(in, nil, "aud-ghost", Options{}); err == nil {
		t.Error("an unknown audience was accepted")
	}
}

func TestLearningReviewDeckGatesStoriesAndMergesTheRecord(t *testing.T) {
	in, sets := statusInput(t)
	sets = append(sets, &toc.ObservationSet{Stories: []toc.Story{
		{ID: "x", Date: "2026-03-01", Title: "Hidden", Summary: "HIDDEN-SUMMARY", Quote: "HIDDEN-QUOTE", Consent: toc.ConsentUnknown, Sensitivity: toc.SensitivityPublic},
	}})
	md := LearningReview(report.LearningInput{StatusInput: in, Sets: sets, Period: "2026-Q3"}, Options{})
	if strings.Contains(md, "HIDDEN") {
		t.Error("an unconsented story reached the learning-review deck")
	}
	if !strings.Contains(md, "# What we decided") || !strings.Contains(md, "Trial two daytime sessions in the north next quarter.") {
		t.Error("the recorded decision is missing")
	}
}

func TestDecksAreDeterministic(t *testing.T) {
	a, _ := statusInput(t)
	b, _ := statusInput(t)
	if Exec(a, Options{}) != Exec(b, Options{}) {
		t.Error("two renders differ")
	}
}

// A chosen headline list is told whole, in its order and wording; without one
// the deck tells the first few indicators reported to the audience.
func TestStoryDeckUsesTheHeadlineList(t *testing.T) {
	in, sets := statusInput(t)
	var aud *toc.Audience
	for i := range in.Doc.Audiences {
		if in.Doc.Audiences[i].ID == "aud-funders" {
			aud = &in.Doc.Audiences[i]
		}
	}
	if aud == nil || len(in.Doc.Indicators) == 0 {
		t.Fatal("example has no aud-funders or no indicators")
	}
	aud.Headline = []toc.HeadlineMetric{{IndicatorID: in.Doc.Indicators[len(in.Doc.Indicators)-1].ID, Label: "DECK HEADLINE LABEL"}}
	md, err := Story(in, sets, "aud-funders", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, "DECK HEADLINE LABEL") {
		t.Errorf("the headline label is not in the deck:\n%s", md)
	}
}
