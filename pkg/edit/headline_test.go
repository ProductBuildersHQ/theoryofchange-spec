package edit

import (
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

// Removing an indicator removes it from every headline list and every focus
// that names it, and leaves the rest of each list alone.
func TestCascadeRemovesAnIndicatorFromHeadlinesAndFocus(t *testing.T) {
	doc := loadFull(t)
	doc.Pathways[0].Focus.GuardrailIDs = []string{"ind-reference-downloads", "ind-idp-count-na"}
	headlineBefore := len(doc.Audiences[0].Headline)
	if headlineBefore < 2 || doc.Audiences[0].Headline[0].IndicatorID != "ind-idp-count" || doc.Pathways[0].Focus.NorthStarID != "ind-idp-count" {
		t.Fatalf("test setup: headline %+v, focus %+v", doc.Audiences[0].Headline, doc.Pathways[0].Focus)
	}

	p, err := RemoveElement(doc, kindOf(t, doc, "ind-idp-count"), "ind-idp-count", true)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Apply(doc, p, Options{})
	if err != nil || !res.Applied {
		t.Fatalf("cascade rejected: %v %s %v", err, res.Reason, res.IssuesIntroduced)
	}
	if validate.HasErrors(validate.Document(res.Document)) {
		t.Fatalf("result has errors: %v", validate.Document(res.Document))
	}
	got := res.Document
	if n := len(got.Audiences[0].Headline); n != headlineBefore-1 {
		t.Errorf("headline has %d entries, want %d", n, headlineBefore-1)
	}
	for _, h := range got.Audiences[0].Headline {
		if h.IndicatorID == "ind-idp-count" {
			t.Error("the removed indicator is still in the headline list")
		}
	}
	if got.Pathways[0].Focus.NorthStarID != "" {
		t.Errorf("north star = %q after the indicator was removed", got.Pathways[0].Focus.NorthStarID)
	}
	if got.Pathways[0].Focus.OneMetricID != "ind-idp-roadmaps" {
		t.Errorf("the one metric changed to %q although another indicator was removed", got.Pathways[0].Focus.OneMetricID)
	}

	// A guardrail is dropped from the list, and the others stay.
	p, err = RemoveElement(got, kindOf(t, got, "ind-reference-downloads"), "ind-reference-downloads", true)
	if err != nil {
		t.Fatal(err)
	}
	res, err = Apply(got, p, Options{})
	if err != nil || !res.Applied {
		t.Fatalf("cascade rejected: %v %s", err, res.Reason)
	}
	if !slices.Equal(res.Document.Pathways[0].Focus.GuardrailIDs, []string{"ind-idp-count-na"}) {
		t.Errorf("guardrails = %v", res.Document.Pathways[0].Focus.GuardrailIDs)
	}
}

// Removing an indicator that a composition is built from removes the
// composition, since its arithmetic no longer holds, and not the indicator that
// carried it.
func TestRemovingAPartRemovesTheComposition(t *testing.T) {
	doc := loadFull(t)
	p, err := RemoveElement(doc, kindOf(t, doc, "ind-adoption-stories"), "ind-adoption-stories", true)
	if err != nil {
		t.Fatal(err)
	}
	res, err := Apply(doc, p, Options{})
	if err != nil || !res.Applied {
		t.Fatalf("cascade rejected: %v %s %v", err, res.Reason, res.IssuesIntroduced)
	}
	ind, ok := res.Document.Indicator("ind-idp-count-eu")
	if !ok {
		t.Fatal("the indicator that carried the composition was removed")
	}
	if ind.ComposedOf != nil {
		t.Errorf("composition survived: %+v", ind.ComposedOf)
	}
}
