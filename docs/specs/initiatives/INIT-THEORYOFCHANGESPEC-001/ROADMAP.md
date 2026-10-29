# ROADMAP — Theory of Change Specification — Go-first model, JSON Schema, measurement, and leadership reporting

**Initiative:** `INIT-THEORYOFCHANGESPEC-001`
**Repository:** `github.com/ProductBuildersHQ/theoryofchange-spec`
**Status:** Planned — 0 of 11 phases completed

> RMI IDs are stable and permanent. Commits implementing an item carry the trailer `Refs: RMI-THEORYOFCHANGESPEC-<NNN>`. Phase status is derived from member RMIs — a phase is complete only when all its RMIs are complete. RMI `0NN` corresponds to PLAN step `P-NN`; phases order the work, so numbers are not sequential across phases. Phases 1–5 are the thin slice (adopter milestones M0 and M1). Requirements in each `Satisfies` line are defined in PRD.md and TRD.md.

## Phase 1 — Core Model and Schema

**Theme:** Scaffold the module, map the Theory of Change vocabulary to Go types with a small minimal-valid document, and publish a lint-clean, embedded JSON Schema
**Status:** Planned — 0 of 6 items completed

- [ ] `RMI-THEORYOFCHANGESPEC-001` Repository scaffold: go.mod, CI workflows (go-ci, go-lint, go-sast), lint config, README, CLAUDE.md, LICENSE
  - Type: enabler · Status: planned
  - Satisfies: — (enabler)
  - Acceptance: `go build ./...` and `golangci-lint run` pass; CI workflows green
- [ ] `RMI-THEORYOFCHANGESPEC-002` Concept-to-type mapping and glossary from the research brief
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-001`
  - Satisfies: PRD G-1
  - Acceptance: every core concept of research §8.4 and §9.8 maps to a planned Go type or a documented deferral
- [ ] `RMI-THEORYOFCHANGESPEC-003` `pkg/ref` reference grammar (`toc:<doc>#<element>`) with slug validation
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-001`
  - Satisfies: PRD-FR-604, TRD-015
  - Acceptance: parse/format round-trip; invalid refs rejected
- [ ] `RMI-THEORYOFCHANGESPEC-004` `pkg/toc` core document model with a minimal required set (goal, outcome, link, indicator), reusing prism-core
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-002`, `RMI-THEORYOFCHANGESPEC-003`
  - Satisfies: PRD-FR-101…110, TRD-001, TRD-016
  - Acceptance: round-trip marshal tests; enum `Valid`/`All` helpers tested; minimal fixture round-trips
- [ ] `RMI-THEORYOFCHANGESPEC-005` ObservationSet, Status, StatusDiff, and ActorRegistry document types
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-004`
  - Satisfies: PRD-FR-111, PRD-FR-406, TRD-002
  - Acceptance: round-trip tests for each document kind
- [ ] `RMI-THEORYOFCHANGESPEC-006` Schema pipeline: `cmd/generate-schema`, embedded `schema/` with Get/Names/Validate, schemakit lint, CI drift check
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-004`, `RMI-THEORYOFCHANGESPEC-005`
  - Satisfies: PRD-FR-201, PRD-FR-202, TRD-003, TRD-004
  - Acceptance: `schemakit lint --property-case camelCase` reports 0 violations; regeneration leaves no diff; every schema loads from embed

## Phase 2 — Load, Validate, and Edit

**Theme:** Trusted documents, a queryable causal graph, and one validated patch engine through which every edit flows
**Status:** Planned — 0 of 6 items completed

- [ ] `RMI-THEORYOFCHANGESPEC-007` `pkg/tocio` JSON/YAML loading with limits, observation-set merge, and canonical formatting
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-006`
  - Satisfies: PRD-FR-203, PRD-FR-704, TRD-005, TRD-031
  - Acceptance: one fixture in YAML and JSON yields equal documents; `Format` idempotent
- [ ] `RMI-THEORYOFCHANGESPEC-008` `pkg/validate` schema, ID uniqueness and format, reference resolution with JSON-pointer issues
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-007`
  - Satisfies: PRD-FR-301, TRD-006
  - Acceptance: table tests per error class
