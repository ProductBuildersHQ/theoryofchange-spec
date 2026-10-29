# PRD — Theory of Change Specification — Go-first model, JSON Schema, measurement, and leadership reporting

**Initiative:** `INIT-THEORYOFCHANGESPEC-001`
**Repository:** `github.com/ProductBuildersHQ/theoryofchange-spec`
**Workflow:** pbhq-lite
**Status:** Draft — pending stakeholder review (pbhq-lite gate: after PRD)

Background research: [`docs/research/theory-of-change.md`](../../../research/theory-of-change.md).

## 1. Problem

A Theory of Change (ToC) states how and why a desired long-term change is
expected to happen: a long-term goal, the web of preconditions (outcomes) that
must hold for it, the interventions that produce them, the assumptions behind
every causal step, and the indicators that show whether change is happening.
The method is mature in philanthropy and international development (Aspen
Roundtable / ActKnowledge, Annie E. Casey Foundation, UNDG) and applies just as
well to corporate and market strategy — for example, driving adoption of one
of several competing standards for agentic authorization.

Today a ToC almost always lives as a diagram in a slide deck plus a narrative
document. That causes four concrete failures:

1. **Not machine-readable.** Pathways, assumptions, and indicators sit in
   pictures and prose. No tool can check that every outcome has an indicator,
   that every causal link has a stated rationale, or that the pathway actually
   reaches the long-term goal. Current state: 0 of these checks are automated
   anywhere in the ProductBuildersHQ / PRISM ecosystem.
2. **Disconnected from measurement.** Indicator baselines, targets, and
   observations are tracked in separate spreadsheets. Leadership cannot see
   "which part of our theory is working, which part is failing, and which
   assumptions have broken" without an analyst rebuilding the picture by hand
   each reporting cycle (typically several analyst-hours per report).
3. **Not composable.** The ToC cannot be joined to the strategy artifacts PRISM
   already models (OKRs, V2MOMs, capabilities, maturity SLIs, market entities),
   so it cannot feed the planned **Unified Strategy Model (USM)** — ToC combined
   with a Balanced Scorecard from one market participant's perspective — or the
   planned **Unified Futures Model (UFM)** — several competing futures, each
   championed by different market actors, measured side by side (e.g.
   competing proposals to extend or replace OAuth for agentic access, such as
   AAuth, AIMS, and Poppy).
4. **Not AI-native.** There is no schema an agent can author against, no
   deterministic validator to keep an agent honest, and no rubric to judge ToC
   quality, so LLM help with ToCs produces unverifiable prose.
5. **Not sustainable for small organizations.** Small-staff organizations —
   especially intermediaries such as nonprofit networks, associations, and
   collective-impact backbones — run several program areas, serve member
   organizations that in turn serve end beneficiaries, and have a few staff
   hours a week for measurement. Their impact evidence is mostly anecdotal and
   lives in multi-tab spreadsheets with heavy manual entry. They need to pick
   a small metric set (typically 8–12) that is consistent across program
   areas, cheap to collect (passive data first, minimal surveys), honest about
   contribution versus attribution, and directly usable for audience-specific
   storytelling and grow / adjust / sunset program decisions. No current tool
   connects the ToC, the metric definitions, the collection burden, the
   stories, and the reports.

**Who is affected:**

- *Strategy owners* (product/portfolio leaders, program officers) who author
  and revise ToCs.
- *Executives and funders* who need a short, trustworthy read-out of progress.
- *Analysts / evaluators* who maintain indicators and test assumptions.
- *AI agents* (Claude Code, MCP clients) that draft, critique, and update ToCs.
- *Downstream model builders* (USM, UFM) that need a ToC building block.
- *Measurement leads at small-staff intermediary organizations* who must
  define a unified metric framework across program areas within a fixed
  staff-time budget.
- *Non-technical analysts* who work in spreadsheets and run repeatable steps
  through an AI assistant rather than a CLI.
- *Communications staff* who turn metrics and stories into audience-specific
  narratives and annual impact reports.
- *Consultants and facilitators* who run ToC workshops and need to produce
  the diagram, glossary, logic models, and metric set quickly.

**Desired state:** a ToC is a versioned JSON (or YAML) document conforming to a
published JSON Schema generated from Go structs. Go tools validate it, analyze
its causal graph, roll up indicator observations into pathway status, judge its
quality against a rubric, and render standard and bespoke reports — executive
summary, pathway diagram, results framework, assumption register — in seconds.

## 2. Goals and Non-Goals

### 2.1 Goals

