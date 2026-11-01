# ROADMAP — Theory of Change Specification — Go-first model, JSON Schema, measurement, and leadership reporting

**Initiative:** `INIT-THEORYOFCHANGESPEC-001`
**Repository:** `github.com/ProductBuildersHQ/theoryofchange-spec`
**Status:** In progress — 14 of 16 phases completed

> RMI IDs are stable and permanent. Commits implementing an item carry the trailer `Refs: RMI-THEORYOFCHANGESPEC-<NNN>`. Phase status is derived from member RMIs — a phase is complete only when all its RMIs are complete. RMI `0NN` corresponds to PLAN step `P-NN`; phases order the work, so numbers are not sequential across phases. Phases 1–5 are the thin slice (adopter milestones M0 and M1). Requirements in each `Satisfies` line are defined in PRD.md and TRD.md.

## Phase 1 — Core Model and Schema

**Theme:** Scaffold the module, map the Theory of Change vocabulary to Go types with a small minimal-valid document, and publish a lint-clean, embedded JSON Schema
**Status:** Completed — 6 of 6 items completed

- [x] `RMI-THEORYOFCHANGESPEC-001` Repository scaffold: go.mod, CI workflows (go-ci, go-lint, go-sast), lint config, README, CLAUDE.md, LICENSE
  - Type: enabler · Status: completed
  - Satisfies: — (enabler)
  - Acceptance: `go build ./...` and `golangci-lint run` pass; CI workflows green
- [x] `RMI-THEORYOFCHANGESPEC-002` Concept-to-type mapping and glossary from the research brief
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-001`
  - Satisfies: PRD G-1
  - Acceptance: every core concept of research §8.4 and §9.8 maps to a planned Go type or a documented deferral
- [x] `RMI-THEORYOFCHANGESPEC-003` `pkg/ref` reference grammar (`toc:<doc>#<element>`) with slug validation
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-001`
  - Satisfies: PRD-FR-604, TRD-015
  - Acceptance: parse/format round-trip; invalid refs rejected
- [x] `RMI-THEORYOFCHANGESPEC-004` `pkg/toc` core document model with a minimal required set (goal, outcome, link, indicator), reusing prism-core
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-002`, `RMI-THEORYOFCHANGESPEC-003`
  - Satisfies: PRD-FR-101…110, TRD-001, TRD-016
  - Acceptance: round-trip marshal tests; enum `Valid`/`All` helpers tested; minimal fixture round-trips
- [x] `RMI-THEORYOFCHANGESPEC-005` ObservationSet, Status, StatusDiff, and ActorRegistry document types
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-004`
  - Satisfies: PRD-FR-111, PRD-FR-406, TRD-002
  - Acceptance: round-trip tests for each document kind
- [x] `RMI-THEORYOFCHANGESPEC-006` Schema pipeline: `cmd/generate-schema`, embedded `schema/` with Get/Names/Validate, schemakit lint, CI drift check
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-004`, `RMI-THEORYOFCHANGESPEC-005`
  - Satisfies: PRD-FR-201, PRD-FR-202, TRD-003, TRD-004
  - Acceptance: `schemakit lint --property-case camelCase` reports 0 violations; regeneration leaves no diff; every schema loads from embed

## Phase 2 — Load, Validate, and Edit

**Theme:** Trusted documents, a queryable causal graph, and one validated patch engine through which every edit flows
**Status:** Completed — 6 of 6 items completed

- [x] `RMI-THEORYOFCHANGESPEC-007` `pkg/tocio` JSON/YAML loading with limits, observation-set merge, and canonical formatting
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-006`
  - Satisfies: PRD-FR-203, PRD-FR-704, TRD-005, TRD-031
  - Acceptance: one fixture in YAML and JSON yields equal documents; `Format` idempotent
- [x] `RMI-THEORYOFCHANGESPEC-008` `pkg/validate` schema, ID uniqueness and format, reference resolution with JSON-pointer issues
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-007`
  - Satisfies: PRD-FR-301, TRD-006
  - Acceptance: table tests per error class
