package toc

import "slices"

// Enumerations are typed strings with snake_case values. Each has an All…
// function listing its values in a stable order and a Valid method. The
// empty string is never valid; optional fields are checked only when set.

// Level is an outcome's position in the pathway of change.
type Level string

const (
	LevelImpact       Level = "impact"
	LevelLongTerm     Level = "long_term"
	LevelIntermediate Level = "intermediate"
	LevelShortTerm    Level = "short_term"
	LevelOutput       Level = "output"
)

// AllLevels returns levels from highest to lowest.
func AllLevels() []Level {
	return []Level{LevelImpact, LevelLongTerm, LevelIntermediate, LevelShortTerm, LevelOutput}
}

// Valid reports whether v is a known Level.
func (v Level) Valid() bool { return slices.Contains(AllLevels(), v) }

// Rank returns the level's height: impact is highest (5), output lowest (1).
// Unknown levels return 0.
func (v Level) Rank() int {
	all := AllLevels()
	if i := slices.Index(all, v); i >= 0 {
		return len(all) - i
	}
	return 0
}

// Sphere is the degree of control an initiative has over an outcome.
type Sphere string

const (
	SphereControl   Sphere = "control"
	SphereInfluence Sphere = "influence"
	SphereInterest  Sphere = "interest"
)

// AllSpheres returns all spheres.
func AllSpheres() []Sphere { return []Sphere{SphereControl, SphereInfluence, SphereInterest} }

// Valid reports whether v is a known Sphere.
func (v Sphere) Valid() bool { return slices.Contains(AllSpheres(), v) }

// LogicType is how preconditions combine.
type LogicType string

const (
	LogicAll     LogicType = "all"
	LogicAny     LogicType = "any"
	LogicAtLeast LogicType = "at_least"
)

// AllLogicTypes returns all logic types.
func AllLogicTypes() []LogicType { return []LogicType{LogicAll, LogicAny, LogicAtLeast} }

// Valid reports whether v is a known LogicType.
func (v LogicType) Valid() bool { return slices.Contains(AllLogicTypes(), v) }

// ChangeType is the kind of change an outcome describes.
type ChangeType string

const (
	ChangeKnowledge  ChangeType = "knowledge"
	ChangeAttitude   ChangeType = "attitude"
	ChangeSkill      ChangeType = "skill"
	ChangeBehavior   ChangeType = "behavior"
	ChangePractice   ChangeType = "practice"
	ChangePolicy     ChangeType = "policy"
	ChangeNorm       ChangeType = "norm"
	ChangeCondition  ChangeType = "condition"
	ChangeInvestment ChangeType = "investment"
	ChangeAdoption   ChangeType = "adoption"
)

// AllChangeTypes returns all change types.
func AllChangeTypes() []ChangeType {
	return []ChangeType{
		ChangeKnowledge, ChangeAttitude, ChangeSkill, ChangeBehavior, ChangePractice,
		ChangePolicy, ChangeNorm, ChangeCondition, ChangeInvestment, ChangeAdoption,
	}
}

// Valid reports whether v is a known ChangeType.
func (v ChangeType) Valid() bool { return slices.Contains(AllChangeTypes(), v) }

// LinkType is the nature of a causal link.
type LinkType string

const (
	LinkCauses             LinkType = "causes"
	LinkEnables            LinkType = "enables"
	LinkContributesTo      LinkType = "contributes_to"
	LinkNecessaryCondition LinkType = "necessary_condition"
)

// AllLinkTypes returns all link types.
func AllLinkTypes() []LinkType {
	return []LinkType{LinkCauses, LinkEnables, LinkContributesTo, LinkNecessaryCondition}
}

// Valid reports whether v is a known LinkType.
func (v LinkType) Valid() bool { return slices.Contains(AllLinkTypes(), v) }

// Polarity is whether a link increases or decreases its target.
type Polarity string

const (
	PolarityPositive Polarity = "positive"
	PolarityNegative Polarity = "negative"
)

// AllPolarities returns all polarities.
func AllPolarities() []Polarity { return []Polarity{PolarityPositive, PolarityNegative} }

// Valid reports whether v is a known Polarity.
func (v Polarity) Valid() bool { return slices.Contains(AllPolarities(), v) }

// LoopType classifies an intentional feedback loop.
type LoopType string

const (
	LoopReinforcing LoopType = "reinforcing"
	LoopBalancing   LoopType = "balancing"
)

