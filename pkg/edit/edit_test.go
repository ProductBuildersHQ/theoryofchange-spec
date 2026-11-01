package edit

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
	"github.com/ProductBuildersHQ/impact-model/pkg/validate"
)

func loadFull(t *testing.T) *impact.Model {
	t.Helper()
	data, err := os.ReadFile("../impact/testdata/full.json")
	if err != nil {
		t.Fatalf("read full.json: %v", err)
	}
	var d impact.Model
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatalf("decode full.json: %v", err)
	}
	return &d
}

func snapshot(t *testing.T, d *impact.Model) string {
	t.Helper()
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err) // test fixtures only
	}
	return v
}

func apply(t *testing.T, d *impact.Model, ops ...Op) *Result {
	t.Helper()
	res, err := Apply(d, Patch(ops), Options{})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return res
}

func TestOperations(t *testing.T) {
	newOutcome := map[string]any{"id": "oc-new", "title": "A new outcome", "level": "short_term"}
	tests := []struct {
		name  string
		ops   []Op
		check func(t *testing.T, d *impact.Model)
	}{
		{"append outcome", []Op{must(Add("/outcomes/-", newOutcome))}, func(t *testing.T, d *impact.Model) {
			if last := d.Outcomes[len(d.Outcomes)-1]; last.ID != "oc-new" {
				t.Fatalf("last outcome = %q", last.ID)
			}
		}},
		{"insert at index", []Op{must(Add("/outcomes/0", newOutcome))}, func(t *testing.T, d *impact.Model) {
			if d.Outcomes[0].ID != "oc-new" || d.Outcomes[1].ID != "oc-safer-agents" {
				t.Fatalf("insert order wrong: %q %q", d.Outcomes[0].ID, d.Outcomes[1].ID)
			}
		}},
		{"replace by id", []Op{must(Replace("/outcomes/@oc-idp-support/title", "Providers ship it"))}, func(t *testing.T, d *impact.Model) {
			if o, _ := d.Outcome("oc-idp-support"); o.Title != "Providers ship it" {
				t.Fatalf("title = %q", o.Title)
			}
		}},
		{"add optional field by id", []Op{must(Add("/outcomes/@oc-spec-stable/description", "Stable means versioned and tested."))}, func(t *testing.T, d *impact.Model) {
			if o, _ := d.Outcome("oc-spec-stable"); o.Description == "" {
				t.Fatal("description not added")
			}
		}},
		{"remove link by id", []Op{Remove("/links/@ln-require-platform-loop")}, func(t *testing.T, d *impact.Model) {
			if _, ok := d.Link("ln-require-platform-loop"); ok {
				t.Fatal("link still present")
			}
		}},
		{"remove optional field", []Op{Remove("/outcomes/@oc-idp-support/tags")}, func(t *testing.T, d *impact.Model) {
			if o, _ := d.Outcome("oc-idp-support"); len(o.Tags) != 0 {
				t.Fatal("tags still present")
			}
		}},
		{"move within array", []Op{Move("/outcomes/@oc-reference-impl", "/outcomes/0")}, func(t *testing.T, d *impact.Model) {
			if d.Outcomes[0].ID != "oc-reference-impl" {
				t.Fatalf("first outcome = %q", d.Outcomes[0].ID)
			}
		}},
		{"copy value", []Op{Copy("/outcomes/@oc-idp-support/title", "/outcomes/@oc-platform-support/description")}, func(t *testing.T, d *impact.Model) {
			if o, _ := d.Outcome("oc-platform-support"); o.Description != "Major identity providers support the standard" {
				t.Fatalf("description = %q", o.Description)
			}
		}},
		{"test then replace", []Op{
			must(Test("/outcomes/@oc-idp-support/level", "medium_term")),
			must(Replace("/outcomes/@oc-idp-support/level", "short_term")),
		}, func(t *testing.T, d *impact.Model) {
			if o, _ := d.Outcome("oc-idp-support"); o.Level != impact.LevelShortTerm {
				t.Fatalf("level = %q", o.Level)
			}
		}},
		{"nested id path", []Op{must(Replace("/indicators/@ind-idp-count/targets/1/value", 6))}, func(t *testing.T, d *impact.Model) {
			if ind, _ := d.Indicator("ind-idp-count"); *ind.Targets[1].Value != 6 {
				t.Fatalf("target = %v", *ind.Targets[1].Value)
			}
		}},
		{"escaped key", []Op{must(Add("/context/market/attributes/"+Pointer("a/b")[1:], "x"))}, func(t *testing.T, d *impact.Model) {
			if d.Context.Market.Attributes["a/b"] != "x" {
				t.Fatalf("attributes = %v", d.Context.Market.Attributes)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := loadFull(t)
			before := snapshot(t, doc)
			res := apply(t, doc, tt.ops...)
			if !res.Applied {
				t.Fatalf("not applied: %s; introduced: %v", res.Reason, res.IssuesIntroduced)
			}
			if !res.Changed {
				t.Fatal("Changed = false")
			}
			tt.check(t, res.Document)
			if snapshot(t, doc) != before {
				t.Fatal("Apply mutated its input")
			}
			if res.Document == doc {
				t.Fatal("result aliases the input")
			}
		})
	}
}

func TestNoOpPatchIsNotChanged(t *testing.T) {
	doc := loadFull(t)
	res := apply(t, doc, must(Test("/id", "open-agent-auth")), must(Replace("/id", "open-agent-auth")))
	if !res.Applied || res.Changed {
		t.Fatalf("Applied=%v Changed=%v", res.Applied, res.Changed)
	}
}

func TestPatchErrors(t *testing.T) {
	tests := []struct {
		name string
		ops  []Op
		want error
	}{
		{"unknown id", []Op{must(Replace("/outcomes/@nope/title", "x"))}, ErrUnknownID},
		{"unknown member", []Op{Remove("/context/nothing")}, ErrPath},
		{"replace missing member", []Op{must(Replace("/context/narrativeX", "x"))}, ErrPath},
		{"index out of range", []Op{Remove("/outcomes/99")}, ErrPath},
		{"non-canonical index", []Op{Remove("/outcomes/01")}, ErrPath},
		{"negative index", []Op{Remove("/outcomes/-1")}, ErrPath},
		{"dash outside add", []Op{Remove("/outcomes/-")}, ErrPath},
		{"add at id", []Op{must(Add("/outcomes/@oc-spec-stable", map[string]any{"id": "x"}))}, ErrPath},
		{"add index past end", []Op{must(Add("/outcomes/99", map[string]any{"id": "x"}))}, ErrPath},
		{"step into scalar", []Op{must(Replace("/id/x", "y"))}, ErrPath},
		{"failed test", []Op{must(Test("/outcomes/@oc-idp-support/level", "impact"))}, ErrTestFailed},
		{"move into itself", []Op{Move("/outcomes/0", "/outcomes/0/title")}, ErrPath},
		{"copy from missing", []Op{Copy("/context/zzz", "/context/narrativeCopy")}, ErrPath},
		{"unknown op", []Op{{Op: "frobnicate", Path: "/id"}}, ErrUnknownOp},
		{"root path", []Op{{Op: OpRemove, Path: ""}}, ErrPath},
		{"relative path", []Op{{Op: OpRemove, Path: "id"}}, ErrPath},
		{"add without value", []Op{{Op: OpAdd, Path: "/outcomes/-"}}, ErrPatch},
		{"move without from", []Op{{Op: OpMove, Path: "/outcomes/0"}}, ErrPatch},
		{"invalid value JSON", []Op{{Op: OpAdd, Path: "/context/narrative", Value: json.RawMessage(`{`)}}, ErrPatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := loadFull(t)
			before := snapshot(t, doc)
			res, err := Apply(doc, Patch(tt.ops), Options{})
			if !errors.Is(err, tt.want) || !errors.Is(err, ErrPatch) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if res != nil {
				t.Fatal("expected a nil Result on error")
			}
			var oe *OpError
			if !errors.As(err, &oe) && !errors.Is(err, ErrPatch) {
				t.Fatalf("expected an *OpError, got %T", err)
			}
			if snapshot(t, doc) != before {
				t.Fatal("document changed despite error")
			}
		})
	}
	if _, err := Apply(nil, nil, Options{}); !errors.Is(err, ErrPatch) {
		t.Fatalf("nil document error = %v", err)
	}
}