- [x] `RMI-THEORYOFCHANGESPEC-009` `pkg/graph` build and queries (upstream, downstream, paths to impact, cycles, levels, orphans, dead ends, ceiling)
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-008`
  - Satisfies: PRD-FR-302, TRD-007
  - Acceptance: unit and property tests; 1,000-node benchmark recorded
- [x] `RMI-THEORYOFCHANGESPEC-046` `pkg/edit` patch engine: RFC 6902 with `@id` addressing, atomic apply, error-delta rejection, dry run
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-008`
  - Satisfies: PRD-FR-701, PRD-FR-702, PRD G-14, TRD-029
  - Acceptance: error-introducing patches rejected with input untouched; unknown `@id` fails the whole patch
- [x] `RMI-THEORYOFCHANGESPEC-011` `pkg/service` and `cmd/theoryofchange` (Cobra) with validate, graph, schema, init, fmt, and patch commands
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-009`, `RMI-THEORYOFCHANGESPEC-046`
  - Satisfies: PRD US-1, PRD US-16, TRD §6.1
  - Acceptance: CLI tests cover exit codes and output formats; `fmt --check` exits 1 on unformatted input
- [x] `RMI-THEORYOFCHANGESPEC-047` Element commands (`add` / `set` / `rm --cascade`) compiled to patches for core element kinds
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-046`, `RMI-THEORYOFCHANGESPEC-011`
  - Satisfies: PRD-FR-703, PRD US-16, TRD-030
  - Acceptance: each command equals its patch equivalent; `rm` without `--cascade` rejects dangling references

## Phase 3 — Pathways, Metric Library, Provenance, and Reference Examples

**Theme:** Program pathways, shared metric definitions, a glossary, and provenance on every element, proven on fictional reference examples
**Status:** Completed — 4 of 4 items completed

- [x] `RMI-THEORYOFCHANGESPEC-032` Pathways (program areas) and parent / child ToC references, with element commands
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-008`, `RMI-THEORYOFCHANGESPEC-047`
  - Satisfies: PRD-FR-112, PRD US-9, TRD-017
  - Acceptance: pathway round-trip; child ToC outcomes resolve against the parent
- [x] `RMI-THEORYOFCHANGESPEC-033` Metric library, glossary, and indicator extensions (metricId, pathwayIds, claim, caveats, rollupFrom, theme_count), with element commands
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-032`
  - Satisfies: PRD-FR-113, PRD-FR-114, PRD-FR-117, PRD US-10, TRD-017
  - Acceptance: roll-up declarations validated against metric aggregation; claim defaults to contribution when omitted
- [x] `RMI-THEORYOFCHANGESPEC-052` Provenance on every element (stated / inferred / hypothesis / confirmed / rejected, cited evidence), indicator dataAvailability, and Consultation records; rejected elements excluded from analysis
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-033`
  - Satisfies: PRD-FR-124, PRD-FR-125, PRD-FR-123, TRD-036
  - Acceptance: round-trip tests; rejected elements absent from graph, measurement, and default reports
- [x] `RMI-THEORYOFCHANGESPEC-010` Reference examples (fictional) with provenance: intermediary organization (three pathways, two impact layers, metric library, glossary) and market adoption (open agent-authorization standard, AAER categories)
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-033`, `RMI-THEORYOFCHANGESPEC-052`
  - Satisfies: PRD G-1, PRD G-9, PRD US-1
  - Acceptance: both examples validate; exercise groups, feedback links, qualitative indicators, and every provenance basis

## Phase 4 — Workshop Pack and Source Drafting

**Theme:** Draft a theory from public sources and render the pathway diagram, logic models, metric definitions, glossary, CSV exports, and a Marp workshop deck from the document alone — adopter milestone M0
**Status:** Completed — 6 of 6 items completed

- [x] `RMI-THEORYOFCHANGESPEC-022` Versioned `report.View` model built from document and graph, status optional, with basis badges
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-009`, `RMI-THEORYOFCHANGESPEC-010`
  - Satisfies: PRD-FR-501, PRD-FR-502, PRD-FR-505, TRD-010
  - Acceptance: view tests with and without a status snapshot; inferred elements carry badges
- [x] `RMI-THEORYOFCHANGESPEC-023` Mermaid pathway renderer with level subgraphs, optional health styling, feedback links, ceiling, and pathway filter
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`
  - Satisfies: PRD-FR-501, PRD G-12
  - Acceptance: golden Mermaid output