// AllLoopTypes returns all loop types.
func AllLoopTypes() []LoopType { return []LoopType{LoopReinforcing, LoopBalancing} }

// Valid reports whether v is a known LoopType.
func (v LoopType) Valid() bool { return slices.Contains(AllLoopTypes(), v) }

// InterventionType classifies an intervention.
type InterventionType string

const (
	InterventionProgram         InterventionType = "program"
	InterventionProduct         InterventionType = "product"
	InterventionPolicy          InterventionType = "policy"
	InterventionAdvocacy        InterventionType = "advocacy"
	InterventionPartnership     InterventionType = "partnership"
	InterventionStandardization InterventionType = "standardization"
	InterventionInvestment      InterventionType = "investment"
	InterventionCommunication   InterventionType = "communication"
	InterventionResearch        InterventionType = "research"
	InterventionOther           InterventionType = "other"
)

// AllInterventionTypes returns all intervention types.
func AllInterventionTypes() []InterventionType {
	return []InterventionType{
		InterventionProgram, InterventionProduct, InterventionPolicy, InterventionAdvocacy,
		InterventionPartnership, InterventionStandardization, InterventionInvestment,
		InterventionCommunication, InterventionResearch, InterventionOther,
	}
}

// Valid reports whether v is a known InterventionType.
func (v InterventionType) Valid() bool { return slices.Contains(AllInterventionTypes(), v) }

// InterventionStatus is an intervention's lifecycle state.
type InterventionStatus string

const (
	InterventionPlanned    InterventionStatus = "planned"
	InterventionInProgress InterventionStatus = "in_progress"
	InterventionCompleted  InterventionStatus = "completed"
	InterventionPaused     InterventionStatus = "paused"
	InterventionCancelled  InterventionStatus = "cancelled"
)

// AllInterventionStatuses returns all intervention statuses.
func AllInterventionStatuses() []InterventionStatus {
	return []InterventionStatus{
		InterventionPlanned, InterventionInProgress, InterventionCompleted,
		InterventionPaused, InterventionCancelled,
	}
}

// Valid reports whether v is a known InterventionStatus.
func (v InterventionStatus) Valid() bool { return slices.Contains(AllInterventionStatuses(), v) }

// AssumptionType classifies an assumption (research brief §2, UNDG).
type AssumptionType string

const (
	AssumptionCausal         AssumptionType = "causal"
	AssumptionContextual     AssumptionType = "contextual"
	AssumptionImplementation AssumptionType = "implementation"
	AssumptionBehavioral     AssumptionType = "behavioral"
	AssumptionExternal       AssumptionType = "external"
	AssumptionPartnership    AssumptionType = "partnership"
)

// AllAssumptionTypes returns all assumption types.
func AllAssumptionTypes() []AssumptionType {
	return []AssumptionType{
		AssumptionCausal, AssumptionContextual, AssumptionImplementation,
		AssumptionBehavioral, AssumptionExternal, AssumptionPartnership,
	}
}

// Valid reports whether v is a known AssumptionType.
func (v AssumptionType) Valid() bool { return slices.Contains(AllAssumptionTypes(), v) }

// AssumptionStatus is the current assessment of an assumption.
type AssumptionStatus string

const (
	AssumptionUntested    AssumptionStatus = "untested"
	AssumptionHolding     AssumptionStatus = "holding"
	AssumptionAtRisk      AssumptionStatus = "at_risk"
	AssumptionInvalidated AssumptionStatus = "invalidated"
	AssumptionValidated   AssumptionStatus = "validated"
)

// AllAssumptionStatuses returns all assumption statuses.
func AllAssumptionStatuses() []AssumptionStatus {
	return []AssumptionStatus{
		AssumptionUntested, AssumptionHolding, AssumptionAtRisk,
		AssumptionInvalidated, AssumptionValidated,
	}
}

// Valid reports whether v is a known AssumptionStatus.
func (v AssumptionStatus) Valid() bool { return slices.Contains(AllAssumptionStatuses(), v) }

// Criticality is how much the theory depends on an assumption.
type Criticality string

const (
	CriticalityLow      Criticality = "low"
	CriticalityMedium   Criticality = "medium"
	CriticalityHigh     Criticality = "high"
	CriticalityCritical Criticality = "critical"
)

// AllCriticalities returns criticalities from lowest to highest.
func AllCriticalities() []Criticality {
	return []Criticality{CriticalityLow, CriticalityMedium, CriticalityHigh, CriticalityCritical}
}

