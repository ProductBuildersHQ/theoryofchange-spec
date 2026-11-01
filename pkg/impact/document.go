// Package toc defines the impact model document model.
//
// A Document holds the impact model: its context, outcomes connected by causal
// links (optionally combined through precondition groups), interventions,
// assumptions, indicators, actors, evidence, external factors, risks,
// alignments to other frameworks, decisions, and a revision log.
// Measurements live in separate ObservationSet documents.
//
// The structs here are the source of truth for the generated JSON Schemas.
// JSON property names are camelCase; enumeration values are snake_case.
package impact

import core "github.com/grokify/prism-core"

// SpecVersion is the version of this specification that documents declare.
const SpecVersion = "0.1"

// Kind identifies a document type.
type Kind string

// Document kinds.
const (
	KindImpactModel    Kind = "ImpactModel"
	KindObservationSet Kind = "ObservationSet"
	KindStatus         Kind = "ImpactModelStatus"
	KindActorRegistry  Kind = "ActorRegistry"
	KindReviewResponse Kind = "ReviewResponse"
)

// Model is an impact model.
type Model struct {
	Schema      string `json:"$schema,omitempty"`
	Kind        Kind   `json:"kind" jsonschema:"required,enum=ImpactModel"`
	SpecVersion string `json:"specVersion" jsonschema:"required"`
	// ID is the document slug used in references (model:<id>).
	ID       string        `json:"id" jsonschema:"required"`
	Metadata core.Metadata `json:"metadata"`
	// Perspective is whose impact model this is.
	Perspective *Perspective `json:"perspective,omitempty"`
	Context     Context      `json:"context" jsonschema:"required"`
	Ceiling     *Ceiling     `json:"ceilingOfAccountability,omitempty"`

	Outcomes        []Outcome           `json:"outcomes" jsonschema:"required,minItems=1"`
	Groups          []PreconditionGroup `json:"preconditionGroups,omitempty"`
	Links           []CausalLink        `json:"links,omitempty"`
	Interventions   []Intervention      `json:"interventions,omitempty"`
	Assumptions     []Assumption        `json:"assumptions,omitempty"`
	Indicators      []Indicator         `json:"indicators,omitempty"`
	Actors          []Actor             `json:"actors,omitempty"`
	ActorRegistry   string              `json:"actorRegistry,omitempty"`
	Evidence        []Evidence          `json:"evidence,omitempty"`
	ExternalFactors []ExternalFactor    `json:"externalFactors,omitempty"`
	Risks           []Risk              `json:"risks,omitempty"`
	Alignments      []Alignment         `json:"alignments,omitempty"`
	Decisions       []Decision          `json:"decisions,omitempty"`
	Revisions       []Revision          `json:"revisions,omitempty"`

	// Parent references the impact model this one elaborates (model:<doc>).
	Parent   string    `json:"parent,omitempty" jsonschema:"pattern=^model:[a-z][a-z0-9]*(-[a-z0-9]+)*$"`
	Pathways []Pathway `json:"pathways,omitempty"`

	// Metrics is the shared metric library; Glossary the shared vocabulary.
	Metrics  []Metric       `json:"metrics,omitempty"`
	Glossary []GlossaryTerm `json:"glossary,omitempty"`

	// Causes are the problem tree; TargetGroups the priority populations.
	Causes       []Cause       `json:"causes,omitempty"`
	TargetGroups []TargetGroup `json:"targetGroups,omitempty"`

	// Audiences say who needs to see results; Themes are the story codebook.
	Audiences []Audience `json:"audiences,omitempty"`
	Themes    []Theme    `json:"themes,omitempty"`

	// SelectionCriteria score candidate indicators; Budget is what the
	// organization can sustain measuring.
	SelectionCriteria []SelectionCriterion `json:"selectionCriteria,omitempty"`
	Budget            *MeasurementBudget   `json:"measurementBudget,omitempty"`

	// Consultations record validation of the impact model with stakeholders.
	Consultations []Consultation `json:"consultations,omitempty"`
}

// NewModel returns a document with Kind and SpecVersion set.
func NewModel(id, problemStatement string) *Model {
	return &Model{
		Kind:        KindImpactModel,
		SpecVersion: SpecVersion,
		ID:          id,
		Context:     Context{ProblemStatement: problemStatement},
	}
}

// Perspective identifies the actor whose impact model this is and, for competing
// futures, the future it belongs to.
type Perspective struct {
	ActorID string          `json:"actorId" jsonschema:"required"`
	Role    PerspectiveRole `json:"role,omitempty"`
	// FutureRef groups this impact model under a competing future (e.g.
	// "ufm:<future-slug>").
	FutureRef string `json:"futureRef,omitempty"`
}

// Context is the problem analysis the impact model responds to.
type Context struct {
	ProblemStatement string       `json:"problemStatement" jsonschema:"required"`
	RootCauses       []string     `json:"rootCauses,omitempty"`
	TargetPopulation string       `json:"targetPopulation,omitempty"`
	Market           *ExternalRef `json:"market,omitempty"`
	Geography        string       `json:"geography,omitempty"`
	TimeHorizon      *TimeHorizon `json:"timeHorizon,omitempty"`
	// Mission is why the organization exists and what it does. It frames the
	// Impact model and is not an outcome: nothing causes it and it is not measured.
	Mission string `json:"mission,omitempty"`
	// Vision is the future the organization works toward. The impact
	// outcomes are how the impact model says it is reached.
	Vision    string `json:"vision,omitempty"`
	Narrative string `json:"narrative,omitempty"`
}

// TimeHorizon is the period the impact model covers.
type TimeHorizon struct {
	Start string `json:"start" jsonschema:"required,format=date"`
	End   string `json:"end" jsonschema:"required,format=date"`
}

// Ceiling is the ceiling of accountability: outcomes above it are tracked
// but not owned. Set either Level (everything above that level) or
// OutcomeIDs (the highest outcomes still owned).
type Ceiling struct {
	Level      Level    `json:"level,omitempty"`
	OutcomeIDs []string `json:"outcomeIds,omitempty"`
	Rationale  string   `json:"rationale,omitempty"`
}

// Decision is a question leadership needs to decide.
type Decision struct {
	ID         string         `json:"id" jsonschema:"required"`
	Question   string         `json:"question" jsonschema:"required"`
	Options    []string       `json:"options,omitempty"`
	DueDate    string         `json:"dueDate,omitempty" jsonschema:"format=date"`
	OwnerID    string         `json:"ownerActorId,omitempty"`
	Status     DecisionStatus `json:"status,omitempty"`
	RelatedIDs []string       `json:"relatedIds,omitempty"`
}

// Revision records a change to the impact model (adaptive management).
type Revision struct {
	Version            string          `json:"version" jsonschema:"required"`
	Date               string          `json:"date" jsonschema:"required,format=date"`
	Summary            string          `json:"summary" jsonschema:"required"`
	Trigger            RevisionTrigger `json:"trigger,omitempty"`
	ChangedIDs         []string        `json:"changedIds,omitempty"`
	TriggerEvidenceIDs []string        `json:"triggerEvidenceIds,omitempty"`
	QualityReview      *QualityReview  `json:"qualityReview,omitempty"`
	Approvers          []core.Approver `json:"approvers,omitempty"`
}

// QualityReview records the plausible / feasible / testable judgment.
type QualityReview struct {
	Plausible QualityRating `json:"plausible,omitempty"`
	Feasible  QualityRating `json:"feasible,omitempty"`
	Testable  QualityRating `json:"testable,omitempty"`
	Notes     string        `json:"notes,omitempty"`
}
