package report

import (
	"strings"
	"testing"
	"time"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
)

func exampleSets(t *testing.T) []*impact.ObservationSet {
	t.Helper()
	sets, err := impactio.LoadObservations("../../examples/intermediary/observations-2026-q3.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return sets
}

func TestStoryReportGolden(t *testing.T) {
	in := exampleInput(t)
	for _, aud := range []string{"aud-funders", "aud-board"} {
		out, err := Story(in, exampleSets(t), aud)
		if err != nil {
			t.Fatal(err)
		}
		golden(t, "story-"+aud+".md", out)
	}
}

func story(id, date, consent, sensitivity, quote string, audiences ...string) impact.Story {
	return impact.Story{
		ID: id, Date: date, Title: "Title " + id, Summary: "Summary " + id, Quote: quote,
		AudienceIDs: audiences, Consent: impact.Consent(consent), Sensitivity: impact.Sensitivity(sensitivity),
	}
}

func TestConsentAndSensitivityGateStoriesAndQuotes(t *testing.T) {
	in := exampleInput(t)
	sets := []*impact.ObservationSet{{Stories: []impact.Story{
		story("ok-public", "2026-03-01", "granted", "public", "QUOTE-OK-PUBLIC", "aud-funders"),
		story("ok-internal", "2026-03-02", "granted", "internal", "QUOTE-OK-INTERNAL", "aud-funders"),
		story("ok-restricted", "2026-03-03", "granted", "restricted", "QUOTE-RESTRICTED", "aud-funders"),
		story("ok-unset", "2026-03-04", "granted", "", "QUOTE-UNSET", "aud-funders"),
		story("no-consent", "2026-03-05", "not_granted", "public", "QUOTE-NO-CONSENT", "aud-funders"),
		story("unknown-consent", "2026-03-06", "unknown", "public", "QUOTE-UNKNOWN", "aud-funders"),
		story("other-audience", "2026-03-07", "granted", "public", "QUOTE-OTHER", "aud-board"),
		story("later", "2027-03-07", "granted", "public", "QUOTE-LATER", "aud-funders"),
	}}}
	out, err := Story(in, sets, "aud-funders")
	if err != nil {
		t.Fatal(err)
	}
	// Quotes appear only for consented stories that may be rendered verbatim.
	for _, want := range []string{"QUOTE-OK-PUBLIC", "QUOTE-OK-INTERNAL"} {
		if !strings.Contains(out, want) {
			t.Errorf("%s should be quoted", want)
		}
	}
	for _, leak := range []string{"QUOTE-RESTRICTED", "QUOTE-UNSET", "QUOTE-NO-CONSENT", "QUOTE-UNKNOWN", "QUOTE-OTHER", "QUOTE-LATER"} {
		if strings.Contains(out, leak) {
			t.Errorf("%s reached the report", leak)
		}
	}
	// A consented restricted story keeps its paraphrase but not its words.
	if !strings.Contains(out, "Summary ok-restricted") || !strings.Contains(out, "Summary ok-unset") {
		t.Error("the summary of a consented story is always shown")
	}
	// Stories without consent are withheld entirely, summary included, and counted.
	for _, gone := range []string{"Summary no-consent", "Title no-consent", "Summary unknown-consent", "Summary other-audience", "Summary later"} {
		if strings.Contains(out, gone) {
			t.Errorf("%q should not appear", gone)
		}
	}
	if !strings.Contains(out, "2 further stories for this audience are not shown because consent has not been given.") {
		t.Errorf("withheld count missing or wrong:\n%s", out)
	}
}

func TestOneWithheldStoryIsSingular(t *testing.T) {
	in := exampleInput(t)
	sets := []*impact.ObservationSet{{Stories: []impact.Story{story("a", "2026-03-01", "not_granted", "public", "", "aud-funders")}}}
	out, _ := Story(in, sets, "aud-funders")
	if !strings.Contains(out, "1 further story for this audience is not shown") {
		t.Errorf("singular form missing:\n%s", out)
	}
	if !strings.Contains(out, "No story that this audience may be told is available yet.") {
		t.Error("an empty audience should say so")
	}
}

func TestStoryReportCarriesCaveatsAndHowToDescribeNumbers(t *testing.T) {
	in := exampleInput(t)
	out, err := Story(in, exampleSets(t), "aud-funders")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "How to describe it") || !strings.Contains(out, "a contribution, not attributable to the work alone") {
		t.Errorf("claims are not carried into the report:\n%s", out)
	}
	if !strings.Contains(out, "Caveats that go with these numbers") {
		t.Errorf("caveats missing:\n%s", out)
	}
}

func TestThemeCountsUseOnlyShownStories(t *testing.T) {
	in := exampleInput(t)
	sets := []*impact.ObservationSet{{Stories: []impact.Story{
		{ID: "a", Date: "2026-03-01", Title: "a", Summary: "s", Consent: impact.ConsentGranted, Sensitivity: impact.SensitivityPublic, AudienceIDs: []string{"aud-funders"}, ThemeIDs: []string{"thm-peer-help"}},
		{ID: "b", Date: "2026-03-02", Title: "b", Summary: "s", Consent: impact.ConsentGranted, Sensitivity: impact.SensitivityPublic, AudienceIDs: []string{"aud-funders"}, ThemeIDs: []string{"thm-peer-help", "thm-shared-savings"}},
		{ID: "c", Date: "2026-03-03", Title: "c", Summary: "s", Consent: impact.ConsentNotGranted, Sensitivity: impact.SensitivityPublic, AudienceIDs: []string{"aud-funders"}, ThemeIDs: []string{"thm-shared-savings"}},
	}}}
	out, _ := Story(in, sets, "aud-funders")
	if !strings.Contains(out, "| Members helping each other | 2 |") || !strings.Contains(out, "| Savings from shared services | 1 |") {
		t.Errorf("theme counts wrong:\n%s", out)
	}
}

func TestUnknownAudience(t *testing.T) {
	if _, err := Story(exampleInput(t), nil, "aud-ghost"); err == nil {
		t.Error("an unknown audience was accepted")
	}
}

func TestDiscloseIsStableAndDeduplicates(t *testing.T) {
	asOf := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	sets := []*impact.ObservationSet{
		{Stories: []impact.Story{story("b", "2026-05-01", "granted", "public", "q"), story("a", "2026-05-01", "granted", "public", "q")}},
		{Stories: []impact.Story{story("a", "2026-05-01", "granted", "public", "q")}},
	}
	shown, withheld := Disclose(sets, asOf, nil)
	if len(shown) != 2 || withheld != 0 || shown[0].Story.ID != "a" || shown[1].Story.ID != "b" {
		t.Errorf("shown = %+v withheld %d", shown, withheld)
	}
}