- [x] `RMI-THEORYOFCHANGESPEC-037` Logic-model (per pathway), metric-definitions, and glossary reports
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`
  - Satisfies: PRD-FR-505, PRD US-12, TRD-022
  - Acceptance: golden files over the intermediary example
- [x] `RMI-THEORYOFCHANGESPEC-049` `pkg/slides` Marp engine and workshop deck (`theoryofchange deck workshop`)
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-023`, `RMI-THEORYOFCHANGESPEC-037`
  - Satisfies: PRD-FR-801, PRD G-15, PRD US-17, TRD-033
  - Acceptance: golden workshop deck; deck numbers and labels equal the reports'
- [x] `RMI-THEORYOFCHANGESPEC-039` `theoryofchange report` definition kinds, `definitions` pack, and `theoryofchange export csv` (indicators, metrics, glossary)
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-023`, `RMI-THEORYOFCHANGESPEC-037`, `RMI-THEORYOFCHANGESPEC-011`
  - Satisfies: PRD-FR-204, PRD-FR-505, PRD G-12, TRD-021
  - Acceptance: CLI tests; CSV golden files (RFC 4180)
- [x] `RMI-THEORYOFCHANGESPEC-053` `theoryofchange evidence add` and the toc-draft-from-sources agent skill
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-047`, `RMI-THEORYOFCHANGESPEC-052`, `RMI-THEORYOFCHANGESPEC-039`
  - Satisfies: PRD-FR-606, PRD US-19, PRD G-17, TRD-037
  - Acceptance: smoke test over saved fixture pages yields a draft where every element has a basis and every `stated` element cites a source

## Phase 5 — Review and Agreement

**Theme:** Assessment report, Pandoc-ready sharing exports with stamps, review packet and sheet, review import as proposals, and approval of the agreed baseline — adopter milestone M1
**Status:** Completed — 5 of 5 items completed

- [x] `RMI-THEORYOFCHANGESPEC-054` Assessment report: provenance mix, gaps, data availability, assumptions needing testing, generated questions, short narrative
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`, `RMI-THEORYOFCHANGESPEC-052`
  - Satisfies: PRD-FR-512, PRD US-19, TRD-038
  - Acceptance: golden file over the intermediary example
- [x] `RMI-THEORYOFCHANGESPEC-056` Pandoc Markdown profile with PNG diagrams, export stamps on every artifact, and `theoryofchange export docx|pptx`
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-039`, `RMI-THEORYOFCHANGESPEC-049`, `RMI-THEORYOFCHANGESPEC-023`
  - Satisfies: PRD-FR-901, PRD-FR-902, TRD-039
  - Acceptance: golden Markdown; stamp present in every format; pandoc export test skipped when absent
- [x] `RMI-THEORYOFCHANGESPEC-057` Review packet (ID-tagged response tables) and review sheet (CSV) exports
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-054`, `RMI-THEORYOFCHANGESPEC-056`
  - Satisfies: PRD-FR-903, PRD US-20, TRD-039
  - Acceptance: golden packet and sheet; every element row carries its stable ID
- [x] `RMI-THEORYOFCHANGESPEC-058` ReviewResponse schema and `theoryofchange import review` (JSON and --from-csv) producing proposed patches and a consultation record
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-057`, `RMI-THEORYOFCHANGESPEC-046`
  - Satisfies: PRD-FR-904, PRD-FR-905, PRD US-20, TRD-040
  - Acceptance: seeded rows from every origin map to ops or flagged notes; stale hashes flagged; document never modified
