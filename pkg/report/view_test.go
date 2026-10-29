package report

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

func fixture(t *testing.T, name string, v any) {
	t.Helper()
	data, err := os.ReadFile("../toc/testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
}

func fullDoc(t *testing.T) *toc.Document {
	t.Helper()
	var d toc.Document
	fixture(t, "full.json", &d)
	return &d
}

func fullStatus(t *testing.T) *toc.Status {
	t.Helper()
	var s toc.Status
	fixture(t, "status.json", &s)
	return &s
}

func build(t *testing.T, doc *toc.Document, st *toc.Status, o Options) *View {
	t.Helper()
	v, err := BuildView(doc, nil, st, o)
	if err != nil {
		t.Fatalf("BuildView: %v", err)
	}
	return v
}

func outcome(t *testing.T, v *View, id string) OutcomeView {
	t.Helper()
	for _, o := range v.Outcomes {
		if o.ID == id {
			return o
		}
	}
	t.Fatalf("outcome %q not in view", id)
	return OutcomeView{}
}

func TestViewWithoutStatus(t *testing.T) {
	v := build(t, fullDoc(t), nil, Options{})
	if v.Version != ViewVersion || v.AsOf != "" || v.Filter != "" {
		t.Fatalf("header = %+v", v)
	}
	d := v.Document
	if d.ID != "open-agent-auth" || d.Title != "Open agent authorization standard adoption" || d.Perspective != "Standards working group" {
		t.Fatalf("document = %+v", d)
	}
	if d.Ceiling != "Long-term outcome and below" || d.CeilingRationale == "" || d.TimeStart != "2026-01-01" {
		t.Fatalf("ceiling/time = %+v", d)
	}
	if len(v.Outcomes) != 6 || len(v.Interventions) != 2 || len(v.Pathways) != 1 {
		t.Fatalf("counts: outcomes=%d interventions=%d pathways=%d", len(v.Outcomes), len(v.Interventions), len(v.Pathways))
	}

	impact := outcome(t, v, "oc-safer-agents")
	if impact.Owned {
		t.Error("the impact outcome sits above the ceiling and is not owned")
	}
	if !outcome(t, v, "oc-enterprise-require").Owned {
		t.Error("a long-term outcome is within the ceiling")
	}
	enterprise := outcome(t, v, "oc-enterprise-require")
	if enterprise.Logic != "at least 1" || !slices.Equal(enterprise.Preconditions, []string{"oc-idp-support", "oc-platform-support"}) {
		t.Fatalf("enterprise = %+v", enterprise)
	}
	if !slices.Equal(enterprise.Indicators, []string{}) && enterprise.Indicators == nil {
		t.Fatal("empty lists must be non-nil so JSON renders []")
	}
	idp := outcome(t, v, "oc-idp-support")
	if !idp.Basis.Marked || idp.Basis.Basis != toc.BasisInferred {
		t.Errorf("inferred outcome must be marked: %+v", idp.Basis)
	}
	if outcome(t, v, "oc-spec-stable").Basis.Marked {
		t.Error("a stated outcome must not be marked")
	}
	if !slices.Equal(idp.Pathways, []string{"pw-supply"}) || !slices.Contains(idp.Indicators, "ind-idp-count") {
		t.Fatalf("idp pathways/indicators = %v / %v", idp.Pathways, idp.Indicators)
	}
	if idp.Status != nil {
		t.Error("no status was supplied")
	}
}

func TestViewIndicatorsAndMetrics(t *testing.T) {
	v := build(t, fullDoc(t), nil, Options{})
	var count IndicatorView
	for _, i := range v.Indicators {
		if i.ID == "ind-idp-count" {
			count = i
		}
	}
	if count.Baseline != "0 providers (as of 2026-01-01)" {
		t.Errorf("baseline = %q", count.Baseline)
	}
	if len(count.Targets) != 2 || count.Targets[0] != "2 providers (by 2026-12-31) — First movers" ||
		count.Targets[1] != "5 providers (by 2027-12-31) (Covers half the market by seats)" {
		t.Errorf("targets = %q", count.Targets)
	}
	if count.Claim != toc.ClaimContribution || count.MeasuresTitle != "Major identity providers support the standard" {
		t.Errorf("claim/measures = %q / %q", count.Claim, count.MeasuresTitle)
	}
	if count.DataAvailability != "public" || len(count.Caveats) != 1 {
		t.Errorf("availability/caveats = %q / %v", count.DataAvailability, count.Caveats)
	}
	for _, i := range v.Indicators {
		if i.ID == "ind-idp-roadmaps" && (len(i.Levels) != 4 || i.Baseline != "exploring (as of 2026-01-01)") {
			t.Errorf("qualitative indicator = %+v", i)
		}
		if i.ID == "ind-idp-total" && (i.Claim != toc.ClaimAttribution || len(i.RollupFrom) != 2) {
			t.Errorf("roll-up indicator = %+v", i)
		}
	}

	var providers MetricView
	for _, m := range v.Metrics {
		if m.ID == "met-providers" {
			providers = m
		}
		if m.ID == "met-provider-share" && m.WeightedBy != "Providers with GA support" {
			t.Errorf("weightedBy = %q", m.WeightedBy)
		}
	}
	if !slices.Equal(providers.Indicators, []string{"ind-idp-count", "ind-idp-count-na", "ind-idp-count-eu", "ind-idp-total"}) {
		t.Errorf("metric indicators = %v", providers.Indicators)
	}
}

func TestViewWithStatus(t *testing.T) {
	v := build(t, fullDoc(t), fullStatus(t), Options{})
	if v.AsOf != "2026-12-31" {
		t.Fatalf("AsOf = %q", v.AsOf)
	}
	idp := outcome(t, v, "oc-idp-support")
	if idp.Status == nil || idp.Status.State != "achieved" || idp.Status.PathwayState != "on_track" || idp.Status.Health != "green" || *idp.Status.Confidence != 0.62 {
		t.Fatalf("outcome status = %+v", idp.Status)
	}
	for _, i := range v.Indicators {
		if i.ID == "ind-idp-count" {
			if i.Status == nil || i.Status.State != "achieved" || i.Latest != "2 providers (on 2026-12-15)" {
				t.Errorf("indicator status = %+v latest %q", i.Status, i.Latest)
			}
		}
	}
	for _, a := range v.Assumptions {
		if a.ID == "asm-idp-incentive" && (a.Status != "holding" || a.StatusSource != "assessment") {
			t.Errorf("assumption status = %q from %q", a.Status, a.StatusSource)
		}
		if a.ID == "asm-no-fork" && a.StatusSource != "document" {
			t.Errorf("an assumption without an assessment keeps the document status: %+v", a)
		}
	}
	for _, l := range v.Links {
		if l.ID == "ln-spec-idp" && l.State != "ok" {
			t.Errorf("link state = %q", l.State)
		}
	}
}

func TestPathwayFilter(t *testing.T) {
	doc := fullDoc(t)
	doc.Pathways = append(doc.Pathways, toc.Pathway{ID: "pw-platform", Title: "Platforms", OutcomeIDs: []string{"oc-platform-support"}})
	v := build(t, doc, nil, Options{Pathway: "pw-platform"})
	got := []string{}
	for _, o := range v.Outcomes {
		got = append(got, o.ID)
	}
	if !slices.Equal(got, []string{"oc-safer-agents", "oc-enterprise-require", "oc-platform-support"}) {
		t.Fatalf("filtered outcomes = %v", got)
	}
	if v.Filter != "pw-platform" || len(v.Pathways) != 1 || v.Pathways[0].ID != "pw-platform" {
		t.Fatalf("pathways = %+v", v.Pathways)
	}
	for _, l := range v.Links {
		if l.From == "oc-spec-stable" {
			t.Fatalf("link outside the filter: %+v", l)
		}
	}
	if len(v.Interventions) != 0 {
		t.Errorf("interventions outside the pathway leaked in: %+v", v.Interventions)
	}

	if _, err := BuildView(doc, nil, nil, Options{Pathway: "pw-ghost"}); err == nil {
		t.Fatal("an unknown pathway should be an error")
	}
}

func TestRejectedElementsAreExcludedAndListed(t *testing.T) {
	doc := fullDoc(t)
	doc.Outcomes[3].Provenance = &toc.Provenance{Basis: toc.BasisRejected, Note: "Reviewers said platforms are out of scope."}
	doc.Risks[0].Provenance = &toc.Provenance{Basis: toc.BasisRejected}
	v := build(t, doc, nil, Options{})

	for _, o := range v.Outcomes {
		if o.ID == "oc-platform-support" {
			t.Fatal("a rejected outcome is still in the view")
		}
	}
	for _, l := range v.Links {
		if l.From == "oc-platform-support" || l.To == "oc-platform-support" {
			t.Fatalf("a link to a rejected outcome is still in the view: %+v", l)
		}
	}
	if len(v.Rejected) != 2 {
		t.Fatalf("rejected = %+v", v.Rejected)
	}
	if v.Rejected[0].ID != "oc-platform-support" || v.Rejected[0].Kind != toc.ElementOutcome || v.Rejected[0].Note == "" {
		t.Fatalf("rejected[0] = %+v", v.Rejected[0])
	}
	if v.BasisCounts[toc.BasisRejected] != 0 {
		t.Error("rejected elements are not counted among active ones")
	}
	// The pathway no longer lists the rejected outcome.
	if slices.Contains(v.Pathways[0].OutcomeIDs, "oc-platform-support") {
		t.Errorf("pathway still lists it: %v", v.Pathways[0].OutcomeIDs)
	}
}

func TestRestrictedEvidenceIsWithheld(t *testing.T) {
	doc := fullDoc(t)
	doc.Evidence = append(doc.Evidence,
		toc.Evidence{ID: "ev-secret", Title: "Interview notes", Sensitivity: toc.SensitivityRestricted, Excerpt: "do not publish", Summary: "also private", URL: "https://example.org/private/notes"},
		toc.Evidence{ID: "ev-unset", Title: "Unclassified", Excerpt: "unset means restricted"},
		toc.Evidence{ID: "ev-internal", Title: "Internal memo", Sensitivity: toc.SensitivityInternal, Excerpt: "fine to show"},
	)
	v := build(t, doc, nil, Options{})
	byID := map[string]EvidenceView{}
	for _, e := range v.Evidence {
		byID[e.ID] = e
	}
	if e := byID["ev-secret"]; !e.Withheld || e.Excerpt != "" || e.Summary != "" || e.URL != "" {
		t.Errorf("restricted evidence leaked: %+v", e)
	}
	if e := byID["ev-unset"]; !e.Withheld || e.Excerpt != "" || e.Sensitivity != "restricted" {
		t.Errorf("unset sensitivity must be treated as restricted: %+v", e)
	}
	if e := byID["ev-internal"]; e.Withheld || e.Excerpt != "fine to show" {
		t.Errorf("internal evidence should be shown: %+v", e)
	}
	if e := byID["ev-survey"]; e.Withheld || e.Excerpt == "" {
		t.Errorf("public evidence should be shown: %+v", e)
	}
	// The raw secret must not appear anywhere in the serialized view.
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"do not publish", "also private", "unset means restricted", "example.org/private/notes"} {
		if strings.Contains(string(b), secret) {
			t.Errorf("serialized view contains withheld text %q", secret)
		}
	}
}