- [ ] `RMI-THEORYOFCHANGESPEC-009` `pkg/graph` build and queries (upstream, downstream, paths to impact, cycles, levels, orphans, dead ends, ceiling)
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-008`
  - Satisfies: PRD-FR-302, TRD-007
  - Acceptance: unit and property tests; 1,000-node benchmark recorded
- [ ] `RMI-THEORYOFCHANGESPEC-046` `pkg/edit` patch engine: RFC 6902 with `@id` addressing, atomic apply, error-delta rejection, dry run
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-008`
  - Satisfies: PRD-FR-701, PRD-FR-702, PRD G-14, TRD-029
  - Acceptance: error-introducing patches rejected with input untouched; unknown `@id` fails the whole patch
- [ ] `RMI-THEORYOFCHANGESPEC-011` `pkg/service` and `cmd/theoryofchange` (Cobra) with validate, graph, schema, init, fmt, and patch commands
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-009`, `RMI-THEORYOFCHANGESPEC-046`
  - Satisfies: PRD US-1, PRD US-16, TRD §6.1
  - Acceptance: CLI tests cover exit codes and output formats; `fmt --check` exits 1 on unformatted input
- [ ] `RMI-THEORYOFCHANGESPEC-047` Element commands (`add` / `set` / `rm --cascade`) compiled to patches for core element kinds
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-046`, `RMI-THEORYOFCHANGESPEC-011`
  - Satisfies: PRD-FR-703, PRD US-16, TRD-030
  - Acceptance: each command equals its patch equivalent; `rm` without `--cascade` rejects dangling references

## Phase 3 — Pathways, Metric Library, Provenance, and Reference Examples

**Theme:** Program pathways, shared metric definitions, a glossary, and provenance on every element, proven on fictional reference examples
**Status:** Planned — 0 of 4 items completed

- [ ] `RMI-THEORYOFCHANGESPEC-032` Pathways (program areas) and parent / child ToC references, with element commands
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-008`, `RMI-THEORYOFCHANGESPEC-047`
  - Satisfies: PRD-FR-112, PRD US-9, TRD-017
  - Acceptance: pathway round-trip; child ToC outcomes resolve against the parent
- [ ] `RMI-THEORYOFCHANGESPEC-033` Metric library, glossary, and indicator extensions (metricId, pathwayIds, claim, caveats, rollupFrom, theme_count), with element commands
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-032`
  - Satisfies: PRD-FR-113, PRD-FR-114, PRD-FR-117, PRD US-10, TRD-017
  - Acceptance: roll-up declarations validated against metric aggregation; claim defaults to contribution when omitted
- [ ] `RMI-THEORYOFCHANGESPEC-052` Provenance on every element (stated / inferred / hypothesis / confirmed / rejected, cited evidence), indicator dataAvailability, and Consultation records; rejected elements excluded from analysis
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-033`
  - Satisfies: PRD-FR-124, PRD-FR-125, PRD-FR-123, TRD-036
  - Acceptance: round-trip tests; rejected elements absent from graph, measurement, and default reports
- [ ] `RMI-THEORYOFCHANGESPEC-010` Reference examples (fictional) with provenance: intermediary organization (three pathways, two impact layers, metric library, glossary) and market adoption (open agent-authorization standard, AAER categories)
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-033`, `RMI-THEORYOFCHANGESPEC-052`
  - Satisfies: PRD G-1, PRD G-9, PRD US-1
  - Acceptance: both examples validate; exercise groups, feedback links, qualitative indicators, and every provenance basis

## Phase 4 — Workshop Pack and Source Drafting

**Theme:** Draft a theory from public sources and render the pathway diagram, logic models, metric definitions, glossary, CSV exports, and a Marp workshop deck from the document alone — adopter milestone M0
**Status:** Planned — 0 of 6 items completed

- [ ] `RMI-THEORYOFCHANGESPEC-022` Versioned `report.View` model built from document and graph, status optional, with basis badges
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-009`, `RMI-THEORYOFCHANGESPEC-010`
  - Satisfies: PRD-FR-501, PRD-FR-502, PRD-FR-505, TRD-010
  - Acceptance: view tests with and without a status snapshot; inferred elements carry badges
- [ ] `RMI-THEORYOFCHANGESPEC-023` Mermaid pathway renderer with level subgraphs, optional health styling, feedback links, ceiling, and pathway filter
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`
  - Satisfies: PRD-FR-501, PRD G-12
  - Acceptance: golden Mermaid output
