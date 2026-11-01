package impact

import core "github.com/grokify/prism-core"

// Outcome is a node in the pathway of change: a condition that must exist
// for the impact model to work and does not yet exist. Incoming causal links are
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
	// Responsibility says how far the perspective actor answers for the
	// outcome; omitted means own. Every outcome is monitored, but only own
	// and shared outcomes count toward accountability.
	Responsibility Responsibility `json:"responsibility,omitempty"`
	// DeliveredBy are the actors who deliver the outcome.
	DeliveredBy []string `json:"deliveredBy,omitempty"`
	// AddressesCauses are the causes in the problem tree the outcome
	// answers.
	AddressesCauses []string `json:"addressesCauses,omitempty"`
	// AddressesEffects are the effects in the problem tree the outcome
	// relieves.
	AddressesEffects []string `json:"addressesEffects,omitempty"`
	// TargetGroupIDs are the priority populations the outcome is for.
	TargetGroupIDs []string `json:"targetGroupIds,omitempty"`
	// Priority records how the outcome scored against the prioritization
	// criteria.
	Priority *Priority `json:"priority,omitempty"`
	// AudienceIDs are the audiences who need to see this outcome reported.
	AudienceIDs []string `json:"audienceIds,omitempty"`
	// ContributesTo lists outcomes in a parent impact model this outcome
	// contributes to, as element references (model:<doc>#<outcome>).
	ContributesTo []string    `json:"contributesTo,omitempty"`
	Tags          []string    `json:"tags,omitempty"`
	Provenance    *Provenance `json:"provenance,omitempty"`
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
	Sufficiency string      `json:"sufficiency,omitempty"`
	Provenance  *Provenance `json:"provenance,omitempty"`
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
	Lag        string      `json:"lag,omitempty"`
	Provenance *Provenance `json:"provenance,omitempty"`
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
	// Priority records how the intervention scored against the
	// prioritization criteria.
	Priority   *Priority   `json:"priority,omitempty"`
	Provenance *Provenance `json:"provenance,omitempty"`
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
	Provenance   *Provenance      `json:"provenance,omitempty"`
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

	// MetricID instantiates a shared metric definition, so the same metric
	// is computed identically in every pathway.
	MetricID string `json:"metricId,omitempty"`
	// PathwayIDs are the pathways this indicator serves.
	PathwayIDs []string `json:"pathwayIds,omitempty"`
	// Claim says whether the result may be attributed to the initiative or
	// only contributed to. Omitted means contribution.
	Claim Claim `json:"claim,omitempty"`
	// Caveats must be shown wherever the indicator is reported, for
	// example "based on sign-ups, not attendance".
	Caveats []string `json:"caveats,omitempty"`
	// RollupFrom lists indicators of the same metric whose values combine
	// into this one using the metric's aggregation.
	RollupFrom []string `json:"rollupFrom,omitempty"`
	// ComposedOf says the value is worked out from other indicators, such as
	// a cost per outcome from a cost and a count. It is arithmetic on
	// measured values and is not a causal link; an indicator that has its own
	// observation uses that and ignores the composition.
	ComposedOf *Composition `json:"composedOf,omitempty"`
	// DataAvailability says whether the data this indicator needs is
	// public, already held internally, new to collect, held by a partner,
	// or not yet known.
	DataAvailability DataAvailability `json:"dataAvailability,omitempty"`
	// TargetGroupIDs are the priority populations the indicator is broken
	// down by.
	TargetGroupIDs []string `json:"targetGroupIds,omitempty"`
	// AudienceIDs are the audiences this indicator is reported to.
	AudienceIDs []string `json:"audienceIds,omitempty"`
	// ThemeIDs are the story themes a theme_count indicator counts.
	ThemeIDs []string `json:"themeIds,omitempty"`
	// Collection says how the data is gathered and what that costs.
	Collection *Collection `json:"collection,omitempty"`
	// Selection records whether the indicator is a candidate, selected, or
	// retired, and how it scored against the selection criteria.
	Selection  *Selection  `json:"selection,omitempty"`
	Provenance *Provenance `json:"provenance,omitempty"`
}

