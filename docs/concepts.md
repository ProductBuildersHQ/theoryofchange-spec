# Concepts → Go Types

How each concept from the
[research brief](research/theory-of-change.md) (§8 entity model, §9.8 UNDG
implications) maps to the Go model in `pkg/impact`. The
[TRD §7](specs/initiatives/INIT-THEORYOFCHANGESPEC-001/TRD.md) is normative;
this page is the index.

The model is one impact model. The methods the research brief describes are
projections of it, drawn by the `report` command:

| Method | Projection shows | Report kinds |
|--------|------|--------------|
| Theory of change | Pathways of change, with the reasons and assumptions behind each link | `pathway` |
| Logic model | Inputs, activities, outputs, and outcomes by level, for each pathway | `logic-model`, `logic-model-layered`, `logic-model-d2` |
| Outcome map | Outcomes by level, with who is responsible for each | `outcome-map` |
| Logframe | A matrix of narrative, indicators, verification, and assumptions | `logframe` |

OKRs and outcome journeys are projections too, written by `export okr` and
`export journey` for the PRISM planning tools. V2MOM and DMAIC projections are
planned.

**Delivered by** names the roadmap item (`RMI-THEORYOFCHANGESPEC-NNN`, shown
as `NNN`) that adds the type or field. Concepts not in v0 are listed under
[Deferred](#deferred).

## Document and context

| Concept (source) | Go type / field | Delivered by |
|------------------|-----------------|--------------|
| Impact model document (holds the theory of change of ActKnowledge and UNDG) | `impact.Model` (`kind: ImpactModel`) | 004 |
| Metadata, owners, status | `Model.Metadata` (`prism-core` `Metadata`) | 004 |
| Whose impact model it is (perspective) | `Model.Perspective` (`actorId`, `role`, `futureRef`) | 004 |
| Problem statement, target population, time horizon, narrative | `Model.Context` | 004 |
| Purpose: why the organization exists and what it works toward (a statement, not a causal element) | `Context.Mission`, `Context.Vision` | 064 |
| Ceiling of accountability (ActKnowledge) | `Model.Ceiling` | 004 |
| Revision / adaptive management, quality review (plausible, feasible, testable) | `Model.Revisions[]` (`Revision`, `QualityReview`) | 004 |
| Decisions requested | `Model.Decisions[]` | 004 |
| Problem tree: immediate / underlying / structural causes (UNDG) | `Model.Causes[]` (`Cause`) | 044 |
| Target groups (UNDG equity) | `Model.TargetGroups[]` | 044 |
| Validation and consultation trail (UNDG) | `Model.Consultations[]` (`Consultation`) | 052 |
| Glossary (shared vocabulary) | `Model.Glossary[]` (`GlossaryTerm`) | 033 |

## Causal backbone

| Concept | Go type / field | Delivered by |
|---------|-----------------|--------------|
| Long-term goal, impact, outcome, precondition | `Outcome` with `Level` (`impact`, `long_term`, `medium_term`, `short_term`, `output`) | 004 |
| Sphere of control / influence / interest (Outcome Mapping, RBM) | `Outcome.Sphere` | 004 |
| AND / OR / k-of-n preconditions, ActKnowledge "cluster" | `Outcome.Logic` (`Logic{Type, K}`) and `PreconditionGroup` | 004 |
| Outcome category taxonomies (ORS, COM-B, MSD AAER) | `Outcome.Category` (open string), `Outcome.ChangeType` | 004 |
| Population that changes; co-owning actors | `Outcome.Population`, `Outcome.ActorIDs` | 004 |
| Responsibility: own / shared / partner / unaddressed (UNDG solution tree) | `Outcome.Responsibility`, `Outcome.DeliveredBy`, `Outcome.AddressesCauses` | 044 |
| Causal link, rationale | `CausalLink` (`From`, `To`, `Type`, `Rationale`) | 004 |
| Interventions placed on connectors (ActKnowledge) | `CausalLink.InterventionIDs` | 004 |
| Polarity, lag, feedback loops (reinforcing / balancing) | `CausalLink.Polarity`, `Lag`, `Feedback`, `LoopType` | 004 |
| Intervention, strategy, activity, outputs | `Intervention` (`Targets`, `Outputs`, `Status`, `Resources`) | 004 |
| Capability links (PRISM) | `Intervention.Capabilities` (`prism-core` `CapabilityRef`) | 004 |
| Pathways / program areas / products | `Model.Pathways[]` (`Pathway`, kind `program` or `product`) | 032, 064 |
| Nested ToCs (overarching vs outcome-level, UNDG) | `Model.Parent`, `Pathway.ChildModel` | 032 |

## Assumptions, risks, context, evidence

| Concept | Go type / field | Delivered by |
|---------|-----------------|--------------|
| Assumption (if X then Y because Z) | `Assumption` (`Statement`, `IfThenBecause`) | 004 |
| Assumption types: causal, contextual, implementation, behavioral, external, partnership | `Assumption.Type` | 004 |
| Criticality, plausibility, needs testing, test method (UNDG three questions) | `Assumption.Criticality`, `Plausibility`, `ExplainsFully`, `NeedsTesting`, `Test` | 004 |
| Assumption status: untested / holding / at risk / invalidated / validated | `Assumption.Status` | 004 |
| Risk (likelihood, impact, mitigation) | `Risk` (embeds `prism-core` `Risk`) | 004 |
| Risk categories (UNDG) and assumption–risk inverse links | `Risk.Category`, `Risk.RelatedAssumptionIDs` | 004 |
| External factors (enablers, constraints, trends, shocks) | `ExternalFactor` | 004 |
| Evidence, strength, contradicting evidence, process-tracing tests | `Evidence` (`Strength`, `Direction`, `ProcessTracingTest`, `Sensitivity`) | 004 |
| Sources for public-information drafts | `Evidence` (`URL`, `Excerpt`, `AccessedAt`) | 004 |
| Stated vs inferred vs confirmed (provenance) | `Provenance` embedded on every element | 052 |

## Actors

| Concept | Go type / field | Delivered by |
|---------|-----------------|--------------|
| Stakeholders, boundary partners, beneficiaries, champions, competitors | `Actor` (`Type`, `Roles`, `Stance`, `Influence`, `Interest`) | 004 |
| Contact person | `Actor.Contact` (`prism-core` `Person`) | 004 |
| Shared actor registry across documents (UFM) | `ActorRegistry` document; `Model.ActorRegistry` | 005 |

## Measurement

| Concept | Go type / field | Delivered by |
|---------|-----------------|--------------|
| Indicator (definition, unit, direction) | `Indicator` | 004 |
| Population, target, threshold, timeline (ActKnowledge) | `Indicator.Population`, `Targets[].Value`, `Targets[].Threshold`, `Targets[].Date` | 004 |
| Baseline, milestones | `Indicator.Baseline`, `Targets[].Milestone` | 004 |
| Quantitative / qualitative / progress markers (expect / like / love) | `Indicator.MeasurementType`, `Levels` | 004 |
| Data source, method, frequency, owner, disaggregation | `Indicator.DataSource`, `Method`, `Frequency`, `OwnerActorID`, `Disaggregation` | 004 |
| Indicator subject: outcome, output, or assumption | `Indicator.Measures` | 004 |
| Shared metric definitions with formulas | `Model.Metrics[]` (`Metric`); `Indicator.MetricID`, `RollupFrom` | 033 |
| Attribution vs contribution | `Indicator.Claim`, `Caveats` | 033 |
| Metric tier: activity, output, outcome, impact (input, output, and driver metrics) | `Model.IndicatorTier` (derived from what the indicator measures; never stored) | 063 |
| North Star, one metric that matters, guardrails, for a program or product | `Pathway.Focus` (`NorthStarID`, `OneMetricID`, `GuardrailIDs`, `Rationale`) | 064 |
| Headline metrics: the few numbers an audience leads with, in its order and wording | `Audience.Headline[]` (`HeadlineMetric`: `IndicatorID`, `Label`, `Note`) | 064 |
| A metric worked out from others, such as cost per outcome | `Indicator.ComposedOf` (`Composition`: `Op` ratio, product, sum, or difference, and `IndicatorIDs`) | 067 |
| Data availability | `Indicator.DataAvailability` | 052 |
| Metric selection, collection burden, budget | `SelectionCriterion`, `Indicator.Selection`, `Indicator.Collection`, `MeasurementBudget` | 034 |
| Observations (time series, data quality) | `ObservationSet.Observations[]` (`Observation`) | 005 |
| Assumption assessments over time | `ObservationSet.Assessments[]` | 005 |
| Most Significant Change stories, theme codebook | `ObservationSet.Stories[]`, `Model.Themes[]` | 035 |
| Learning reviews (MEL) | `ObservationSet.Reviews[]` | 035 |
| Computed status, diffs between periods | `Status`, `StatusDiff` | 005 (types), 017–020 (computation) |

## Alignment and integration

| Concept | Go type / field | Delivered by |
|---------|-----------------|--------------|
| Links to OKRs, Balanced Scorecard, V2MOM, capabilities, maturity SLIs, market entities, VSEM, USM, UFM, SDGs | `Alignment` with `ExternalRef` | 004 |
| Cross-repository references | `pkg/ref` (`model:<doc>#<element>`) | 003 |
| Audiences for reports and stories | `Model.Audiences[]` | 035 |

## Deferred

| Concept (source) | Reason | Reserved name |
|------------------|--------|---------------|
| Contribution claims (Mayne's Contribution Analysis) | Needs evaluation workflow beyond v0 | `contributionClaims` |
| Substantiated observed outcomes (Outcome Harvesting) | Stories cover v0; substantiation workflow later | `observedOutcomes` |
| Weighted precondition logic | `at_least` k-of-n covers v0 | — |

## Glossary

| Term | Meaning in this specification |
|------|-------------------------------|
| Impact | The highest-level change, usually above the ceiling of accountability; contribution is claimed, not achievement |
| Long-term outcome | The goal the initiative holds itself accountable for |
| Outcome | A condition that must exist for the theory to work and does not yet exist, phrased as a state, not an activity |
| Output | A direct product of an intervention, within the organization's control |
| Precondition | An outcome that must hold for a higher outcome; incoming links to an outcome are its preconditions |
| Precondition group | An explicit AND / OR / k-of-n combination of preconditions |
| Causal link | A directed claim that one outcome (or intervention) contributes to another, with a rationale |
| Intervention | Something the actors do to bring about outcomes (program, product, policy, advocacy, partnership…) |
| Assumption | A condition believed to hold for a link or outcome to work, which the initiative does not bring about |
| Indicator | Measurable evidence that an outcome, output, or assumption is changing |
| Ceiling of accountability | The line above which outcomes are tracked but not owned |
| Pathway | A named grouping of outcomes, interventions, and indicators, such as a program area |
| Provenance basis | Whether an element was stated by a source, inferred, proposed as a hypothesis, confirmed, or rejected in review |
| Observation set | A separate document of measured values, assumption assessments, stories, and learning reviews for a period |