func TestBasisCountsAndUnmarked(t *testing.T) {
	v := build(t, fullDoc(t), nil, Options{})
	total := v.Unmarked
	for _, n := range v.BasisCounts {
		total += n
	}
	want := len(v.Pathways) + len(v.Outcomes) + len(v.Links) + len(v.Groups) + len(v.Interventions) +
		len(v.Assumptions) + len(v.Indicators) + len(v.Metrics) + len(v.Actors)
	if total != want {
		t.Fatalf("basis counts + unmarked = %d, want %d", total, want)
	}
	if v.BasisCounts[toc.BasisStated] == 0 || v.BasisCounts[toc.BasisInferred] == 0 {
		t.Errorf("counts = %v", v.BasisCounts)
	}
}

func TestViewIsDeterministicSerializableAndNonMutating(t *testing.T) {
	doc := fullDoc(t)
	before, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	a, b := build(t, doc, fullStatus(t), Options{}), build(t, doc, fullStatus(t), Options{})
	if !reflect.DeepEqual(a, b) {
		t.Fatal("BuildView is not deterministic")
	}
	after, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("BuildView modified the document")
	}
	out, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out, []byte(":null")) {
		t.Errorf("the view serializes a null; lists must be [] so templates can range over them")
	}
}

func TestMinimalDocumentBuildsAView(t *testing.T) {
	var d toc.Document
	fixture(t, "minimal.json", &d)
	v := build(t, &d, nil, Options{})
	if len(v.Outcomes) != 2 || v.Document.Title != "Minimal theory" || v.Document.Ceiling != "" {
		t.Fatalf("view = %+v", v.Document)
	}
	if v.Unmarked == 0 || len(v.BasisCounts) != 0 {
		t.Errorf("a document without provenance has only unmarked elements: %v / %d", v.BasisCounts, v.Unmarked)
	}
}