- [x] `RMI-THEORYOFCHANGESPEC-055` `theoryofchange approve` with approvers, revision, version label, and refusal checks
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-058`, `RMI-THEORYOFCHANGESPEC-052`
  - Satisfies: PRD-FR-707, PRD US-21, TRD-041
  - Acceptance: approval sets status and revision through the patch engine; refuses on errors or open hypotheses without --force

## Phase 6 — Measurement Framework Design

**Theme:** Metric selection and collection burden, audiences and stories, UNDG change analysis, and the full intermediary example — adopter milestone M1.5 (collection plan)
**Status:** Completed — 5 of 5 items completed

- [x] `RMI-THEORYOFCHANGESPEC-034` Metric selection criteria and scores, collection cost, measurement budget, and `pkg/burden` totals
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-033`
  - Satisfies: PRD-FR-115, PRD G-10, PRD US-11, TRD-020
  - Acceptance: burden arithmetic table tests (minutes per week, questions per respondent per year, per pathway and mode)
- [x] `RMI-THEORYOFCHANGESPEC-035` Audiences, theme codebook, and Story / LearningReview observation types
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-033`
  - Satisfies: PRD-FR-116, PRD-FR-117, PRD-FR-118, TRD-017, TRD-018
  - Acceptance: unknown theme IDs and missing consent rejected by validation
- [x] `RMI-THEORYOFCHANGESPEC-044` Change-analysis model: problem-tree causes, outcome responsibility and delivering actors, prioritization criteria for outcomes and interventions, target groups, assumption and risk enrichment
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-032`, `RMI-THEORYOFCHANGESPEC-034`
  - Satisfies: PRD-FR-105, PRD-FR-108, PRD-FR-119…122, TRD-025
  - Acceptance: round-trip tests; unknown cause / target-group / actor references rejected by validation
- [x] `RMI-THEORYOFCHANGESPEC-036` Intermediary example extended with framework-design and change-analysis data; schemas regenerated lint-clean
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-034`, `RMI-THEORYOFCHANGESPEC-035`, `RMI-THEORYOFCHANGESPEC-044`
  - Satisfies: PRD G-9 success metric
  - Acceptance: example with 8–12 selected metrics, budget, audiences, codebook, sample stories, problem tree, partner-delivered outcomes, and consultations validates; `schemakit lint` 0 violations
- [x] `RMI-THEORYOFCHANGESPEC-038` Metric-selection and collection-plan reports and CSV exports
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`, `RMI-THEORYOFCHANGESPEC-034`, `RMI-THEORYOFCHANGESPEC-039`
  - Satisfies: PRD-FR-505, PRD US-11, PRD US-12, TRD-022
  - Acceptance: golden files; collection-plan totals equal `pkg/burden` results

## Phase 7 — Quality: Lint, Rubric, and Judging

**Theme:** Deterministic lint with profiles, the LLM-as-a-Judge rubric with a development profile, and reproducible file-based judging with multi-judge agreement
**Status:** Completed — 6 of 6 items completed

- [x] `RMI-THEORYOFCHANGESPEC-012` Lint engine: rule registry, draft / standard / full profiles, options, structured-evaluation SummaryReport output
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-009`
  - Satisfies: PRD-FR-303, PRD-FR-304, PRD-FR-804, TRD-008, TRD-035
  - Acceptance: registry and report-shape tests; minimal fixture is GO under `draft`
- [x] `RMI-THEORYOFCHANGESPEC-013` Structure and logic rules (TOC-STR-*, TOC-LOG-*) with seeded-defect fixtures
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-012`
  - Satisfies: PRD-FR-303, TRD §6.4
  - Acceptance: each rule triggered by exactly one fixture; reference examples report 0 errors
- [x] `RMI-THEORYOFCHANGESPEC-014` Assumption, measurement, intervention, accountability, actor, evidence, pathway, metric, claim, budget, audience, cause, responsibility, risk, equity, validation, and provenance rules with seeded-defect fixtures
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-012`, `RMI-THEORYOFCHANGESPEC-036`
  - Satisfies: PRD-FR-303, PRD US-10, PRD US-11, TRD §6.4
  - Acceptance: one fixture per rule; ≥ 15 rules total implemented and tested
- [x] `RMI-THEORYOFCHANGESPEC-015` ToC spec template, core RubricSet (sector-neutral UNDG checklist plus plausible / feasible / testable), and development rubric profile for LLM-as-a-Judge, embedded in `spec/`
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-004`
  - Satisfies: PRD-FR-601, TRD-012, TRD-028
  - Acceptance: both rubrics parse into `rubric.RubricSet`, `Validate()` passes, weights sum to 1.0; every UNDG checklist item maps to a criterion
