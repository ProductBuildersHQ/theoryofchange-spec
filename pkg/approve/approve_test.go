package approve

import (
	"errors"
	"slices"
	"strings"
	"testing"

	core "github.com/grokify/prism-core"

	"github.com/ProductBuildersHQ/impact-model/pkg/edit"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
)

func load(t *testing.T) *impact.Model {
	t.Helper()
	doc, err := impactio.LoadDocument("../../examples/intermediary/impact-model.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func request() Request {
	return Request{
		Version: "1.0.0", Date: "2026-04-01", Summary: "Agreed with the network's staff.",
		Approvers: []core.Approver{{Person: core.Person{Name: "Board chair", Role: "chair"}}},
	}
}

func refusal(t *testing.T, err error, code string) Refusal {
	t.Helper()
	var re *RefusedError
	if !errors.As(err, &re) || !errors.Is(err, ErrRefused) {
		t.Fatalf("error = %v, want a refusal", err)
	}
	for _, r := range re.Refusals {
		if r.Code == code {
			return r
		}
	}
	t.Fatalf("no %q refusal in %v", code, re.Refusals)
	return Refusal{}
}

// cleared confirms every hypothesis so the impact model is ready.
func cleared(t *testing.T, doc *impact.Model) {
	t.Helper()
	ids, err := Hypotheses(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		ref, _ := doc.Element(id)
		kind := kindFor(t, ref.Path)
		p, err := edit.SetFields(doc, kind, id, map[string]any{"provenance.basis": "confirmed"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		res, err := edit.Apply(doc, p, edit.Options{})
		if err != nil || !res.Applied {
			t.Fatalf("clear %s: %v %+v", id, err, res)
		}
		*doc = *res.Document
	}
}

func TestApprove(t *testing.T) {
	doc := load(t)
	cleared(t, doc)
	p, err := Build(doc, request())
	if err != nil {
		t.Fatal(err)
	}
	res, err := edit.Apply(doc, p, edit.Options{})
	if err != nil || !res.Applied {
		t.Fatalf("apply: %v %+v", err, res)
	}
	got := res.Document
	if got.Metadata.Status != StatusApproved || got.Metadata.Version != "1.0.0" {
		t.Errorf("metadata = %+v", got.Metadata)
	}
	rev := got.Revisions[len(got.Revisions)-1]
	if rev.Version != "1.0.0" || rev.Trigger != impact.TriggerReview || len(rev.Approvers) != 1 ||
		!rev.Approvers[0].Approved || rev.Approvers[0].ApprovedAt.Format("2006-01-02") != "2026-04-01" {
		t.Errorf("revision = %+v", rev)
	}
	if len(got.Revisions) != len(doc.Revisions)+1 {
		t.Error("revision not appended")
	}
}

func TestApproveFirstRevision(t *testing.T) {
	doc := load(t)
	cleared(t, doc)
	doc.Revisions = nil
	doc.Metadata.Version = ""
	p, err := Build(doc, request())
	if err != nil {
		t.Fatal(err)
	}
	if res, err := edit.Apply(doc, p, edit.Options{}); err != nil || !res.Applied || len(res.Document.Revisions) != 1 {
		t.Fatalf("apply: %v %+v", err, res)
	}
}

func TestRefusesOpenHypotheses(t *testing.T) {
	doc := load(t)
	ids, _ := Hypotheses(doc)
	if len(ids) == 0 {
		t.Skip("the example has no hypotheses")
	}
	_, err := Build(doc, request())
	if r := refusal(t, err, "hypotheses"); !slices.Equal(r.IDs, ids) {
		t.Errorf("ids = %v, want %v", r.IDs, ids)
	}

	forced := request()
	forced.Force = true
	if _, err := Build(doc, forced); err == nil {
		t.Fatal("force without a reason was accepted")
	}
	forced.Reason = "board asked for sign-off before the next funding round"
	p, err := Build(doc, forced)
	if err != nil {
		t.Fatal(err)
	}
	res, err := edit.Apply(doc, p, edit.Options{})
	if err != nil || !res.Applied {
		t.Fatalf("apply: %v %+v", err, res)
	}
	last := res.Document.Revisions[len(res.Document.Revisions)-1]
	if !strings.Contains(last.Summary, "board asked for sign-off") {
		t.Errorf("summary = %q", last.Summary)
	}
}

func TestRefusesValidationErrors(t *testing.T) {
	doc := load(t)
	cleared(t, doc)
	doc.Links[0].To = "no-such-outcome"
	_, err := Build(doc, request())
	refusal(t, err, "errors")

	forced := request()
	forced.Force, forced.Reason = true, "x"
	if _, err := Build(doc, forced); err == nil {
		t.Error("force overrode validation errors")
	}
}

func TestRefusesIncompleteRequest(t *testing.T) {
	doc := load(t)
	cleared(t, doc)
	for name, mutate := range map[string]func(*Request){
		"no version":      func(r *Request) { r.Version = "" },
		"same version":    func(r *Request) { r.Version = currentVersion(doc) },
		"bad date":        func(r *Request) { r.Date = "April" },
		"no approvers":    func(r *Request) { r.Approvers = nil },
		"force no reason": func(r *Request) { r.Force = true },
	} {
		r := request()
		mutate(&r)
		_, err := Build(doc, r)
		if name == "same version" && currentVersion(doc) == "" {
			continue
		}
		if name == "force no reason" {
			refusal(t, err, "request")
			continue
		}
		refusal(t, err, "request")
	}
}

func TestRejectedHypothesesDoNotCount(t *testing.T) {
	doc := load(t)
	ids, _ := Hypotheses(doc)
	if len(ids) == 0 {
		t.Skip("the example has no hypotheses")
	}
	cleared(t, doc)
	if ids, _ := Hypotheses(doc); len(ids) != 0 {
		t.Errorf("hypotheses remain: %v", ids)
	}
}

// kindFor finds the editing kind by the array a JSON pointer such as
// "/actors/2" lives in.
func kindFor(t *testing.T, path string) edit.ElementKind {
	t.Helper()
	field := strings.Split(path, "/")[1]
	for _, k := range edit.ElementKinds() {
		if k.Field == field {
			return k
		}
	}
	t.Fatalf("no kind for %s", path)
	return edit.ElementKind{}
}