func TestAtomicity(t *testing.T) {
	doc := loadFull(t)
	before := snapshot(t, doc)
	_, err := Apply(doc, Patch{
		must(Replace("/outcomes/@oc-idp-support/title", "Changed")),
		Remove("/outcomes/@missing"),
	}, Options{})
	var oe *OpError
	if !errors.As(err, &oe) || oe.Index != 1 {
		t.Fatalf("expected an OpError at index 1, got %v", err)
	}
	if snapshot(t, doc) != before {
		t.Fatal("first operation leaked despite later failure")
	}
}

func TestRejectsErrorIntroducingPatch(t *testing.T) {
	doc := loadFull(t)
	before := snapshot(t, doc)
	bad := must(Add("/links/-", map[string]any{"id": "ln-bad", "from": "oc-spec-stable", "to": "oc-ghost"}))

	res := apply(t, doc, bad)
	if res.Applied || res.Document != nil || res.Reason == "" {
		t.Fatalf("expected rejection with a reason: %+v", res)
	}
	if len(res.IssuesIntroduced) != 1 || res.IssuesIntroduced[0].Code != "VAL-REF-001" {
		t.Fatalf("IssuesIntroduced = %v", res.IssuesIntroduced)
	}
	if snapshot(t, doc) != before {
		t.Fatal("rejected patch changed the document")
	}

	allowed, err := Apply(doc, Patch{bad}, Options{AllowErrors: true})
	if err != nil {
		t.Fatal(err)
	}
	if !allowed.Applied || allowed.Document == nil || len(allowed.IssuesIntroduced) != 1 {
		t.Fatalf("AllowErrors should accept and still report: %+v", allowed)
	}
}

