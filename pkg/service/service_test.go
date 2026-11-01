package service

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/edit"
	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/impactio"
)

const fixtures = "../impact/testdata/"

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(fixtures + name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

// copyFixture writes a fixture into a temp dir and returns its path.
func copyFixture(t *testing.T, name, as string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), as)
	if err := os.WriteFile(path, fixture(t, name), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func fullDoc(t *testing.T) *impact.Model {
	t.Helper()
	doc, err := impactio.ParseDocument(fixture(t, "full.json"), impactio.FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestValidateBytes(t *testing.T) {
	r := ValidateBytes("full.json", fixture(t, "full.json"), impactio.FormatJSON, nil)
	if !r.Valid() || r.Kind != impact.KindImpactModel || len(r.Issues) != 0 || r.Issues == nil {
		t.Fatalf("valid document report = %+v", r)
	}

	dangling := strings.Replace(string(fixture(t, "minimal.json")), `"to": "oc-funding"`, `"to": "oc-nowhere"`, 1)
	r = ValidateBytes("x", []byte(dangling), impactio.FormatJSON, nil)
	if r.Valid() || len(r.Issues) != 1 || r.Issues[0].Code != "VAL-REF-001" {
		t.Fatalf("dangling report = %+v", r)
	}

	r = ValidateBytes("x", []byte(`{"kind":"ImpactModel"}`), impactio.FormatJSON, nil)
	if r.Valid() || r.Kind != impact.KindImpactModel {
		t.Fatalf("schema-invalid report = %+v", r)
	}
	r = ValidateBytes("x", []byte(`{`), impactio.FormatJSON, nil)
	if r.Valid() || r.Kind != "" {
		t.Fatalf("unparseable report = %+v", r)
	}
}

func TestValidateObservationsAgainstTheory(t *testing.T) {
	obs := fixture(t, "observations.json")
	if r := ValidateBytes("obs", obs, impactio.FormatJSON, nil); !r.Valid() || r.Kind != impact.KindObservationSet {
		t.Fatalf("schema-only report = %+v", r)
	}
	if r := ValidateBytes("obs", obs, impactio.FormatJSON, fullDoc(t)); !r.Valid() {
		t.Fatalf("report against impact model = %+v", r)
	}
	other := strings.Replace(string(obs), `"indicatorId": "ind-idp-count"`, `"indicatorId": "ind-ghost"`, 1)
	r := ValidateBytes("obs", []byte(other), impactio.FormatJSON, fullDoc(t))
	if r.Valid() || r.Issues[0].Code != "VAL-REF-001" {
		t.Fatalf("unknown indicator should fail against the impact model: %+v", r)
	}
}

func TestValidateFile(t *testing.T) {
	path := copyFixture(t, "full.json", "doc.json")
	r, err := ValidateFile(path, nil)
	if err != nil || !r.Valid() || r.Source != path {
		t.Fatalf("ValidateFile = %+v, %v", r, err)
	}
	if _, err := ValidateFile(filepath.Join(t.TempDir(), "missing.json"), nil); err == nil {
		t.Fatal("missing file should be an error")
	}
	if _, err := ValidateFile("doc.txt", nil); !errors.Is(err, impactio.ErrUnsupportedFormat) {
		t.Fatalf("unsupported extension error = %v", err)
	}
}

func TestRunGraphQuery(t *testing.T) {
	g, err := BuildGraph(fullDoc(t))
	if err != nil {
		t.Fatal(err)
	}
	asJSON := func(v any) string {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	tests := []struct {
		q    GraphQuery
		id   string
		want string
	}{
		{QueryUpstream, "oc-enterprise-require", `["oc-idp-support","oc-platform-support","oc-spec-stable","oc-reference-impl","int-interop","int-reference"]`},
		{QueryDownstream, "oc-idp-support", `["oc-safer-agents","oc-enterprise-require"]`},
		{QueryPaths, "oc-spec-stable", `[["oc-spec-stable","oc-idp-support","oc-enterprise-require","oc-safer-agents"]]`},
		{QueryCycles, "", `[]`},
		{QueryOrphans, "", `[]`},
		{QueryDeadEnds, "", `[]`},
	}
	for _, tt := range tests {
		t.Run(string(tt.q), func(t *testing.T) {
			got, err := RunGraphQuery(g, tt.q, tt.id, 0)
			if err != nil {
				t.Fatal(err)
			}
			if asJSON(got) != tt.want {
				t.Fatalf("%s %s = %s, want %s", tt.q, tt.id, asJSON(got), tt.want)
			}
		})
	}

	sum, err := RunGraphQuery(g, QuerySummary, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := sum.(Summary)
	if !ok || s.Nodes != 8 || s.FeedbackEdges != 1 || len(s.ImpactOutcomes) != 1 || s.Cycles == nil || s.Orphans == nil {
		t.Fatalf("summary = %+v", sum)
	}

	levels, err := RunGraphQuery(g, QueryLevels, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if le := levels.([]LevelEntry); le[0].ID != "oc-safer-agents" || le[0].Level != 0 {
		t.Fatalf("levels = %+v", le)
	}

	all, err := RunGraphQuery(g, QueryGraph, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if e := all.(Export); len(e.Nodes) != 8 || len(e.Edges) == 0 {
		t.Fatalf("export = %+v", e)
	}

	for name, args := range map[string]struct {
		q  GraphQuery
		id string
	}{
		"unknown query": {"explode", ""},
		"missing id":    {QueryUpstream, ""},
		"unknown id":    {QueryPaths, "oc-ghost"},
	} {
		if _, err := RunGraphQuery(g, args.q, args.id, 0); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	if !QueryUpstream.NeedsID() || QuerySummary.NeedsID() || GraphQuery("x").Valid() {
		t.Error("query predicates mismatch")
	}
}

func TestBuildGraphRejectsInvalidDocument(t *testing.T) {
	doc := fullDoc(t)
	doc.Links[0].To = "oc-ghost"
	if _, err := BuildGraph(doc); err == nil {
		t.Fatal("expected an error for a dangling reference")
	}
}

func TestDOT(t *testing.T) {
	g, err := BuildGraph(fullDoc(t))
	if err != nil {
		t.Fatal(err)
	}
	dot := DOT(g)
	for _, want := range []string{
		"digraph impact-model {",
		`"oc-safer-agents" [shape=box, label="oc-safer-agents\n(impact)"];`,
		`"int-interop" [shape=ellipse`,
		`"oc-enterprise-require" -> "oc-platform-support" [style=dashed];`,
		`"int-reference" -> "oc-reference-impl" [style=dotted];`,
	} {
		if !strings.Contains(dot, want) {
			t.Errorf("DOT output missing %q:\n%s", want, dot)
		}
	}
	if DOT(g) != dot {
		t.Error("DOT is not deterministic")
	}
}

func TestInitDocument(t *testing.T) {
	doc, err := InitDocument(InitOptions{Template: "minimal", ID: "starter", Name: "Starter", Problem: "Real problem."})
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID != "starter" || doc.Metadata.Name != "Starter" || doc.Context.ProblemStatement != "Real problem." {
		t.Fatalf("doc = %+v", doc)
	}
	// It must round-trip through the YAML writer and parser.
	out, err := impactio.Canonical(doc, impactio.FormatYAML)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := impactio.ParseDocument(out, impactio.FormatYAML); err != nil {
		t.Fatalf("starter does not parse back: %v\n%s", err, out)
	}
	def, err := InitDocument(InitOptions{Template: "minimal"})
	if err != nil || def.ID != "my-impact-model" || !strings.HasPrefix(def.Context.ProblemStatement, "TODO") {
		t.Fatalf("defaults = %+v, %v", def, err)
	}
	if _, err := InitDocument(InitOptions{Template: "nope"}); !errors.Is(err, ErrUnknownTemplate) {
		t.Fatalf("unknown template error = %v", err)
	}
	if got := Templates(); len(got) == 0 || got[0] != "minimal" {
		t.Fatalf("Templates() = %v", got)
	}
}

func TestSortElements(t *testing.T) {
	doc := fullDoc(t)
	SortElements(doc)
	for i := 1; i < len(doc.Outcomes); i++ {
		if doc.Outcomes[i-1].ID > doc.Outcomes[i].ID {
			t.Fatalf("outcomes not sorted: %q before %q", doc.Outcomes[i-1].ID, doc.Outcomes[i].ID)
		}
	}
	if doc.Links[0].ID > doc.Links[1].ID {
		t.Fatal("links not sorted")
	}
}

func TestFormatFile(t *testing.T) {
	// A hand-indented file is not canonical.
	messy := copyFixture(t, "minimal.json", "doc.json")
	data, err := os.ReadFile(messy)
	if err != nil {
		t.Fatal(err)
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatal(err)
	}
	compact, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(messy, compact, 0o600); err != nil {
		t.Fatal(err)
	}

	res, err := FormatFile(messy, true, false)
	if err != nil || !res.Changed || res.Written {
		t.Fatalf("check mode = %+v, %v", res, err)
	}
	if got, _ := os.ReadFile(messy); string(got) != string(compact) {
		t.Fatal("check mode wrote the file")
	}
	res, err = FormatFile(messy, false, false)
	if err != nil || !res.Changed || !res.Written {
		t.Fatalf("write mode = %+v, %v", res, err)
	}
	res, err = FormatFile(messy, true, false)
	if err != nil || res.Changed {
		t.Fatalf("second run should be a no-op: %+v, %v", res, err)
	}

	// --sort reorders by ID, and is idempotent too.
	full := copyFixture(t, "full.json", "full.yaml")
	if err := os.WriteFile(full, mustCanonical(t, fullDoc(t), impactio.FormatYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	if res, err = FormatFile(full, false, true); err != nil || !res.Changed {
		t.Fatalf("sort should change the file: %+v, %v", res, err)
	}
	if res, err = FormatFile(full, true, true); err != nil || res.Changed {
		t.Fatalf("sort should be idempotent: %+v, %v", res, err)
	}

	// Non-document kinds format without sorting.
	status := copyFixture(t, "status.json", "status.json")
	if _, err := FormatFile(status, false, true); err != nil {
		t.Fatalf("status file: %v", err)
	}

	bad := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(bad, []byte(`{"kind":"ImpactModel"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := FormatFile(bad, false, false); err == nil || !strings.Contains(err.Error(), "bad.json") {
		t.Fatalf("invalid file error = %v", err)
	}
	if _, err := FormatFile(filepath.Join(t.TempDir(), "nope.json"), false, false); err == nil {
		t.Fatal("missing file should be an error")
	}
}

func mustCanonical(t *testing.T, v any, f impactio.Format) []byte {
	t.Helper()
	b, err := impactio.Canonical(v, f)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err) // test fixtures only
	}
	return v
}

func TestPatchFile(t *testing.T) {
	path := copyFixture(t, "full.json", "doc.json")
	rename := must(edit.Replace("/outcomes/@oc-idp-support/title", "Providers ship it"))

	// Dry run: nothing written, document returned.
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	res, err := PatchFile(path, edit.Patch{rename}, PatchOptions{DryRun: true})
	if err != nil || !res.Applied || !res.Changed || res.Written || !res.DryRun || !strings.Contains(string(res.Document), "Providers ship it") {
		t.Fatalf("dry run = %+v, %v", res, err)
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatal("dry run wrote the file")
	}

	// Real run writes canonical output in the file's format.
	res, err = PatchFile(path, edit.Patch{rename}, PatchOptions{})
	if err != nil || !res.Applied || !res.Written {
		t.Fatalf("write = %+v, %v", res, err)
	}
	doc, err := impactio.LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if o, _ := doc.Outcome("oc-idp-support"); o.Title != "Providers ship it" {
		t.Fatalf("title = %q", o.Title)
	}

	// Applying the same patch again changes nothing and writes nothing.
	res, err = PatchFile(path, edit.Patch{rename}, PatchOptions{})
	if err != nil || !res.Applied || res.Changed || res.Written {
		t.Fatalf("no-op = %+v, %v", res, err)
	}

	// A patch that adds a structural error is rejected and writes nothing.
	snapshot, _ := os.ReadFile(path)
	bad := must(edit.Add("/links/-", map[string]any{"id": "ln-bad", "from": "oc-spec-stable", "to": "oc-ghost"}))
	res, err = PatchFile(path, edit.Patch{bad}, PatchOptions{})
	if err != nil || res.Applied || res.Written || res.Reason == "" || len(res.IssuesIntroduced) != 1 || res.IssuesResolved == nil {
		t.Fatalf("rejection = %+v, %v", res, err)
	}
	if after, _ := os.ReadFile(path); string(after) != string(snapshot) {
		t.Fatal("rejected patch wrote the file")
	}
	res, err = PatchFile(path, edit.Patch{bad}, PatchOptions{AllowErrors: true})
	if err != nil || !res.Applied || !res.Written {
		t.Fatalf("AllowErrors = %+v, %v", res, err)
	}

	// A failing operation is an error and writes nothing.
	snapshot, _ = os.ReadFile(path)
	if _, err := PatchFile(path, edit.Patch{edit.Remove("/outcomes/@ghost")}, PatchOptions{}); !errors.Is(err, edit.ErrUnknownID) {
		t.Fatalf("failing op error = %v", err)
	}
	if after, _ := os.ReadFile(path); string(after) != string(snapshot) {
		t.Fatal("failing patch wrote the file")
	}
	if _, err := PatchFile(filepath.Join(t.TempDir(), "nope.json"), edit.Patch{rename}, PatchOptions{}); err == nil {
		t.Fatal("missing file should be an error")
	}
}

func TestPatchFilePreservesYAML(t *testing.T) {
	doc := fullDoc(t)
	path := filepath.Join(t.TempDir(), "doc.yaml")
	if err := os.WriteFile(path, mustCanonical(t, doc, impactio.FormatYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	rename := must(edit.Replace("/outcomes/@oc-idp-support/title", "Providers ship it"))
	if res, err := PatchFile(path, edit.Patch{rename}, PatchOptions{}); err != nil || !res.Written {
		t.Fatalf("patch YAML = %+v, %v", res, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(string(data), "{") || !strings.Contains(string(data), "Providers ship it") {
		t.Fatalf("file is no longer YAML:\n%s", data)
	}
}
