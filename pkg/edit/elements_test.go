package edit

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/validate"
)

func TestKindsAreUnique(t *testing.T) {
	names, fields := map[string]bool{}, map[string]bool{}
	for _, k := range ElementKinds() {
		if names[k.Name] || fields[k.Field] {
			t.Errorf("duplicate kind %+v", k)
		}
		names[k.Name], fields[k.Field] = true, true
		if got, ok := KindByName(k.Name); !ok || got != k {
			t.Errorf("KindByName(%q) = %+v, %v", k.Name, got, ok)
		}
	}
	if _, ok := KindByName("nonsense"); ok {
		t.Error("KindByName accepted an unknown kind")
	}
	// Every kind must correspond to a toc.Elements() kind with that array.
	doc := loadFull(t)
	have := map[string]bool{}
	for _, e := range doc.Elements() {
		have[strings.Split(e.Path, "/")[1]] = true
	}
	for _, k := range ElementKinds() {
		if !have[k.Field] {
			t.Errorf("kind %s: field %q is not an identified-element array in the model", k.Name, k.Field)
		}
	}
}

func TestParseValue(t *testing.T) {
	tests := map[string]any{
		"3":          float64(3),
		"true":       true,
		"null":       nil,
		`["a","b"]`:  []any{"a", "b"},
		`{"x":1}`:    map[string]any{"x": float64(1)},
		`"2026"`:     "2026",
		"plain text": "plain text",
		"":           "",
		"oc-impact":  "oc-impact",
	}
	for in, want := range tests {
		if got := ParseValue(in); !reflect.DeepEqual(got, want) {
			t.Errorf("ParseValue(%q) = %#v, want %#v", in, got, want)
		}
	}
}