| ID | Goal | Success criterion |
|----|------|-------------------|
| G-1 | Capture the canonical ToC vocabulary as a Go-first data model | Every core concept in the research brief's entity model (§8) maps to a Go type and a JSON Schema definition; mapping table published in docs |
| G-2 | Publish a lint-clean, embedded JSON Schema | Schemas generated from Go types only; `schemakit lint --property-case camelCase` reports 0 violations; schemas embedded via `//go:embed` |
| G-3 | Make ToC quality checkable without an LLM | ≥ 15 deterministic lint rules covering structure, causal logic, and measurability, each with a stable rule ID and unit test |
| G-4 | Turn observations into pathway status | Given a ToC and an observation set, compute indicator, outcome, assumption, and pathway status deterministically (same inputs → byte-identical output) |
| G-5 | Generate leadership-ready reports | Executive summary, pathway diagram (Mermaid), results framework, and assumption register render from one command in < 1 s for a 200-node ToC |
| G-6 | Make ToC authoring and review AI-native | ToC spec template + structured-evaluation rubric for LLM-as-a-Judge; MCP server exposing validate / lint / status / report tools |
| G-7 | Reuse PRISM rather than duplicate it | Shared primitives come from `prism-core`; OKR and outcome-journey exports go through `prism-roadmap` types; no parallel Person / Risk / health types |
| G-8 | Be a stable building block for USM and UFM | A published integration contract (typed references, perspective, alignment links, actor registry); a test fixture composes three competing-future ToCs with a shared actor registry without schema changes |
| G-9 | Model program portfolios and two-layer (intermediary) impact | A reference ToC with ≥ 3 program pathways, an organization-level impact, member-organization and end-beneficiary layers, and contribution-labelled downstream indicators validates with 0 schema changes |
| G-10 | Make measurement sustainable | Shared metric definitions with fixed formulas; metric-selection scoring; per-indicator collection mode and staff-minute cost; budget checks warn when the selected set exceeds the declared staff-time or respondent-question budget |
| G-11 | Connect measurement to story and learning | Audiences, a story theme codebook, Most-Significant-Change-style stories, and learning reviews are modelled; audience-specific story, learning-review, and portfolio-decision reports render from data |
| G-12 | Produce workshop-ready definition artifacts before any data exists | Pathway diagram, per-pathway logic models, glossary, metric-selection matrix, results-framework definitions, and collection plan render from a ToC document alone |
| G-13 | Work where non-technical staff work | CSV import of observations and CSV export of definition tables; packaged agent skills let an AI assistant run each step through the deterministic tools |
| G-14 | Change the theory safely through the CLI, SDK, and MCP | Every edit goes through one patch engine that applies atomically and rejects edits introducing validation errors; ergonomic `theoryofchange <element> add/set/rm` commands; canonical formatting for readable diffs; semantic theory diffs that draft revision-log entries |
| G-15 | Produce presentation-ready artifacts | Marp slide decks (workshop, executive, learning review, audience story) generated from the same report view as the Markdown reports, with optional SVG / PDF / PPTX export |
| G-17 | Draft a credible ToC from public sources, honestly labelled | Every element carries a basis (`stated`, `inferred`, `hypothesis`, `confirmed`, `rejected`) and source citations; reports and decks mark inferred elements; an assessment report lists gaps and questions for the organization |
| G-18 | Run a review-and-agreement round trip with people who never touch the tools | Review packet (DOCX via Pandoc) and review sheet (CSV) go out with stable element IDs and a version stamp; responses come back as a `ReviewResponse` and become proposed patches plus a consultation record; `theoryofchange approve` produces the agreed baseline version |
| G-16 | Make LLM-as-a-Judge results reproducible and trustworthy | Judge bundles and verdicts are files alongside the ToC; multi-judge runs record inter-rater agreement; lint profiles keep early drafts from drowning in warnings |

### 2.2 Non-Goals

- **Building USM or UFM.** This initiative defines the ToC building block and
  its integration contract only. Balanced Scorecard modeling, futures
  comparison, and market-share forecasting belong to those initiatives.
- **A diagram editor or web UI.** Output is files (Markdown, Mermaid, JSON,
  HTML via existing PRISM/site tooling later). Visual editing is out of scope.
- **Causal inference or statistics.** We record evidence strength and compute
  rule-based confidence; we do not estimate effect sizes, run counterfactuals,
  or claim attribution.
- **Data collection.** Observations are supplied as files (JSON, YAML, or
  CSV); connectors to CRMs, form tools, analytics systems, survey tools, or
  GitHub are out of scope. Record linkage (matching sign-ups to contacts and
  organizations) and survey administration happen upstream; this module
  consumes the resulting aggregate observations.
- **Automated theme coding.** The module defines the theme codebook and
  validates coded stories; assigning themes to free text (by people or an
  LLM) happens outside it.
- **Google (or any cloud) API dependencies inside this module.** Sharing to
  Google Drive and reading responses back use external tooling invoked by an
  agent skill; this module produces and consumes files only.
- **LLM calls inside this module.** The module ships templates, rubrics, and
  prompts; judging is executed by visionspec / visionstudio or an MCP client.
  No model SDKs become dependencies of this repo.
- **Persistence / multi-user storage.** No database. Documents live in git.
- **Logframe-only workflows.** The canonical model is the ToC outcomes
  framework. Per-pathway logic-model *views* are projections of it
  (PRD-FR-505); logframe matrices may come later.

## 3. Personas and User Stories

Acceptance criteria are written to be testable by unit tests, golden files, or
CLI exit codes.

**US-1 — Author a ToC as data.**
As a *strategy owner*, I want to write my ToC in YAML or JSON against a
published schema, so that editors give me completion and errors as I type.

- Given a document with `$schema` set to the published schema URL, VS Code /
  any JSON-Schema-aware editor validates it (manual check, documented).
- `theoryofchange validate my.toc.yaml` exits 0 for a valid document and non-zero with
  JSON-pointer locations for each schema or reference error.

**US-2 — Catch weak theories early.**
As a *strategy owner*, I want an automated critique of my ToC's logic, so that
gaps (outcomes with no indicator, links with no rationale, pathways that never
reach the goal) are found before review.

- `theoryofchange lint` reports each finding with rule ID, severity, element ID, and a
  fix hint; output is a structured-evaluation GO/NO-GO summary report in JSON
  or a human table.
- A seeded-defect fixture triggers every lint rule exactly once.

**US-3 — See which part of the theory is working.**
As an *executive*, I want a one-page summary of progress toward the long-term
goal, so that I know where the theory is holding, where it is failing, and
what decision is needed.