- [x] `RMI-THEORYOFCHANGESPEC-050` `pkg/judge` and `theoryofchange judge show | record | summary`: bundle hashing, verdict validation, multi-judge agreement
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-015`, `RMI-THEORYOFCHANGESPEC-039`
  - Satisfies: PRD-FR-803, PRD G-16, PRD US-18, TRD-034
  - Acceptance: bundle golden; verdicts with mismatched bundle hash rejected; aggregation reports per-category agreement
- [x] `RMI-THEORYOFCHANGESPEC-016` `theoryofchange lint` CLI and rubric discrimination eval with at least two judges and agreement, recorded in `docs/evals/`
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-013`, `RMI-THEORYOFCHANGESPEC-014`, `RMI-THEORYOFCHANGESPEC-050`
  - Satisfies: PRD US-2, PRD success metrics "LLM-judge discrimination" and "Judge agreement recorded"
  - Acceptance: no judge raises a high or critical finding on the reference example, after which it goes to human review; seeded weak ToC fails ≥ 3 rubric categories; agreement reported

## Phase 8 — Measurement

**Theme:** Turn spreadsheet data and observations into indicator, outcome, pathway, and assumption status with reproducible roll-ups and confidence — adopter milestone M2
**Status:** Completed — 6 of 6 items completed

- [x] `RMI-THEORYOFCHANGESPEC-017` Indicator progress, expected progress, status (including baseline_only), and health mapping
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-008`, `RMI-THEORYOFCHANGESPEC-033`
  - Satisfies: PRD-FR-401, PRD-FR-402, PRD-FR-509, TRD-009, TRD-019
  - Acceptance: table tests cover every status branch
- [x] `RMI-THEORYOFCHANGESPEC-018` Outcome, pathway-status, and cross-pathway metric roll-ups through precondition logic and nested groups
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-009`, `RMI-THEORYOFCHANGESPEC-017`
  - Satisfies: PRD-FR-403, PRD-FR-113, TRD-019
  - Acceptance: tests for all / any / at_least / nested groups and each metric aggregation
- [x] `RMI-THEORYOFCHANGESPEC-019` Assumption propagation and link / path confidence scoring
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-018`
  - Satisfies: PRD-FR-404, PRD-FR-405
  - Acceptance: invalidated assumption flags every downstream outcome (US-5); confidence tests with documented weights
- [x] `RMI-THEORYOFCHANGESPEC-020` `measure.Evaluate` status snapshot, theme counts, and `Diff` between periods; determinism
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-019`, `RMI-THEORYOFCHANGESPEC-035`
  - Satisfies: PRD-FR-406, TRD INV-3
  - Acceptance: golden snapshots byte-identical across two runs
- [x] `RMI-THEORYOFCHANGESPEC-040` CSV observation import with mapping file and unmatched-row reporting
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-007`, `RMI-THEORYOFCHANGESPEC-033`
  - Satisfies: PRD-FR-204, PRD US-13, TRD-021
  - Acceptance: seeded bad rows all reported; 0 silent drops
- [x] `RMI-THEORYOFCHANGESPEC-021` `theoryofchange status` and `theoryofchange import` CLI; performance benchmarks
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-020`, `RMI-THEORYOFCHANGESPEC-040`, `RMI-THEORYOFCHANGESPEC-011`
  - Satisfies: PRD US-4, PRD US-13, TRD §8
  - Acceptance: CLI tests; 200-node benchmark under 1 s

## Phase 9 — Reporting, Story, and Presentations

**Theme:** Executive, audience story, learning-review, portfolio, narrative, and division-of-labour reports, plus exec / learning-review / story decks with optional export — adopter milestone M3
**Status:** Completed — 8 of 8 items completed

