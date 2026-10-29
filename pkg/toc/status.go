package toc

import "slices"

// State is the computed status of an indicator, outcome, or pathway.
type State string

const (
	StateAchieved     State = "achieved"
	StateOnTrack      State = "on_track"
	StateAtRisk       State = "at_risk"
	StateOffTrack     State = "off_track"
	StateNotStarted   State = "not_started"
	StateNoData       State = "no_data"
	StateStale        State = "stale"
	StateBaselineOnly State = "baseline_only"
	// StateUnmeasured applies to outcomes with no indicators.
	StateUnmeasured State = "unmeasured"
)

// AllStates returns all states.
func AllStates() []State {
	return []State{
		StateAchieved, StateOnTrack, StateAtRisk, StateOffTrack, StateNotStarted,
		StateNoData, StateStale, StateBaselineOnly, StateUnmeasured,
	}
}

// Valid reports whether v is a known State.
func (v State) Valid() bool { return slices.Contains(AllStates(), v) }

// LinkState is the computed state of a causal link.
type LinkState string

const (
	LinkStateOK     LinkState = "ok"
	LinkStateAtRisk LinkState = "at_risk"
	LinkStateBroken LinkState = "broken"
)

// AllLinkStates returns all link states.
func AllLinkStates() []LinkState { return []LinkState{LinkStateOK, LinkStateAtRisk, LinkStateBroken} }

// Valid reports whether v is a known LinkState.
func (v LinkState) Valid() bool { return slices.Contains(AllLinkStates(), v) }

// RollupMethod is how an outcome's indicator states combine.
type RollupMethod string

const (
	RollupWorst RollupMethod = "worst"
	RollupMean  RollupMethod = "mean"
)

// AllRollupMethods returns all roll-up methods.
func AllRollupMethods() []RollupMethod { return []RollupMethod{RollupWorst, RollupMean} }

// Valid reports whether v is a known RollupMethod.
func (v RollupMethod) Valid() bool { return slices.Contains(AllRollupMethods(), v) }

// AssumptionSource is where an assumption's effective status came from.
type AssumptionSource string

const (
	SourceDocument   AssumptionSource = "document"
	SourceAssessment AssumptionSource = "assessment"
)

// AllAssumptionSources returns all assumption sources.
func AllAssumptionSources() []AssumptionSource {
	return []AssumptionSource{SourceDocument, SourceAssessment}
}

// Valid reports whether v is a known AssumptionSource.
func (v AssumptionSource) Valid() bool { return slices.Contains(AllAssumptionSources(), v) }

// Status is a computed snapshot of a theory's progress at a date. It is
// produced by measurement, never written by hand, and is diffable across
// periods.
type Status struct {
	Schema      string        `json:"$schema,omitempty"`
	Kind        Kind          `json:"kind" jsonschema:"required,enum=TheoryOfChangeStatus"`
	SpecVersion string        `json:"specVersion" jsonschema:"required"`
	TocID       string        `json:"tocId" jsonschema:"required"`
	AsOf        string        `json:"asOf" jsonschema:"required,format=date"`
	Options     StatusOptions `json:"options" jsonschema:"required"`

	Indicators  []IndicatorStatus `json:"indicators,omitempty"`
	Outcomes    []OutcomeStatus   `json:"outcomes,omitempty"`
	Links       []LinkStatus      `json:"links,omitempty"`
	Assumptions []AssumptionState `json:"assumptions,omitempty"`
	Overall     []ImpactStatus    `json:"overall,omitempty"`
}

// StatusOptions records the settings a snapshot was computed with, so every
// score can be traced to its inputs.
type StatusOptions struct {
	// Tolerance is the on-track tolerance against expected progress.
	Tolerance float64           `json:"tolerance"`
	Rollup    RollupMethod      `json:"rollup" jsonschema:"required"`
	Weights   ConfidenceWeights `json:"confidenceWeights"`
}

// ConfidenceWeights are the rule-based link-confidence weights.
type ConfidenceWeights struct {
	Evidence    float64 `json:"evidence"`
	Assumptions float64 `json:"assumptions"`
	Rationale   float64 `json:"rationale"`
}

// DefaultStatusOptions returns the documented defaults (TRD §6.5).
func DefaultStatusOptions() StatusOptions {
	return StatusOptions{
		Tolerance: 0.10,
		Rollup:    RollupWorst,
		Weights:   ConfidenceWeights{Evidence: 0.4, Assumptions: 0.4, Rationale: 0.2},
	}
}

// IndicatorStatus is an indicator's computed state.
type IndicatorStatus struct {
	IndicatorID string `json:"indicatorId" jsonschema:"required"`
	State       State  `json:"state" jsonschema:"required"`
	// Health is a prism-core health value (green, yellow, red) or empty.
	Health string `json:"health,omitempty"`
	// Progress and Expected are fractions of the way from baseline to the
	// final target.
	Progress         *float64     `json:"progress,omitempty"`
	Expected         *float64     `json:"expected,omitempty"`
	Latest           *Observation `json:"latest,omitempty"`
	ObservationCount int          `json:"observationCount,omitempty"`
}

// OutcomeStatus is an outcome's computed state from its own indicators and
// from its preconditions.
type OutcomeStatus struct {
	OutcomeID    string   `json:"outcomeId" jsonschema:"required"`
	State        State    `json:"state" jsonschema:"required"`
	PathwayState State    `json:"pathwayState" jsonschema:"required"`
	Health       string   `json:"health,omitempty"`
	Confidence   *float64 `json:"confidence,omitempty"`
	IndicatorIDs []string `json:"indicatorIds,omitempty"`
	// AssumptionBroken lists invalidated or at-risk assumptions on links
	// upstream of this outcome.
	AssumptionBroken []string `json:"assumptionBroken,omitempty"`
}

// LinkStatus is a causal link's computed state and confidence.
type LinkStatus struct {
	LinkID     string    `json:"linkId" jsonschema:"required"`
	State      LinkState `json:"state" jsonschema:"required"`
	Confidence float64   `json:"confidence"`
}

// AssumptionState is an assumption's effective status and where it came
// from.
type AssumptionState struct {
	AssumptionID string           `json:"assumptionId" jsonschema:"required"`
	Status       AssumptionStatus `json:"status" jsonschema:"required"`
	Source       AssumptionSource `json:"source" jsonschema:"required"`
	AssessedAt   string           `json:"assessedAt,omitempty" jsonschema:"format=date"`
}

// ImpactStatus is the headline status for one impact outcome.
type ImpactStatus struct {
	OutcomeID    string   `json:"outcomeId" jsonschema:"required"`
	PathwayState State    `json:"pathwayState" jsonschema:"required"`
	Health       string   `json:"health,omitempty"`
	Confidence   *float64 `json:"confidence,omitempty"`
}

// StatusDiff lists what changed between two status snapshots.
type StatusDiff struct {
	TocID    string         `json:"tocId" jsonschema:"required"`
	FromAsOf string         `json:"fromAsOf" jsonschema:"required,format=date"`
	ToAsOf   string         `json:"toAsOf" jsonschema:"required,format=date"`
	Changes  []StatusChange `json:"changes,omitempty"`
}

// StatusChange is one changed value.
type StatusChange struct {
	ElementID   string      `json:"elementId" jsonschema:"required"`
	ElementKind ElementKind `json:"elementKind" jsonschema:"required"`
	Field       string      `json:"field" jsonschema:"required"`
	From        string      `json:"from,omitempty"`
	To          string      `json:"to,omitempty"`
}
