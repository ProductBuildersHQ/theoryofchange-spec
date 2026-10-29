package review

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/edit"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

const (
	hash       = "sha256:current"
	inferredID = "oc-resident-access"
	statedID   = "oc-st-leader-skills"
)

func load(t *testing.T) *toc.Document {
	t.Helper()
	doc, err := tocio.LoadDocument("../../examples/intermediary/theory.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func response(rows ...toc.ReviewItem) *toc.ReviewResponse {
	return &toc.ReviewResponse{
		Kind: toc.KindReviewResponse, SpecVersion: "0.1", TocID: "riverbend-network", ContentHash: hash,
		Reviewers: []string{"program staff"}, Method: toc.ReviewSheet, Date: "2026-03-01", Rows: rows,
	}
}

func opts(c ...Candidate) Options { return Options{ContentHash: hash, Candidates: c} }

func TestAgreeConfirmsBasis(t *testing.T) {
	doc := load(t)
	res, err := Import(doc, response(toc.ReviewItem{ElementID: statedID, Decision: toc.DecisionAgree, Origin: toc.OriginSheet}), opts())
	if err != nil {
		t.Fatal(err)
	}
	if res.NeedsHuman() || res.Mapped != 1 {
		t.Fatalf("result = %+v", res)
	}
	applied, err := edit.Apply(doc, res.Patch, edit.Options{})
	if err != nil || !applied.Applied {
		t.Fatalf("apply: %v %+v", err, applied)
	}
	o, _ := applied.Document.Outcome(statedID)
	p := o.Provenance
	if p.Basis != toc.BasisConfirmed || p.DecidedAt != "2026-03-01" || p.DecidedBy != "program staff" {
		t.Errorf("provenance = %+v", p)
	}
	if len(p.EvidenceIDs) != 1 {
		t.Errorf("evidence ids were lost: %+v", p)
	}
}

func TestRemoveRejects(t *testing.T) {
	doc := load(t)
	res, _ := Import(doc, response(toc.ReviewItem{ElementID: inferredID, Decision: toc.DecisionRemove, Origin: toc.OriginTable}), opts())
	applied, err := edit.Apply(doc, res.Patch, edit.Options{AllowErrors: true})
	if err != nil || !applied.Applied {
		t.Fatalf("apply: %v %+v", err, applied)
	}
	o, _ := applied.Document.Outcome(inferredID)
	if !o.Provenance.IsRejected() {
		t.Errorf("provenance = %+v", o.Provenance)
	}
}

func TestStructuredChange(t *testing.T) {
	doc := load(t)
	row := toc.ReviewItem{
		ElementID: statedID, Decision: toc.DecisionChange, Origin: toc.OriginSuggestion,
		SuggestedChange: "title: Leaders gain coaching skills\ncategory: leadership",
	}
	res, _ := Import(doc, response(row), opts())
	if res.NeedsHuman() || res.Mapped != 1 {
		t.Fatalf("result = %+v", res)
	}
	applied, err := edit.Apply(doc, res.Patch, edit.Options{})
	if err != nil || !applied.Applied {
		t.Fatalf("apply: %v %+v", err, applied)
	}
	o, _ := applied.Document.Outcome(statedID)
	if o.Title != "Leaders gain coaching skills" || o.Category != "leadership" {
		t.Errorf("outcome = %+v", o)
	}
}

func TestFlaggedRows(t *testing.T) {
	doc := load(t)
	rows := []toc.ReviewItem{
		{ElementID: statedID, Decision: toc.DecisionChange, SuggestedChange: "Make this about confidence too", Origin: toc.OriginSheet},
		{ElementID: statedID, Decision: toc.DecisionUnsure, Origin: toc.OriginSheet},
		{ElementID: statedID, Comment: "Who funds this?", Origin: toc.OriginComment},
		{ElementID: "no-such-id", Decision: toc.DecisionAgree, Origin: toc.OriginSheet},
		{Comment: "What about the funders?", Origin: toc.OriginComment},
		{QuotedText: "text that is nowhere", Decision: toc.DecisionAgree, Origin: toc.OriginComment},
		{ElementID: statedID, Decision: toc.DecisionChange, SuggestedChange: "id: other", Origin: toc.OriginSheet},
	}
	res, _ := Import(doc, response(rows...), opts())
	if res.Mapped != 0 || len(res.Patch) != 0 || len(res.Flags) != len(rows) {
		t.Fatalf("mapped %d, patch %d, flags %d", res.Mapped, len(res.Patch), len(res.Flags))
	}
	for i, f := range res.Flags {
		if f.Row != i+1 || f.Reason == "" {
			t.Errorf("flag %d = %+v", i, f)
		}
	}
	if !res.NeedsHuman() {
		t.Error("flags must need a human")
	}
}

func TestQuotedTextResolution(t *testing.T) {
	doc := load(t)
	cands := []Candidate{
		{ID: statedID, Statement: "Leaders gain coaching-based skills and confidence"},
		{ID: "a", Statement: "Same words"}, {ID: "b", Statement: "Same  words"},
	}
	rows := []toc.ReviewItem{
		{QuotedText: "  Leaders gain coaching-based\nskills and confidence ", Decision: toc.DecisionAgree, Origin: toc.OriginComment},
		{QuotedText: "Same words", Decision: toc.DecisionAgree, Origin: toc.OriginComment},
	}
	res, _ := Import(doc, response(rows...), opts(cands...))
	if res.Mapped != 1 || len(res.Flags) != 1 || !strings.Contains(res.Flags[0].Reason, "several") || res.Flags[0].Row != 2 {
		t.Fatalf("result = %+v", res)
	}
	if !strings.Contains(string(mustJSON(t, res.Patch)), statedID) {
		t.Errorf("patch does not address %s", statedID)
	}
}

func TestStaleHashFlagsEveryRow(t *testing.T) {
	doc := load(t)
	resp := response(
		toc.ReviewItem{ElementID: statedID, Decision: toc.DecisionAgree, Origin: toc.OriginSheet},
		toc.ReviewItem{Comment: "hello", Origin: toc.OriginComment},
	)
	resp.ContentHash = "sha256:older"
	res, _ := Import(doc, resp, opts())
	if !res.Stale || res.Mapped != 1 || len(res.Patch) == 0 {
		t.Fatalf("result = %+v", res)
	}
	rows := []int{}
	for _, f := range res.Flags {
		rows = append(rows, f.Row)
	}
	if !slices.Equal(rows, []int{1, 2}) {
		t.Errorf("flagged rows = %v", rows)
	}
}

func TestLaterRowsSeeEarlierOnes(t *testing.T) {
	doc := load(t)
	rows := []toc.ReviewItem{
		{ElementID: statedID, Decision: toc.DecisionAgree, Origin: toc.OriginSheet},
		{ElementID: statedID, Decision: toc.DecisionChange, SuggestedChange: "title: Better title", Origin: toc.OriginSheet},
	}
	res, _ := Import(doc, response(rows...), opts())
	if res.Mapped != 2 || res.NeedsHuman() {
		t.Fatalf("result = %+v", res)
	}
	applied, err := edit.Apply(doc, res.Patch, edit.Options{})
	if err != nil || !applied.Applied {
		t.Fatalf("apply: %v %+v", err, applied)
	}
}

func TestDocumentNeverModified(t *testing.T) {
	doc := load(t)
	before := mustJSON(t, doc)
	_, err := Import(doc, response(
		toc.ReviewItem{ElementID: statedID, Decision: toc.DecisionRemove, Origin: toc.OriginSheet},
		toc.ReviewItem{ElementID: inferredID, Decision: toc.DecisionAgree, Origin: toc.OriginSheet},
	), opts())
	if err != nil {
		t.Fatal(err)
	}
	if string(mustJSON(t, doc)) != string(before) {
		t.Error("Import modified the document")
	}
}

func TestConsultation(t *testing.T) {
	doc := load(t)
	resp := response(
		toc.ReviewItem{ElementID: statedID, Decision: toc.DecisionAgree, Origin: toc.OriginSheet},
		toc.ReviewItem{ElementID: inferredID, Decision: toc.DecisionAgree, Origin: toc.OriginSheet},
		toc.ReviewItem{ElementID: inferredID, Decision: toc.DecisionRemove, Origin: toc.OriginSheet},
		toc.ReviewItem{Comment: "note", Origin: toc.OriginComment},
	)
	res, _ := Import(doc, resp, opts())
	c := res.Consultation
	if c.ID != "consult-review-2026-03-01" || c.Step != toc.StepWhole || c.Method != toc.MethodDocumentReview ||
		!slices.Equal(c.StakeholderGroups, []string{"program staff"}) {
		t.Errorf("consultation = %+v", c)
	}
	want := []string{"agree: 2", "remove: 1", "comment only: 1"}
	if !slices.Equal(c.Findings, want) {
		t.Errorf("findings = %v, want %v", c.Findings, want)
	}

	doc.Consultations = append(doc.Consultations, toc.Consultation{ID: c.ID, Date: "2026-03-01", Step: toc.StepWhole})
	if got := consultationID(doc, "2026-03-01"); got != "consult-review-2026-03-01-2" {
		t.Errorf("collision id = %q", got)
	}
	resp.Method = toc.ReviewWorkshop
	if res, _ := Import(doc, resp, opts()); res.Consultation.Method != toc.MethodWorkshop {
		t.Errorf("workshop method = %q", res.Consultation.Method)
	}
}

func TestWrongTheoryRefused(t *testing.T) {
	resp := response()
	resp.TocID = "another-theory"
	if _, err := Import(load(t), resp, opts()); err == nil {
		t.Fatal("a response for another theory was accepted")
	}
}

func TestDeterministic(t *testing.T) {
	doc := load(t)
	resp := response(
		toc.ReviewItem{ElementID: statedID, Decision: toc.DecisionAgree, Origin: toc.OriginSheet},
		toc.ReviewItem{Comment: "x", Origin: toc.OriginComment},
	)
	a, _ := Import(doc, resp, opts())
	b, _ := Import(doc, resp, opts())
	if string(mustJSON(t, a)) != string(mustJSON(t, b)) {
		t.Error("two imports differ")
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
