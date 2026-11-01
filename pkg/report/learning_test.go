package report

import (
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/burden"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

func learningInput(t *testing.T) LearningInput {
	t.Helper()
	return LearningInput{StatusInput: exampleInput(t), Sets: exampleSets(t), Period: "2026-Q3"}
}

func TestLearningReviewGolden(t *testing.T) {
	golden(t, "learning-review.md", LearningReview(learningInput(t)))
}

func TestPortfolioGolden(t *testing.T) {
	in := exampleInput(t)
	cost, err := burden.Compute(in.Doc)
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "portfolio.md", Portfolio(in, cost))
}

func TestLearningReviewSections(t *testing.T) {
	out := LearningReview(learningInput(t))
	for _, want := range []string{
		"# Riverbend Nonprofit Network theory of change: learning review 2026-Q3", "## What changed", "## Surprises", "### Ahead of plan",
		"### Behind plan", "## Assumptions that moved", "## What the stories are about", "## How far to trust the data",
		"## Questions to discuss", "## What was recorded", "Trial two daytime sessions in the north next quarter.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestLearningReviewOnlyMergesTheNamedPeriod(t *testing.T) {
	in := learningInput(t)
	in.Period = "2025-Q1"
	if strings.Contains(LearningReview(in), "## What was recorded") {
		t.Error("a review from another period was merged")
	}
}

func TestLearningReviewGatesStoriesLikeTheStoryReport(t *testing.T) {
	in := learningInput(t)
	in.Sets = []*impact.ObservationSet{{Stories: []impact.Story{
		{ID: "a", Date: "2026-03-01", Title: "Granted", Summary: "Told", Quote: "QUOTE-A", Consent: impact.ConsentGranted, Sensitivity: impact.SensitivityPublic, ThemeIDs: []string{"thm-peer-help"}},
		{ID: "b", Date: "2026-03-02", Title: "Refused", Summary: "SUMMARY-B", Quote: "QUOTE-B", Consent: impact.ConsentNotGranted, Sensitivity: impact.SensitivityPublic, ThemeIDs: []string{"thm-peer-help"}},
		{ID: "c", Date: "2026-03-03", Title: "Restricted", Summary: "Told too", Quote: "QUOTE-C", Consent: impact.ConsentGranted, Sensitivity: impact.SensitivityRestricted},
	}}}
	out := LearningReview(in)
	if !strings.Contains(out, "QUOTE-A") || strings.Contains(out, "QUOTE-B") || strings.Contains(out, "QUOTE-C") || strings.Contains(out, "SUMMARY-B") {
		t.Errorf("quote gating failed:\n%s", out)
	}
	// Theme counts count every coded story, shown or not.
	if !strings.Contains(out, "| Members helping each other | 2 |") {
		t.Errorf("theme counts:\n%s", out)
	}
	if !strings.Contains(out, "1 story is not shown because consent has not been given.") {
		t.Errorf("withheld note:\n%s", out)
	}
}

func TestNewThemesAreMarked(t *testing.T) {
	in := learningInput(t)
	story := func(id, theme string) impact.Story {
		return impact.Story{ID: id, Date: "2026-03-01", Title: id, Summary: "s", Consent: impact.ConsentGranted, Sensitivity: impact.SensitivityPublic, ThemeIDs: []string{theme}}
	}
	in.Sets = []*impact.ObservationSet{{Stories: []impact.Story{story("a", "thm-peer-help"), story("b", "thm-shared-savings")}}}
	in.PrevSets = []*impact.ObservationSet{{Stories: []impact.Story{story("old", "thm-peer-help")}}}
	out := LearningReview(in)
	if !strings.Contains(out, "| Savings from shared services (new) | 1 |") || strings.Contains(out, "Members helping each other (new)") {
		t.Errorf("new themes:\n%s", out)
	}
	in.PrevSets = nil
	if strings.Contains(LearningReview(in), "(new)") {
		t.Error("without a previous period nothing is new")
	}
}

func TestLearningReviewWithoutADiffOrObservations(t *testing.T) {
	in := learningInput(t)
	in.Diff, in.Sets = nil, nil
	out := LearningReview(in)
	for _, want := range []string{"No earlier snapshot was given", "There are no observations in this period.", "No stories were coded this period."} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestPortfolioDescribesWithoutRecommending(t *testing.T) {
	in := exampleInput(t)
	cost, _ := burden.Compute(in.Doc)
	out := Portfolio(in, cost)
	for _, row := range []string{"Learning Exchange", "Shared Services", "Civic Voice"} {
		if !strings.Contains(out, row) {
			t.Errorf("no row for %s", row)
		}
	}
	lower := strings.ToLower(out)
	for _, forbidden := range []string{"recommend", "should grow", "should sunset", "we suggest"} {
		if strings.Contains(lower, forbidden) {
			t.Errorf("the portfolio report makes a recommendation (%q)", forbidden)
		}
	}
	if !strings.Contains(out, "does not rank them") {
		t.Error("the reading note is missing")
	}
}

func TestPortfolioWithoutPathways(t *testing.T) {
	in := exampleInput(t)
	in.View.Pathways = nil
	cost, _ := burden.Compute(in.Doc)
	if !strings.Contains(Portfolio(in, cost), "no pathways") {
		t.Error("An impact model with no pathways should say so")
	}
}

func TestFreeDataShare(t *testing.T) {
	in := exampleInput(t)
	got := freeDataShare(in.Doc, "pw-learning")
	if !strings.Contains(got, " of ") || !strings.HasSuffix(got, "%)") {
		t.Errorf("share = %q", got)
	}
	if freeDataShare(in.Doc, "pw-ghost") != "—" {
		t.Error("an unknown pathway has no share")
	}
}
