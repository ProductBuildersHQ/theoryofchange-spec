package tocio

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	oscfs "github.com/grokify/oscompat/fs"
)

const fixtures = "../toc/testdata/"

func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func TestFormatFromPath(t *testing.T) {
	tests := map[string]Format{"a.json": FormatJSON, "a.YAML": FormatYAML, "dir/a.yml": FormatYAML}
	for path, want := range tests {
		got, err := FormatFromPath(path)
		if err != nil || got != want {
			t.Errorf("FormatFromPath(%q) = %q, %v; want %q", path, got, err, want)
		}
	}
	if _, err := FormatFromPath("a.txt"); !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("FormatFromPath(a.txt) error = %v", err)
	}
	if !FormatJSON.Valid() || Format("toml").Valid() {
		t.Error("Format.Valid mismatch")
	}
}

func TestParseDocumentJSONAndYAMLAgree(t *testing.T) {
	for _, name := range []string{"minimal.json", "full.json"} {
		t.Run(name, func(t *testing.T) {
			fromJSON, err := ParseDocument(readFixture(t, fixtures+name), FormatJSON)
			if err != nil {
				t.Fatalf("ParseDocument JSON: %v", err)
			}
			yamlBytes, err := Canonical(fromJSON, FormatYAML)
			if err != nil {
				t.Fatalf("Canonical YAML: %v", err)
			}
			fromYAML, err := ParseDocument(yamlBytes, FormatYAML)
			if err != nil {
				t.Fatalf("ParseDocument YAML: %v\n%s", err, yamlBytes)
			}
			if !reflect.DeepEqual(fromJSON, fromYAML) {
				t.Fatalf("YAML and JSON decode differently:\n%s", yamlBytes)
			}
			h1, err := ContentHash(fromJSON)
			if err != nil {
				t.Fatal(err)
			}
			h2, err := ContentHash(fromYAML)
			if err != nil {
				t.Fatal(err)
			}
			if h1 != h2 || len(h1) != 64 {
				t.Fatalf("content hashes differ or malformed: %q vs %q", h1, h2)
			}
		})
	}
}

func TestHandwrittenYAML(t *testing.T) {
	doc, err := ParseDocument(readFixture(t, "testdata/handwritten.yaml"), FormatYAML)
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	if doc.Context.TimeHorizon == nil || doc.Context.TimeHorizon.Start != "2026-01-01" || doc.Context.TimeHorizon.End != "2027-12-31" {
		t.Fatalf("unquoted dates not converted: %+v", doc.Context.TimeHorizon)
	}
	if doc.Indicators[0].Baseline.Date != "2026-01-31" {
		t.Fatalf("baseline date = %q", doc.Indicators[0].Baseline.Date)
	}
	if !strings.Contains(doc.Context.Narrative, "Line one") || !strings.HasSuffix(doc.Context.Narrative, "\n") {
		t.Fatalf("literal block lost: %q", doc.Context.Narrative)
	}
	if strings.Contains(doc.Context.ProblemStatement, "\n") {
		t.Fatalf("folded scalar kept newline: %q", doc.Context.ProblemStatement)
	}
}

func TestCanonicalIsIdempotent(t *testing.T) {
	for _, f := range []Format{FormatJSON, FormatYAML} {
		t.Run(string(f), func(t *testing.T) {
			doc, err := ParseDocument(readFixture(t, fixtures+"full.json"), FormatJSON)
			if err != nil {
				t.Fatal(err)
			}
			first, err := Canonical(doc, f)
			if err != nil {
				t.Fatal(err)
			}
			again, err := Reformat(first, f)
			if err != nil {
				t.Fatalf("Reformat: %v", err)
			}
			if string(first) != string(again) {
				t.Fatalf("formatting is not idempotent:\n--- first\n%s\n--- again\n%s", first, again)
			}
			if first[len(first)-1] != '\n' {
				t.Fatal("missing trailing newline")
			}
		})
	}
}

func TestCanonicalYAMLQuotesAmbiguousStrings(t *testing.T) {
	doc := toc.NewDocument("doc", "true")
	doc.Outcomes = []toc.Outcome{{ID: "oc-1", Title: "2026-01-01", Level: toc.LevelImpact, Description: "123"}}
	out, err := Canonical(doc, FormatYAML)
	if err != nil {
		t.Fatal(err)
	}
	back, err := ParseDocument(out, FormatYAML)
	if err != nil {
		t.Fatalf("ParseDocument: %v\n%s", err, out)
	}
	if back.Context.ProblemStatement != "true" || back.Outcomes[0].Title != "2026-01-01" || back.Outcomes[0].Description != "123" {
		t.Fatalf("ambiguous strings changed type:\n%s", out)
	}
}