- [x] `RMI-THEORYOFCHANGESPEC-024` Status reports: exec, results, assumptions, interventions, evidence-gaps, full
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-020`, `RMI-THEORYOFCHANGESPEC-022`, `RMI-THEORYOFCHANGESPEC-023`
  - Satisfies: PRD-FR-501, PRD US-3
  - Acceptance: golden files; every reported status cites its inputs
- [x] `RMI-THEORYOFCHANGESPEC-041` Audience story report with consent-gated quotes
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-024`, `RMI-THEORYOFCHANGESPEC-035`
  - Satisfies: PRD-FR-506, PRD US-14, TRD-023
  - Acceptance: golden files; no-consent and restricted quotes absent
- [x] `RMI-THEORYOFCHANGESPEC-042` Learning-review and portfolio reports
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-024`, `RMI-THEORYOFCHANGESPEC-034`
  - Satisfies: PRD-FR-507, PRD-FR-508, PRD US-14
  - Acceptance: golden files over two observation periods
- [x] `RMI-THEORYOFCHANGESPEC-045` Narrative (programme / strategy document scaffold) and division-of-labour reports; accountable vs monitored-only separation
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`, `RMI-THEORYOFCHANGESPEC-044`
  - Satisfies: PRD-FR-510, PRD-FR-511, PRD-FR-120, TRD-026, TRD-027
  - Acceptance: golden files over a mixed-responsibility fixture
- [x] `RMI-THEORYOFCHANGESPEC-060` Projection artifacts: outcome map (outcomes framework), logframe matrix, and logic-model diagram, as definition reports and in the definitions pack
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`, `RMI-THEORYOFCHANGESPEC-044`
  - Satisfies: PRD-FR-513
  - Acceptance: golden files; every row names the element IDs it was projected from; a one-pathway theory and a multi-pathway theory both render
- [x] `RMI-THEORYOFCHANGESPEC-025` Custom report templates and restricted-evidence suppression
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-024`
  - Satisfies: PRD-FR-502, PRD-FR-503, TRD-011
  - Acceptance: custom template test; restricted excerpt absent from all outputs
- [x] `RMI-THEORYOFCHANGESPEC-051` Exec, learning-review, and story decks; `--export pdf|pptx|html` via optional `mmdc` and `marp`
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-049`, `RMI-THEORYOFCHANGESPEC-024`, `RMI-THEORYOFCHANGESPEC-041`, `RMI-THEORYOFCHANGESPEC-042`
  - Satisfies: PRD-FR-801, PRD-FR-802, PRD US-17, TRD-033
  - Acceptance: golden decks; export test produces files when tools are present and exits 3 with guidance when absent
- [x] `RMI-THEORYOFCHANGESPEC-026` `theoryofchange report` status kinds in CLI and executive-report benchmark
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-025`, `RMI-THEORYOFCHANGESPEC-041`, `RMI-THEORYOFCHANGESPEC-042`, `RMI-THEORYOFCHANGESPEC-045`, `RMI-THEORYOFCHANGESPEC-021`
  - Satisfies: PRD G-5
  - Acceptance: CLI tests; exec report for 200-node ToC under 1 s

## Phase 10 — Agents and Revisions

**Theme:** MCP tools (including validated patching), agent skills, semantic theory diffs, and automated Google sharing — adopter milestone M4
**Status:** In progress — 3 of 4 items completed

- [x] `RMI-THEORYOFCHANGESPEC-027` MCP server exposing schema, validate, lint, graph, status, report, patch, deck, and judge-bundle tools with root confinement
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-026`, `RMI-THEORYOFCHANGESPEC-046`, `RMI-THEORYOFCHANGESPEC-050`
  - Satisfies: PRD-FR-602, PRD-FR-706, PRD US-6, TRD-013
  - Acceptance: in-memory transport test drives every tool; path traversal rejected; `toc_patch` rejects error-introducing patches
- [x] `RMI-THEORYOFCHANGESPEC-048` `pkg/diff` semantic theory diff and `theoryofchange diff --revision`
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-046`, `RMI-THEORYOFCHANGESPEC-009`
  - Satisfies: PRD-FR-705, PRD US-16, PRD US-21, TRD-032
  - Acceptance: golden diffs keyed by element ID, including rejected elements; draft Revision lists changed IDs