- [ ] `RMI-THEORYOFCHANGESPEC-037` Logic-model (per pathway), metric-definitions, and glossary reports
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`
  - Satisfies: PRD-FR-505, PRD US-12, TRD-022
  - Acceptance: golden files over the intermediary example
- [ ] `RMI-THEORYOFCHANGESPEC-049` `pkg/slides` Marp engine and workshop deck (`theoryofchange deck workshop`)
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-023`, `RMI-THEORYOFCHANGESPEC-037`
  - Satisfies: PRD-FR-801, PRD G-15, PRD US-17, TRD-033
  - Acceptance: golden workshop deck; deck numbers and labels equal the reports'
- [ ] `RMI-THEORYOFCHANGESPEC-039` `theoryofchange report` definition kinds, `definitions` pack, and `theoryofchange export csv` (indicators, metrics, glossary)
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-023`, `RMI-THEORYOFCHANGESPEC-037`, `RMI-THEORYOFCHANGESPEC-011`
  - Satisfies: PRD-FR-204, PRD-FR-505, PRD G-12, TRD-021
  - Acceptance: CLI tests; CSV golden files (RFC 4180)
- [ ] `RMI-THEORYOFCHANGESPEC-053` `theoryofchange evidence add` and the toc-draft-from-sources agent skill
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-047`, `RMI-THEORYOFCHANGESPEC-052`, `RMI-THEORYOFCHANGESPEC-039`
  - Satisfies: PRD-FR-606, PRD US-19, PRD G-17, TRD-037
  - Acceptance: smoke test over saved fixture pages yields a draft where every element has a basis and every `stated` element cites a source

## Phase 5 — Review and Agreement

**Theme:** Assessment report, Pandoc-ready sharing exports with stamps, review packet and sheet, review import as proposals, and approval of the agreed baseline — adopter milestone M1
**Status:** Planned — 0 of 5 items completed

- [ ] `RMI-THEORYOFCHANGESPEC-054` Assessment report: provenance mix, gaps, data availability, assumptions needing testing, generated questions, short narrative
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`, `RMI-THEORYOFCHANGESPEC-052`
  - Satisfies: PRD-FR-512, PRD US-19, TRD-038
  - Acceptance: golden file over the intermediary example
- [ ] `RMI-THEORYOFCHANGESPEC-056` Pandoc Markdown profile with PNG diagrams, export stamps on every artifact, and `theoryofchange export docx|pptx`
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-039`, `RMI-THEORYOFCHANGESPEC-049`, `RMI-THEORYOFCHANGESPEC-023`
  - Satisfies: PRD-FR-901, PRD-FR-902, TRD-039
  - Acceptance: golden Markdown; stamp present in every format; pandoc export test skipped when absent
- [ ] `RMI-THEORYOFCHANGESPEC-057` Review packet (ID-tagged response tables) and review sheet (CSV) exports
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-054`, `RMI-THEORYOFCHANGESPEC-056`
  - Satisfies: PRD-FR-903, PRD US-20, TRD-039
  - Acceptance: golden packet and sheet; every element row carries its stable ID
- [ ] `RMI-THEORYOFCHANGESPEC-058` ReviewResponse schema and `theoryofchange import review` (JSON and --from-csv) producing proposed patches and a consultation record
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-057`, `RMI-THEORYOFCHANGESPEC-046`
  - Satisfies: PRD-FR-904, PRD-FR-905, PRD US-20, TRD-040
  - Acceptance: seeded rows from every origin map to ops or flagged notes; stale hashes flagged; document never modified
- [ ] `RMI-THEORYOFCHANGESPEC-055` `theoryofchange approve` with approvers, revision, version label, and refusal checks
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-058`, `RMI-THEORYOFCHANGESPEC-052`
  - Satisfies: PRD-FR-707, PRD US-21, TRD-041
  - Acceptance: approval sets status and revision through the patch engine; refuses on errors or open hypotheses without --force

## Phase 6 — Measurement Framework Design

**Theme:** Metric selection and collection burden, audiences and stories, UNDG change analysis, and the full intermediary example — adopter milestone M1.5 (collection plan)
**Status:** Planned — 0 of 5 items completed

- [ ] `RMI-THEORYOFCHANGESPEC-034` Metric selection criteria and scores, collection cost, measurement budget, and `pkg/burden` totals
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-033`
  - Satisfies: PRD-FR-115, PRD G-10, PRD US-11, TRD-020
  - Acceptance: burden arithmetic table tests (minutes per week, questions per respondent per year, per pathway and mode)