- `theoryofchange report exec` produces Markdown containing: overall status, pathway
  status by level, top at-risk outcomes, broken or untested critical
  assumptions, changes since the previous period, and decisions requested.
- Every status in the report traces to an indicator observation or an
  assumption assessment ID (no unsupported claims).

**US-4 — Update measurements without editing the theory.**
As an *analyst*, I want to record observations in a separate file per period,
so that the slow-changing theory and the fast-changing measurements have
independent git histories.

- Observation sets reference indicators by ID; `theoryofchange status` merges any number
  of observation files; references to unknown indicators are errors.

**US-5 — Test assumptions explicitly.**
As an *evaluator*, I want each critical assumption to carry its test, its
evidence, and a current assessment, so that broken assumptions automatically
flag every downstream outcome as at risk.

- Marking an assumption `invalidated` changes the pathway status of every
  outcome that depends on the links it underpins, verified by unit test.

**US-6 — Let an agent do the drafting.**
As a *strategy owner working with an AI agent*, I want the agent to draft and
revise my ToC through tools that validate every edit, so that agent output is
always schema-valid and lint-checked.

- MCP server exposes `toc_validate`, `toc_lint`, `toc_graph`, `toc_status`,
  `toc_report`, and `toc_schema`; each returns structured JSON; an MCP
  integration test drives every tool.
- A ToC spec template and rubric are published so visionspec / visionstudio
  can run LLM-as-a-Judge on a ToC narrative.

**US-7 — Connect the theory to plans.**
As a *portfolio leader*, I want ToC outcomes and indicators exported as OKRs
and interventions linked to PRISM capabilities, so that execution plans trace
back to the theory.

- `theoryofchange export okr` emits a `prism-roadmap` OKR document that passes
  prism-roadmap's own validation.
- Interventions reference capabilities with `prism-core` `CapabilityRef`.

**US-8 — Compare competing futures (contract only).**
As a *UFM model builder*, I want several ToCs — each describing one competing
future from its champions' perspective — to share an actor registry and be
referenced by stable IDs, so that UFM can compare them without each ToC
knowing about the others.

- Fixture: three ToCs for three competing approaches share one actor
  registry; all three validate; cross-document references resolve via the
  published reference format.

**US-9 — One theory, several program areas.**
As a *measurement lead at an intermediary organization*, I want one
organization-level ToC with a pathway (and logic model) per program area, so
that each program keeps its own outcomes while everything rolls up to the
organization's impact.

- Pathways group outcomes, interventions, and indicators; reports filter by
  `--pathway`; organization-level indicators roll up from pathway indicators
  where the metric's roll-up rule allows it.
- A program may instead be its own ToC document whose outcomes contribute to
  outcomes in a parent ToC via references; both styles validate.

**US-10 — Honest about two layers of impact.**
As an *intermediary*, I want indicators for changes in member organizations
marked as attributable and indicators for end beneficiaries marked as
contribution, so that reports never overclaim.

- Every indicator has a `claim` of `attribution` or `contribution`; reports
  label contribution indicators; lint warns when an indicator above the
  ceiling of accountability claims attribution.

**US-11 — Choose a small, sustainable metric set.**
As a *measurement lead*, I want to score candidate metrics against
selection criteria and see the total collection burden of the selected set,
so that we measure what matters within our staff-time budget.

- Candidate indicators carry criterion scores and a selection status;
  `theoryofchange report metric-selection` ranks them; `theoryofchange report collection-plan`
  totals staff minutes per period and survey questions per respondent per
  year against the declared budget; lint warns on budget overrun.

**US-12 — Workshop artifacts on day one.**
As a *facilitator*, I want the pathway diagram, logic models, glossary,
metric definitions, and collection plan generated from the draft ToC during
and right after a workshop, so that the group reviews the same artifacts the
system will later measure against.

- All definition reports render with zero observations; CSV export of each
  definition table opens in a spreadsheet.

**US-13 — Data from spreadsheets, stories from people.**
As a *non-technical analyst*, I want to load monthly numbers from a CSV and
record coded stories, so that I never hand-edit JSON.

- `theoryofchange import observations --csv` maps columns to indicators and periods and
  reports unmatched rows; stories carry themes from the document's codebook,
  consent and sensitivity; theme counts become qualitative indicator values.

**US-14 — Learn and tell the story.**
As a *leadership team*, I want a quarterly learning-review draft, audience-
specific impact stories, and a portfolio view of each program's progress,
burden, and confidence, so that we learn, communicate, and decide to grow,
adjust, or sunset programs from the same data.

- `theoryofchange report learning-review`, `theoryofchange report story --audience <id>`, and
  `theoryofchange report portfolio` render from the ToC, observations, and stories;
  every number cites its source; restricted stories are never quoted.

**US-15 — An assistant runs the routine.**
As a *non-technical analyst using an AI assistant*, I want packaged agent
skills that call the deterministic tools in a fixed order with fixed
definitions, so that monthly results are identical no matter which run or
model produced them.

- Skills for authoring a ToC, refreshing status from CSV, drafting a learning
  review, and drafting audience stories; each skill states its inputs,
  commands, checks (row counts, unmatched rows, change versus previous run),
  and outputs.

**US-16 — Edit the theory without hand-editing JSON.**
As a *strategy owner or AI agent*, I want to add, change, and remove
outcomes, links, assumptions, indicators, and actors with commands or small
patches that are validated before they are written, so that the document is
never left in a broken state and every change is a readable diff.