func TestCanonicalYAMLKeepsFieldOrder(t *testing.T) {
	doc, err := ParseDocument(readFixture(t, fixtures+"minimal.json"), FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Canonical(doc, FormatYAML)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	prev := -1
	for _, key := range []string{"kind:", "specVersion:", "id:", "outcomes:"} {
		i := strings.Index(s, key)
		if i <= prev {
			t.Fatalf("YAML key %q is out of model order:\n%s", key, s)
		}
		prev = i
	}
}

func TestParseDocumentErrors(t *testing.T) {
	tests := map[string]struct {
		data   string
		format Format
		want   error
	}{
		"bad json":       {`{"kind":`, FormatJSON, ErrInvalidDocument},
		"bad yaml":       {"kind: [", FormatYAML, ErrInvalidDocument},
		"two yaml docs":  {"a: 1\n---\nb: 2\n", FormatYAML, ErrInvalidDocument},
		"non-string key": {"1: a\n", FormatYAML, ErrInvalidDocument},
		"schema failure": {`{"kind":"TheoryOfChange"}`, FormatJSON, ErrInvalidDocument},
		"unknown field":  {strings.Replace(string(readFixtureNoT("minimal.json")), `"id": "minimal",`, `"id": "minimal", "bogus": true,`, 1), FormatJSON, ErrInvalidDocument},
		"bad format":     {`{}`, Format("toml"), ErrUnsupportedFormat},
		"too large":      {strings.Repeat(" ", MaxFileBytes+1), FormatJSON, ErrTooLarge},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ParseDocument([]byte(tt.data), tt.format)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func readFixtureNoT(name string) []byte {
	data, err := os.ReadFile(fixtures + name)
	if err != nil {
		panic(err)
	}
	return data
}

func TestSchemaErrorIncludesLocation(t *testing.T) {
	bad := strings.Replace(string(readFixtureNoT("minimal.json")), `"level": "impact"`, `"level": "huge"`, 1)
	_, err := ParseDocument([]byte(bad), FormatJSON)
	if err == nil || !strings.Contains(err.Error(), "/outcomes/0/level") {
		t.Fatalf("expected error naming /outcomes/0/level, got %v", err)
	}
}

func TestDetectKindAndSchemaName(t *testing.T) {
	tests := map[string]toc.Kind{
		"minimal.json":      toc.KindTheoryOfChange,
		"observations.json": toc.KindObservationSet,
		"status.json":       toc.KindStatus,
		"registry.json":     toc.KindActorRegistry,
	}
	for name, want := range tests {
		got, err := DetectKind(readFixture(t, fixtures+name), FormatJSON)
		if err != nil || got != want {
			t.Errorf("DetectKind(%s) = %q, %v; want %q", name, got, err, want)
		}
		if _, err := SchemaName(got); err != nil {
			t.Errorf("SchemaName(%q): %v", got, err)
		}
	}
	if _, err := DetectKind([]byte(`{"a":1}`), FormatJSON); !errors.Is(err, ErrInvalidDocument) {
		t.Errorf("missing kind error = %v", err)
	}
	if _, err := SchemaName("Nope"); !errors.Is(err, ErrInvalidDocument) {
		t.Errorf("unknown kind error = %v", err)
	}
}

func TestParseOtherKinds(t *testing.T) {
	if s, err := ParseObservationSet(readFixture(t, fixtures+"observations.json"), FormatJSON); err != nil || len(s.Observations) != 2 {
		t.Fatalf("ParseObservationSet: %v %+v", err, s)
	}
	if s, err := ParseStatus(readFixture(t, fixtures+"status.json"), FormatJSON); err != nil || s.TocID != "open-agent-auth" {
		t.Fatalf("ParseStatus: %v", err)
	}
	if r, err := ParseActorRegistry(readFixture(t, fixtures+"registry.json"), FormatJSON); err != nil || len(r.Actors) != 2 {
		t.Fatalf("ParseActorRegistry: %v", err)
	}
}

func TestReformatEveryKind(t *testing.T) {
	for _, name := range []string{"full.json", "observations.json", "status.json", "registry.json"} {
		out, err := Reformat(readFixture(t, fixtures+name), FormatJSON)
		if err != nil {
			t.Fatalf("Reformat(%s): %v", name, err)
		}
		again, err := Reformat(out, FormatJSON)
		if err != nil || string(out) != string(again) {
			t.Fatalf("Reformat(%s) not idempotent (err %v)", name, err)
		}
	}
	if _, err := Reformat([]byte(`{"kind":"Nope"}`), FormatJSON); !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("unknown kind error = %v", err)
	}
}

func TestLoadDocumentAndObservations(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "doc.yaml")
	if err := WriteFile(docPath, readFixture(t, "testdata/handwritten.yaml")); err != nil {
		t.Fatal(err)
	}
	doc, err := LoadDocument(docPath)
	if err != nil || doc.ID != "handwritten" {
		t.Fatalf("LoadDocument: %v %+v", err, doc)
	}
	if _, err := LoadDocument(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("LoadDocument accepted a missing file")
	}
	if _, err := LoadDocument(filepath.Join(dir, "doc.txt")); !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("unsupported extension error = %v", err)
	}
	bad := filepath.Join(dir, "bad.json")
	if err := WriteFile(bad, []byte(`{"kind":"TheoryOfChange"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDocument(bad); err == nil || !strings.Contains(err.Error(), "bad.json") {
		t.Fatalf("error should name the file: %v", err)
	}

	obsPath := filepath.Join(dir, "obs.json")
	if err := WriteFile(obsPath, readFixture(t, fixtures+"observations.json")); err != nil {
		t.Fatal(err)
	}
	sets, err := LoadObservations(obsPath, obsPath)
	if err != nil || len(sets) != 2 {
		t.Fatalf("LoadObservations: %v", err)
	}
	if _, err := LoadObservations(bad); err == nil {
		t.Fatal("LoadObservations accepted a document as an observation set")
	}
}

func TestMergeObservations(t *testing.T) {
	a := toc.NewObservationSet("doc", "2026-Q3")
	a.Observations = []toc.Observation{{IndicatorID: "ind-1", Date: "2026-09-30"}}
	b := toc.NewObservationSet("doc", "2026-Q4")
	b.Observations = []toc.Observation{{IndicatorID: "ind-1", Date: "2026-12-31"}}
	b.Notes = []toc.Note{{Text: "n"}}

	m, err := MergeObservations(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Observations) != 2 || m.Observations[0].Date != "2026-09-30" || len(m.Notes) != 1 || m.Period != "" {
		t.Fatalf("unexpected merge: %+v", m)
	}
	same, err := MergeObservations(a, a)
	if err != nil || same.Period != "2026-Q3" || len(same.Observations) != 2 {
		t.Fatalf("same-period merge: %v %+v", err, same)
	}
	if _, err := MergeObservations(a, toc.NewObservationSet("other", "")); err == nil {
		t.Fatal("merged sets for different documents")
	}
	if _, err := MergeObservations(); err == nil {
		t.Fatal("merged nothing")
	}
}

func TestElementLimit(t *testing.T) {
	doc := toc.NewDocument("big", "p")
	doc.Outcomes = make([]toc.Outcome, MaxElements+1)
	for i := range doc.Outcomes {
		doc.Outcomes[i] = toc.Outcome{ID: "oc-" + strings.Repeat("a", 1) + "-" + itoa(i), Title: "t", Level: toc.LevelOutput}
	}
	data, err := Canonical(doc, FormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > MaxFileBytes {
		t.Skipf("fixture exceeds byte limit first (%d bytes)", len(data))
	}
	if _, err := ParseDocument(data, FormatJSON); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("error = %v, want ErrTooLarge", err)
	}
}

func itoa(i int) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var b []byte
	for ; i > 0; i /= 10 {
		b = append([]byte{digits[i%10]}, b...)
	}
	return string(b)
}

func TestCanonicalUnsupportedFormat(t *testing.T) {
	if _, err := Canonical(toc.NewDocument("d", "p"), Format("toml")); !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("error = %v", err)
	}
}

func TestWriteFileIsAtomicAndKeepsMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.json")
	if err := WriteFile(path, []byte("one")); err != nil {
		t.Fatal(err)
	}
	modes := oscfs.PermissionsEnforced() // Windows has no Unix permission bits
	if info, err := os.Stat(path); err != nil || (modes && info.Mode().Perm() != 0o644) {
		t.Fatalf("new file mode = %v (%v)", info.Mode().Perm(), err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(path, []byte("two")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "two" {
		t.Fatalf("content = %q (%v)", got, err)
	}
	if info, _ := os.Stat(path); modes && info.Mode().Perm() != 0o600 {
		t.Fatalf("existing mode not preserved: %v", info.Mode().Perm())
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v (%v)", entries, err)
	}
	if err := WriteFile(filepath.Join(dir, "missing-dir", "x.json"), []byte("x")); err == nil {
		t.Fatal("expected an error writing into a missing directory")
	}
}
