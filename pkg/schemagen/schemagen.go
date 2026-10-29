// Package schemagen generates the JSON Schemas for the Theory of Change
// document kinds from the Go types in pkg/toc.
//
// Go structs are the source of truth. Enumeration types are emitted as
// string enums whose values come from the pkg/toc All… functions, so the
// schema and the code cannot drift.
package schemagen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"github.com/invopop/jsonschema"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
)

// BaseURL is the prefix of every schema $id.
const BaseURL = "https://github.com/ProductBuildersHQ/theoryofchange-spec/schema/v0/"

// Definition describes one generated schema.
type Definition struct {
	// Name is the file stem: <Name>.schema.json.
	Name        string
	Title       string
	Description string
	Value       any
}

// Definitions returns the schemas this module publishes, in a stable order.
func Definitions() []Definition {
	return []Definition{
		{"theory-of-change", "Theory of Change", "A Theory of Change document: context, outcomes, causal links, interventions, assumptions, indicators, actors, and evidence.", &toc.Document{}},
		{"observation-set", "Theory of Change Observation Set", "Indicator observations, assumption assessments, and notes for one Theory of Change and period.", &toc.ObservationSet{}},
		{"theory-of-change-status", "Theory of Change Status", "A computed status snapshot of indicators, outcomes, pathways, links, and assumptions at a date.", &toc.Status{}},
		{"status-diff", "Theory of Change Status Diff", "Changes between two status snapshots.", &toc.StatusDiff{}},
		{"actor-registry", "Theory of Change Actor Registry", "A shared list of actors referenced by several theories.", &toc.ActorRegistry{}},
	}
}

// ID returns the $id URL for a schema name.
func ID(name string) string {
	return BaseURL + name + ".schema.json"
}

// FileName returns the file name for a schema name.
func FileName(name string) string {
	return name + ".schema.json"
}

// File is one generated schema file.
type File struct {
	Name string
	Data []byte
}

// Generate returns every schema as indented JSON with a trailing newline,
// in Definitions order. Output is deterministic.
func Generate() ([]File, error) {
	defs := Definitions()
	files := make([]File, 0, len(defs))
	for _, d := range defs {
		data, err := GenerateOne(d)
		if err != nil {
			return nil, err
		}
		files = append(files, File{Name: FileName(d.Name), Data: data})
	}
	return files, nil
}

// GenerateOne returns one schema as indented JSON with a trailing newline.
func GenerateOne(d Definition) ([]byte, error) {
	r := &jsonschema.Reflector{
		RequiredFromJSONSchemaTags: true,
		Mapper:                     enumMapper,
	}
	s := r.Reflect(d.Value)
	if s == nil {
		return nil, fmt.Errorf("schemagen: reflect %s: nil schema", d.Name)
	}
	s.ID = jsonschema.ID(ID(d.Name))
	s.Title = d.Title
	s.Description = d.Description

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return nil, fmt.Errorf("schemagen: encode %s: %w", d.Name, err)
	}
	return buf.Bytes(), nil
}

// EnumValues maps each pkg/toc enumeration type to its values.
func EnumValues() map[reflect.Type][]string {
	m := map[reflect.Type][]string{}
	add := func(t reflect.Type, values []string) { m[t] = values }
	add(typeOf[toc.Level](), strs(toc.AllLevels()))
	add(typeOf[toc.Sphere](), strs(toc.AllSpheres()))
	add(typeOf[toc.LogicType](), strs(toc.AllLogicTypes()))
	add(typeOf[toc.ChangeType](), strs(toc.AllChangeTypes()))
	add(typeOf[toc.LinkType](), strs(toc.AllLinkTypes()))
	add(typeOf[toc.Polarity](), strs(toc.AllPolarities()))
	add(typeOf[toc.LoopType](), strs(toc.AllLoopTypes()))
	add(typeOf[toc.InterventionType](), strs(toc.AllInterventionTypes()))
	add(typeOf[toc.InterventionStatus](), strs(toc.AllInterventionStatuses()))
	add(typeOf[toc.AssumptionType](), strs(toc.AllAssumptionTypes()))
	add(typeOf[toc.AssumptionStatus](), strs(toc.AllAssumptionStatuses()))
	add(typeOf[toc.Criticality](), strs(toc.AllCriticalities()))
	add(typeOf[toc.Rating](), strs(toc.AllRatings()))
	add(typeOf[toc.MeasurementType](), strs(toc.AllMeasurementTypes()))
	add(typeOf[toc.Direction](), strs(toc.AllDirections()))
	add(typeOf[toc.ActorType](), strs(toc.AllActorTypes()))
	add(typeOf[toc.ActorRole](), strs(toc.AllActorRoles()))
	add(typeOf[toc.Stance](), strs(toc.AllStances()))
	add(typeOf[toc.PerspectiveRole](), strs(toc.AllPerspectiveRoles()))
	add(typeOf[toc.EvidenceType](), strs(toc.AllEvidenceTypes()))
	add(typeOf[toc.Strength](), strs(toc.AllStrengths()))
	add(typeOf[toc.EvidenceDirection](), strs(toc.AllEvidenceDirections()))
	add(typeOf[toc.ProcessTracingTest](), strs(toc.AllProcessTracingTests()))
	add(typeOf[toc.Sensitivity](), strs(toc.AllSensitivities()))
	add(typeOf[toc.FactorKind](), strs(toc.AllFactorKinds()))
	add(typeOf[toc.RiskCategory](), strs(toc.AllRiskCategories()))
	add(typeOf[toc.Framework](), strs(toc.AllFrameworks()))
	add(typeOf[toc.Relation](), strs(toc.AllRelations()))
	add(typeOf[toc.RevisionTrigger](), strs(toc.AllRevisionTriggers()))
	add(typeOf[toc.QualityRating](), strs(toc.AllQualityRatings()))
	add(typeOf[toc.DecisionStatus](), strs(toc.AllDecisionStatuses()))
	add(typeOf[toc.DataQuality](), strs(toc.AllDataQualities()))
	add(typeOf[toc.State](), strs(toc.AllStates()))
	add(typeOf[toc.LinkState](), strs(toc.AllLinkStates()))
	add(typeOf[toc.RollupMethod](), strs(toc.AllRollupMethods()))
	add(typeOf[toc.AssumptionSource](), strs(toc.AllAssumptionSources()))
	add(typeOf[toc.ElementKind](), strs(toc.AllElementKinds()))
	return m
}

// enumMapper emits enumeration types as string enums. The document Kind
// type is left to per-field enum tags, since each document kind allows one
// value.
func enumMapper(t reflect.Type) *jsonschema.Schema {
	values, ok := EnumValues()[t]
	if !ok {
		return nil
	}
	enum := make([]any, len(values))
	for i, v := range values {
		enum[i] = v
	}
	return &jsonschema.Schema{Type: "string", Enum: enum}
}

func typeOf[T any]() reflect.Type {
	return reflect.TypeFor[T]()
}

func strs[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = string(v)
	}
	return slices.Clip(out)
}