- `theoryofchange patch` applies an RFC 6902 JSON Patch atomically; if the result has
  more `error`-severity validation issues than the input, nothing is written
  and the new issues are reported.
- `theoryofchange outcome add --id … --title … --level …`, `theoryofchange link add --from … --to …`,
  and siblings for every element kind produce the same result as the
  equivalent patch.
- `theoryofchange fmt` rewrites a document in canonical form; running it twice is a
  no-op.
- `theoryofchange diff old new` lists added / removed / changed elements by ID and can
  emit a draft `Revision` entry.

**US-17 — Present it.**
As a *facilitator or executive sponsor*, I want slide decks generated from
the ToC and its status, so that workshops, board meetings, and funder
briefings use the same numbers as the reports.

- `theoryofchange deck workshop|exec|learning-review|story` writes Marp Markdown; with
  the optional Marp and Mermaid CLIs installed, `--export pdf|pptx|html`
  produces a finished file, and diagrams render to SVG.

**US-18 — Judge quality reproducibly.**
As a *reviewer*, I want the judge input and each judge's verdict stored as
files, and agreement across judges reported, so that a rubric score is
auditable rather than a one-off chat answer.

- `theoryofchange judge show` emits the judge bundle (rubric, ToC JSON, generated
  narrative, lint summary); `theoryofchange judge record` validates and stores a
  structured-evaluation verdict; `theoryofchange judge summary` aggregates several
  verdicts with agreement metrics.

**US-19 — Draft from public sources.**
As a *consultant or analyst*, I want an assistant to draft an organization's
ToC from its public website and reports, with every element labelled as
stated by the organization or inferred by us and cited to its source, so
that the draft is a credible, honest conversation starter.

- `theoryofchange evidence add --url … --title … --excerpt … --accessed …` records
  sources; the `toc-draft-from-sources` skill creates evidence records and
  builds the draft only through validated patches; lint reports elements
  without a basis and `stated` elements without a citation.
- `theoryofchange report assessment` lists stated vs inferred elements, outcomes with no
  indicator, indicators whose data is not available, assumptions needing
  testing, and generated questions for the organization.

**US-20 — Review without touching the tools.**
As a *staff member at the organization*, I want to review the draft as a
familiar document or spreadsheet in Google Drive and mark each element agree
/ change / remove / unsure with comments, so that my feedback reaches the
model without learning a new tool.

- `theoryofchange export review-packet` produces Markdown that Pandoc converts to DOCX,
  with one ID-tagged table per element kind and empty response columns;
  `theoryofchange export review-sheet` produces the same rows as CSV.
- Every exported artifact carries the ToC ID, version, and content hash.
- Responses read back (by a person or an assistant using Google tooling
  outside this module) are normalized into a `ReviewResponse` file;
  `theoryofchange import review` turns it into a proposed patch and a consultation
  record. Nothing is applied until someone runs `theoryofchange patch` on the proposal.
  Responses against a different content hash are flagged.

**US-21 — Agree on the baseline.**
As an *organization's leadership*, I want to approve the reviewed ToC as a
named version, so that both teams work from the same agreed theory.

- `theoryofchange approve --version v1.0 --approver …` sets status `approved`, records
  approvers and a `Revision`, and refuses while any `error` lint finding or
  unresolved `hypothesis` element remains at an accountable level unless
  `--force` with a reason.
- `theoryofchange diff hypothesis.json v1.0.json` shows what changed through review,
  including rejected elements.

## 4. Functional Requirements

### 4.1 Model (PRD-FR-1xx)

- **PRD-FR-101** The model MUST represent a ToC document with: identity and
  metadata, perspective (the actor whose theory it is), scope and context
  (problem statement, root causes, target population or market, time horizon),
  one or more impact-level goals, and a ceiling of accountability.
- **PRD-FR-102** The model MUST represent outcomes as graph nodes with a level
  (`impact`, `long_term`, `intermediate`, `short_term`, `output`), a sphere
  (`control`, `influence`, `interest`), an open category taxonomy (e.g. ORS
  impact / influence / leverage, COM-B, market-systems Adopt–Adapt–Expand–
  Respond), and a precondition logic (`all`, `any`, or `at_least` k-of-n) over
  their incoming links, with explicit, nestable precondition groups for
  combinations such as `(A AND B) OR C`.
- **PRD-FR-103** The model MUST represent causal links as directed edges with
  type, rationale, linked assumptions, linked evidence, interventions placed
  on the link, and an optional `feedback` flag (reinforcing / balancing) that
  permits intentional loops such as network effects.
- **PRD-FR-104** The model MUST represent interventions (strategies /
  activities) with the outcomes they target, their outputs, owning actor,
  status, and optional capability and initiative references.
- **PRD-FR-105** The model MUST represent assumptions with type (`causal`,
  `contextual`, `implementation`, `behavioral`, `external`, `partnership`),
  an optional structured *if / then / because* statement, criticality,
  plausibility, whether testing is needed, the elements they apply to, how
  they will be tested, and a current status (the UNDG three questions:
  complete? plausible? needs testing?).
- **PRD-FR-106** The model MUST represent indicators attached to outcomes,
  intervention outputs, or assumptions, carrying ActKnowledge's four
  components (population, target, threshold, timeline) plus
  definition, unit, direction, baseline, time-phased targets / milestones,
  data source, collection method, frequency, owner, population, and
  measurement type (quantitative, qualitative scale, or hybrid).
- **PRD-FR-107** The model MUST represent actors (stakeholders, boundary
  partners, beneficiaries, champions, competitors) with type, role, stance,
  influence, and interest.