// Valid reports whether v is a known Criticality.
func (v Criticality) Valid() bool { return slices.Contains(AllCriticalities(), v) }

// Rating is a three-point low / medium / high scale used for plausibility,
// influence, interest, and likelihood.
type Rating string

const (
	RatingLow    Rating = "low"
	RatingMedium Rating = "medium"
	RatingHigh   Rating = "high"
)

// AllRatings returns ratings from lowest to highest.
func AllRatings() []Rating { return []Rating{RatingLow, RatingMedium, RatingHigh} }

// Valid reports whether v is a known Rating.
func (v Rating) Valid() bool { return slices.Contains(AllRatings(), v) }

// MeasurementType is how an indicator is measured.
type MeasurementType string

const (
	MeasurementQuantitative MeasurementType = "quantitative"
	MeasurementQualitative  MeasurementType = "qualitative"
	MeasurementHybrid       MeasurementType = "hybrid"
)

// AllMeasurementTypes returns all measurement types.
func AllMeasurementTypes() []MeasurementType {
	return []MeasurementType{MeasurementQuantitative, MeasurementQualitative, MeasurementHybrid}
}

// Valid reports whether v is a known MeasurementType.
func (v MeasurementType) Valid() bool { return slices.Contains(AllMeasurementTypes(), v) }

// Direction is the desired direction of change for an indicator.
type Direction string

const (
	DirectionIncrease Direction = "increase"
	DirectionDecrease Direction = "decrease"
	DirectionMaintain Direction = "maintain"
)

// AllDirections returns all directions.
func AllDirections() []Direction {
	return []Direction{DirectionIncrease, DirectionDecrease, DirectionMaintain}
}

// Valid reports whether v is a known Direction.
func (v Direction) Valid() bool { return slices.Contains(AllDirections(), v) }

// ActorType classifies an actor.
type ActorType string

const (
	ActorOrganization    ActorType = "organization"
	ActorPerson          ActorType = "person"
	ActorGroup           ActorType = "group"
	ActorCommunity       ActorType = "community"
	ActorCustomerSegment ActorType = "customer_segment"
	ActorStandardsBody   ActorType = "standards_body"
	ActorRegulator       ActorType = "regulator"
	ActorVendor          ActorType = "vendor"
	ActorPlatform        ActorType = "platform"
	ActorInvestor        ActorType = "investor"
	ActorOther           ActorType = "other"
)

// AllActorTypes returns all actor types.
func AllActorTypes() []ActorType {
	return []ActorType{
		ActorOrganization, ActorPerson, ActorGroup, ActorCommunity, ActorCustomerSegment,
		ActorStandardsBody, ActorRegulator, ActorVendor, ActorPlatform, ActorInvestor, ActorOther,
	}
}

// Valid reports whether v is a known ActorType.
func (v ActorType) Valid() bool { return slices.Contains(AllActorTypes(), v) }

// ActorRole is a role an actor plays in the theory.
type ActorRole string

const (
	RoleImplementer     ActorRole = "implementer"
	RoleBeneficiary     ActorRole = "beneficiary"
	RoleBoundaryPartner ActorRole = "boundary_partner"
	RoleFunder          ActorRole = "funder"
	RoleChampion        ActorRole = "champion"
	RoleInfluencer      ActorRole = "influencer"
	RoleCompetitor      ActorRole = "competitor"
	RoleAdopter         ActorRole = "adopter"
	RoleGatekeeper      ActorRole = "gatekeeper"
)

// AllActorRoles returns all actor roles.
func AllActorRoles() []ActorRole {
	return []ActorRole{
		RoleImplementer, RoleBeneficiary, RoleBoundaryPartner, RoleFunder, RoleChampion,
		RoleInfluencer, RoleCompetitor, RoleAdopter, RoleGatekeeper,
	}
}

// Valid reports whether v is a known ActorRole.
func (v ActorRole) Valid() bool { return slices.Contains(AllActorRoles(), v) }

// Stance is an actor's position toward the theory's goal.
type Stance string

const (
	StanceChampion  Stance = "champion"
	StanceSupporter Stance = "supporter"
	StanceNeutral   Stance = "neutral"
	StanceSkeptic   Stance = "skeptic"
	StanceOpponent  Stance = "opponent"
	StanceUnknown   Stance = "unknown"
)

// AllStances returns all stances.
func AllStances() []Stance {
	return []Stance{StanceChampion, StanceSupporter, StanceNeutral, StanceSkeptic, StanceOpponent, StanceUnknown}
}

