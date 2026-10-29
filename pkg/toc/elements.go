package toc

import core "github.com/grokify/prism-core"

// Outcome is a node in the pathway of change: a condition that must exist
// for the theory to work and does not yet exist. Incoming causal links are
// its preconditions.
type Outcome struct {
	ID          string `json:"id" jsonschema:"required"`
	Title       string `json:"title" jsonschema:"required"`
	Description string `json:"description,omitempty"`
	Level       Level  `json:"level" jsonschema:"required"`
	Sphere      Sphere `json:"sphere,omitempty"`
	// Logic combines incoming links that are not members of a
	// PreconditionGroup. The default (nil) is all-of.
	Logic *Logic `json:"logic,omitempty"`
	// Category is an open taxonomy, e.g. ORS impact / influence / leverage /
	// capacity, COM-B stages, or market-systems adopt / adapt / expand /
	// respond.
	Category   string     `json:"category,omitempty"`
	ChangeType ChangeType `json:"changeType,omitempty"`
	// Population describes who or what changes.
	Population string `json:"population,omitempty"`
	// ActorIDs are actors who change or co-own the outcome.
	ActorIDs []string `json:"actorIds,omitempty"`
	// ContributesTo lists outcomes in a parent theory this outcome
	// contributes to, as element references (toc:<doc>#<outcome>).
	ContributesTo []string `json:"contributesTo,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// Logic is how preconditions combine. K is used only with LogicAtLeast.
type Logic struct {
	Type LogicType `json:"type" jsonschema:"required"`
	K    int       `json:"k,omitempty"`
}

// PreconditionGroup is an explicit AND / OR / k-of-n combination of
// preconditions of one target outcome. Members are outcome or group IDs, so
// groups nest: (A AND B) OR C.
type PreconditionGroup struct {
	ID      string   `json:"id" jsonschema:"required"`
	Target  string   `json:"target" jsonschema:"required"`
	Members []string `json:"members" jsonschema:"required,minItems=2"`
	Logic   Logic    `json:"logic" jsonschema:"required"`
	// Sufficiency records the judgment that the members are jointly
	// sufficient for the target.
	Sufficiency string `json:"sufficiency,omitempty"`
}

// CausalLink is a directed claim that From (an outcome or intervention)
// contributes to To (an outcome).
type CausalLink struct {
	ID        string   `json:"id" jsonschema:"required"`
	From      string   `json:"from" jsonschema:"required"`
	To        string   `json:"to" jsonschema:"required"`
	Type      LinkType `json:"type,omitempty"`
	Rationale string   `json:"rationale,omitempty"`
	// AssumptionIDs underpin this link.
	AssumptionIDs []string `json:"assumptionIds,omitempty"`
	EvidenceIDs   []string `json:"evidenceIds,omitempty"`
	// InterventionIDs are placed on this connector: what it takes for From
	// to produce To.
	InterventionIDs []string `json:"interventionIds,omitempty"`
	Polarity        Polarity `json:"polarity,omitempty"`
	// Feedback marks an intentional loop; feedback links are excluded from
	// cycle checks and pathway roll-ups.
	Feedback bool     `json:"feedback,omitempty"`
	LoopType LoopType `json:"loopType,omitempty"`
	// Lag is an ISO 8601 duration, e.g. "P6M".
	Lag string `json:"lag,omitempty"`
}

// Intervention is something actors do to bring about outcomes.
type Intervention struct {
	ID          string             `json:"id" jsonschema:"required"`
	Title       string             `json:"title" jsonschema:"required"`
	Description string             `json:"description,omitempty"`
	Type        InterventionType   `json:"type,omitempty"`
	Targets     []string           `json:"targets" jsonschema:"required,minItems=1"`
	Outputs     []string           `json:"outputs,omitempty"`
	OwnerID     string             `json:"ownerActorId,omitempty"`
	Status      InterventionStatus `json:"status,omitempty"`
	// Capabilities references PRISM capabilities this intervention enables,
	// improves, or depends on.
	Capabilities []core.CapabilityRef `json:"capabilities,omitempty"`
	// InitiativeRefs points at delivery initiatives (e.g. VisionStudio).
	InitiativeRefs []ExternalRef `json:"initiativeRefs,omitempty"`
	Resources      string        `json:"resources,omitempty"`
}

// IfThenBecause is the structured form of an assumption: if X, then Y,
// because Z.
type IfThenBecause struct {
	If      string `json:"if" jsonschema:"required"`
	Then    string `json:"then" jsonschema:"required"`
	Because string `json:"because,omitempty"`
}

// Assumption is a condition believed to hold for a link, outcome, or
// intervention to work, which the initiative does not itself bring about.
type Assumption struct {
	ID            string         `json:"id" jsonschema:"required"`
	Statement     string         `json:"statement" jsonschema:"required"`
	IfThenBecause *IfThenBecause `json:"ifThenBecause,omitempty"`
	Type          AssumptionType `json:"type,omitempty"`
	// AppliesTo lists link, outcome, or intervention IDs.
	AppliesTo    []string    `json:"appliesTo,omitempty"`
	Criticality  Criticality `json:"criticality,omitempty"`
	Plausibility Rating      `json:"plausibility,omitempty"`
	// ExplainsFully records whether the assumption fully explains the
	// expected change (UNDG question 1).
	ExplainsFully *bool `json:"explainsFully,omitempty"`
	// NeedsTesting records whether the assumption must be tested (UNDG
	// question 3).
	NeedsTesting bool `json:"needsTesting,omitempty"`
	// Test describes how the assumption will be tested.
	Test         string           `json:"test,omitempty"`
	IndicatorIDs []string         `json:"indicatorIds,omitempty"`
	EvidenceIDs  []string         `json:"evidenceIds,omitempty"`
	Status       AssumptionStatus `json:"status,omitempty"`
	OwnerID      string           `json:"ownerActorId,omitempty"`
}

// Indicator is measurable evidence that an outcome, output, or assumption is
// changing.
type Indicator struct {
	ID string `json:"id" jsonschema:"required"`
	// Measures is the outcome, intervention (output indicator), or
	// assumption (monitoring indicator) this indicator measures.
	Measures        string          `json:"measures" jsonschema:"required"`
	Name            string          `json:"name" jsonschema:"required"`
	Definition      string          `json:"definition,omitempty"`
	MeasurementType MeasurementType `json:"measurementType,omitempty"`
	Unit            string          `json:"unit,omitempty"`
	Direction       Direction       `json:"direction,omitempty"`
	// Tolerance is the acceptable band for maintain-direction indicators.
	Tolerance float64 `json:"tolerance,omitempty"`
	// Levels are ordered qualitative levels from lowest to highest.
	Levels         []string    `json:"levels,omitempty"`
	Baseline       *Baseline   `json:"baseline,omitempty"`
	Targets        []Target    `json:"targets,omitempty"`
	Population     *Population `json:"population,omitempty"`
	Disaggregation []string    `json:"disaggregation,omitempty"`
	DataSource     string      `json:"dataSource,omitempty"`
	Method         string      `json:"method,omitempty"`
	// Frequency is an ISO 8601 duration, e.g. "P1M".
	Frequency string `json:"frequency,omitempty"`
	OwnerID   string `json:"ownerActorId,omitempty"`
	// ProxyFor lists outcome IDs this indicator also stands in for.
	ProxyFor     []string      `json:"proxyFor,omitempty"`
	ExternalRefs []ExternalRef `json:"externalRefs,omitempty"`
}

// Baseline is an indicator's starting point.
type Baseline struct {
	Value  *float64 `json:"value,omitempty"`
	Level  string   `json:"level,omitempty"`
	Date   string   `json:"date,omitempty" jsonschema:"format=date"`
	Source string   `json:"source,omitempty"`
}

// Target is a time-phased target or milestone for an indicator. Value is
// the target (how many), Threshold the level of change judged sufficient,
// and Date the timeline.
type Target struct {
	Value     *float64 `json:"value,omitempty"`
	Level     string   `json:"level,omitempty"`
	Threshold string   `json:"threshold,omitempty"`
	Date      string   `json:"date" jsonschema:"required,format=date"`
	Label     string   `json:"label,omitempty"`
	Milestone bool     `json:"milestone,omitempty"`
}

// Population is who or what an indicator counts.
type Population struct {
	Description string   `json:"description" jsonschema:"required"`
	ActorIDs    []string `json:"actorIds,omitempty"`
	Size        *float64 `json:"size,omitempty"`
}

// Actor is a stakeholder, partner, beneficiary, champion, or competitor.
type Actor struct {
	ID           string        `json:"id" jsonschema:"required"`
	Name         string        `json:"name" jsonschema:"required"`
	Type         ActorType     `json:"type,omitempty"`
	Roles        []ActorRole   `json:"roles,omitempty"`
	Stance       Stance        `json:"stance,omitempty"`
	Influence    Rating        `json:"influence,omitempty"`
	Interest     Rating        `json:"interest,omitempty"`
	Description  string        `json:"description,omitempty"`
	Contact      *core.Person  `json:"contact,omitempty"`
	ExternalRefs []ExternalRef `json:"externalRefs,omitempty"`
}

// Evidence is a source supporting or contradicting a claim: research, data,
// testimony, or a web page used when drafting from public information.
type Evidence struct {
	ID                 string             `json:"id" jsonschema:"required"`
	Title              string             `json:"title" jsonschema:"required"`
	Type               EvidenceType       `json:"type,omitempty"`
	Strength           Strength           `json:"strength,omitempty"`
	Direction          EvidenceDirection  `json:"direction,omitempty"`
	ProcessTracingTest ProcessTracingTest `json:"processTracingTest,omitempty"`
	Source             string             `json:"source,omitempty"`
	URL                string             `json:"url,omitempty" jsonschema:"format=uri"`
	Summary            string             `json:"summary,omitempty"`
	Excerpt            string             `json:"excerpt,omitempty"`
	// Sensitivity controls verbatim rendering; the zero value is treated as
	// restricted.
	Sensitivity Sensitivity `json:"sensitivity,omitempty"`
	AccessedAt  string      `json:"accessedAt,omitempty" jsonschema:"format=date"`
}

// ExternalFactor is a condition outside the theory's control that helps or
// hinders change.
type ExternalFactor struct {
	ID          string     `json:"id" jsonschema:"required"`
	Description string     `json:"description" jsonschema:"required"`
	Kind        FactorKind `json:"kind,omitempty"`
	Affects     []string   `json:"affects,omitempty"`
	Likelihood  Rating     `json:"likelihood,omitempty"`
	// Trend is a prism-core trend direction (up, down, flat, unknown).
	Trend string `json:"trend,omitempty" jsonschema:"enum=up,enum=down,enum=flat,enum=unknown"`
}

// Risk is a prism-core risk attached to elements of the theory.
type Risk struct {
	core.Risk
	AppliesTo            []string     `json:"appliesTo,omitempty"`
	Category             RiskCategory `json:"category,omitempty"`
	RelatedAssumptionIDs []string     `json:"relatedAssumptionIds,omitempty"`
}

// Pathway is a named grouping of outcomes and interventions, such as a
// program area or a strategy. Detail for a pathway may live in a child
// theory that points back with the document's parent field.
type Pathway struct {
	ID              string      `json:"id" jsonschema:"required"`
	Title           string      `json:"title" jsonschema:"required"`
	Kind            PathwayKind `json:"kind,omitempty"`
	Description     string      `json:"description,omitempty"`
	OwnerID         string      `json:"ownerActorId,omitempty"`
	OutcomeIDs      []string    `json:"outcomeIds,omitempty"`
	InterventionIDs []string    `json:"interventionIds,omitempty"`
	// ChildToc references a theory that elaborates this pathway
	// (toc:<doc>).
	ChildToc string `json:"childToc,omitempty" jsonschema:"pattern=^toc:[a-z][a-z0-9]*(-[a-z0-9]+)*$"`
}

// ExternalRef points at an element of another framework or system.
type ExternalRef struct {
	Framework Framework `json:"framework" jsonschema:"required"`
	// Ref is opaque to this module (e.g. a market-spec "{type}:{slug}" or a
	// VisionStudio initiative ID).
	Ref        string            `json:"ref" jsonschema:"required"`
	URL        string            `json:"url,omitempty" jsonschema:"format=uri"`
	Relation   Relation          `json:"relation,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// Alignment links an element of this theory to an external element, such as
// a Balanced Scorecard objective or an OKR key result.
type Alignment struct {
	ElementID string      `json:"elementId" jsonschema:"required"`
	Target    ExternalRef `json:"target" jsonschema:"required"`
}