- [ ] `RMI-THEORYOFCHANGESPEC-035` Audiences, theme codebook, and Story / LearningReview observation types
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-033`
  - Satisfies: PRD-FR-116, PRD-FR-117, PRD-FR-118, TRD-017, TRD-018
  - Acceptance: unknown theme IDs and missing consent rejected by validation
- [ ] `RMI-THEORYOFCHANGESPEC-044` Change-analysis model: problem-tree causes, outcome responsibility and delivering actors, prioritization criteria for outcomes and interventions, target groups, assumption and risk enrichment
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-032`, `RMI-THEORYOFCHANGESPEC-034`
  - Satisfies: PRD-FR-105, PRD-FR-108, PRD-FR-119…122, TRD-025
  - Acceptance: round-trip tests; unknown cause / target-group / actor references rejected by validation
- [ ] `RMI-THEORYOFCHANGESPEC-036` Intermediary example extended with framework-design and change-analysis data; schemas regenerated lint-clean
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-034`, `RMI-THEORYOFCHANGESPEC-035`, `RMI-THEORYOFCHANGESPEC-044`
  - Satisfies: PRD G-9 success metric
  - Acceptance: example with 8–12 selected metrics, budget, audiences, codebook, sample stories, problem tree, partner-delivered outcomes, and consultations validates; `schemakit lint` 0 violations
- [ ] `RMI-THEORYOFCHANGESPEC-038` Metric-selection and collection-plan reports and CSV exports
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`, `RMI-THEORYOFCHANGESPEC-034`, `RMI-THEORYOFCHANGESPEC-039`
  - Satisfies: PRD-FR-505, PRD US-11, PRD US-12, TRD-022
  - Acceptance: golden files; collection-plan totals equal `pkg/burden` results

## Phase 7 — Quality: Lint, Rubric, and Judging

**Theme:** Deterministic lint with profiles, the LLM-as-a-Judge rubric with a development profile, and reproducible file-based judging with multi-judge agreement
**Status:** Planned — 0 of 6 items completed

- [ ] `RMI-THEORYOFCHANGESPEC-012` Lint engine: rule registry, draft / standard / full profiles, options, structured-evaluation SummaryReport output
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-009`
  - Satisfies: PRD-FR-303, PRD-FR-304, PRD-FR-804, TRD-008, TRD-035
  - Acceptance: registry and report-shape tests; minimal fixture is GO under `draft`
- [ ] `RMI-THEORYOFCHANGESPEC-013` Structure and logic rules (TOC-STR-*, TOC-LOG-*) with seeded-defect fixtures
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-012`
  - Satisfies: PRD-FR-303, TRD §6.4
  - Acceptance: each rule triggered by exactly one fixture; reference examples report 0 errors
- [ ] `RMI-THEORYOFCHANGESPEC-014` Assumption, measurement, intervention, accountability, actor, evidence, pathway, metric, claim, budget, audience, cause, responsibility, risk, equity, validation, and provenance rules with seeded-defect fixtures
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-012`, `RMI-THEORYOFCHANGESPEC-036`
  - Satisfies: PRD-FR-303, PRD US-10, PRD US-11, TRD §6.4
  - Acceptance: one fixture per rule; ≥ 15 rules total implemented and tested
- [ ] `RMI-THEORYOFCHANGESPEC-015` ToC spec template, core RubricSet (sector-neutral UNDG checklist plus plausible / feasible / testable), and development rubric profile for LLM-as-a-Judge, embedded in `spec/`
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-004`
  - Satisfies: PRD-FR-601, TRD-012, TRD-028
  - Acceptance: both rubrics parse into `rubric.RubricSet`, `Validate()` passes, weights sum to 1.0; every UNDG checklist item maps to a criterion
- [ ] `RMI-THEORYOFCHANGESPEC-050` `pkg/judge` and `theoryofchange judge show | record | summary`: bundle hashing, verdict validation, multi-judge agreement
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-015`, `RMI-THEORYOFCHANGESPEC-039`
  - Satisfies: PRD-FR-803, PRD G-16, PRD US-18, TRD-034
  - Acceptance: bundle golden; verdicts with mismatched bundle hash rejected; aggregation reports per-category agreement
- [ ] `RMI-THEORYOFCHANGESPEC-016` `theoryofchange lint` CLI and rubric discrimination eval with at least two judges and agreement, recorded in `docs/evals/`
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-013`, `RMI-THEORYOFCHANGESPEC-014`, `RMI-THEORYOFCHANGESPEC-050`
  - Satisfies: PRD US-2, PRD success metrics "LLM-judge discrimination" and "Judge agreement recorded"
  - Acceptance: reference example passes; seeded weak ToC fails ≥ 3 rubric categories; agreement reported