// Valid reports whether v is a known Stance.
func (v Stance) Valid() bool { return slices.Contains(AllStances(), v) }

// PerspectiveRole is the perspective actor's relationship to the theory.
type PerspectiveRole string

const (
	PerspectiveProponent   PerspectiveRole = "proponent"
	PerspectiveFunder      PerspectiveRole = "funder"
	PerspectiveImplementer PerspectiveRole = "implementer"
	PerspectiveObserver    PerspectiveRole = "observer"
)

// AllPerspectiveRoles returns all perspective roles.
func AllPerspectiveRoles() []PerspectiveRole {
	return []PerspectiveRole{PerspectiveProponent, PerspectiveFunder, PerspectiveImplementer, PerspectiveObserver}
}

// Valid reports whether v is a known PerspectiveRole.
func (v PerspectiveRole) Valid() bool { return slices.Contains(AllPerspectiveRoles(), v) }

// EvidenceType classifies an evidence record.
type EvidenceType string

const (
	EvidenceResearch             EvidenceType = "research"
	EvidenceEvaluation           EvidenceType = "evaluation"
	EvidenceMonitoringData       EvidenceType = "monitoring_data"
	EvidenceExpertOpinion        EvidenceType = "expert_opinion"
	EvidenceStakeholderTestimony EvidenceType = "stakeholder_testimony"
	EvidenceCaseStudy            EvidenceType = "case_study"
	EvidenceInternalMetric       EvidenceType = "internal_metric"
	EvidenceMarketSignal         EvidenceType = "market_signal"
	EvidenceOutcomeHarvest       EvidenceType = "outcome_harvest"
	EvidenceAnecdote             EvidenceType = "anecdote"
)

// AllEvidenceTypes returns all evidence types.
func AllEvidenceTypes() []EvidenceType {
	return []EvidenceType{
		EvidenceResearch, EvidenceEvaluation, EvidenceMonitoringData, EvidenceExpertOpinion,
		EvidenceStakeholderTestimony, EvidenceCaseStudy, EvidenceInternalMetric,
		EvidenceMarketSignal, EvidenceOutcomeHarvest, EvidenceAnecdote,
	}
}

// Valid reports whether v is a known EvidenceType.
func (v EvidenceType) Valid() bool { return slices.Contains(AllEvidenceTypes(), v) }

// Strength is how strongly evidence supports (or contradicts) a claim.
type Strength string

const (
	StrengthStrong    Strength = "strong"
	StrengthModerate  Strength = "moderate"
	StrengthWeak      Strength = "weak"
	StrengthAnecdotal Strength = "anecdotal"
)

// AllStrengths returns strengths from strongest to weakest.
func AllStrengths() []Strength {
	return []Strength{StrengthStrong, StrengthModerate, StrengthWeak, StrengthAnecdotal}
}

// Valid reports whether v is a known Strength.
func (v Strength) Valid() bool { return slices.Contains(AllStrengths(), v) }

// EvidenceDirection is whether evidence supports or contradicts.
type EvidenceDirection string

const (
	EvidenceSupports    EvidenceDirection = "supports"
	EvidenceContradicts EvidenceDirection = "contradicts"
)

// AllEvidenceDirections returns all evidence directions.
func AllEvidenceDirections() []EvidenceDirection {
	return []EvidenceDirection{EvidenceSupports, EvidenceContradicts}
}

// Valid reports whether v is a known EvidenceDirection.
func (v EvidenceDirection) Valid() bool { return slices.Contains(AllEvidenceDirections(), v) }

// ProcessTracingTest is the process-tracing test an evidence item passes.
type ProcessTracingTest string

const (
	TestStrawInTheWind ProcessTracingTest = "straw_in_the_wind"
	TestHoop           ProcessTracingTest = "hoop"
	TestSmokingGun     ProcessTracingTest = "smoking_gun"
	TestDoublyDecisive ProcessTracingTest = "doubly_decisive"
)

// AllProcessTracingTests returns all process-tracing tests.
func AllProcessTracingTests() []ProcessTracingTest {
	return []ProcessTracingTest{TestStrawInTheWind, TestHoop, TestSmokingGun, TestDoublyDecisive}
}

// Valid reports whether v is a known ProcessTracingTest.
func (v ProcessTracingTest) Valid() bool { return slices.Contains(AllProcessTracingTests(), v) }

// Sensitivity controls whether content may be rendered verbatim.
type Sensitivity string