func TestExpandDotted(t *testing.T) {
	got, err := ExpandDotted(map[string]any{"id": "g1", "logic.type": "at_least", "logic.k": 2})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"id": "g1", "logic": map[string]any{"type": "at_least", "k": 2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ExpandDotted = %#v", got)
	}
	for name, in := range map[string]map[string]any{
		"empty segment": {"a..b": 1},
		"conflict":      {"a": 1, "a.b": 2},
		"conflict rev":  {"a.b": 2, "a": 1},
	} {
		if _, err := ExpandDotted(in); !errors.Is(err, ErrPatch) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
}

func outcomeKind(t *testing.T) ElementKind {
	t.Helper()
	k, ok := KindByName("outcome")
	if !ok {
		t.Fatal("outcome kind missing")
	}
	return k
}

func TestAddElementEqualsHandWrittenPatch(t *testing.T) {
	fields := map[string]any{"id": "oc-new", "title": "A new outcome", "level": "short_term", "logic.type": "any"}
	p, err := AddElement(outcomeKind(t), fields)
	if err != nil {
		t.Fatal(err)
	}
	doc := loadFull(t)
	viaCommand := apply(t, doc, p...)

	hand := must(Add("/outcomes/-", map[string]any{"id": "oc-new", "title": "A new outcome", "level": "short_term", "logic": map[string]any{"type": "any"}}))
	viaPatch := apply(t, doc, hand)
	if !viaCommand.Applied || snapshot(t, viaCommand.Document) != snapshot(t, viaPatch.Document) {
		t.Fatalf("command and patch differ:\n%s\n%s", snapshot(t, viaCommand.Document), snapshot(t, viaPatch.Document))
	}

	if _, err := AddElement(outcomeKind(t), map[string]any{"title": "no id"}); !errors.Is(err, ErrPatch) {
		t.Fatalf("missing id error = %v", err)
	}
	if _, err := AddElement(outcomeKind(t), map[string]any{"id": "x", "a.": 1}); !errors.Is(err, ErrPatch) {
		t.Fatalf("bad dotted key error = %v", err)
	}

	// A duplicate ID is rejected by validation, not by the builder.
	dup, err := AddElement(outcomeKind(t), map[string]any{"id": "oc-idp-support", "title": "dup", "level": "output"})
	if err != nil {
		t.Fatal(err)
	}
	if res := apply(t, doc, dup...); res.Applied {
		t.Fatal("a duplicate ID should be rejected")
	}
}

func TestSetFields(t *testing.T) {
	doc := loadFull(t)
	p, err := SetFields(outcomeKind(t), "oc-idp-support",
		map[string]any{"title": "Providers ship it", "description": "Added text", "logic.type": "any"},
		[]string{"tags"})
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 4 || p[0].Op != OpAdd || p[3].Op != OpRemove {
		t.Fatalf("patch = %+v", p)
	}
	if _, err := Apply(doc, p, Options{}); !errors.Is(err, ErrPath) {
		// logic does not exist on this outcome, so setting logic.type must fail loudly.
		t.Fatalf("setting a nested field on a missing parent: error = %v", err)
	}
}

func TestSetFieldsOnExistingNestedAndPlainFields(t *testing.T) {
	doc := loadFull(t)
	p, err := SetFields(outcomeKind(t), "oc-enterprise-require",
		map[string]any{"title": "Enterprises require it", "logic.k": 2},
		[]string{"population"})
	if err != nil {
		t.Fatal(err)
	}
	res := apply(t, doc, p...)
	if !res.Applied {
		t.Fatalf("not applied: %s %v", res.Reason, res.IssuesIntroduced)
	}
	o, _ := res.Document.Outcome("oc-enterprise-require")
	if o.Title != "Enterprises require it" || o.Logic.K != 2 || o.Population != "" {
		t.Fatalf("outcome = %+v", o)
	}
}

func TestSetFieldsErrors(t *testing.T) {
	k := outcomeKind(t)
	for name, call := range map[string]func() error{
		"set id":     func() error { _, err := SetFields(k, "oc-x", map[string]any{"id": "y"}, nil); return err },
		"unset id":   func() error { _, err := SetFields(k, "oc-x", nil, []string{"id"}); return err },
		"no changes": func() error { _, err := SetFields(k, "oc-x", nil, nil); return err },
	} {
		if err := call(); !errors.Is(err, ErrPatch) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
	// An unknown element fails when the patch is applied.
	p, err := SetFields(k, "oc-ghost", map[string]any{"title": "t"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(loadFull(t), p, Options{}); !errors.Is(err, ErrUnknownID) {
		t.Fatalf("unknown element error = %v", err)
	}
}

func kindOf(t *testing.T, doc *toc.Document, id string) ElementKind {
	t.Helper()
	e, ok := doc.Element(id)
	if !ok {
		t.Fatalf("no element %q", id)
	}
	field := strings.Split(e.Path, "/")[1]
	for _, k := range ElementKinds() {
		if k.Field == field {
			return k
		}
	}
	t.Fatalf("no editable kind for %s", field)
	return ElementKind{}
}

func mentions(v any, id, key string) bool {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			if k == "changedIds" {
				continue // revision history may name removed elements
			}
			if mentions(child, id, k) {
				return true
			}
		}
	case []any:
		for _, child := range x {
			if mentions(child, id, key) {
				return true
			}
		}
	case string:
		return x == id && key != ""
	}
	return false
}

// TestCascadeRemovesEveryElementCleanly is the drift guard for refSites:
// removing any element with cascade must apply without adding errors and
// leave no reference to it behind.
func TestCascadeRemovesEveryElementCleanly(t *testing.T) {
	doc := loadFull(t)
	for _, e := range doc.Elements() {
		k := kindOf(t, doc, e.ID)
		t.Run(e.ID, func(t *testing.T) {
			p, err := RemoveElement(doc, k, e.ID, true)
			if err != nil {
				t.Fatalf("RemoveElement: %v", err)
			}
			res, err := Apply(doc, p, Options{})
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}
			if !res.Applied {
				t.Fatalf("cascade removal rejected: %s; introduced: %v", res.Reason, res.IssuesIntroduced)
			}
			if validate.HasErrors(validate.Document(res.Document)) {
				t.Fatalf("result has errors: %v", validate.Document(res.Document))
			}
			if _, still := res.Document.Element(e.ID); still {
				t.Fatalf("%s is still present", e.ID)
			}
			raw, err := json.Marshal(res.Document)
			if err != nil {
				t.Fatal(err)
			}
			tree, err := decodeTree(raw)
			if err != nil {
				t.Fatal(err)
			}
			if mentions(tree, e.ID, "") {
				t.Fatalf("a reference to %s survived the cascade", e.ID)
			}
		})
	}
}

func TestRemoveWithoutCascade(t *testing.T) {
	doc := loadFull(t)

	// An unreferenced element removes cleanly.
	k := kindOf(t, doc, "rsk-competing")
	p, err := RemoveElement(doc, k, "rsk-competing", false)
	if err != nil || len(p) != 1 {
		t.Fatalf("patch = %+v, %v", p, err)
	}
	if res := apply(t, doc, p...); !res.Applied {
		t.Fatalf("unreferenced removal rejected: %s", res.Reason)
	}

	// A referenced element is rejected, and the rejection names the references.
	k = kindOf(t, doc, "oc-idp-support")
	p, err = RemoveElement(doc, k, "oc-idp-support", false)
	if err != nil {
		t.Fatal(err)
	}
	res := apply(t, doc, p...)
	if res.Applied || len(res.IssuesIntroduced) == 0 {
		t.Fatalf("referenced removal should be rejected: %+v", res)
	}
	errorsSeen := 0
	for _, i := range res.IssuesIntroduced {
		if i.Severity != validate.SeverityError {
			continue // a stale revision changedIds entry is only a warning
		}
		errorsSeen++
		if i.Code != "VAL-REF-001" && i.Code != "VAL-INT-001" && i.Code != "VAL-GRP-001" {
			t.Errorf("unexpected issue %s", i)
		}
	}
	if errorsSeen == 0 {
		t.Errorf("rejection should report the dangling references: %v", res.IssuesIntroduced)
	}
}

func TestCascadeConsequences(t *testing.T) {
	doc := loadFull(t)

	// Removing an outcome removes its links, shrinks the group below two
	// members (removing the group), and drops it from intervention targets.
	k := kindOf(t, doc, "oc-idp-support")
	p, err := RemoveElement(doc, k, "oc-idp-support", true)
	if err != nil {
		t.Fatal(err)
	}
	res := apply(t, doc, p...)
	if !res.Applied {
		t.Fatalf("rejected: %s", res.Reason)
	}
	d := res.Document
	if _, ok := d.Link("ln-spec-idp"); ok {
		t.Error("link into the removed outcome survived")
	}
	if _, ok := d.Link("ln-idp-require"); ok {
		t.Error("link out of the removed outcome survived")
	}
	if _, ok := d.Group("grp-supply"); ok {
		t.Error("a group left with one member should be removed")
	}
	in, _ := d.Intervention("int-interop")
	if len(in.Targets) != 1 || in.Targets[0] != "oc-spec-stable" {
		t.Errorf("intervention targets = %v", in.Targets)
	}
	if _, ok := d.Indicator("ind-idp-count"); ok {
		t.Error("indicator of the removed outcome survived")
	}
	// The assumption lost its link and its evidence is untouched.
	a, _ := d.Assumption("asm-idp-incentive")
	if len(a.AppliesTo) != 0 {
		t.Errorf("assumption still applies to %v", a.AppliesTo)
	}
	// Its monitoring indicator measures the assumption, not the removed outcome.
	if len(a.IndicatorIDs) != 1 || a.IndicatorIDs[0] != "ind-idp-roadmaps" {
		t.Errorf("assumption indicators = %v", a.IndicatorIDs)
	}

	// Removing the last target of an intervention removes the intervention.
	k = kindOf(t, doc, "oc-reference-impl")
	p, err = RemoveElement(doc, k, "oc-reference-impl", true)
	if err != nil {
		t.Fatal(err)
	}
	if res = apply(t, doc, p...); !res.Applied {
		t.Fatalf("rejected: %s", res.Reason)
	}
	if _, ok := res.Document.Intervention("int-reference"); ok {
		t.Error("an intervention left with no targets should be removed")
	}
	if _, ok := res.Document.Indicator("ind-reference-downloads"); ok {
		t.Error("an indicator of a removed intervention should be removed")
	}

	// Removing the perspective actor removes the perspective; removing the
	// accountability ceiling's outcome clears the ceiling.
	k = kindOf(t, doc, "act-proponent")
	p, err = RemoveElement(doc, k, "act-proponent", true)
	if err != nil {
		t.Fatal(err)
	}
	if res = apply(t, doc, p...); !res.Applied || res.Document.Perspective != nil {
		t.Fatalf("perspective should be removed with its actor (applied=%v)", res.Applied)
	}

	withCeiling := loadFull(t)
	withCeiling.Ceiling = &toc.Ceiling{OutcomeIDs: []string{"oc-enterprise-require"}}
	k = kindOf(t, withCeiling, "oc-enterprise-require")
	p, err = RemoveElement(withCeiling, k, "oc-enterprise-require", true)
	if err != nil {
		t.Fatal(err)
	}
	if res = apply(t, withCeiling, p...); !res.Applied || res.Document.Ceiling != nil {
		t.Fatalf("ceiling should be cleared (applied=%v)", res.Applied)
	}

	// Alignments have no ID and are addressed by index.
	k = kindOf(t, doc, "oc-enterprise-require")
	p, err = RemoveElement(doc, k, "oc-enterprise-require", true)
	if err != nil {
		t.Fatal(err)
	}
	if res = apply(t, doc, p...); !res.Applied || len(res.Document.Alignments) != 0 {
		t.Fatalf("alignment should be removed (applied=%v)", res.Applied)
	}
}

func TestRemoveUnknownElement(t *testing.T) {
	doc := loadFull(t)
	if _, err := RemoveElement(doc, outcomeKind(t), "oc-ghost", true); !errors.Is(err, ErrUnknownID) {
		t.Fatalf("error = %v", err)
	}
	// An ID that exists, but as a different kind.
	if _, err := RemoveElement(doc, outcomeKind(t), "rsk-competing", false); !errors.Is(err, ErrUnknownID) {
		t.Fatalf("wrong-kind error = %v", err)
	}
}

func TestCascadeIsDeterministic(t *testing.T) {
	doc := loadFull(t)
	k := kindOf(t, doc, "oc-idp-support")
	a, err := RemoveElement(doc, k, "oc-idp-support", true)
	if err != nil {
		t.Fatal(err)
	}
	b, err := RemoveElement(doc, k, "oc-idp-support", true)
	if err != nil {
		t.Fatal(err)
	}
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	if string(ja) != string(jb) {
		t.Fatal("cascade patches differ between runs")
	}
}