## Phase 8 — Measurement

**Theme:** Turn spreadsheet data and observations into indicator, outcome, pathway, and assumption status with reproducible roll-ups and confidence — adopter milestone M2
**Status:** Planned — 0 of 6 items completed

- [ ] `RMI-THEORYOFCHANGESPEC-017` Indicator progress, expected progress, status (including baseline_only), and health mapping
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-008`, `RMI-THEORYOFCHANGESPEC-033`
  - Satisfies: PRD-FR-401, PRD-FR-402, PRD-FR-509, TRD-009, TRD-019
  - Acceptance: table tests cover every status branch
- [ ] `RMI-THEORYOFCHANGESPEC-018` Outcome, pathway-status, and cross-pathway metric roll-ups through precondition logic and nested groups
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-009`, `RMI-THEORYOFCHANGESPEC-017`
  - Satisfies: PRD-FR-403, PRD-FR-113, TRD-019
  - Acceptance: tests for all / any / at_least / nested groups and each metric aggregation
- [ ] `RMI-THEORYOFCHANGESPEC-019` Assumption propagation and link / path confidence scoring
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-018`
  - Satisfies: PRD-FR-404, PRD-FR-405
  - Acceptance: invalidated assumption flags every downstream outcome (US-5); confidence tests with documented weights
- [ ] `RMI-THEORYOFCHANGESPEC-020` `measure.Evaluate` status snapshot, theme counts, and `Diff` between periods; determinism
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-019`, `RMI-THEORYOFCHANGESPEC-035`
  - Satisfies: PRD-FR-406, TRD INV-3
  - Acceptance: golden snapshots byte-identical across two runs
- [ ] `RMI-THEORYOFCHANGESPEC-040` CSV observation import with mapping file and unmatched-row reporting
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-007`, `RMI-THEORYOFCHANGESPEC-033`
  - Satisfies: PRD-FR-204, PRD US-13, TRD-021
  - Acceptance: seeded bad rows all reported; 0 silent drops
- [ ] `RMI-THEORYOFCHANGESPEC-021` `theoryofchange status` and `theoryofchange import` CLI; performance benchmarks
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-020`, `RMI-THEORYOFCHANGESPEC-040`, `RMI-THEORYOFCHANGESPEC-011`
  - Satisfies: PRD US-4, PRD US-13, TRD §8
  - Acceptance: CLI tests; 200-node benchmark under 1 s

## Phase 9 — Reporting, Story, and Presentations

**Theme:** Executive, audience story, learning-review, portfolio, narrative, and division-of-labour reports, plus exec / learning-review / story decks with optional export — adopter milestone M3
**Status:** Planned — 0 of 8 items completed

- [ ] `RMI-THEORYOFCHANGESPEC-024` Status reports: exec, results, assumptions, interventions, evidence-gaps, full
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-020`, `RMI-THEORYOFCHANGESPEC-022`, `RMI-THEORYOFCHANGESPEC-023`
  - Satisfies: PRD-FR-501, PRD US-3
  - Acceptance: golden files; every reported status cites its inputs
- [ ] `RMI-THEORYOFCHANGESPEC-041` Audience story report with consent-gated quotes
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-024`, `RMI-THEORYOFCHANGESPEC-035`
  - Satisfies: PRD-FR-506, PRD US-14, TRD-023
  - Acceptance: golden files; no-consent and restricted quotes absent
- [ ] `RMI-THEORYOFCHANGESPEC-042` Learning-review and portfolio reports
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-024`, `RMI-THEORYOFCHANGESPEC-034`
  - Satisfies: PRD-FR-507, PRD-FR-508, PRD US-14
  - Acceptance: golden files over two observation periods
- [ ] `RMI-THEORYOFCHANGESPEC-045` Narrative (programme / strategy document scaffold) and division-of-labour reports; accountable vs monitored-only separation
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`, `RMI-THEORYOFCHANGESPEC-044`
  - Satisfies: PRD-FR-510, PRD-FR-511, PRD-FR-120, TRD-026, TRD-027
  - Acceptance: golden files over a mixed-responsibility fixture
- [ ] `RMI-THEORYOFCHANGESPEC-060` Projection artifacts: outcome map (outcomes framework), logframe matrix, and logic-model diagram, as definition reports and in the definitions pack
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-022`, `RMI-THEORYOFCHANGESPEC-044`
  - Satisfies: PRD-FR-513
  - Acceptance: golden files; every row names the element IDs it was projected from; a one-pathway theory and a multi-pathway theory both render
- [ ] `RMI-THEORYOFCHANGESPEC-025` Custom report templates and restricted-evidence suppression
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-024`
  - Satisfies: PRD-FR-502, PRD-FR-503, TRD-011
  - Acceptance: custom template test; restricted excerpt absent from all outputs