func TestPatchThatResolvesErrors(t *testing.T) {
	doc := loadFull(t)
	doc.Links[0].To = "oc-ghost"
	if !validate.HasErrors(validate.Document(doc)) {
		t.Fatal("test setup: document should start invalid")
	}
	res := apply(t, doc, must(Replace("/links/0/to", "oc-idp-support")))
	if !res.Applied || len(res.IssuesResolved) != 1 || len(res.IssuesIntroduced) != 0 {
		t.Fatalf("fix should apply and resolve one issue: %+v", res)
	}
	if validate.HasErrors(validate.Document(res.Document)) {
		t.Fatal("fixed document still has errors")
	}
}

func TestEditsToAnAlreadyInvalidDocumentAreAllowedWhenNoWorse(t *testing.T) {
	doc := loadFull(t)
	doc.Links[0].To = "oc-ghost"
	res := apply(t, doc, must(Replace("/outcomes/@oc-idp-support/title", "Unrelated improvement")))
	if !res.Applied {
		t.Fatalf("an unrelated edit to an already-invalid document should apply: %s", res.Reason)
	}
}

func TestRejectsSchemaViolation(t *testing.T) {
	doc := loadFull(t)
	before := snapshot(t, doc)
	for _, opts := range []Options{{}, {AllowErrors: true}} {
		res, err := Apply(doc, Patch{must(Replace("/outcomes/@oc-idp-support/level", 5))}, opts)
		if err != nil {
			t.Fatal(err)
		}
		if res.Applied || res.Document != nil {
			t.Fatalf("schema-violating patch applied (AllowErrors=%v)", opts.AllowErrors)
		}
		found := false
		for _, i := range res.IssuesIntroduced {
			if i.Code == "VAL-SCH-001" && strings.HasPrefix(i.Path, "/outcomes/") {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected a schema issue under /outcomes: %v", res.IssuesIntroduced)
		}
	}
	if snapshot(t, doc) != before {
		t.Fatal("document changed")
	}
	// An unknown property is a schema violation too.
	res, err := Apply(doc, Patch{must(Add("/outcomes/@oc-idp-support/bogus", true))}, Options{})
	if err != nil || res.Applied {
		t.Fatalf("unknown property should be rejected: err=%v applied=%v", err, res != nil && res.Applied)
	}
}

func TestNumbersSurvivePatching(t *testing.T) {
	doc := loadFull(t)
	res := apply(t, doc,
		must(Add("/indicators/@ind-idp-count/tolerance", 0.125)),
		must(Replace("/indicators/@ind-idp-count/population/size", 12345678901)),
	)
	if !res.Applied {
		t.Fatalf("not applied: %s", res.Reason)
	}
	ind, _ := res.Document.Indicator("ind-idp-count")
	if ind.Tolerance != 0.125 || *ind.Population.Size != 12345678901 {
		t.Fatalf("numbers changed: tolerance=%v size=%v", ind.Tolerance, *ind.Population.Size)
	}
}

func TestDeterministic(t *testing.T) {
	doc := loadFull(t)
	op := must(Replace("/outcomes/@oc-idp-support/title", "Same"))
	a, b := apply(t, doc, op), apply(t, doc, op)
	if snapshot(t, a.Document) != snapshot(t, b.Document) {
		t.Fatal("Apply is not deterministic")
	}
}

func TestParsePatch(t *testing.T) {
	good := `[
  {"op":"add","path":"/outcomes/-","value":{"id":"oc-x","title":"t","level":"output"}},
  {"op":"remove","path":"/links/@ln-1"},
  {"op":"move","from":"/outcomes/0","path":"/outcomes/1"},
  {"op":"test","path":"/id","value":null}
]`
	p, err := ParsePatch([]byte(good))
	if err != nil || len(p) != 4 {
		t.Fatalf("ParsePatch: %v (%d ops)", err, len(p))
	}
	if string(p[3].Value) != "null" {
		t.Fatalf("null value lost: %q", p[3].Value)
	}
	for name, bad := range map[string]string{
		"not an array":  `{"op":"add"}`,
		"unknown field": `[{"op":"remove","path":"/id","extra":1}]`,
		"unknown op":    `[{"op":"explode","path":"/id"}]`,
		"missing path":  `[{"op":"remove"}]`,
		"missing value": `[{"op":"replace","path":"/id"}]`,
		"missing from":  `[{"op":"copy","path":"/id"}]`,
		"malformed":     `[{`,
		"relative from": `[{"op":"move","from":"id","path":"/id"}]`,
	} {
		if _, err := ParsePatch([]byte(bad)); !errors.Is(err, ErrPatch) {
			t.Errorf("%s: error = %v, want ErrPatch", name, err)
		}
	}
}

func TestBuildersRoundTripThroughParse(t *testing.T) {
	p := Patch{
		must(Add("/outcomes/-", map[string]any{"id": "oc-x", "title": "t", "level": "output"})),
		must(Replace("/id", "renamed")),
		must(Test("/id", "open-agent-auth")),
		Remove("/links/@ln-1"),
		Move("/outcomes/0", "/outcomes/1"),
		Copy("/id", "/context/narrative"),
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	back, err := ParsePatch(data)
	if err != nil || len(back) != len(p) {
		t.Fatalf("round trip: %v (%d ops)", err, len(back))
	}
	if _, err := Add("/x", make(chan int)); !errors.Is(err, ErrPatch) {
		t.Fatalf("unencodable value error = %v", err)
	}
}

func TestPointerEscaping(t *testing.T) {
	if got := Pointer("a/b", "c~d", "@id", "-"); got != "/a~1b/c~0d/@id/-" {
		t.Fatalf("Pointer = %q", got)
	}
}

func TestEqualHandlesNumbers(t *testing.T) {
	a, b := json.Number("1"), json.Number("1.0")
	if !equal(a, b) || equal(json.Number("1"), json.Number("2")) {
		t.Fatal("numeric equality mismatch")
	}
	if equal(map[string]any{"a": a}, map[string]any{"b": a}) || equal([]any{a}, []any{a, a}) || equal(a, "1") {
		t.Fatal("structural inequality not detected")
	}
}

func TestOpErrorText(t *testing.T) {
	e := &OpError{Index: 2, Op: Op{Op: OpRemove, Path: "/x"}, Err: ErrPath}
	if !strings.Contains(e.Error(), "patch op 2") || !errors.Is(e, ErrPath) {
		t.Fatalf("OpError = %v", e)
	}
}