- [x] `RMI-THEORYOFCHANGESPEC-043` Agent skills: toc-author (edits via patches), toc-status-refresh, toc-learning-review, toc-audience-story
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-026`, `RMI-THEORYOFCHANGESPEC-039`, `RMI-THEORYOFCHANGESPEC-047`
  - Satisfies: PRD-FR-605, PRD US-15, PRD G-13, TRD-024
  - Acceptance: smoke test runs each skill's command sequence twice with identical outputs
- [ ] `RMI-THEORYOFCHANGESPEC-059` toc-share-google agent skill (publish, read responses, import) over external Google tooling, with manual fallback
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-056`, `RMI-THEORYOFCHANGESPEC-057`, `RMI-THEORYOFCHANGESPEC-058`, `RMI-GOGOOGLE-006`, `RMI-MCPGOOGLE-004`, `RMI-MCPGOOGLE-005`
  - Satisfies: PRD-FR-607, PRD US-20, PRD G-18, TRD-042
  - Acceptance: smoke test with recorded gogoogle and mcp-google outputs produces a valid ReviewResponse and proposals; manual fallback documented and tested via --from-csv

## Phase 11 — Ecosystem Integration and Release

**Theme:** PRISM exports, the USM/UFM integration contract and competing-futures fixture, and the v0.1.0 release
**Status:** In progress — 3 of 4 items completed

- [x] `RMI-THEORYOFCHANGESPEC-028` `pkg/prismx` OKR and outcome-journey export with import-boundary test
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-020`
  - Satisfies: PRD-FR-603, PRD US-7, TRD-014
  - Acceptance: exported OKR documents pass prism-roadmap validation; core packages do not import prism-roadmap
- [x] `RMI-THEORYOFCHANGESPEC-029` Actor registry resolution, cross-document references, and competing-futures fixture
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-008`, `RMI-THEORYOFCHANGESPEC-032`
  - Satisfies: PRD G-8, PRD US-8, TRD §7.3
  - Acceptance: three competing-future ToCs sharing one actor registry validate with no schema changes
- [ ] `RMI-THEORYOFCHANGESPEC-030` USM/UFM integration contract document
  - Type: capability · Status: in_progress
  - Depends on: `RMI-THEORYOFCHANGESPEC-029`
  - Satisfies: PRD G-8
  - Acceptance: contract reviewed by USM/UFM owners
- [x] `RMI-THEORYOFCHANGESPEC-031` Documentation, CHANGELOG, v0.1.0 release, and spec-type sync follow-up for specification-workflow-spec
  - Type: release · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-027`, `RMI-THEORYOFCHANGESPEC-043`, `RMI-THEORYOFCHANGESPEC-048`, `RMI-THEORYOFCHANGESPEC-059`, `RMI-THEORYOFCHANGESPEC-028`, `RMI-THEORYOFCHANGESPEC-030`, `RMI-THEORYOFCHANGESPEC-051`
  - Satisfies: all PRD success metrics (release gate)
  - Acceptance: CI green before tag; publication scan clean; follow-up RMI filed in specification-workflow-spec

## Phase 12 — Attribution and Notices

**Theme:** Credit the methods and projects the spec builds on, and carry the notices its dependencies require
**Status:** Completed — 1 of 1 items completed

- [x] `RMI-THEORYOFCHANGESPEC-061` Credits, third-party notices, and a dependency license audit
  - Type: capability · Status: completed
  - Satisfies: — (documentation and compliance)
  - Acceptance: README and docs credit the methods and projects used; THIRD_PARTY_NOTICES.md carries each distributed dependency's license and notice text; every module compiled into the CLI is permissively licensed

## Phase 13 — Headline Metrics

**Theme:** Tie the causal chain to the metrics each program steers by and each audience is told: a North Star per program, guardrails, curated headline numbers per audience, and metrics worked out from others
**Status:** Completed — 7 of 7 items completed

- [x] `RMI-THEORYOFCHANGESPEC-062` Headline metrics per audience and a North Star per program, with a metric-chain diagram
  - Type: capability · Status: completed
  - Satisfies: PRD G-8
  - Acceptance: The README and docs index show how resources, activities, outputs, outcomes, impact, North Star metrics, and audience headline metrics link; audiences can list an ordered, worded set of headline indicators; the story report and deck use it; lint checks it

- [x] `RMI-THEORYOFCHANGESPEC-063` Shared pathway and tier helpers used by every rule and report
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-062`
  - Acceptance: One function says which pathways an indicator serves and which tier it is; PTH-002/003 and MET-001/003 agree; no rule re-derives membership

