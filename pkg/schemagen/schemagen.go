// Package schemagen generates the JSON Schemas for the impact model
// document kinds from the Go types in pkg/impact.
//
// Go structs are the source of truth. Enumeration types are emitted as
// string enums whose values come from the pkg/impact All… functions, so the
// schema and the code cannot drift.
package schemagen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"github.com/invopop/jsonschema"

	"github.com/ProductBuildersHQ/impact-model/pkg/impact"
)

// BaseURL is the prefix of every schema $id.
const BaseURL = "https://github.com/ProductBuildersHQ/impact-model/schema/v0/"

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
		{"impact-model", "Impact Model", "An impact model document: context, outcomes, causal links, interventions, assumptions, indicators, actors, and evidence.", &impact.Model{}},
		{"observation-set", "Impact Model Observation Set", "Indicator observations, assumption assessments, and notes for one impact model and period.", &impact.ObservationSet{}},
		{"impact-model-status", "Impact Model Status", "A computed status snapshot of indicators, outcomes, pathways, links, and assumptions at a date.", &impact.Status{}},
		{"status-diff", "Impact Model Status Diff", "Changes between two status snapshots.", &impact.StatusDiff{}},
		{"actor-registry", "Impact Model Actor Registry", "A shared list of actors referenced by several impact models.", &impact.ActorRegistry{}},
		{"csv-mapping", "Impact Model CSV Mapping", "How the columns of a spreadsheet map onto indicator observations.", &impact.CSVMapping{}},
		{"review-response", "Impact Model Review Response", "What reviewers said about an impact model, by element, independent of the channel it was collected through.", &impact.ReviewResponse{}},
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

// EnumValues maps each pkg/impact enumeration type to its values.
func EnumValues() map[reflect.Type][]string {
	m := map[reflect.Type][]string{}
	add := func(t reflect.Type, values []string) { m[t] = values }
	add(typeOf[impact.Level](), strs(impact.AllLevels()))
	add(typeOf[impact.Sphere](), strs(impact.AllSpheres()))
	add(typeOf[impact.LogicType](), strs(impact.AllLogicTypes()))
	add(typeOf[impact.ChangeType](), strs(impact.AllChangeTypes()))
	add(typeOf[impact.LinkType](), strs(impact.AllLinkTypes()))
	add(typeOf[impact.Polarity](), strs(impact.AllPolarities()))
	add(typeOf[impact.LoopType](), strs(impact.AllLoopTypes()))
	add(typeOf[impact.InterventionType](), strs(impact.AllInterventionTypes()))
	add(typeOf[impact.InterventionStatus](), strs(impact.AllInterventionStatuses()))
	add(typeOf[impact.AssumptionType](), strs(impact.AllAssumptionTypes()))
	add(typeOf[impact.AssumptionStatus](), strs(impact.AllAssumptionStatuses()))
	add(typeOf[impact.Criticality](), strs(impact.AllCriticalities()))
	add(typeOf[impact.Rating](), strs(impact.AllRatings()))
	add(typeOf[impact.MeasurementType](), strs(impact.AllMeasurementTypes()))
	add(typeOf[impact.Direction](), strs(impact.AllDirections()))
	add(typeOf[impact.ActorType](), strs(impact.AllActorTypes()))
	add(typeOf[impact.ActorRole](), strs(impact.AllActorRoles()))
	add(typeOf[impact.Stance](), strs(impact.AllStances()))
	add(typeOf[impact.PerspectiveRole](), strs(impact.AllPerspectiveRoles()))
	add(typeOf[impact.EvidenceType](), strs(impact.AllEvidenceTypes()))
	add(typeOf[impact.Strength](), strs(impact.AllStrengths()))
	add(typeOf[impact.EvidenceDirection](), strs(impact.AllEvidenceDirections()))
	add(typeOf[impact.ProcessTracingTest](), strs(impact.AllProcessTracingTests()))
	add(typeOf[impact.Sensitivity](), strs(impact.AllSensitivities()))
	add(typeOf[impact.FactorKind](), strs(impact.AllFactorKinds()))
	add(typeOf[impact.RiskCategory](), strs(impact.AllRiskCategories()))
	add(typeOf[impact.PathwayKind](), strs(impact.AllPathwayKinds()))
	add(typeOf[impact.Projection](), strs(impact.AllProjections()))
	add(typeOf[impact.Aggregation](), strs(impact.AllAggregations()))
	add(typeOf[impact.CompositionOp](), strs(impact.AllCompositionOps()))
	add(typeOf[impact.MetricLayer](), strs(impact.AllMetricLayers()))
	add(typeOf[impact.Claim](), strs(impact.AllClaims()))
	add(typeOf[impact.Basis](), strs(impact.AllBases()))
	add(typeOf[impact.DataAvailability](), strs(impact.AllDataAvailabilities()))
	add(typeOf[impact.ConsultationStep](), strs(impact.AllConsultationSteps()))
	add(typeOf[impact.Responsibility](), strs(impact.AllResponsibilities()))
	add(typeOf[impact.CauseLevel](), strs(impact.AllCauseLevels()))
	add(typeOf[impact.Consent](), strs(impact.AllConsents()))
	add(typeOf[impact.CollectionMode](), strs(impact.AllCollectionModes()))
	add(typeOf[impact.SelectionStatus](), strs(impact.AllSelectionStatuses()))
	add(typeOf[impact.SelectionSubject](), strs(impact.AllSelectionSubjects()))
	add(typeOf[impact.ReviewMethod](), strs(impact.AllReviewMethods()))
	add(typeOf[impact.ReviewDecision](), strs(impact.AllReviewDecisions()))
	add(typeOf[impact.ReviewOrigin](), strs(impact.AllReviewOrigins()))
	add(typeOf[impact.ConsultationMethod](), strs(impact.AllConsultationMethods()))
	add(typeOf[impact.Framework](), strs(impact.AllFrameworks()))
	add(typeOf[impact.Relation](), strs(impact.AllRelations()))
	add(typeOf[impact.RevisionTrigger](), strs(impact.AllRevisionTriggers()))
	add(typeOf[impact.QualityRating](), strs(impact.AllQualityRatings()))
	add(typeOf[impact.DecisionStatus](), strs(impact.AllDecisionStatuses()))
	add(typeOf[impact.DataQuality](), strs(impact.AllDataQualities()))
	add(typeOf[impact.State](), strs(impact.AllStates()))
	add(typeOf[impact.LinkState](), strs(impact.AllLinkStates()))
	add(typeOf[impact.RollupMethod](), strs(impact.AllRollupMethods()))
	add(typeOf[impact.AssumptionSource](), strs(impact.AllAssumptionSources()))
	add(typeOf[impact.ElementKind](), strs(impact.AllElementKinds()))
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
