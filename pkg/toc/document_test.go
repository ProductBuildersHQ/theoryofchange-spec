package toc

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func loadFixture(t *testing.T, name string) (*Document, []byte) {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var doc Document
	if err := dec.Decode(&doc); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	return &doc, data
}

func normalize(t *testing.T, data []byte) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return v
}

func TestFixturesRoundTrip(t *testing.T) {
	for _, name := range []string{"minimal.json", "full.json"} {
		t.Run(name, func(t *testing.T) {
			doc, data := loadFixture(t, name)
			out, err := json.Marshal(doc)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if !reflect.DeepEqual(normalize(t, data), normalize(t, out)) {
				t.Fatalf("round trip changed the document:\n in: %s\nout: %s", data, out)
			}
		})
	}
}

func TestFullFixtureExercisesEveryField(t *testing.T) {
	doc, _ := loadFixture(t, "full.json")
	assertAllFieldsSet(t, reflect.ValueOf(*doc), "Document", map[string]bool{
		// Mutually exclusive or registry-only fields, covered elsewhere.
		"Document.ActorRegistry":        true,
		"Document.Ceiling.OutcomeIDs":   true,
		"Document.Metadata.ID":          true,
		"Document.Metadata.Title":       true,
		"Document.Metadata.Description": true,
		"Document.Metadata.Version":     true,
		"Document.Metadata.Domain":      true,
		"Document.Metadata.Team":        true,
		"Document.Metadata.Authors":     true,
		"Document.Metadata.CreatedAt":   true,
		"Document.Metadata.UpdatedAt":   true,
	})
}

// assertAllFieldsSet checks that every exported field is non-zero in at
// least one element of each slice, so the fixture keeps covering the model
// as it grows.
func assertAllFieldsSet(t *testing.T, v reflect.Value, path string, skip map[string]bool) {
	t.Helper()
	if skip[path] {
		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			t.Errorf("%s is nil in full.json", path)
			return
		}
		assertAllFieldsSet(t, v.Elem(), path, skip)
	case reflect.Struct:
		typ := v.Type()
		for i := range typ.NumField() {
			f := typ.Field(i)
			if !f.IsExported() {
				continue
			}
			name := path + "." + f.Name
			if f.Anonymous {
				name = path
			}
			assertAllFieldsSet(t, v.Field(i), name, skip)
		}
	case reflect.Slice:
		if v.Len() == 0 {
			t.Errorf("%s is empty in full.json", path)
			return
		}
		if v.Type().Elem().Kind() != reflect.Struct {
			return
		}
		// Each field must be set in at least one element.
		elemType := v.Type().Elem()
		for i := range elemType.NumField() {
			f := elemType.Field(i)
			if !f.IsExported() {
				continue
			}
			name := path + "[]." + f.Name
			if skip[name] {
				continue
			}
			set := false
			for j := range v.Len() {
				if !v.Index(j).Field(i).IsZero() {
					set = true
					break
				}
			}
			if !set && !f.Anonymous {
				t.Errorf("%s is never set in full.json", name)
			}
		}
	default:
		if v.IsZero() && v.Kind() != reflect.Bool {
			t.Errorf("%s is zero in full.json", path)
		}
	}
}

var camelCase = regexp.MustCompile(`^(\$schema|[a-z][a-zA-Z0-9]*)$`)

func TestJSONKeysAreCamelCase(t *testing.T) {
	_, data := loadFixture(t, "full.json")
	var walk func(path string, v any)
	walk = func(path string, v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, child := range x {
				// Attribute maps hold user-defined keys.
				if !strings.HasSuffix(path, ".attributes") && !camelCase.MatchString(k) {
					t.Errorf("key %q at %s is not camelCase", k, path)
				}
				walk(path+"."+k, child)
			}
		case []any:
			for _, child := range x {
				walk(path+"[]", child)
			}
		}
	}
	walk("", normalize(t, data))
}

var snakeCase = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

func TestEnums(t *testing.T) {
	enums := map[string][]string{
		"Level":              strs(AllLevels()),
		"Sphere":             strs(AllSpheres()),
		"LogicType":          strs(AllLogicTypes()),
		"ChangeType":         strs(AllChangeTypes()),
		"LinkType":           strs(AllLinkTypes()),
		"Polarity":           strs(AllPolarities()),
		"LoopType":           strs(AllLoopTypes()),
		"InterventionType":   strs(AllInterventionTypes()),
		"InterventionStatus": strs(AllInterventionStatuses()),
		"AssumptionType":     strs(AllAssumptionTypes()),
		"AssumptionStatus":   strs(AllAssumptionStatuses()),
		"Criticality":        strs(AllCriticalities()),
		"Rating":             strs(AllRatings()),
		"MeasurementType":    strs(AllMeasurementTypes()),
		"Direction":          strs(AllDirections()),
		"ActorType":          strs(AllActorTypes()),
		"ActorRole":          strs(AllActorRoles()),
		"Stance":             strs(AllStances()),
		"PerspectiveRole":    strs(AllPerspectiveRoles()),
		"EvidenceType":       strs(AllEvidenceTypes()),
		"Strength":           strs(AllStrengths()),
		"EvidenceDirection":  strs(AllEvidenceDirections()),
		"ProcessTracingTest": strs(AllProcessTracingTests()),
		"Sensitivity":        strs(AllSensitivities()),
		"FactorKind":         strs(AllFactorKinds()),
		"RiskCategory":       strs(AllRiskCategories()),
		"Framework":          strs(AllFrameworks()),
		"Relation":           strs(AllRelations()),
		"RevisionTrigger":    strs(AllRevisionTriggers()),
		"QualityRating":      strs(AllQualityRatings()),
		"DecisionStatus":     strs(AllDecisionStatuses()),
	}
	for name, values := range enums {
		seen := map[string]bool{}
		for _, v := range values {
			if !snakeCase.MatchString(v) {
				t.Errorf("%s value %q is not snake_case", name, v)
			}
			if seen[v] {
				t.Errorf("%s value %q is duplicated", name, v)
			}
			seen[v] = true
		}
	}

	// Valid accepts every listed value and rejects empty and unknown values.
	for _, v := range AllLevels() {
		if !v.Valid() {
			t.Errorf("Level %q not valid", v)
		}
	}
	if Level("").Valid() || Level("huge").Valid() {
		t.Error("Level accepted an invalid value")
	}
	if Framework("").Valid() || !FrameworkBSC.Valid() {
		t.Error("Framework validity mismatch")
	}
}