- [ ] `RMI-THEORYOFCHANGESPEC-051` Exec, learning-review, and story decks; `--export pdf|pptx|html` via optional `mmdc` and `marp`
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-049`, `RMI-THEORYOFCHANGESPEC-024`, `RMI-THEORYOFCHANGESPEC-041`, `RMI-THEORYOFCHANGESPEC-042`
  - Satisfies: PRD-FR-801, PRD-FR-802, PRD US-17, TRD-033
  - Acceptance: golden decks; export test produces files when tools are present and exits 3 with guidance when absent
- [ ] `RMI-THEORYOFCHANGESPEC-026` `theoryofchange report` status kinds in CLI and executive-report benchmark
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-025`, `RMI-THEORYOFCHANGESPEC-041`, `RMI-THEORYOFCHANGESPEC-042`, `RMI-THEORYOFCHANGESPEC-045`, `RMI-THEORYOFCHANGESPEC-021`
  - Satisfies: PRD G-5
  - Acceptance: CLI tests; exec report for 200-node ToC under 1 s

## Phase 10 — Agents and Revisions

**Theme:** MCP tools (including validated patching), agent skills, semantic theory diffs, and automated Google sharing — adopter milestone M4
**Status:** Planned — 0 of 4 items completed

- [ ] `RMI-THEORYOFCHANGESPEC-027` MCP server exposing schema, validate, lint, graph, status, report, patch, deck, and judge-bundle tools with root confinement
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-026`, `RMI-THEORYOFCHANGESPEC-046`, `RMI-THEORYOFCHANGESPEC-050`
  - Satisfies: PRD-FR-602, PRD-FR-706, PRD US-6, TRD-013
  - Acceptance: in-memory transport test drives every tool; path traversal rejected; `toc_patch` rejects error-introducing patches
- [ ] `RMI-THEORYOFCHANGESPEC-048` `pkg/diff` semantic theory diff and `theoryofchange diff --revision`
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-046`, `RMI-THEORYOFCHANGESPEC-009`
  - Satisfies: PRD-FR-705, PRD US-16, PRD US-21, TRD-032
  - Acceptance: golden diffs keyed by element ID, including rejected elements; draft Revision lists changed IDs
- [ ] `RMI-THEORYOFCHANGESPEC-043` Agent skills: toc-author (edits via patches), toc-status-refresh, toc-learning-review, toc-audience-story
  - Type: capability · Status: planned
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
**Status:** Planned — 0 of 4 items completed

- [ ] `RMI-THEORYOFCHANGESPEC-028` `pkg/prismx` OKR and outcome-journey export with import-boundary test
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-020`
  - Satisfies: PRD-FR-603, PRD US-7, TRD-014
  - Acceptance: exported OKR documents pass prism-roadmap validation; core packages do not import prism-roadmap
- [ ] `RMI-THEORYOFCHANGESPEC-029` Actor registry resolution, cross-document references, and competing-futures fixture
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-008`, `RMI-THEORYOFCHANGESPEC-032`
  - Satisfies: PRD G-8, PRD US-8, TRD §7.3
  - Acceptance: three competing-future ToCs sharing one actor registry validate with no schema changes
- [ ] `RMI-THEORYOFCHANGESPEC-030` USM/UFM integration contract document
  - Type: capability · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-029`
  - Satisfies: PRD G-8
  - Acceptance: contract reviewed by USM/UFM owners
- [ ] `RMI-THEORYOFCHANGESPEC-031` Documentation, CHANGELOG, v0.1.0 release, and spec-type sync follow-up for specification-workflow-spec
  - Type: release · Status: planned
  - Depends on: `RMI-THEORYOFCHANGESPEC-027`, `RMI-THEORYOFCHANGESPEC-043`, `RMI-THEORYOFCHANGESPEC-048`, `RMI-THEORYOFCHANGESPEC-059`, `RMI-THEORYOFCHANGESPEC-028`, `RMI-THEORYOFCHANGESPEC-030`, `RMI-THEORYOFCHANGESPEC-051`
  - Satisfies: all PRD success metrics (release gate)
  - Acceptance: CI green before tag; publication scan clean; follow-up RMI filed in specification-workflow-spec