- **PRD-FR-108** The model MUST represent evidence (source, type, strength,
  sensitivity) and external factors / risks, reusing `prism-core` types where
  they exist; risks carry a category (`environmental_political`,
  `opportunity`, `design`, `partnership`, `market`, `technical`,
  `operational`) and the assumptions they mirror.
- **PRD-FR-109** The model MUST represent alignments from any element to an
  external framework element (OKR, Balanced Scorecard objective, V2MOM
  method, PRISM capability, maturity SLI, market-spec entity, USM, UFM,
  SDG, custom).
- **PRD-FR-110** The model MUST record a revision log (version, summary,
  trigger) to support adaptive management.
- **PRD-FR-111** Observations MUST live in a separate observation-set document
  type referencing indicators and assumptions by ID, holding quantitative
  values, qualitative levels, assumption assessments, timestamps, sources, and
  notes.

- **PRD-FR-112** The model MUST represent pathways (named groupings such as
  program areas or strategies) over outcomes, interventions, and indicators,
  and MAY reference a parent ToC whose outcomes this document's outcomes
  contribute to.
- **PRD-FR-113** The model MUST represent shared metric definitions (name,
  definition, exact formula, unit, aggregation / roll-up rule) that
  indicators instantiate per pathway, so one metric is computed identically
  everywhere.
- **PRD-FR-114** Every indicator has a claim (`attribution` or
  `contribution`; `contribution` when omitted, so reports never overclaim)
  and MAY declare caveats (e.g. "based on sign-ups, not attendance") that
  reports must render.
- **PRD-FR-115** The model MUST represent metric selection (configurable
  criteria, per-indicator scores, status `candidate` / `selected` /
  `retired`, rationale) and collection cost (mode: `passive`, `form`,
  `survey`, `interview`, `manual`, `derived`; source system; staff minutes per
  period; questions per respondent), plus a document-level measurement budget.
- **PRD-FR-116** The model MUST represent audiences (who, what they need to
  see) and allow outcomes, indicators, and stories to target audiences.
- **PRD-FR-117** The model MUST represent a glossary and a story theme
  codebook (theme ID, label, definition, examples, related outcomes).
- **PRD-FR-118** Observation sets MUST additionally carry stories (summary,
  optional quote, date, actor, domain of change, themes, related outcomes,
  consent, sensitivity), learning reviews (period, questions, findings,
  decisions, actions), and per-observation data quality (`verified`,
  `reported`, `estimated`).

- **PRD-FR-119** The model MUST represent a problem tree: causes with level
  (`immediate`, `underlying`, `structural`), parent cause, and affected
  groups; outcomes MUST be able to declare which causes they address, so the
  solution tree can be checked against the problem tree at every level.
- **PRD-FR-120** Outcomes MUST declare responsibility (`own`, `shared`,
  `partner`, `unaddressed`) and the actors delivering them. All outcomes are
  monitored; only `own` and `shared` outcomes count toward the perspective
  actor's accountability.
- **PRD-FR-121** Prioritization criteria (e.g. equity, comparative
  advantage, feasibility, evidence of what works) MUST be scorable on
  outcomes and interventions as well as indicators, with the subject kind
  declared on each criterion.
- **PRD-FR-122** The model MUST represent target groups (priority
  populations) in the context and let indicators declare the target groups
  they disaggregate by.
- **PRD-FR-123** The model MUST record consultations / validations: date,
  ToC step validated (`focus`, `change_analysis`, `assumptions_risks`,
  `actors`, `measurement`, `whole`), method (workshop, interview, survey,
  document review, evidence review), stakeholder groups (not individuals),
  findings, and resulting changes.

- **PRD-FR-124** Every element (outcome, group, link, intervention,
  assumption, indicator, metric, actor, cause, pathway, risk, external factor)
  MUST support provenance: basis (`stated`, `inferred`, `hypothesis`,
  `confirmed`, `rejected`), cited evidence IDs, and a note. Rejected elements
  remain in the document (excluded from graph analysis, measurement, and
  default reports) so the review history stays visible.
- **PRD-FR-125** Indicators MUST declare data availability: `public`,
  `internal_existing`, `internal_new`, `partner_held`, or `unknown`.

### 4.2 Schema (PRD-FR-2xx)

- **PRD-FR-201** JSON Schemas MUST be generated from the Go types; no
  hand-written schema files.
- **PRD-FR-202** Generated schemas MUST be embedded and retrievable by name at
  runtime and via `theoryofchange schema <name>`.
- **PRD-FR-203** Documents MAY be authored in YAML; YAML MUST be converted to
  JSON before schema validation so both formats validate identically.
- **PRD-FR-204** Observations MUST be importable from CSV through a declared
  column mapping, with unmatched rows reported (never dropped silently);
  definition tables (indicators, metrics, selection, collection plan,
  glossary) MUST be exportable as CSV.

### 4.3 Validation, graph analysis, and lint (PRD-FR-3xx)

- **PRD-FR-301** Validation MUST check schema conformance, ID uniqueness, ID
  format, and resolution of every internal reference.
- **PRD-FR-302** Graph analysis MUST compute: upstream/downstream sets,
  every path from an element to the impact goal(s), topological levels,
  unintended cycles (excluding `feedback` links), orphans, and dead ends.
- **PRD-FR-303** Lint MUST implement deterministic quality rules derived from
  the "plausible, doable, testable" criteria and the research brief, each with
  a stable ID (`TOC-<AREA>-<NNN>`), severity, and fix hint.