func strs[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = string(v)
	}
	return out
}

func TestLevelRank(t *testing.T) {
	if LevelImpact.Rank() <= LevelLongTerm.Rank() || LevelShortTerm.Rank() <= LevelOutput.Rank() {
		t.Fatal("levels not ranked from impact down to output")
	}
	if LevelOutput.Rank() != 1 || Level("x").Rank() != 0 {
		t.Fatalf("unexpected ranks: output=%d unknown=%d", LevelOutput.Rank(), Level("x").Rank())
	}
}

func TestSensitivityRenderable(t *testing.T) {
	if !SensitivityPublic.Renderable() || !SensitivityInternal.Renderable() {
		t.Fatal("public and internal must be renderable")
	}
	if SensitivityRestricted.Renderable() || Sensitivity("").Renderable() {
		t.Fatal("restricted and unset must not be renderable")
	}
}

func TestNewDocument(t *testing.T) {
	d := NewDocument("doc", "problem")
	if d.Kind != KindTheoryOfChange || d.SpecVersion != SpecVersion || d.Context.ProblemStatement != "problem" {
		t.Fatalf("unexpected document %+v", d)
	}
}

func TestAccessors(t *testing.T) {
	doc, _ := loadFixture(t, "full.json")

	elems := doc.Elements()
	if len(elems) == 0 || elems[0].ID != "oc-safer-agents" || elems[0].Path != "/outcomes/0" {
		t.Fatalf("unexpected first element %+v", elems[0])
	}
	if e, ok := doc.Element("rsk-competing"); !ok || e.Kind != ElementRisk || e.Path != "/risks/0" {
		t.Fatalf("Element(rsk-competing) = %+v, %v", e, ok)
	}
	if _, ok := doc.Element("missing"); ok {
		t.Fatal("Element found a missing ID")
	}

	if o, ok := doc.Outcome("oc-idp-support"); !ok || o.Level != LevelIntermediate {
		t.Fatalf("Outcome lookup failed: %+v %v", o, ok)
	}
	if _, ok := doc.Group("grp-supply"); !ok {
		t.Fatal("Group lookup failed")
	}
	if l, ok := doc.Link("ln-spec-idp"); !ok || l.Lag != "P6M" {
		t.Fatal("Link lookup failed")
	}
	if _, ok := doc.Intervention("int-interop"); !ok {
		t.Fatal("Intervention lookup failed")
	}
	if a, ok := doc.Assumption("asm-idp-incentive"); !ok || !a.NeedsTesting {
		t.Fatal("Assumption lookup failed")
	}
	if _, ok := doc.Indicator("ind-idp-count"); !ok {
		t.Fatal("Indicator lookup failed")
	}
	if _, ok := doc.Actor("act-idp"); !ok {
		t.Fatal("Actor lookup failed")
	}
	if _, ok := doc.EvidenceRecord("ev-survey"); !ok {
		t.Fatal("Evidence lookup failed")
	}
	if _, ok := doc.Outcome("nope"); ok {
		t.Fatal("Outcome lookup found a missing ID")
	}

	if got := doc.IndicatorsMeasuring("asm-idp-incentive"); len(got) != 1 || got[0].ID != "ind-idp-roadmaps" {
		t.Fatalf("IndicatorsMeasuring = %+v", got)
	}
	if got := doc.LinksTo("oc-enterprise-require"); len(got) != 2 {
		t.Fatalf("LinksTo = %d links", len(got))
	}
	if got := doc.LinksFrom("oc-enterprise-require"); len(got) != 2 {
		t.Fatalf("LinksFrom = %d links", len(got))
	}
	if got := doc.ImpactOutcomes(); len(got) != 1 || got[0].ID != "oc-safer-agents" {
		t.Fatalf("ImpactOutcomes = %+v", got)
	}

	enterprise, _ := doc.Outcome("oc-enterprise-require")
	if l := enterprise.EffectiveLogic(); l.Type != LogicAtLeast || l.K != 1 {
		t.Fatalf("EffectiveLogic = %+v", l)
	}
	idp, _ := doc.Outcome("oc-idp-support")
	if l := idp.EffectiveLogic(); l.Type != LogicAll {
		t.Fatalf("default EffectiveLogic = %+v", l)
	}

	var nilDoc *Document
	if nilDoc.Elements() != nil {
		t.Fatal("nil document should have no elements")
	}
}