- [x] `RMI-THEORYOFCHANGESPEC-064` Model headline metrics: audience headline lists, pathway focus (North Star, one metric, guardrails), product kind, mission and vision
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-063`
  - Acceptance: Go types, JSON Schemas regenerated, references validated; existing documents still validate

- [x] `RMI-THEORYOFCHANGESPEC-065` Lint the activity-to-impact chain, headline lists, and North Stars
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-064`
  - Acceptance: Rules report an activity that leads nowhere, a link and intervention that disagree, an audience with no numbers, a North Star that is not an outcome at or below the ceiling

- [x] `RMI-THEORYOFCHANGESPEC-066` Reports: metric tree, headline use in story and deck, headline report, estimates shown as estimates
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-065`
  - Acceptance: A metric-tree report per pathway; story and deck use the audience headline list and drop the hardcoded five; an estimated value reads as an estimate

- [x] `RMI-THEORYOFCHANGESPEC-067` Structured metric composition and derived indicator values
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-066`
  - Acceptance: A metric can be a ratio or product of other metrics; a derived value is computed when its parts are observed; no cycles

- [x] `RMI-THEORYOFCHANGESPEC-068` Examples, guides, MCP text, and the metric-chain diagram updated for headline metrics
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-066`, `RMI-THEORYOFCHANGESPEC-067`
  - Acceptance: The Riverbend example uses the new fields; a guide explains tiers, North Star, and headline; the diagram's planned boxes become solid

## Phase 14 — Vocabulary

**Theme:** Make the terms consistent and easy to learn
**Status:** Completed — 1 of 1 items completed

- [x] `RMI-THEORYOFCHANGESPEC-069` Rename the intermediate outcome level to medium_term so short, medium, and long form one series
  - Type: capability · Status: completed
  - Acceptance: the level is medium_term in the model, schema, reports, examples, and docs; every test passes; no alias is kept

## Phase 15 — Layered Logic Model Diagram

**Theme:** Draw the logic model as stage columns with real edges, from one shared diagram spec
**Status:** Completed — 3 of 3 items completed

- [x] `RMI-THEORYOFCHANGESPEC-070` Derive a layered logic model graph (stages, nodes, edges) from a view
  - Type: capability · Status: completed
  - Acceptance: inputs, activities, outputs, and outcomes by level are columns; edges come from resources, interventions, and links; order is deterministic and reduces crossings; tests cover each edge kind

- [x] `RMI-THEORYOFCHANGESPEC-071` Emit the logic model graph as D2 and Mermaid through d2vision and add report kinds
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-070`
  - Acceptance: logic-model-layered and logic-model-d2 reports; the existing logic-model diagram is unchanged; the D2 compiles and keeps every label; stage columns align

- [x] `RMI-THEORYOFCHANGESPEC-072` Use the layered logic model diagram in the workshop deck beside the existing slide
  - Type: capability · Status: completed
  - Depends on: `RMI-THEORYOFCHANGESPEC-071`
  - Acceptance: the deck has the layered diagram as a slide and the existing table and diagram remain; docs describe both and how they differ

## Phase 16 — Release v0.2.0

**Theme:** Prepare the second release: changelog, release notes, and guides
**Status:** Completed — 1 of 1 items completed

- [x] `RMI-THEORYOFCHANGESPEC-073` Prepare v0.2.0: changelog, release notes, and guide updates
  - Type: capability · Status: completed
  - Acceptance: CHANGELOG.json has v0.2.0 and CHANGELOG.md is generated from it; docs/releases/v0.2.0.md exists; the guides cover headline metrics, composed indicators, and the layered diagram; mkdocs build --strict passes