// Collection is how an indicator's data is gathered and what that costs.
type Collection struct {
	Mode CollectionMode `json:"mode" jsonschema:"required"`
	// System is where the data lives or is entered.
	System string `json:"system,omitempty"`
	// MinutesPerPeriod is the staff time one collection takes; Period is the
	// ISO 8601 duration it repeats over, e.g. "P1M".
	MinutesPerPeriod float64 `json:"minutesPerPeriod,omitempty" jsonschema:"minimum=0"`
	Period           string  `json:"period,omitempty"`
	// QuestionsPerRespondent is how many questions one respondent answers in
	// one collection.
	QuestionsPerRespondent int `json:"questionsPerRespondent,omitempty" jsonschema:"minimum=0"`
	RespondentsPerYear     int `json:"respondentsPerYear,omitempty" jsonschema:"minimum=0"`
	// SampleRate is the share of respondents asked, from above 0 to 1;
	// omitted means everyone.
	SampleRate float64 `json:"sampleRate,omitempty" jsonschema:"minimum=0,maximum=1"`
}

// Selection records where an indicator stands in metric selection.
type Selection struct {
	Status    SelectionStatus  `json:"status" jsonschema:"required"`
	Scores    []CriterionScore `json:"scores,omitempty"`
	Rationale string           `json:"rationale,omitempty"`
}

// CriterionScore is one score of a subject against a selection criterion.
type CriterionScore struct {
	CriterionID string `json:"criterionId" jsonschema:"required"`
	Score       int    `json:"score" jsonschema:"required,minimum=1,maximum=5"`
	Note        string `json:"note,omitempty"`
}

// SelectionCriterion is a question asked of every candidate, such as "does
// this help a decision?". Scores run 1 (poor) to 5 (strong); the weight sets
// how much the criterion counts in the weighted score.
type SelectionCriterion struct {
	ID        string           `json:"id" jsonschema:"required"`
	Name      string           `json:"name" jsonschema:"required"`
	Question  string           `json:"question,omitempty"`
	Weight    float64          `json:"weight,omitempty" jsonschema:"minimum=0"`
	AppliesTo SelectionSubject `json:"appliesTo" jsonschema:"required"`
	// Inverted marks a criterion where a low score is better, such as cost
	// to collect.
	Inverted bool `json:"inverted,omitempty"`
}

// MeasurementBudget is what the organization declares it can sustain.
type MeasurementBudget struct {
	StaffMinutesPerWeek              float64 `json:"staffMinutesPerWeek,omitempty" jsonschema:"minimum=0"`
	MaxQuestionsPerRespondentPerYear int     `json:"maxQuestionsPerRespondentPerYear,omitempty" jsonschema:"minimum=0"`
	TargetSelectedMetrics            *Range  `json:"targetSelectedMetrics,omitempty"`
}

// Range is an inclusive count range.
type Range struct {
	Min int `json:"min,omitempty" jsonschema:"minimum=0"`
	Max int `json:"max,omitempty" jsonschema:"minimum=0"`
}

// EffectiveClaim returns the indicator's claim, defaulting to contribution.
func (i Indicator) EffectiveClaim() Claim {
	if i.Claim == "" {
		return ClaimContribution
	}
	return i.Claim
}

// Composition builds an indicator's value from the latest values of others.
type Composition struct {
	Op CompositionOp `json:"op" jsonschema:"required"`
	// IndicatorIDs are the parts. A ratio and a difference take exactly two,
	// in order: the first is the dividend or the minuend.
	IndicatorIDs []string `json:"indicatorIds" jsonschema:"required"`
}

// Metric is a shared metric definition with an exact formula. Indicators
// instantiate it per pathway, so one metric means one thing everywhere.
type Metric struct {
	ID         string `json:"id" jsonschema:"required"`
	Name       string `json:"name" jsonschema:"required"`
	Definition string `json:"definition" jsonschema:"required"`
	// Formula states exactly how the value is computed, in words, for
	// example "sum over sessions of attendees times session hours".
	Formula     string      `json:"formula" jsonschema:"required"`
	Unit        string      `json:"unit,omitempty"`
	Direction   Direction   `json:"direction,omitempty"`
	Aggregation Aggregation `json:"aggregation,omitempty"`
	// WeightMetricID names the metric that weights a weighted_mean.
	WeightMetricID string      `json:"weightMetricId,omitempty"`
	Layer          MetricLayer `json:"layer,omitempty"`
	Provenance     *Provenance `json:"provenance,omitempty"`
}