const (
	SensitivityPublic     Sensitivity = "public"
	SensitivityInternal   Sensitivity = "internal"
	SensitivityRestricted Sensitivity = "restricted"
)

// AllSensitivities returns all sensitivities.
func AllSensitivities() []Sensitivity {
	return []Sensitivity{SensitivityPublic, SensitivityInternal, SensitivityRestricted}
}

// Valid reports whether v is a known Sensitivity.
func (v Sensitivity) Valid() bool { return slices.Contains(AllSensitivities(), v) }

// Renderable reports whether content at this sensitivity may be shown
// verbatim in human-facing output. The zero value is treated as restricted.
func (v Sensitivity) Renderable() bool {
	return v == SensitivityPublic || v == SensitivityInternal
}

// FactorKind classifies an external factor.
type FactorKind string

const (
	FactorEnabler    FactorKind = "enabler"
	FactorConstraint FactorKind = "constraint"
	FactorTrend      FactorKind = "trend"
	FactorShock      FactorKind = "shock"
)

// AllFactorKinds returns all factor kinds.
func AllFactorKinds() []FactorKind {
	return []FactorKind{FactorEnabler, FactorConstraint, FactorTrend, FactorShock}
}

// Valid reports whether v is a known FactorKind.
func (v FactorKind) Valid() bool { return slices.Contains(AllFactorKinds(), v) }

// RiskCategory classifies a risk (UNDG categories plus market and delivery).
type RiskCategory string

const (
	RiskEnvironmentalPolitical RiskCategory = "environmental_political"
	RiskOpportunity            RiskCategory = "opportunity"
	RiskDesign                 RiskCategory = "design"
	RiskPartnership            RiskCategory = "partnership"
	RiskMarket                 RiskCategory = "market"
	RiskTechnical              RiskCategory = "technical"
	RiskOperational            RiskCategory = "operational"
)

// AllRiskCategories returns all risk categories.
func AllRiskCategories() []RiskCategory {
	return []RiskCategory{
		RiskEnvironmentalPolitical, RiskOpportunity, RiskDesign, RiskPartnership,
		RiskMarket, RiskTechnical, RiskOperational,
	}
}

// Valid reports whether v is a known RiskCategory.
func (v RiskCategory) Valid() bool { return slices.Contains(AllRiskCategories(), v) }

// PathwayKind classifies what a pathway groups.
type PathwayKind string

const (
	PathwayProgram   PathwayKind = "program"
	PathwayStrategy  PathwayKind = "strategy"
	PathwayTheme     PathwayKind = "theme"
	PathwayGeography PathwayKind = "geography"
	PathwaySegment   PathwayKind = "segment"
)

// AllPathwayKinds returns all pathway kinds.
func AllPathwayKinds() []PathwayKind {
	return []PathwayKind{PathwayProgram, PathwayStrategy, PathwayTheme, PathwayGeography, PathwaySegment}
}

// Valid reports whether v is a known PathwayKind.
func (v PathwayKind) Valid() bool { return slices.Contains(AllPathwayKinds(), v) }

// Aggregation is how indicator values of one metric combine across
// pathways.
type Aggregation string

const (
	AggregationSum          Aggregation = "sum"
	AggregationMean         Aggregation = "mean"
	AggregationWeightedMean Aggregation = "weighted_mean"
	AggregationMin          Aggregation = "min"
	AggregationMax          Aggregation = "max"
	AggregationLatest       Aggregation = "latest"
	// AggregationNone means values cannot be combined, for example distinct
	// counts across overlapping populations; organization-level values must
	// be observed directly.
	AggregationNone Aggregation = "none"
)

// AllAggregations returns all aggregations.
func AllAggregations() []Aggregation {
	return []Aggregation{
		AggregationSum, AggregationMean, AggregationWeightedMean, AggregationMin,
		AggregationMax, AggregationLatest, AggregationNone,
	}
}

// Valid reports whether v is a known Aggregation.
func (v Aggregation) Valid() bool { return slices.Contains(AllAggregations(), v) }

// MetricLayer says whether a metric applies across all pathways or to
// particular ones.
type MetricLayer string

const (
	LayerCommon  MetricLayer = "common"
	LayerPathway MetricLayer = "pathway"
)

// AllMetricLayers returns all metric layers.
func AllMetricLayers() []MetricLayer { return []MetricLayer{LayerCommon, LayerPathway} }