- **PRD-FR-304** Lint output MUST be a structured-evaluation summary report
  (GO / NO-GO) so it plugs into existing ecosystem gates.

### 4.4 Measurement (PRD-FR-4xx)

- **PRD-FR-401** Indicator progress MUST be computed direction-aware from
  baseline, current value, and target, with expected progress interpolated
  from time-phased targets at an `asOf` date.
- **PRD-FR-402** Indicator status MUST be one of `achieved`, `on_track`,
  `at_risk`, `off_track`, `not_started`, `no_data`, or `stale` (no observation
  within the indicator's frequency window), mapped to `prism-core` health.
- **PRD-FR-403** Outcome status MUST roll up from its own indicators and,
  separately, a pathway status MUST roll up from preconditions using the
  outcome's precondition logic.
- **PRD-FR-404** Assumption status MUST propagate: an `invalidated` or
  `at_risk` assumption marks the links it underpins and every downstream
  outcome's pathway status accordingly.
- **PRD-FR-405** Each causal link and each path to impact MUST receive a
  rule-based confidence score from evidence strength and assumption status,
  documented and reproducible.
- **PRD-FR-406** The computed status snapshot MUST itself be a schema-defined
  JSON document, diffable across periods ("what changed since last period").

### 4.5 Reporting (PRD-FR-5xx)

- **PRD-FR-501** Standard reports MUST include: executive summary, pathway
  diagram (Mermaid flowchart with status styling), results framework table,
  assumption register, intervention status, and evidence-gap list.
- **PRD-FR-502** Bespoke reports MUST be possible through user-supplied Go
  `text/template` files over a documented report view model.
- **PRD-FR-503** Evidence marked `restricted` MUST NOT be rendered verbatim in
  any report; it is cited by ID with an access note.
- **PRD-FR-504** Reports MUST be deterministic for fixed inputs and `asOf`.
- **PRD-FR-505** Definition reports MUST render from a ToC document with no
  observations: per-pathway logic model (inputs → activities → outputs →
  outcomes → impact projection), results-framework definitions, metric
  selection matrix, collection plan with budget totals, and glossary.
- **PRD-FR-506** An audience story report MUST combine the indicators and
  stories tagged for one audience into a narrative scaffold with every number
  cited; stories without consent or marked `restricted` MUST NOT be quoted.
- **PRD-FR-507** A learning-review report MUST draft the period's review:
  status changes, surprises (largest deviations from expected progress),
  assumption changes, theme counts, and open questions.
- **PRD-FR-508** A portfolio report MUST show, per pathway, pathway status,
  confidence, collection burden, and audience coverage, as input to grow /
  adjust / sunset decisions (the report presents evidence; it does not
  recommend).
- **PRD-FR-510** A narrative report MUST scaffold a programme / strategy
  document from the ToC: causal analysis (problem tree and how outcomes
  address it), priorities and rationale (prioritization scores), assumptions
  and risks with mitigation, partner-delivered and unaddressed elements,
  implementation arrangements (who does what), M&E provisions (assumptions
  needing testing, monitoring cadence, evaluation questions from evidence
  gaps), and key messages.
- **PRD-FR-511** A division-of-labour report MUST show actors × outcomes /
  interventions with roles, flagging outcomes needing collaboration (≥ 2
  delivering actors) and outcomes with no delivering actor.
- **PRD-FR-512** An assessment report MUST render from the document alone:
  provenance mix (counts and lists of stated / inferred / hypothesis /
  confirmed / rejected), outcomes without indicators, indicators by data
  availability, assumptions needing testing, open decisions, and generated
  questions for the organization (one per inferred or hypothesis element at
  an accountable level, plus one per gap), with a short definitions-only
  narrative summary.
- **PRD-FR-509** Indicators with a baseline and no target MUST report
  `baseline_only` rather than a failure, so a first measurement year is
  presented as establishing a baseline.

### 4.6 AI-native and ecosystem integration (PRD-FR-6xx)

- **PRD-FR-601** The module MUST publish a ToC spec template (Markdown) and a
  structured-evaluation `RubricSet` for LLM-as-a-Judge review, embedded and
  syncable into specification-workflow-spec. The core rubric MUST cover the
  sector-neutral items of the UNDG quality-assurance checklist (participatory
  process, logical flow without leaps of faith, root causes addressed,
  explicit assumptions and risks, who does what, evidence and its limits,
  changes specified for target groups, diagram plus narrative); an optional
  `development` rubric profile MUST add the sector-specific items (gender,
  leaving no one behind, resilience to shocks, poverty–environment links,
  structural causes of violence).
- **PRD-FR-602** An MCP server (official Go SDK) MUST expose validate, lint,
  graph, status, report, and schema tools over stdio.
- **PRD-FR-603** Export to `prism-roadmap` OKR documents and outcome journeys
  MUST be provided in an adapter package that the core model does not import.
- **PRD-FR-604** A published reference format MUST identify a ToC document and
  any element within it across repositories, for use by USM and UFM.
- **PRD-FR-605** The repository MUST ship agent skills (instruction files plus
  helper templates) for: authoring / revising a ToC, refreshing status from
  CSV, drafting a learning review, and drafting audience stories. Skills MUST
  delegate every computation to the deterministic CLI or MCP tools.
- **PRD-FR-606** The repository MUST ship a `toc-draft-from-sources` skill:
  given source URLs or documents, it records evidence with excerpts, drafts
  elements through patches with provenance, and renders the workshop pack and
  assessment report.
- **PRD-FR-607** The repository MUST ship a `toc-share-google` skill that
  exports the review packet, review sheet, and decks, publishes them with
  external Google tooling (an SDK CLI or an MCP server, not a dependency of
  this module), reads responses back (response columns, comments, and
  suggestions), writes a `ReviewResponse`, and runs `theoryofchange import review`. The
  skill MUST also document a manual fallback (download/upload by hand, CSV
  review sheet).

### 4.7 Editing (PRD-FR-7xx)

- **PRD-FR-701** All document mutations (CLI, SDK, MCP) MUST go through one
  patch engine applying RFC 6902 operations atomically, with element
  addressing by ID (e.g. `/outcomes/@oc-adoption/title`) as well as by index.
- **PRD-FR-702** A patch MUST be rejected (nothing written) when it increases
  the number of `error`-severity validation issues, unless `--allow-errors`
  is given; the response lists issues introduced and resolved.
- **PRD-FR-703** Element commands (`add`, `set`, `rm`, plus `link`/`unlink`
  where relevant) MUST exist for outcomes, groups, links, interventions,
  assumptions, indicators, metrics, pathways, actors, evidence, causes, and
  glossary terms; removing an element MUST report (and with `--cascade`,
  remove) references to it.
- **PRD-FR-704** `theoryofchange fmt` MUST produce a canonical serialization (field
  order, indentation, element order preserved unless `--sort`), idempotent,
  for JSON and YAML.
- **PRD-FR-705** `theoryofchange diff` MUST report a semantic diff between two ToC
  versions keyed by element ID and MAY emit a draft `Revision` entry.
- **PRD-FR-706** The MCP server MUST expose `toc_patch` using the same engine.
- **PRD-FR-707** `theoryofchange approve` MUST be the only path to status `approved`:
  it records approvers (prism-core `Approver`), a `Revision` with trigger
  `review`, and the version label, and enforces the checks in US-21.

### 4.9 Sharing and review round trip (PRD-FR-9xx)

- **PRD-FR-901** A `pandoc` Markdown profile MUST be available for every
  report and the review packet: diagrams pre-rendered to PNG, pipe tables
  without nested formatting, a YAML metadata block (title, version, date),
  and explicit page breaks; `theoryofchange export docx|pptx` MUST wrap an optional
  external `pandoc` with an optional reference template, from the CLI only.
- **PRD-FR-902** Every exported artifact (Markdown, DOCX, PPTX, CSV, deck)
  MUST carry an export stamp: ToC ID, version label, content hash (SHA-256
  of canonical JSON), and as-of date.
- **PRD-FR-903** `theoryofchange export review-packet` and `theoryofchange export review-sheet`
  MUST list elements by kind with stable element ID, statement, basis,
  sources, and empty response columns `decision` (agree / change / remove /
  unsure), `suggested_change`, and `comment`.
- **PRD-FR-904** A schema-defined `ReviewResponse` MUST capture: ToC ID,
  version, content hash, reviewer group (not individual names required),
  method, and rows (element ID, decision, suggested change, comment, origin:
  `sheet`, `table`, `comment`, `suggestion`, plus quoted anchor text).
- **PRD-FR-905** `theoryofchange import review` MUST validate a `ReviewResponse`,
  produce a proposed patch (agree → basis `confirmed`; remove → basis
  `rejected`; change → field edits when the suggested change is structured,
  otherwise a note flagged for human decision), and a `Consultation` record;
  it MUST NOT modify the document, and MUST flag responses whose content hash
  differs from the current document.

### 4.8 Presentations and judging (PRD-FR-8xx)

- **PRD-FR-801** Slide decks MUST be generated as Marp Markdown from the
  report view model: `workshop` (problem, pathway, logic models, assumptions,
  metric definitions, glossary, open questions), `exec`, `learning-review`,
  and `story` (per audience).
- **PRD-FR-802** Diagram and deck export (Mermaid → SVG, Marp → PDF / PPTX /
  HTML) MUST be optional and performed by external CLIs invoked from the
  `theoryofchange` command only; the library emits text and stays network- and
  binary-free.
- **PRD-FR-803** `theoryofchange judge show | record | summary` MUST produce the judge
  bundle, validate and store structured-evaluation verdicts beside the
  document, and aggregate multiple verdicts with inter-rater agreement.
- **PRD-FR-804** Lint MUST support profiles: `draft` (errors and structural
  rules only), `standard` (default), and `full` (all rules including info),
  so a minimal valid document — goal, one outcome, one link, one indicator —
  lints GO under `draft`.

## 5. Success Metrics (required)

New capability, so most baselines are zero. Each metric is checked in CI or
by a documented, repeatable procedure.

| Metric | Baseline | Target | Measurement Method |
|--------|----------|--------|--------------------|
| Research-brief entity concepts mapped to Go types | 0 % | 100 % of core concepts | Concept-to-type mapping table in docs; reviewed at Phase 1 exit |
| Schema lint violations | n/a (no schema) | 0 | `schemakit lint --property-case camelCase schema/*.schema.json` in CI |
| Deterministic lint rules with unit tests | 0 | ≥ 15 rules, 100 % tested | Rule registry test asserts every rule ID has a triggering fixture |
| Seeded-defect detection | 0 % | 100 % of seeded defects detected; 0 findings on the clean reference example | Golden test over `testdata/defects/` and `examples/` |
| Test coverage of library packages (`pkg/...`) | 0 % | ≥ 80 % | `go test -cover ./pkg/...`; coverage badge |
| Time to produce an executive status report from data | Several analyst-hours per cycle (manual) | < 1 s compute, one command | Benchmark `BenchmarkExecReport200Nodes`; CLI timing in docs |
| Report determinism | n/a | 100 % byte-identical across runs | Golden-file tests run twice with the same `asOf` |
| LLM-judge rubric discrimination | n/a | Reference example scores `pass`; seeded weak ToC scores `fail` on ≥ 3 categories | `visionstudio spec judge` / visionspec eval run recorded in `docs/evals/` |
| MCP tool coverage | 0 | 6 of 6 tools exercised by integration test | `go test ./pkg/mcp/...` |
| PRISM export validity | n/a | 100 % of exported OKR docs pass prism-roadmap validation | Adapter test runs prism-roadmap's validator on exports |
| UFM composability | n/a | 3 competing-future ToCs + shared actor registry validate with 0 schema changes | `examples/futures/` fixture test |
| Intermediary portfolio fit | n/a | Reference intermediary ToC (≥ 3 pathways, two impact layers, 8–12 selected metrics, budget, audiences, codebook) validates and lints clean with 0 schema changes | `examples/intermediary/` fixture test |
| Definition-artifact pack from a document alone | Manual (slides + spreadsheet, hours per revision) | 6 definition reports + CSV exports from one command, 0 observations required | Golden tests over the intermediary example |
| CSV import fidelity | n/a | 100 % of rows either imported or listed as unmatched; 0 silent drops | Import tests with seeded bad rows |
| Collection-burden visibility | Unknown (not tracked) | Staff minutes / period and questions / respondent / year computed for 100 % of selected indicators that declare collection cost; budget overrun flagged | Collection-plan golden test; lint rule test |
| Agent-skill determinism | n/a | Two runs of each skill's command sequence on the same inputs produce byte-identical outputs | Skill smoke test runs the documented commands twice |
| Edit safety | n/a | 100 % of error-introducing patches rejected with nothing written; command edits equal their patch equivalents | Patch engine tests; command ↔ patch equivalence tests |
| Canonical formatting | n/a | `theoryofchange fmt` idempotent on 100 % of fixtures | Fixture test runs fmt twice |
| Thin-slice time to workshop pack | Manual diagram + slides, hours per revision | Minimal ToC → diagram, logic models, metric definitions, glossary, and workshop deck in one command | Golden test over the intermediary example |
| Provenance coverage | n/a | 100 % of elements in reference examples carry a basis; 100 % of `stated` elements cite a source | Lint rules TOC-PRV-001/002 over examples |
| Review round-trip fidelity | Manual transcription of feedback | 100 % of review rows become a proposed patch op or a flagged note; 0 silent drops; stale-hash responses flagged | Import tests with seeded sheet, table, comment, and suggestion rows |
| Export stamp coverage | n/a | 100 % of exported artifacts carry ID, version, hash | Export tests per format |
| Judge agreement recorded | n/a | Discrimination eval runs ≥ 2 judges and reports agreement per category | `docs/evals/` record includes agreement metrics |

## 6. Assumptions and Risks

| ID | Assumption / Risk | Impact if wrong | Mitigation |
|----|-------------------|-----------------|------------|
| A-1 | The canonical ToC vocabulary (Aspen/ActKnowledge, AECF, UNDG) is stable enough to freeze in a v0 schema | Breaking schema changes after adopters start | Version the schema (`$id` with major version); additive-only within v0.x minors; revision log |
| A-2 | Rule-based confidence scoring is useful without statistics | Leaders over-trust a number | Label scores as "rule-based confidence", show inputs, never call it attribution |
| A-3 | USM and UFM can be built on top of the reference format and alignment links without changing the ToC model | Rework when USM/UFM land | UFM fixture in Phase 5; contract reviewed with USM/UFM owners before v1 |
| A-4 | Non-technical users will reach the tools through an AI assistant with agent skills and spreadsheets (CSV), not the CLI directly | Tools unused by the people who most need them | Skills and CSV paths are first-class requirements (PRD-FR-204, 605); success metric for skill determinism |
| A-5 | A per-indicator staff-minute estimate is a good-enough proxy for collection burden | Budget checks mislead | Estimates are labelled as estimates; actual time logs can be recorded as observations of a burden indicator and compared |
| R-1 | `prism-roadmap` pulls a large dependency tree | Heavier builds for core users | Keep adapters in a separate package; core packages import only `prism-core` and the standard library plus schema tooling |
| R-2 | Over-modeling makes authoring painful | Low adoption | Minimal valid document is small (goal + one outcome + one link + one indicator); most fields optional; `draft` lint profile; edit commands; templates and examples |
| R-3 | YAML comments are lost when tools rewrite a document | Authors lose notes | Documented: tool-written files are canonical; notes belong in `description` / `narrative` fields; `--dry-run` shows the patch result before writing |

## 7. Out of Scope / Future

- USM (ToC + Balanced Scorecard from a market participant's perspective) and
  UFM (competing futures with actor coalitions) — separate initiatives that
  consume this module.
- Logframe matrix and IATI results exports.
- Web viewer (likely via `prism` sitegen or the `@grokify/prism` viewer).
- Data connectors (CRMs, form tools, analytics, surveys, GitHub, visionstudio
  delivery evidence as automatic output indicators) and record linkage.
- XLSX import/export (CSV covers spreadsheets for v0.1) and hosted
  dashboards.
- A hosted multi-organization service for intermediaries.
- Probabilistic or Bayesian scoring of pathways; contribution-analysis
  workflows beyond the evidence and assumption records.
- GTM and growth-phase artifacts for the tool itself.