// GlossaryTerm defines a term of the shared vocabulary, for example the
// difference between an output and an outcome for this organization.
type GlossaryTerm struct {
	ID         string   `json:"id" jsonschema:"required"`
	Term       string   `json:"term" jsonschema:"required"`
	Definition string   `json:"definition" jsonschema:"required"`
	Examples   []string `json:"examples,omitempty"`
	RelatedIDs []string `json:"relatedIds,omitempty"`
}

// Priority records how an outcome or intervention scored against the
// prioritization criteria, and why it was chosen or set aside.
type Priority struct {
	Scores    []CriterionScore `json:"scores,omitempty"`
	Rationale string           `json:"rationale,omitempty"`
}

// Cause is a root of the problem tree: a condition that leads to the core
// problem (context.problemStatement), directly or through other causes.
// Outcomes name the causes they address, so the solution can be checked
// against the problem at every level.
type Cause struct {
	ID        string     `json:"id" jsonschema:"required"`
	Statement string     `json:"statement" jsonschema:"required"`
	Level     CauseLevel `json:"level" jsonschema:"required"`
	// Parents are the causes this one contributes to; empty for a cause that
	// leads straight to the core problem. Causes form a network, and a tree
	// is the case of at most one parent each.
	Parents          []string `json:"parents,omitempty"`
	AffectedGroupIDs []string `json:"affectedGroupIds,omitempty"`
	EvidenceIDs      []string `json:"evidenceIds,omitempty"`
	// OutOfScope records that the cause is real but not addressed here.
	OutOfScope *OutOfScope `json:"outOfScope,omitempty"`
	Provenance *Provenance `json:"provenance,omitempty"`
}

// Effect is a branch of the problem tree: a consequence of the core problem,
// directly or through other effects. Outcomes name the effects they relieve,
// so the impact model can be checked against the harm it is meant to reduce.
type Effect struct {
	ID        string `json:"id" jsonschema:"required"`
	Statement string `json:"statement" jsonschema:"required"`
	// Parents are the effects this one follows from; empty for an effect of
	// the core problem itself. Effects form a network, as causes do.
	Parents          []string `json:"parents,omitempty"`
	AffectedGroupIDs []string `json:"affectedGroupIds,omitempty"`
	EvidenceIDs      []string `json:"evidenceIds,omitempty"`
	// OutOfScope records that the effect is real but not relieved here.
	OutOfScope *OutOfScope `json:"outOfScope,omitempty"`
	Provenance *Provenance `json:"provenance,omitempty"`
}

// OutOfScope says why a cause or an effect is left alone.
type OutOfScope struct {
	Reason string `json:"reason" jsonschema:"required"`
}

// TargetGroup is a priority population, such as under-resourced member
// organizations or residents of low-access areas.
type TargetGroup struct {
	ID          string   `json:"id" jsonschema:"required"`
	Name        string   `json:"name" jsonschema:"required"`
	Description string   `json:"description,omitempty"`
	Size        *float64 `json:"size,omitempty"`
	ActorIDs    []string `json:"actorIds,omitempty"`
}

// Audience is a group who needs to see results, and what they need to see.
type Audience struct {
	ID    string `json:"id" jsonschema:"required"`
	Name  string `json:"name" jsonschema:"required"`
	Needs string `json:"needs,omitempty"`
	// Channels are where they are reached, e.g. annual report, grant
	// proposal, board deck, newsletter.
	Channels []string `json:"channels,omitempty"`
	// Headline is the short list of indicators this audience is told, in the
	// order they are shown. It is drawn from the indicators reported to the
	// audience. Every audience sees the same figure for an indicator; only the
	// label and the note differ.
	Headline []HeadlineMetric `json:"headline,omitempty"`
}