// Valid reports whether v is a known MetricLayer.
func (v MetricLayer) Valid() bool { return slices.Contains(AllMetricLayers(), v) }

// Claim is how strongly an indicator's result may be credited to the
// initiative. When omitted, contribution is assumed, so reports never
// overclaim.
type Claim string

const (
	ClaimAttribution  Claim = "attribution"
	ClaimContribution Claim = "contribution"
)

// AllClaims returns all claims.
func AllClaims() []Claim { return []Claim{ClaimAttribution, ClaimContribution} }

// Valid reports whether v is a known Claim.
func (v Claim) Valid() bool { return slices.Contains(AllClaims(), v) }

// Framework identifies an external framework an alignment points into.
type Framework string

const (
	FrameworkOKR          Framework = "okr"
	FrameworkBSC          Framework = "bsc"
	FrameworkV2MOM        Framework = "v2mom"
	FrameworkCapability   Framework = "capability"
	FrameworkMaturitySLI  Framework = "maturity_sli"
	FrameworkMarketSpec   Framework = "market_spec"
	FrameworkUSM          Framework = "usm"
	FrameworkUFM          Framework = "ufm"
	FrameworkSDG          Framework = "sdg"
	FrameworkVSEM         Framework = "vsem"
	FrameworkVisionStudio Framework = "visionstudio"
	FrameworkCustom       Framework = "custom"
)

// AllFrameworks returns all frameworks.
func AllFrameworks() []Framework {
	return []Framework{
		FrameworkOKR, FrameworkBSC, FrameworkV2MOM, FrameworkCapability, FrameworkMaturitySLI,
		FrameworkMarketSpec, FrameworkUSM, FrameworkUFM, FrameworkSDG, FrameworkVSEM,
		FrameworkVisionStudio, FrameworkCustom,
	}
}

// Valid reports whether v is a known Framework.
func (v Framework) Valid() bool { return slices.Contains(AllFrameworks(), v) }

// Relation is how an element relates to an external element.
type Relation string

const (
	RelationSupports      Relation = "supports"
	RelationMeasures      Relation = "measures"
	RelationImplements    Relation = "implements"
	RelationContributesTo Relation = "contributes_to"
	RelationSameAs        Relation = "same_as"
)

// AllRelations returns all relations.
func AllRelations() []Relation {
	return []Relation{RelationSupports, RelationMeasures, RelationImplements, RelationContributesTo, RelationSameAs}
}

// Valid reports whether v is a known Relation.
func (v Relation) Valid() bool { return slices.Contains(AllRelations(), v) }

// RevisionTrigger is why a revision was made.
type RevisionTrigger string

const (
	TriggerLearning       RevisionTrigger = "learning"
	TriggerEvidence       RevisionTrigger = "evidence"
	TriggerContextChange  RevisionTrigger = "context_change"
	TriggerStrategyChange RevisionTrigger = "strategy_change"
	TriggerReview         RevisionTrigger = "review"
)

// AllRevisionTriggers returns all revision triggers.
func AllRevisionTriggers() []RevisionTrigger {
	return []RevisionTrigger{TriggerLearning, TriggerEvidence, TriggerContextChange, TriggerStrategyChange, TriggerReview}
}

// Valid reports whether v is a known RevisionTrigger.
func (v RevisionTrigger) Valid() bool { return slices.Contains(AllRevisionTriggers(), v) }

// QualityRating is a pass / partial / fail judgment.
type QualityRating string

const (
	QualityPass    QualityRating = "pass"
	QualityPartial QualityRating = "partial"
	QualityFail    QualityRating = "fail"
)

// AllQualityRatings returns all quality ratings.
func AllQualityRatings() []QualityRating {
	return []QualityRating{QualityPass, QualityPartial, QualityFail}
}

// Valid reports whether v is a known QualityRating.
func (v QualityRating) Valid() bool { return slices.Contains(AllQualityRatings(), v) }

// DecisionStatus is the state of a decision request.
type DecisionStatus string

const (
	DecisionOpen     DecisionStatus = "open"
	DecisionDecided  DecisionStatus = "decided"
	DecisionDeferred DecisionStatus = "deferred"
)

// AllDecisionStatuses returns all decision statuses.
func AllDecisionStatuses() []DecisionStatus {
	return []DecisionStatus{DecisionOpen, DecisionDecided, DecisionDeferred}
}

// Valid reports whether v is a known DecisionStatus.
func (v DecisionStatus) Valid() bool { return slices.Contains(AllDecisionStatuses(), v) }
