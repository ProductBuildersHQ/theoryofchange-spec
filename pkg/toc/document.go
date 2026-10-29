// Package toc defines the Theory of Change document model.
//
// A Document holds the theory: its context, outcomes connected by causal
// links (optionally combined through precondition groups), interventions,
// assumptions, indicators, actors, evidence, external factors, risks,
// alignments to other frameworks, decisions, and a revision log.
// Measurements live in separate ObservationSet documents.
//
// The structs here are the source of truth for the generated JSON Schemas.
// JSON property names are camelCase; enumeration values are snake_case.
package toc

import core "github.com/grokify/prism-core"

// SpecVersion is the version of this specification that documents declare.
const SpecVersion = "0.1"

// Kind identifies a document type.
type Kind string

// Document kinds.
const (
	KindTheoryOfChange Kind = "TheoryOfChange"
	KindObservationSet Kind = "ObservationSet"
	KindStatus         Kind = "TheoryOfChangeStatus"
	KindActorRegistry  Kind = "ActorRegistry"
	KindReviewResponse Kind = "ReviewResponse"
)

// Document is a Theory of Change.
type Document struct {
	Schema      string `json:"$schema,omitempty"`
	Kind        Kind   `json:"kind" jsonschema:"required,enum=TheoryOfChange"`
	SpecVersion string `json:"specVersion" jsonschema:"required"`
	// ID is the document slug used in references (toc:<id>).
	ID       string        `json:"id" jsonschema:"required"`
	Metadata core.Metadata `json:"metadata"`
	// Perspective is whose theory this is.
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

	// Parent references the theory this one elaborates (toc:<doc>).
	Parent   string    `json:"parent,omitempty" jsonschema:"pattern=^toc:[a-z][a-z0-9]*(-[a-z0-9]+)*$"`
	Pathways []Pathway `json:"pathways,omitempty"`

	// Metrics is the shared metric library; Glossary the shared vocabulary.
	Metrics  []Metric       `json:"metrics,omitempty"`
	Glossary []GlossaryTerm `json:"glossary,omitempty"`

	// Consultations record validation of the theory with stakeholders.
	Consultations []Consultation `json:"consultations,omitempty"`
}

// NewDocument returns a document with Kind and SpecVersion set.
func NewDocument(id, problemStatement string) *Document {
	return &Document{
		Kind:        KindTheoryOfChange,
		SpecVersion: SpecVersion,
		ID:          id,
		Context:     Context{ProblemStatement: problemStatement},
	}
}

// Perspective identifies the actor whose theory this is and, for competing
// futures, the future it belongs to.
type Perspective struct {
	ActorID string          `json:"actorId" jsonschema:"required"`
	Role    PerspectiveRole `json:"role,omitempty"`
	// FutureRef groups this theory under a competing future (e.g.
	// "ufm:<future-slug>").
	FutureRef string `json:"futureRef,omitempty"`
}

// Context is the problem analysis the theory responds to.
type Context struct {
	ProblemStatement string       `json:"problemStatement" jsonschema:"required"`
	RootCauses       []string     `json:"rootCauses,omitempty"`
	TargetPopulation string       `json:"targetPopulation,omitempty"`
	Market           *ExternalRef `json:"market,omitempty"`
	Geography        string       `json:"geography,omitempty"`
	TimeHorizon      *TimeHorizon `json:"timeHorizon,omitempty"`
	Narrative        string       `json:"narrative,omitempty"`
}

// TimeHorizon is the period the theory covers.
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

// Revision records a change to the theory (adaptive management).
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