// HeadlineMetric is one number in an audience's headline list.
type HeadlineMetric struct {
	// IndicatorID is the indicator whose figure is shown.
	IndicatorID string `json:"indicatorId" jsonschema:"required"`
	// Label words the indicator for this audience, for example "organizations
	// we worked with" in place of "Distinct organizations reached". It must not
	// change what is counted.
	Label string `json:"label,omitempty"`
	// Note says why this audience is shown this number.
	Note string `json:"note,omitempty"`
}

// Theme is an entry of the story codebook, so stories are coded the same
// way by every person or model that codes them.
type Theme struct {
	ID         string   `json:"id" jsonschema:"required"`
	Label      string   `json:"label" jsonschema:"required"`
	Definition string   `json:"definition" jsonschema:"required"`
	Includes   []string `json:"includes,omitempty"`
	Excludes   []string `json:"excludes,omitempty"`
	Examples   []string `json:"examples,omitempty"`
	// OutcomeIDs are the outcomes the theme is evidence about.
	OutcomeIDs []string `json:"outcomeIds,omitempty"`
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
	Provenance   *Provenance   `json:"provenance,omitempty"`
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

// ExternalFactor is a condition outside the impact model's control that helps or
// hinders change.
type ExternalFactor struct {
	ID          string     `json:"id" jsonschema:"required"`
	Description string     `json:"description" jsonschema:"required"`
	Kind        FactorKind `json:"kind,omitempty"`
	Affects     []string   `json:"affects,omitempty"`
	Likelihood  Rating     `json:"likelihood,omitempty"`
	// Trend is a prism-core trend direction (up, down, flat, unknown).
	Trend      string      `json:"trend,omitempty" jsonschema:"enum=up,enum=down,enum=flat,enum=unknown"`
	Provenance *Provenance `json:"provenance,omitempty"`
}

// Risk is a prism-core risk attached to elements of the impact model.
type Risk struct {
	core.Risk
	AppliesTo            []string     `json:"appliesTo,omitempty"`
	Category             RiskCategory `json:"category,omitempty"`
	RelatedAssumptionIDs []string     `json:"relatedAssumptionIds,omitempty"`
	Provenance           *Provenance  `json:"provenance,omitempty"`
}

// Provenance records why an element is in the impact model: its basis and the
// evidence behind it. Most elements carry one, so a draft built from public
// sources can say what a source stated and what an analyst inferred.
type Provenance struct {
	Basis Basis `json:"basis" jsonschema:"required"`
	// EvidenceIDs cite the sources behind the basis.
	EvidenceIDs []string `json:"evidenceIds,omitempty"`
	Note        string   `json:"note,omitempty"`
	// DecidedAt and DecidedBy record when and by whom (a group or role) a
	// basis of confirmed or rejected was decided.
	DecidedAt string `json:"decidedAt,omitempty" jsonschema:"format=date"`
	DecidedBy string `json:"decidedBy,omitempty"`
}

// IsRejected reports whether the provenance marks the element rejected.
// It is safe on a nil Provenance.
func (p *Provenance) IsRejected() bool { return p != nil && p.Basis == BasisRejected }

// Rejected reports whether the element was reviewed and rejected. Rejected
// elements stay in the document but are excluded from analysis.
func (o Outcome) Rejected() bool { return o.Provenance.IsRejected() }

// Rejected reports whether the group was rejected.
func (g PreconditionGroup) Rejected() bool { return g.Provenance.IsRejected() }

// Rejected reports whether the link was rejected.
func (l CausalLink) Rejected() bool { return l.Provenance.IsRejected() }

// Rejected reports whether the intervention was rejected.
func (i Intervention) Rejected() bool { return i.Provenance.IsRejected() }

// Rejected reports whether the assumption was rejected.
func (a Assumption) Rejected() bool { return a.Provenance.IsRejected() }

// Rejected reports whether the indicator was rejected.
func (i Indicator) Rejected() bool { return i.Provenance.IsRejected() }

// Rejected reports whether the actor was rejected.
func (a Actor) Rejected() bool { return a.Provenance.IsRejected() }

// Rejected reports whether the external factor was rejected.
func (f ExternalFactor) Rejected() bool { return f.Provenance.IsRejected() }

// Rejected reports whether the risk was rejected.
func (r Risk) Rejected() bool { return r.Provenance.IsRejected() }

// Rejected reports whether the cause was rejected in review.
func (c Cause) Rejected() bool { return c.Provenance.IsRejected() }

// Rejected reports whether the effect was rejected in review.
func (e Effect) Rejected() bool { return e.Provenance.IsRejected() }

// Rejected reports whether the metric was rejected.
func (m Metric) Rejected() bool { return m.Provenance.IsRejected() }

// Rejected reports whether the pathway was rejected.
func (p Pathway) Rejected() bool { return p.Provenance.IsRejected() }

// Consultation records one validation of the impact model with stakeholders: the
// step validated, how, who took part (as groups), what they said, and what
// changed as a result.
type Consultation struct {
	ID     string             `json:"id" jsonschema:"required"`
	Date   string             `json:"date" jsonschema:"required,format=date"`
	Step   ConsultationStep   `json:"step" jsonschema:"required"`
	Method ConsultationMethod `json:"method,omitempty"`
	// StakeholderGroups name groups or roles, not individuals.
	StakeholderGroups []string `json:"stakeholderGroups,omitempty"`
	// Count is how many people took part.
	Count      int      `json:"count,omitempty"`
	Findings   []string `json:"findings,omitempty"`
	ChangedIDs []string `json:"changedIds,omitempty"`
}

// Pathway is a named grouping of outcomes and interventions, such as a
// program area or a strategy. Detail for a pathway may live in a child
// Impact model that points back with the document's parent field.
type Pathway struct {
	ID              string      `json:"id" jsonschema:"required"`
	Title           string      `json:"title" jsonschema:"required"`
	Kind            PathwayKind `json:"kind,omitempty"`
	Description     string      `json:"description,omitempty"`
	OwnerID         string      `json:"ownerActorId,omitempty"`
	OutcomeIDs      []string    `json:"outcomeIds,omitempty"`
	InterventionIDs []string    `json:"interventionIds,omitempty"`
	// ChildModel references an impact model that elaborates this pathway
	// (model:<doc>).
	ChildModel string `json:"childModel,omitempty" jsonschema:"pattern=^model:[a-z][a-z0-9]*(-[a-z0-9]+)*$"`
	// Focus names the metrics this program is steered by.
	Focus *Focus `json:"focus,omitempty"`
	// Projections names the projections this pathway is meant for beyond the
	// ones every pathway gets. dmaic marks a program whose metrics are
	// measured often enough to improve with DMAIC: lint then checks their
	// cadence and sample size, and export dmaic includes the pathway.
	Projections []Projection `json:"projections,omitempty"`
	Provenance  *Provenance  `json:"provenance,omitempty"`
}

// Focus names the metrics a program, product, or other pathway is steered by.
// It is for steering, and is not the same as what an audience is told: see
// Audience.Headline.
type Focus struct {
	// NorthStarID is the one indicator that best captures the value the pathway
	// delivers to the people it serves. It measures an outcome at or below the
	// ceiling of accountability, never an output or an activity.
	NorthStarID string `json:"northStarId,omitempty"`
	// OneMetricID is the indicator to focus on at the pathway's current stage.
	// It changes as the stage does, and may be the same as the North Star.
	OneMetricID string `json:"oneMetricId,omitempty"`
	// GuardrailIDs are indicators watched so that pursuing the North Star does
	// not harm something else, such as equity, quality, or staff burden.
	GuardrailIDs []string `json:"guardrailIds,omitempty"`
	// Rationale says why these were chosen.
	Rationale string `json:"rationale,omitempty"`
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

// Alignment links an element of this impact model to an external element, such as
// a Balanced Scorecard objective or an OKR key result.
type Alignment struct {
	ElementID string      `json:"elementId" jsonschema:"required"`
	Target    ExternalRef `json:"target" jsonschema:"required"`
}
