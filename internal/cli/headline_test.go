package cli

import (
	"slices"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/edit"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
)

// The focus flags set the North Star, the one metric, and the guardrails, and
// the product kind is accepted.
func TestPathwayFocusFlags(t *testing.T) {
	path := fixtureCopy(t, "full.json", "doc.json")
	r := run(t, "", "pathway", "set", path, "pw-supply", "--kind", "product",
		"--north-star", "ind-idp-count", "--one-metric", "ind-idp-roadmaps", "--guardrails", "ind-reference-downloads,ind-idp-count-na")
	wantCode(t, r, 0)
	doc, err := impactio.LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	pw, ok := doc.Pathway("pw-supply")
	if !ok || pw.Focus == nil {
		t.Fatalf("pathway = %+v", pw)
	}
	if pw.Kind != impact.PathwayProduct || pw.Focus.NorthStarID != "ind-idp-count" || pw.Focus.OneMetricID != "ind-idp-roadmaps" ||
		!slices.Equal(pw.Focus.GuardrailIDs, []string{"ind-reference-downloads", "ind-idp-count-na"}) {
		t.Errorf("kind %q, focus %+v", pw.Kind, pw.Focus)
	}
}

// A reference to an indicator that does not exist is refused, and the file is
// left as it was.
func TestPathwayFocusRejectsAnUnknownIndicator(t *testing.T) {
	path := fixtureCopy(t, "full.json", "doc.json")
	before := mustRead(t, path)
	r := run(t, "", "pathway", "set", path, "pw-supply", "--north-star", "ind-ghost")
	wantCode(t, r, ExitInvalid)
	if mustRead(t, path) != before {
		t.Error("the file changed although the change was rejected")
	}
}

// A headline list given as JSON through --set is the same change as the
// equivalent patch.
func TestHeadlineListSetEqualsItsPatch(t *testing.T) {
	viaCommand := fixtureCopy(t, "full.json", "a.json")
	viaPatch := fixtureCopy(t, "full.json", "b.json")
	list := `[{"indicatorId":"ind-idp-count","label":"Providers we reached","note":"What the board asks for first"},{"indicatorId":"ind-idp-roadmaps"}]`

	wantCode(t, run(t, "", "audience", "set", viaCommand, "aud-board", "--set", "headline="+list), 0)
	ops := []edit.Op{must(edit.Add("/audiences/@aud-board/headline", []any{
		map[string]any{"indicatorId": "ind-idp-count", "label": "Providers we reached", "note": "What the board asks for first"},
		map[string]any{"indicatorId": "ind-idp-roadmaps"},
	}))}
	wantCode(t, run(t, patchJSON(t, ops...), "patch", viaPatch, "-"), 0)

	if mustRead(t, viaCommand) != mustRead(t, viaPatch) {
		t.Fatalf("command and patch produced different files:\n--- command\n%s\n--- patch\n%s", mustRead(t, viaCommand), mustRead(t, viaPatch))
	}
	doc, err := impactio.LoadDocument(viaCommand)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := doc.Audience("aud-board")
	if len(a.Headline) != 2 || a.Headline[0].Label != "Providers we reached" || a.Headline[1].IndicatorID != "ind-idp-roadmaps" {
		t.Errorf("headline = %+v", a.Headline)
	}
}

// An indicator is composed of others through two flags, and a composition
// that does not fit its operation is refused.
func TestIndicatorCompositionFlags(t *testing.T) {
	path := fixtureCopy(t, "full.json", "doc.json")
	wantCode(t, run(t, "", "indicator", "set", path, "ind-idp-count-na", "--composed-op", "difference",
		"--composed-of", "ind-adoption-stories,ind-reference-downloads"), 0)
	doc, err := impactio.LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	ind, ok := doc.Indicator("ind-idp-count-na")
	if !ok || ind.ComposedOf == nil || ind.ComposedOf.Op != impact.CompositionDifference ||
		!slices.Equal(ind.ComposedOf.IndicatorIDs, []string{"ind-adoption-stories", "ind-reference-downloads"}) {
		t.Fatalf("indicator = %+v", ind)
	}

	before := mustRead(t, path)
	wantCode(t, run(t, "", "indicator", "set", path, "ind-idp-count-na", "--composed-of", "ind-adoption-stories"), ExitInvalid)
	if mustRead(t, path) != before {
		t.Error("the file changed although the change was rejected")
	}
}