func TestInvalidDocumentIsAnError(t *testing.T) {
	doc := fullDoc(t)
	doc.Links[0].To = "oc-ghost"
	if _, err := BuildView(doc, nil, nil, Options{}); err == nil {
		t.Fatal("expected an error for an invalid document")
	}
	if _, err := BuildView(nil, nil, nil, Options{}); err == nil {
		t.Fatal("expected an error for a nil document")
	}
}

func TestLabels(t *testing.T) {
	if Label("long_term") != "Long term" || Label("") != "" || Label("impact") != "Impact" {
		t.Error("Label mismatch")
	}
	if LevelLabel(toc.LevelOutput) != "Output" || LevelLabel(toc.LevelLongTerm) != "Long-term outcome" || LevelLabel("custom_level") != "Custom level" {
		t.Error("LevelLabel mismatch")
	}
	if logicText(toc.Logic{Type: toc.LogicAtLeast, K: 2}) != "at least 2" || logicText(toc.Logic{Type: toc.LogicAny}) != "any" {
		t.Error("logicText mismatch")
	}
	if valueText(nil, "", "x") != "" || withDate("", "", "by ") != "" || withDate("", "2026-01-01", "by ") != "by 2026-01-01" || withDate("5", "", "by ") != "5" {
		t.Error("value formatting mismatch")
	}
}

func TestIndicatorOfAPathwayFollowsThePathwayFilter(t *testing.T) {
	doc := fullDoc(t)
	doc.Indicators[5].Measures = "pw-supply"
	doc.Pathways = append(doc.Pathways, toc.Pathway{ID: "pw-other", Title: "Other", OutcomeIDs: []string{"oc-platform-support"}})
	in := func(v *View, id string) bool {
		for _, i := range v.Indicators {
			if i.ID == id {
				return true
			}
		}
		return false
	}
	if v := build(t, doc, nil, Options{}); !in(v, "ind-idp-total") {
		t.Error("an unfiltered view includes a pathway-level indicator")
	}
	if v := build(t, doc, nil, Options{Pathway: "pw-supply"}); !in(v, "ind-idp-total") {
		t.Error("the pathway's own view includes its pathway-level indicator")
	}
	if v := build(t, doc, nil, Options{Pathway: "pw-other"}); in(v, "ind-idp-total") {
		t.Error("another pathway's view must not include it")
	}
}
