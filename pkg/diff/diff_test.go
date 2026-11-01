package diff

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/edit"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
)

var update = flag.Bool("update", false, "rewrite golden files")

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil { //nolint:gosec // test data directory
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v (run with -update)", err)
	}
	if string(want) != got {
		t.Errorf("%s differs (run with -update to accept):\n--- got\n%s\n--- want\n%s", name, got, want)
	}
}

func load(t *testing.T) *impact.Model {
	t.Helper()
	d, err := impactio.LoadDocument("../impact/testdata/full.json")
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// edited returns the full fixture changed in every way the diff reports.
func edited(t *testing.T) *impact.Model {
	t.Helper()
	d := load(t)
	d.Metadata.Version = "0.2"
	d.Context.ProblemStatement += " It is getting worse."
	d.Ceiling.Rationale = "Moved after review."
	d.Outcomes[2].Title = "Major identity providers ship support"                                                       // changed
	d.Outcomes[2].Provenance = &impact.Provenance{Basis: impact.BasisConfirmed}                                         // set
	d.Links[0].Provenance = &impact.Provenance{Basis: impact.BasisRejected}                                             // rejected
	d.Outcomes = append(d.Outcomes, impact.Outcome{ID: "oc-new", Title: "A new outcome", Level: impact.LevelShortTerm}) // added
	d.Indicators = slices.Delete(d.Indicators, len(d.Indicators)-1, len(d.Indicators))                                  // removed
	return d
}

func TestCompareGolden(t *testing.T) {
	res, err := Compare(load(t), edited(t))
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "full-edited.md", res.Markdown())
	j, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	golden(t, "full-edited.json", string(j)+"\n")
}

func TestChangesAreKeyedByElementID(t *testing.T) {
	res, _ := Compare(load(t), edited(t))
	by := map[string]Change{}
	for _, c := range res.Changes {
		by[c.ElementID] = c
	}
	if c := by["oc-idp-support"]; c.Op != Changed || !hasField(c, "title") || !hasField(c, "provenance.basis") {
		t.Errorf("a renamed and confirmed outcome: %+v", c)
	}
	if c := by["oc-new"]; c.Op != Added || c.Kind != "outcome" {
		t.Errorf("added: %+v", c)
	}
	if c := by["ind-adoption-stories"]; c.Op != Removed {
		t.Errorf("removed: %+v", c)
	}
	if c := by["ln-spec-idp"]; c.Op != Changed || !c.Rejected {
		t.Errorf("a rejection must be visible as one: %+v", c)
	}
	if len(res.Document) != 3 {
		t.Errorf("document-level changes = %+v", res.Document)
	}
}

func TestIdenticalAndReformattedTheoriesDoNotDiffer(t *testing.T) {
	res, err := Compare(load(t), load(t))
	if err != nil || !res.Empty() {
		t.Fatalf("%v %+v", err, res)
	}
	if !strings.Contains(res.Markdown(), "Nothing changed.") {
		t.Error("an empty diff should say so")
	}
	// The same impact model read from YAML and from JSON is the same impact model.
	canon, _ := impactio.Canonical(load(t), impactio.FormatYAML)
	fromYAML, err := impactio.ParseDocument(canon, impactio.FormatYAML)
	if err != nil {
		t.Fatal(err)
	}
	if res, _ := Compare(load(t), fromYAML); !res.Empty() {
		t.Errorf("format changed the meaning: %+v", res)
	}
}

func TestRevisionLogIsNotPartOfTheDiff(t *testing.T) {
	cur := load(t)
	cur.Revisions = append(cur.Revisions, impact.Revision{Version: "9", Date: "2026-01-01", Summary: "x"})
	if res, _ := Compare(load(t), cur); !res.Empty() {
		t.Errorf("a new log entry is not a change to the impact model: %+v", res)
	}
}

func TestDeterministic(t *testing.T) {
	a, _ := Compare(load(t), edited(t))
	b, _ := Compare(load(t), edited(t))
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if !bytes.Equal(ja, jb) {
		t.Error("two comparisons differ")
	}
}

func TestDraftRevisionListsEveryChangedIDAndApplies(t *testing.T) {
	old, cur := load(t), edited(t)
	res, _ := Compare(old, cur)
	rev := res.DraftRevision("0.2", "2026-03-01", impact.TriggerLearning)
	if rev.Version != "0.2" || rev.Trigger != impact.TriggerLearning || !strings.Contains(rev.Summary, "1 element(s) added, 2 changed, 1 removed") {
		t.Errorf("revision = %+v", rev)
	}
	for _, id := range []string{"oc-new", "oc-idp-support", "ln-spec-idp", "ind-adoption-stories"} {
		if !slices.Contains(rev.ChangedIDs, id) {
			t.Errorf("changedIds is missing %s: %v", id, rev.ChangedIDs)
		}
	}
	patch, err := RevisionPatch(rev, len(cur.Revisions) > 0)
	if err != nil {
		t.Fatal(err)
	}
	applied, err := edit.Apply(cur, patch, edit.Options{AllowErrors: true})
	if err != nil || !applied.Applied {
		t.Fatalf("the revision patch does not apply: %v %+v", err, applied)
	}
	if got := applied.Document.Revisions; len(got) != len(cur.Revisions)+1 || got[len(got)-1].Version != "0.2" {
		t.Errorf("revisions = %+v", got)
	}
	// An impact model with no revision log gets one.
	bare := load(t)
	bare.Revisions = nil
	p2, _ := RevisionPatch(rev, false)
	if a, err := edit.Apply(bare, p2, edit.Options{AllowErrors: true}); err != nil || !a.Applied || len(a.Document.Revisions) != 1 {
		t.Errorf("first revision: %v %+v", err, a)
	}
}

func hasField(c Change, path string) bool {
	return slices.ContainsFunc(c.Fields, func(f FieldChange) bool { return f.Path == path })
}
