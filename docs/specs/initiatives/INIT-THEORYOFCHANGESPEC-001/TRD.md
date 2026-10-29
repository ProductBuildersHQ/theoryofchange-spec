# TRD — Theory of Change Specification — Go-first model, JSON Schema, measurement, and leadership reporting

**Initiative:** `INIT-THEORYOFCHANGESPEC-001`
**Repository:** `github.com/ProductBuildersHQ/theoryofchange-spec`
**Version:** 0.1
**Status:** Draft

This TRD is a technical contract: what the module MUST guarantee. MUST / MUST
NOT / SHOULD / MAY are used in the RFC 2119 sense. Conditions that do not
apply are answered `Not applicable — <rationale>`.

## 1. Introduction

### 1.1 Purpose

Define the Go module `github.com/ProductBuildersHQ/theoryofchange-spec`: a
Go-struct-first Theory of Change (ToC) model, its generated JSON Schemas, and
the library, CLI, and MCP tooling that validate, analyze, measure, judge, and
report on ToC documents. This TRD lets implementation proceed phase by phase
(PLAN.md) without inventing product decisions.

### 1.2 Scope

In scope: data model, schema pipeline, validation, causal-graph analysis,
deterministic lint, measurement roll-up, report rendering, spec template and
LLM-judge rubric, MCP server, PRISM adapters, and the USM/UFM integration
contract.

Out of scope: see PRD §2.2 (no USM/UFM implementation, no UI, no LLM calls,
no persistence, no statistical inference).

### 1.3 References

| Document | Link |
|----------|------|
| PRD | [PRD.md](PRD.md) |
| PLAN | [PLAN.md](PLAN.md) |
| ROADMAP | [ROADMAP.md](ROADMAP.md) |
| Research brief | [docs/research/theory-of-change.md](../../../research/theory-of-change.md) |
| prism-core | `github.com/grokify/prism-core` (Person, Metadata, Risk, health, trend, CapabilityRef) |
| prism-roadmap | `github.com/grokify/prism-roadmap` (OKR, OutcomeJourney) |
| structured-evaluation | `github.com/plexusone/structured-evaluation` (RubricSet, SummaryReport) |
| market-spec | `github.com/ProductBuildersHQ/market-spec` (EntityRef convention, actor/market nouns) |

## 2. Technical Requirements and Traceability

| TRD ID | Requirement | Source | Verification |
|--------|-------------|--------|--------------|
| TRD-001 | `pkg/toc` MUST define `Document` and all element types in §7 with camelCase JSON tags | PRD-FR-101…110 | Round-trip marshal tests; schema lint |
| TRD-002 | `pkg/toc` MUST define `ObservationSet` as a separate document type | PRD-FR-111 | Round-trip test; schema test |
| TRD-003 | `cmd/generate-schema` MUST be the only producer of `schema/*.schema.json`; CI MUST fail if regenerating changes any file | PRD-FR-201 | `go generate ./schema && git diff --exit-code schema/` |
| TRD-004 | `schema` package MUST embed every generated schema and expose `Get(name)` and `Names()` | PRD-FR-202 | `schema_test.go` loads every name |
| TRD-005 | `pkg/tocio` MUST load `.json`, `.yaml`, `.yml` by converting YAML → JSON before decoding and validating | PRD-FR-203 | Same fixture in both formats yields equal `Document` and equal issues |
| TRD-006 | `pkg/validate` MUST report schema, uniqueness, ID-format, and reference errors with JSON-pointer paths | PRD-FR-301 | Table-driven tests per error class |
| TRD-007 | `pkg/graph` MUST implement the queries in §6.2 in O(V+E) per query (paths enumeration excepted, bounded by `MaxPaths`) | PRD-FR-302 | Unit tests; benchmark on 1,000-node graph |
| TRD-008 | `pkg/lint` MUST implement the rules in §6.4 with stable IDs and emit a structured-evaluation `summary.SummaryReport` | PRD-FR-303, 304 | Rule registry test; seeded-defect golden test |
| TRD-009 | `pkg/measure` MUST implement §6.5 algorithms deterministically for a given `asOf` | PRD-FR-401…406 | Table tests; golden status snapshots |
| TRD-010 | `pkg/report` MUST render the standard reports in §6.6 and accept user templates over `report.View` | PRD-FR-501, 502, 504 | Golden-file tests, run twice |
| TRD-011 | Renderers MUST NOT emit `restricted` evidence excerpts | PRD-FR-503 | Test with restricted evidence asserts absence |
| TRD-012 | `spec/` MUST embed `toc.md` template and `toc.rubric.yaml` that parses into `rubric.RubricSet` and passes its `Validate()` | PRD-FR-601 | `spec_test.go` |
| TRD-013 | `pkg/mcp` MUST expose the six tools in §6.3 over stdio with the official Go SDK | PRD-FR-602 | In-memory transport integration test |
| TRD-014 | `pkg/prismx` MUST export OKR documents and outcome journeys; no package under `pkg/toc`, `pkg/graph`, `pkg/validate`, `pkg/lint`, `pkg/measure` MAY import `prism-roadmap` | PRD-FR-603 | `go list -deps` import-boundary test |
| TRD-015 | `pkg/ref` MUST parse and format the reference grammar in §6.7 | PRD-FR-604 | Parse/format round-trip and invalid-input tests |
| TRD-016 | Core types MUST reuse `prism-core` `Person`, `Risk`, `CapabilityRef`, health and trend constants instead of redefining them | PRD G-7 | Code review; grep test for duplicate type names |
| TRD-017 | `pkg/toc` MUST define `Pathway`, `Metric`, `Audience`, `GlossaryTerm`, `Theme`, `SelectionCriterion`, `MeasurementBudget`, and the indicator extensions in §7.1 | PRD-FR-112…117 | Round-trip tests; schema lint |
| TRD-018 | `ObservationSet` MUST carry `stories`, `learningReviews`, and per-observation `quality` (§7.2) | PRD-FR-118 | Round-trip tests; validation of theme and consent rules |
| TRD-019 | `pkg/measure` MUST compute metric roll-ups across pathways per §6.5 and return `baseline_only` for indicators without targets | PRD-FR-113, 509 | Table tests |
| TRD-020 | `pkg/burden` MUST total collection cost per period and per respondent and compare against `MeasurementBudget` | PRD-FR-115 | Unit tests; collection-plan golden |
| TRD-021 | `pkg/tabular` MUST import observations from CSV via a mapping file and export definition tables as CSV (RFC 4180) | PRD-FR-204 | Import tests with seeded bad rows; export golden files |
| TRD-022 | `pkg/report` MUST render the definition reports with no observations and the story, learning-review, and portfolio reports in §6.6 | PRD-FR-505…508 | Golden-file tests |
| TRD-023 | Story quotes MUST be rendered only when `consent` is `granted` and sensitivity is renderable | PRD-FR-506 | Restricted / no-consent absence tests |
| TRD-024 | `skills/` MUST contain the agent skills in §6.11; each skill's command sequence MUST be executable by a smoke test | PRD-FR-605 | Skill smoke test runs commands twice and compares outputs |
| TRD-025 | `pkg/toc` MUST define `Cause`, `Consultation`, `TargetGroup`, outcome `responsibility` / `deliveredBy` / `addressesCauses`, `priority` scores, and the assumption and risk fields in §7.1 | PRD-FR-105, 108, 119…123 | Round-trip tests; schema lint |
| TRD-026 | Measurement MUST monitor every outcome regardless of responsibility, and reports MUST separate accountable (`own`, `shared`) from monitored-only (`partner`, `unaddressed`) outcomes | PRD-FR-120 | Status and report tests over mixed-responsibility fixture |
| TRD-027 | `pkg/report` MUST render the `narrative` and `actors` reports in §6.6 | PRD-FR-510, 511 | Golden files |
| TRD-028 | `spec/` MUST ship the core rubric plus a `development` rubric profile covering the UNDG checklist split in research §9.6 | PRD-FR-601 | Both parse and `Validate()`; checklist-coverage test maps every checklist item to a criterion ID |
| TRD-029 | `pkg/edit` MUST implement the patch engine in §6.12 (RFC 6902 + ID addressing, atomic apply, error-delta rejection) and be the only mutation path used by CLI, SDK helpers, and MCP | PRD-FR-701, 702, 706 | Engine tests; grep test that `cmd/` and `pkg/mcp` never mutate `toc.Document` directly |
| TRD-030 | Element commands MUST compile to patches; `rm` MUST report dangling references and honor `--cascade` | PRD-FR-703 | Command ↔ patch equivalence tests; reference tests |
| TRD-031 | `pkg/tocio` MUST provide canonical serialization (`Format`) for JSON and YAML; idempotent | PRD-FR-704 | Fixture test runs twice |
| TRD-032 | `pkg/diff` MUST compute an ID-keyed semantic diff and draft a `Revision` | PRD-FR-705 | Golden diffs |
| TRD-033 | `pkg/slides` MUST render Marp decks in §6.13 from `report.View`; export runs only in `cmd/toc` via external CLIs | PRD-FR-801, 802; INV-7 | Golden decks; export test skipped when CLIs absent |
| TRD-034 | `pkg/judge` MUST build judge bundles, validate verdicts as structured-evaluation reports, and aggregate with agreement metrics (§6.14) | PRD-FR-803 | Bundle golden; verdict validation tests; aggregation tests |
| TRD-035 | `pkg/lint` MUST support `draft`, `standard`, `full` profiles; the minimal fixture is GO under `draft` | PRD-FR-804 | Profile tests |
| TRD-036 | Every element type MUST embed `Provenance` (§7.1); indicators carry `dataAvailability`; `rejected` elements are excluded from graph, measurement, and default reports | PRD-FR-124, 125 | Round-trip; exclusion tests |
| TRD-037 | `toc evidence add` and the `toc-draft-from-sources` skill per §6.11 | PRD-FR-606, US-19 | Skill smoke test over saved fixture pages |
| TRD-038 | `assessment` report per §6.6 | PRD-FR-512 | Golden file |
| TRD-039 | `pkg/share` MUST produce the `pandoc` Markdown profile, review packet, review sheet, and export stamps (§6.15) | PRD-FR-901…903 | Golden files; stamp tests per format |
| TRD-040 | `pkg/review` MUST validate `ReviewResponse` and produce a proposed patch plus `Consultation` without modifying the document (§6.15) | PRD-FR-904, 905 | Import tests incl. stale hash |
| TRD-041 | `toc approve` per §6.15 is the only writer of status `approved` | PRD-FR-707, US-21 | Approval tests; grep test |
| TRD-042 | `skills/toc-share-google` per §6.11, with manual fallback | PRD-FR-607 | Skill smoke test with recorded tool outputs |

## 3. System Invariants

- **INV-1** Every element ID MUST be unique within a document across all
  element kinds (outcomes, links, interventions, assumptions, indicators,
  actors, evidence, factors, risks).
- **INV-2** No library function MAY mutate its input `Document` or
  `ObservationSet`; derived data is returned in new values.
- **INV-3** For identical inputs (documents, observation sets, `asOf`,
  options), every computed output (issues, status snapshot, report bytes) MUST
  be byte-identical. Map iteration MUST NOT influence output order; ordering is
  by document order, then ID.
- **INV-4** Unknown references MUST be reported, never silently dropped.
- **INV-5** A score shown to a reader MUST be accompanied by its inputs or a
  link to them (indicator observations, assumption assessments, evidence IDs).
- **INV-6** `restricted` evidence MUST NOT be rendered verbatim in any output
  intended for humans (Markdown, HTML, Mermaid labels).
- **INV-7** The library MUST NOT perform network I/O or execute external
  programs. Only `cmd/` and `pkg/mcp` read or write files, and only paths
  given by the caller; only `cmd/toc` may invoke the optional external
  export tools (`marp`, `mmdc`, `pandoc`). The module has no Google or other
  cloud API dependency.
- **INV-10** Review imports never modify a document; they produce proposals
  that are applied only through `pkg/edit` (INV-9).
- **INV-8** Schema files MUST only change via `cmd/generate-schema`.
- **INV-9** Every change to a ToC document made by this module goes through
  `pkg/edit`; no other code path writes a modified document.

## 4. Architecture

### 4.1 High-Level Architecture

```
            ┌──────────────── authoring ────────────────┐
  human ──► toc.yaml / toc.json      observations-*.yaml ◄── analyst
  agent ──► (via MCP tools)                │
            └───────────────┬──────────────┘
                            ▼
   pkg/tocio  load (YAML→JSON), decode, schema-validate (schema/ embed)
                            │
          ┌─────────────────┼──────────────────────────┐
          ▼                 ▼                          ▼
   pkg/validate        pkg/graph ◄──────────── pkg/lint (rules use graph)
   (refs, IDs)         (paths, cycles,          → structured-evaluation
                        levels, ceiling)          SummaryReport (GO/NO-GO)
                            │
                            ▼
                      pkg/measure  (indicator → outcome → pathway,
                            │       assumption propagation, confidence)
                            ▼
                      toc.Status snapshot (JSON, schema-defined)
                            │
              ┌─────────────┼──────────────┐
              ▼             ▼              ▼
         pkg/report     pkg/prismx      spec/ (template + RubricSet
         (MD, Mermaid,  (OKR, outcome     for LLM-as-a-Judge, run by
          tables, user   journey export)  visionspec / visionstudio)
          templates)
              ▲
   thin adapters: cmd/toc (Cobra CLI) and pkg/mcp (MCP stdio server)
   both call one service layer: pkg/service
```

### 4.2 Component Responsibilities

| Component | Purpose | Responsibilities |
|-----------|---------|------------------|
| `pkg/toc` | Data model | Go structs, enums with `Valid()`/`All…()` helpers, small accessors (`OutcomeByID`, `IndicatorsFor`). No I/O. |
| `pkg/ref` | References | Grammar in §6.7: parse, format, validate slugs. |
| `schema/` | Schemas | `//go:embed *.schema.json`; `Get`, `Names`, `Validate(name, []byte)` using `santhosh-tekuri/jsonschema/v6`. |
| `cmd/generate-schema` | Schema pipeline | Reflect Go types with `invopop/jsonschema`; write files; deterministic output. |
| `pkg/tocio` | Load / save | File and byte loading for JSON/YAML; merge multiple observation sets; pretty JSON writer. |
| `pkg/validate` | Structural validation | Schema + INV-1 + reference resolution + enum validity → `[]Issue`. |
| `pkg/graph` | Causal graph | Build from links + intervention targets; queries in §6.2. |
| `pkg/lint` | Quality rules | Rule registry, §6.4 rules, options (disable / severity override), SummaryReport output. |
| `pkg/measure` | Measurement | §6.5 algorithms; produces `toc.Status`; `Diff(prev, cur)`. |
| `pkg/report` | Rendering | View model, standard templates (embedded), Mermaid builder, user templates. |
| `pkg/prismx` | PRISM adapters | `ToOKR`, `ToOutcomeJourneys`; prism-roadmap import is confined here. |
| `pkg/service` | Service layer | One function per user operation (Validate, Lint, Graph, Status, Report, Export, Schema) used by CLI and MCP. |
| `pkg/mcp` | MCP server | Tool registration and JSON I/O over `pkg/service`. |
| `spec/` | Spec content | `toc.md` template, `toc.rubric.yaml`, judge instructions; embedded. |
| `cmd/toc` | CLI | Cobra commands in §6.1; no business logic. |
| `pkg/burden` | Collection burden | Staff-minute and respondent-question totals per period, pathway, and mode; budget comparison. |
| `pkg/tabular` | Spreadsheet interop | CSV observation import via mapping; CSV export of definition tables. |
| `skills/` | Agent skills | `SKILL.md` instruction files plus templates; every computation delegated to `toc` CLI / MCP tools. |
| `pkg/edit` | Mutations | Patch engine (RFC 6902 + ID addressing), element-command builders, reference cascade. The only way documents change. |
| `pkg/diff` | Theory diff | ID-keyed semantic diff between two documents; draft `Revision`. |
| `pkg/slides` | Presentations | Marp Markdown decks from `report.View`; Mermaid sources embedded or referenced as SVG paths. |
| `pkg/share` | Sharing exports | `pandoc` Markdown profile, review packet, review sheet, export stamps; DOCX/PPTX conversion invoked only by `cmd/toc`. |
| `pkg/review` | Review import | `ReviewResponse` validation, mapping rows to proposed patch ops and flagged notes, `Consultation` record. |
| `pkg/judge` | LLM-as-a-Judge I/O | Judge bundle assembly, verdict validation (structured-evaluation), multi-judge aggregation and agreement. No model calls. |
| `examples/` | Reference ToCs | Intermediary (multi-pathway, two-layer) example, market-adoption example, competing-futures fixture. |

### 4.3 Data Flow

1. Load: `tocio.LoadDocument(path)` → bytes → (YAML→JSON) → schema validate →
   `json.Unmarshal` into `toc.Document`.
2. Validate: `validate.Document(doc)` → issues. Lint and measure refuse to run
   on documents with `error`-severity validation issues (they return
   `ErrInvalidDocument` wrapping the issues).
3. Graph: `graph.Build(doc)`; reused by lint, measure, report.
4. Measure: `measure.Evaluate(doc, g, obsSets, opts)` → `toc.Status`.
5. Report: `report.Build(doc, g, status, opts)` → `report.View` → template →
   bytes.

## 5. State and Lifecycle

**Applicability:** Applicable.

| Entity | States | Trigger | Invalid Transitions |
|--------|--------|---------|---------------------|
| Document (`metadata.status`) | `draft` → `in_review` → `approved` → `active` → `archived` (prism-core status constants) | Author edits field | None enforced by library (git is the audit trail); lint warns when `active` documents have open `error` findings |
| Intervention | `planned`, `in_progress`, `completed`, `paused`, `cancelled` | Author edits | None enforced |
| Assumption | `untested`, `holding`, `at_risk`, `invalidated`, `validated` | Latest `AssumptionAssessment` in observation sets overrides the document's static `status` | None enforced; history kept in observation sets |
| Indicator status | Computed: `achieved`, `on_track`, `at_risk`, `off_track`, `not_started`, `no_data`, `stale`, `baseline_only` | `measure.Evaluate` | Not stored by authors |
| Indicator selection | `candidate` → `selected` → `retired` (and `candidate` → `retired`) | Author edits | `retired` → `selected` requires a `Revision` entry (lint info) |

- **Idempotency:** All operations are pure functions of their inputs; re-running is safe.
- **Concurrency:** Not applicable — no shared mutable state; files are owned by git.
- **Retention / deletion:** Not applicable — documents live in git; the revision log records intent.
- **Terminal and recovery states:** `archived` documents are still loadable and reportable.

## 6. Interface Contracts

### 6.1 CLI (`cmd/toc`)

| Command | Description | Exit codes |
|---------|-------------|------------|
| `toc validate <file>...` | Schema + structural validation (ToC or observation set auto-detected via `kind`) | 0 valid; 1 invalid; 2 usage / I/O error |
| `toc lint <toc> [--format table\|json] [--disable RULE] [--fail-on error\|warning]` | Deterministic quality rules; JSON is a `SummaryReport` | 0 GO; 1 NO-GO; 2 error |
| `toc graph <toc> [--format mermaid\|json\|dot] [--from ID] [--paths]` | Graph export and queries | 0; 2 |
| `toc status <toc> -o <obs.yaml>... [--as-of DATE] [--prev status.json]` | Compute status snapshot (JSON) and optional diff | 0; 2 |
| `toc report <kind> <toc> [-o <obs>...] [--as-of] [--pathway ID] [--audience ID] [--template file] [--out file]` | `kind` (status): `exec`, `results`, `assumptions`, `interventions`, `evidence-gaps`, `story`, `learning-review`, `portfolio`, `narrative`, `actors`, `full`, `custom`; `kind` (definition, no observations needed): `pathway`, `logic-model`, `metrics`, `metric-selection`, `collection-plan`, `glossary`, `definitions` (all definition reports) | 0; 2 |
| `toc import observations --csv <file> --map <mapping.yaml> --toc <toc> [--out obs.yaml]` | CSV → ObservationSet; unmatched rows to stderr and `--unmatched file` | 0 all rows imported; 1 some rows unmatched; 2 error |
| `toc export csv <table> <toc> [-o obs...]` | `table`: `indicators`, `metrics`, `metric-selection`, `collection-plan`, `glossary`, `results`, `themes` | 0; 2 |
| `toc export okr\|journey <toc> [-o obs...]` | PRISM exports | 0; 2 |
| `toc schema [name]` | List or print embedded schemas | 0; 2 |
| `toc init [--template program\|market]` | Write a starter document | 0; 2 |
| `toc mcp` | Start MCP server on stdio | 0 on clean shutdown |
| `toc patch <toc> <patch.json\|-> [--dry-run] [--allow-errors]` | Apply RFC 6902 patch atomically (§6.12) | 0 applied; 1 rejected; 2 error |
| `toc <kind> add\|set\|rm …` (`kind` ∈ outcome, group, link, intervention, assumption, indicator, metric, pathway, actor, evidence, cause, term) | Element commands compiled to patches; `rm --cascade`; all accept `--dry-run` | 0; 1 rejected; 2 |
| `toc fmt <file>... [--check] [--sort]` | Canonical serialization; `--check` exits 1 if a file would change | 0; 1; 2 |
| `toc diff <old> <new> [--format md\|json] [--revision]` | Semantic theory diff; optional draft `Revision` | 0; 2 |
| `toc deck <workshop\|exec\|learning-review\|story> <toc> [-o obs...] [--audience] [--export pdf\|pptx\|html] [--out file]` | Marp deck; export via external `marp` and `mmdc` when installed | 0; 2 (3 when export tool missing) |
| `toc judge show <toc> [--profile core\|development] [-o obs...]` | Emit judge bundle (§6.14) | 0; 2 |
| `toc judge record <toc> <verdict.json> [--judge NAME]` | Validate and store a verdict | 0; 1 invalid; 2 |
| `toc judge summary <toc>` | Aggregate stored verdicts with agreement | 0; 2 |
| `toc evidence add <toc> --url --title [--excerpt] [--accessed DATE] [--type] [--sensitivity]` | Element command for sources (compiles to a patch) | 0; 1; 2 |
| `toc export review-packet <toc> [--format md\|docx] [--reference-doc file]` | ID-tagged review document (§6.15) | 0; 2 (3 when `pandoc` missing for docx) |
| `toc export review-sheet <toc> [--out file.csv]` | Review rows as CSV | 0; 2 |
| `toc export docx\|pptx <kind> <toc> [-o obs...] [--reference-doc file]` | Any report (docx) or deck (pptx) through the `pandoc` profile | 0; 2; 3 |
| `toc import review <toc> <review.json>|--from-csv <sheet.csv> [--out proposed.patch.json] [--consultation-out c.json]` | Proposed patch + consultation; never writes the document | 0 all rows mapped; 1 rows need human decision or stale hash; 2 error |
| `toc approve <toc> --version LABEL --approver NAME[:ROLE]... [--note] [--force --reason]` | Sets `approved`, records approvers and revision | 0; 1 refused; 2 |

Global flags: `--as-of` (RFC 3339 date; default: today in UTC), `--format`,
`--quiet`. All machine output goes to stdout; diagnostics to stderr.

### 6.2 Library API (Go)

```go
// pkg/tocio
func LoadDocument(path string) (*toc.Document, error)
func ParseDocument(data []byte, format Format) (*toc.Document, error)
func LoadObservations(paths ...string) ([]*toc.ObservationSet, error)

// pkg/validate
type Severity string // "error" | "warning" | "info"
type Issue struct {
    Code     string   `json:"code"`     // e.g. "VAL-REF-001"
    Severity Severity `json:"severity"`
    Path     string   `json:"path"`     // JSON pointer, e.g. "/links/3/to"
    Element  string   `json:"element,omitempty"`
    Message  string   `json:"message"`
}
func Document(doc *toc.Document) []Issue
func Observations(doc *toc.Document, sets ...*toc.ObservationSet) []Issue

// pkg/graph
func Build(doc *toc.Document) (*Graph, error)
func (g *Graph) Upstream(id string) []string          // transitive preconditions
func (g *Graph) Downstream(id string) []string        // transitive consequences
func (g *Graph) PathsToImpact(id string, max int) [][]string
func (g *Graph) Cycles() [][]string                   // excludes feedback links
func (g *Graph) Levels() map[string]int               // longest distance to impact
func (g *Graph) Orphans() []string                    // no incoming link and no intervention
func (g *Graph) DeadEnds() []string                   // no path to any impact node
func (g *Graph) BelowCeiling(id string) bool          // within ceiling of accountability

// pkg/lint
type Rule interface { ID() string; Severity() validate.Severity; Check(*Context) []Finding }
func Run(doc *toc.Document, g *graph.Graph, opts Options) (*Result, error) // opts.Profile: draft|standard|full
func (r *Result) SummaryReport(project, version string) *summary.SummaryReport

// pkg/measure
type Options struct { AsOf time.Time; AtRiskTolerance float64; Rollup RollupMethod }
func Evaluate(doc *toc.Document, g *graph.Graph, obs []*toc.ObservationSet, opts Options) (*toc.Status, error)
func Diff(prev, cur *toc.Status) *toc.StatusDiff

// pkg/report
func BuildView(doc *toc.Document, g *graph.Graph, st *toc.Status, opts ViewOptions) (*View, error)
func Render(kind Kind, v *View) ([]byte, error)
func RenderTemplate(tmpl string, v *View) ([]byte, error)
func Mermaid(doc *toc.Document, st *toc.Status, opts MermaidOptions) string

// pkg/prismx
func ToOKR(doc *toc.Document, st *toc.Status) (*okr.OKRDocument, error)
func ToOutcomeJourneys(doc *toc.Document, st *toc.Status) ([]journey.OutcomeJourney, error)
```

Errors: functions return wrapped sentinel errors —
`ErrInvalidDocument`, `ErrUnknownReference`, `ErrCycle` (only from APIs that
require a DAG), `ErrUnsupportedFormat`. Callers use `errors.Is`.

### 6.3 MCP tools (`pkg/mcp`)

| Tool | Input | Output |
|------|-------|--------|
| `toc_schema` | `{ name? }` | schema JSON or list of names |
| `toc_validate` | `{ document \| path, kind? }` | `{ valid, issues[] }` |
| `toc_lint` | `{ document \| path, disable?[] }` | `SummaryReport` JSON |
| `toc_graph` | `{ document \| path, query: "paths"\|"upstream"\|"downstream"\|"cycles"\|"mermaid", id? }` | query result |
| `toc_status` | `{ document \| path, observations[] \| paths[], asOf? }` | `Status` JSON |
| `toc_report` | `{ document \| path, observations…, kind, asOf? }` | `{ format: "markdown", content }` |
| `toc_patch` | `{ document \| path, patch[], dryRun?, allowErrors? }` | `{ applied, document?, issuesIntroduced[], issuesResolved[] }`; with `path` and not `dryRun`, writes canonical file |
| `toc_deck` | `{ document \| path, observations…, kind, audience? }` | `{ format: "marp", content }` (no export) |
| `toc_judge_bundle` | `{ document \| path, profile? }` | judge bundle JSON |

Inline `document` is preferred for agents; `path` inputs MUST resolve inside
the server's configured root (`--root`, default: working directory) and MUST
reject `..` escapes and symlinks leaving the root.

### 6.4 Lint rules (initial set)

Severity `error` blocks GO; `warning` is reported; `info` is advisory.

| Rule ID | Severity | Check |
|---------|----------|-------|
| TOC-STR-001 | error | At least one outcome with level `impact` |
| TOC-STR-002 | error | No unintended cycles (cycles not marked `feedback`) |
| TOC-STR-003 | error | Every non-impact outcome has a path to an impact outcome (no dead ends) |
| TOC-STR-004 | warning | Every outcome has ≥ 1 precondition or a targeting intervention (no orphans) |
| TOC-STR-005 | warning | `at_least` logic has `k` between 1 and number of preconditions |
| TOC-STR-006 | warning | Links respect level order (no link from a higher level to a lower level unless `feedback`; same-level links allowed) |
| TOC-STR-007 | warning | Precondition groups are well-formed: ≥ 2 members, all members are preconditions of the group target, no group cycles |
| TOC-LOG-004 | info | Outcomes with ≥ 2 incoming links state their logic explicitly (non-default `logic` or a group), per Davies' AND/OR ambiguity critique |
| TOC-LOG-001 | warning | Every causal link into an `intermediate`, `long_term`, or `impact` outcome has a rationale |
| TOC-LOG-002 | warning | Every link into an outcome above `short_term` has ≥ 1 linked assumption |
| TOC-LOG-003 | info | Impact outcomes have ≥ 2 preconditions (single-chain theories are usually incomplete) |
| TOC-ASM-001 | warning | Every `high`/`critical` assumption has a `test` description |
| TOC-ASM-002 | warning | Every `high`/`critical` assumption has evidence or an assessment within 180 days |
| TOC-ASM-003 | info | At least one `contextual` or `external` assumption exists |
| TOC-MEA-001 | error | Every outcome within the ceiling of accountability has ≥ 1 indicator |
| TOC-MEA-002 | warning | Every indicator has baseline, ≥ 1 target, data source, and frequency |
| TOC-MEA-003 | warning | Quantitative indicators have a unit and direction |
| TOC-MEA-004 | warning | Target dates fall within the document's time horizon |
| TOC-MEA-005 | warning | Timeline consistency: an outcome's final target date is not earlier than the final target dates of its preconditions |
| TOC-ACC-002 | warning | Outcomes above the ceiling of accountability have no indicators owned by the perspective actor marked as accountable targets (contribution, not achievement, is claimed above the ceiling) |
| TOC-INT-001 | error | Every intervention targets ≥ 1 outcome |
| TOC-INT-002 | warning | Every intervention has an owner actor |
| TOC-ACC-001 | warning | A ceiling of accountability is declared |
| TOC-ACT-001 | info | Outcomes in the `influence` sphere name ≥ 1 actor (boundary partner) |
| TOC-EVD-001 | info | Links into impact-level outcomes cite ≥ 1 evidence record |
| TOC-PTH-001 | warning | Every pathway contains ≥ 1 outcome with a path to an impact outcome |
| TOC-PTH-002 | info | Every pathway has ≥ 1 selected indicator |
| TOC-MET-001 | warning | Every `common`-layer metric is instantiated by an indicator in every pathway, or the omission is explained in the indicator selection rationale |
| TOC-MET-002 | error | `rollupFrom` is only used when the metric's aggregation is not `none`, and all roll-up sources instantiate the same metric |
| TOC-CLM-001 | warning | Indicators measuring outcomes above the ceiling of accountability claim `contribution`, not `attribution` |
| TOC-BUD-001 | warning | Selected indicators' staff minutes per week ≤ `measurementBudget.staffMinutesPerWeek` |
| TOC-BUD-002 | warning | Survey / form questions per respondent per year ≤ `maxQuestionsPerRespondentPerYear` |
| TOC-BUD-003 | info | Number of selected indicators within `targetSelectedMetrics` |
| TOC-BUD-004 | info | Selected indicators declare `collection.mode`; `passive` and `derived` preferred over `survey` where an equivalent exists (reported as share passive) |
| TOC-AUD-001 | info | Every audience is targeted by ≥ 1 selected indicator or story theme |
| TOC-CAU-001 | warning | Every `underlying` and `structural` cause is addressed by ≥ 1 outcome or marked `outOfScope` with a reason |
| TOC-CAU-002 | info | When causes are declared, outcomes address ≥ 1 non-`immediate` cause (root causes, not only immediate ones) |
| TOC-RSP-001 | warning | Outcomes with responsibility `partner` or `shared` name ≥ 1 delivering actor; `shared` names ≥ 2 |
| TOC-RSP-002 | info | `partner` and `unaddressed` outcomes on a path to impact still have ≥ 1 indicator (monitor the whole solution tree) |
| TOC-ASM-004 | warning | Assumptions with `needsTesting: true` have a `test` and either a monitoring indicator or a linked decision / evaluation question |
| TOC-ASM-005 | info | Assumptions of type `partnership` exist when any outcome is `shared` or `partner` |
| TOC-RSK-001 | info | `high` / `critical` assumptions have ≥ 1 related risk with mitigation |
| TOC-EQU-001 | warning | Indicators on outcomes that target groups disaggregate by those groups (specify changes for target groups, not assume them) |
| TOC-VAL-001 | info | Each of `focus`, `change_analysis`, `assumptions_risks`, `actors` has ≥ 1 consultation record |
| TOC-PRV-001 | info (warning under `full`) | Every element declares a provenance basis |
| TOC-PRV-002 | warning | Every `stated` element cites ≥ 1 evidence record with a URL or document reference |
| TOC-PRV-003 | warning | An `approved` document has no `hypothesis` elements at accountable levels (responsibility `own` or `shared`) |
| TOC-PRV-004 | info | Selected indicators declare `dataAvailability` |

Rule IDs are permanent. Retired rules keep their ID reserved.

### 6.5 Measurement algorithms

**Indicator progress** (quantitative). With baseline `b`, current `c`, final
target `t`, direction `d`:

- `increase`: `p = (c − b) / (t − b)`; `decrease`: `p = (b − c) / (b − t)`.
- `maintain`: `p = 1` if `|c − t| ≤ tolerance` else `0`.
- If `t = b`, `p = 1` when the target condition holds, else 0.
- `p` is clamped to `[0, 1.5]` for display; `achieved` when `p ≥ 1`.

**Expected progress at `asOf`.** Targets form a time series
`(baseline.date, 0) → (target_i.date, p_i)`. Expected `e` is linear
interpolation at `asOf` (0 before the baseline date, 1 after the final
target). Indicator status:

| Condition | Status |
|-----------|--------|
| no targets defined (baseline only) | `baseline_only` (latest value reported, no progress computed) |
| no observation and `asOf` < first target date | `not_started` |
| no observation and `asOf` ≥ first target date | `no_data` |
| last observation older than `2 × frequency` | `stale` |
| `p ≥ 1` | `achieved` |
| `p ≥ e − tol` (default `tol = 0.10`) | `on_track` |
| `p ≥ e − 2·tol` | `at_risk` |
| otherwise | `off_track` |

Qualitative indicators use the ordinal position of the observed level in the
indicator's ordered `levels` as `c`, with `b` and `t` as level positions.

**Health mapping** (prism-core): `achieved`, `on_track` → `green`;
`at_risk`, `stale` → `yellow`; `off_track` → `red`; `not_started`, `no_data`,
`baseline_only` → no health (rendered grey). `baseline_only` ranks with
`not_started` in roll-ups.

**Metric roll-ups.** For an indicator with `rollupFrom`, the value at a date
is computed from the latest source observations at or before that date using
the metric's `aggregation` (`weighted_mean` weights by the latest value of
`weightMetricId`'s indicator in the same pathway). If any source lacks a value
the roll-up is `no_data` and lists the missing sources. Direct observations of
a roll-up indicator take precedence and the computed value is reported
alongside for comparison.

**Theme counts.** A `theme_count` indicator's value for a period is the number
of stories in that period's observation sets tagged with any of its
`themeIds` (optionally filtered by `pathwayIds`); stories are counted once.

**Collection burden** (`pkg/burden`). Per selected indicator, staff minutes
are normalized to minutes per week from `collection.minutesPerPeriod` and
`collection.period`; respondent questions per year =
`questionsPerRespondent × (collections per year) × sampleRate`. Totals are
reported overall, per pathway, and per mode; budget comparisons feed
TOC-BUD-*.

**Outcome status** = roll-up of its indicators by `Rollup` option: `worst`
(default) or `mean` of progress. Outcomes without indicators get `unmeasured`.

**Pathway status.** Processed in topological order from the leaves upward. A
node's precondition signal is computed over incoming non-feedback links:
grouped sources are first reduced through their `PreconditionGroup` (nested
groups innermost first), then the group results and ungrouped sources are
combined with the outcome's logic: `all` → worst; `any` → best;
`at_least k` → k-th best. Status ordering for worst/best is
`off_track < at_risk < stale < no_data < not_started < on_track < achieved`. Pathway status = worse of (own outcome status,
precondition signal), where `unmeasured` defers to the precondition signal.

**Assumption propagation.** Effective assumption status = latest assessment
in observation sets, else the document's static status. A link whose
assumptions include `invalidated` is `broken`; `at_risk` makes it `at_risk`.
Every outcome downstream of a broken link gets the flag
`assumptionBroken: [assumptionIDs]` and its pathway status is capped at
`at_risk` (critical assumption: `off_track`).

**Link confidence** ∈ [0, 1]: `0.4·evidence + 0.4·assumptions + 0.2·rationale`,
where evidence = max strength of supporting evidence (`strong` 1.0,
`moderate` 0.7, `weak` 0.4, `anecdotal` 0.2, none 0) minus half the max
strength of contradicting evidence, floored at 0, assumptions = min over linked assumption statuses
(`validated` 1.0, `holding` 0.8, `untested` 0.5, `at_risk` 0.25,
`invalidated` 0), and rationale = 1 if present else 0. **Path confidence** =
product of link confidences along the path; an outcome's confidence = max over
its paths to impact. Weights are options with these defaults and are printed
in every report that shows a confidence value (INV-5).

**Diff.** `Diff(prev, cur)` lists status changes for indicators, outcomes,
assumptions, and pathways, ordered per INV-3.

### 6.6 Reports

| Kind | Content |
|------|---------|
| `exec` | Headline (goal, overall pathway status, as-of), status by level, top 5 at-risk / off-track outcomes with driving indicators, broken or untested critical assumptions, changes since previous snapshot, decisions requested (from `Document.decisions` and auto-generated prompts for broken assumptions), confidence footnote |
| `pathway` | Mermaid `flowchart BT` with subgraphs by level, nodes styled by health, dashed feedback links, ceiling of accountability as a subgraph boundary |
| `results` | Results framework: outcome → indicator → baseline → targets → latest → status |
| `assumptions` | Assumption register: statement, type, criticality, applies to, test, evidence, status |
| `interventions` | Intervention → owner → status → targeted outcomes → pathway status |
| `evidence-gaps` | Links and critical assumptions lacking evidence; outcomes lacking indicators |
| `story` | For one `--audience`: audience needs, the indicators and outcomes tagged for it with latest values and status, consented stories tagged for it (summary always; quote only per TRD-023), theme counts, and caveats — a narrative scaffold for communications staff to finish |
| `learning-review` | For a period: status changes since previous snapshot, largest deviations from expected progress (surprises), assumption status changes, theme counts and new themes, data-quality notes, and draft questions; merges any recorded `LearningReview` for the period |
| `portfolio` | Per pathway: pathway status, outcome-status distribution, confidence, collection burden, share of passive data, audiences covered, and open decisions — evidence for grow / adjust / sunset decisions, without a recommendation |
| `narrative` | Programme / strategy document scaffold following the UNDG section mapping (research §9.7): causal analysis (problem tree → outcomes addressing each cause level), priorities and rationale (priority scores), assumptions and risks with mitigation, partner-delivered and unaddressed elements, implementation arrangements (who does what), M&E provisions (assumptions needing testing, monitoring cadence, evaluation questions from evidence gaps), validation trail (consultations), key messages. Works without observations; adds status when given |
| `assessment` | Definition-phase assessment (no observations needed): provenance mix with lists, outcomes without indicators, indicators by data availability, assumptions needing testing, open decisions, generated questions for the organization, short narrative summary. Inferred and hypothesis elements are marked in every report and deck by a basis badge |
| `actors` | Division of labour: actors × outcomes / interventions with roles; outcomes needing collaboration; outcomes with no delivering actor; partner-delivered outcomes being monitored |
| `full` | All status reports in one document |
| `custom` | User `text/template` over `report.View` |

Definition reports (no observations required):

| Kind | Content |
|------|---------|
| `pathway` | As above, without status styling when no status is given; `--pathway` restricts to one pathway plus its route to impact |
| `logic-model` | Per pathway: inputs (intervention resources) → activities (interventions) → outputs (`output`-level outcomes and intervention outputs) → outcomes (short / intermediate / long-term) → impact, as a table |
| `metrics` | Metric definitions: name, definition, formula, unit, aggregation, layer, and which pathways instantiate each |
| `metric-selection` | Candidate indicators × selection criteria with weighted score, status, rationale; sorted by score |
| `collection-plan` | Selected indicators: definition, source system, mode, owner, frequency, minutes per period, questions per respondent; totals vs budget |
| `glossary` | Glossary terms and theme codebook |
| `definitions` | All definition reports in one document (the workshop pack) |

`report.View` is a documented, versioned struct (`ViewVersion = "1"`);
changing field names is a breaking change.

### 6.7 Reference grammar (`pkg/ref`)

```
ref      = "toc:" docSlug [ "#" elementId ]
docSlug  = slug                ; e.g. toc:agent-auth-open-standard
elementId= slug                ; e.g. toc:agent-auth-open-standard#oc-idp-support
slug     = [a-z][a-z0-9]*(-[a-z0-9]+)*    ; ≤ 128 chars
```

Element IDs inside a document are slugs (INV-1). External references use
`ExternalRef { framework, ref, url?, relation }`; `ref` is opaque to this
module (for market-spec it is that repo's `{type}:{slug}` form). This grammar
is the USM/UFM integration contract together with §7.3.

### 6.8 Error Semantics

Not applicable as HTTP — there is no network API. CLI exit codes are in §6.1.
MCP tool errors return `isError: true` with a JSON body
`{ code, message, issues? }`; codes mirror sentinel errors
(`invalid_document`, `unknown_reference`, `unsupported_format`, `path_outside_root`).

### 6.9 Versioning

- Schema `$id`: `https://github.com/ProductBuildersHQ/theoryofchange-spec/schema/v0/<name>.schema.json`.
- Every document carries `kind` (`TheoryOfChange` | `ObservationSet` |
  `TheoryOfChangeStatus`) and `specVersion` (`"0.1"`).
- Within v0.x, minor versions are additive; removals or renames require a
  major bump and a migration note in the CHANGELOG.

### 6.10 Events

Not applicable — no event bus; file outputs only.

### 6.11 CSV mapping and agent skills

**CSV mapping file** (YAML, schema-defined as `csv-mapping`):

```yaml
tocId: example-network
columns:
  indicator: metric_id        # column holding indicator ID, or…
  indicatorFixed: ""          # …a fixed indicator ID for single-indicator sheets
  date: period_end
  value: value
  pathway: program_area       # optional
  quality: data_quality       # optional
  source: source              # optional
indicatorAliases:             # map spreadsheet labels to indicator IDs
  "Attendee hours": ind-attendee-hours
dateFormat: "2006-01-02"
```

Rows whose indicator, date, or value cannot be resolved are written to the
unmatched list with row number and reason (INV-4).

**Agent skills** (`skills/<name>/SKILL.md` plus templates). Each skill
states purpose, inputs, the exact command sequence, checks to report, and
outputs. AI judgment is limited to drafting text and asking the user; every
number comes from a command.

| Skill | Command sequence (abridged) | Checks reported |
|-------|-----------------------------|-----------------|
| `toc-author` | `toc init` → edit → `toc validate` → `toc lint` → `toc report definitions` | validation issues, lint GO/NO-GO, definition pack written |
| `toc-status-refresh` | `toc import observations --csv` → `toc validate` → `toc status --prev` → `toc report exec` | rows imported / unmatched, indicators with new data, changes vs previous snapshot |
| `toc-learning-review` | `toc status` → `toc report learning-review` → assistant drafts discussion questions | surprises listed, assumptions changed |
| `toc-audience-story` | `toc report story --audience` → assistant drafts narrative from the scaffold | every number in the draft present in the scaffold; no unconsented quotes |
| `toc-draft-from-sources` | `toc init` → for each source: assistant reads it, `toc evidence add --url … --excerpt …` → assistant proposes elements → `toc patch` (each element with `provenance.basis` and `evidenceIds`) → `toc validate` → `toc lint --profile draft` → `toc report definitions` and `toc report assessment` | sources recorded, elements by basis, `stated` elements without citation (must be 0), lint GO under `draft` |
| `toc-share-google` | `toc export review-packet --format docx`, `toc export review-sheet`, `toc deck … --export` / `toc export pptx` → upload with conversion (gogoogle CLI `drive upload --convert`) → after review: read Sheet records, Doc tables, comments, and suggestions (gogoogle CLI or mcp-google tools) → assistant writes `ReviewResponse` → `toc import review` → human runs `toc patch` on accepted proposals | files published with links; rows read per origin; stale-hash warnings; rows needing human decision |

Manual fallback for `toc-share-google`: upload the DOCX and CSV by hand, ask
reviewers to use the response columns, download the sheet as CSV, and convert
it to a `ReviewResponse` with `toc import review --from-csv`.

Skills that change the theory (`toc-author`) MUST use `toc patch` or element
commands, never write the file directly.

### 6.12 Editing contract (`pkg/edit`)

```go
type Patch []Op // RFC 6902: add, remove, replace, move, copy, test
type Result struct {
    Document         *toc.Document
    IssuesIntroduced []validate.Issue
    IssuesResolved   []validate.Issue
    Applied          bool
}
func Apply(doc *toc.Document, p Patch, opts Options) (*Result, error) // opts.AllowErrors
```

- **Addressing.** Standard JSON Pointers plus an ID segment: `@<id>` selects
  the element with that `id` in the array, e.g.
  `/outcomes/@oc-adoption/title` or `/links/@ln-3`. ID segments are resolved
  to indexes before applying; an unknown ID fails the whole patch.
- **Atomicity.** Operations apply to a deep copy (INV-2). If any operation
  fails, or the result fails schema validation, or `error`-severity issues
  increase versus the input (unless `AllowErrors`), `Applied` is false and the
  original is untouched.
- **Element commands** build patches: `add` appends (`/outcomes/-`) after
  checking ID uniqueness; `set` emits `replace`/`add` per flag; `rm` emits
  `remove` and, with `--cascade`, removes or rewrites references (links
  touching a removed outcome; IDs in `appliesTo`, `assumptionIds`, etc.).
  Without `--cascade`, dangling references make the patch rejected (INV-4).
- **Writing.** The CLI writes the result in canonical form (`toc fmt`), in
  the input's format. YAML comments are not preserved (PRD R-3).
- **Canonical form.** JSON: two-space indent, struct field order, no HTML
  escaping, trailing newline. YAML: generated from canonical JSON. Element
  order is preserved; `--sort` orders elements by ID.

### 6.13 Presentations (`pkg/slides`)

Decks are Marp Markdown with a front-matter theme (`default`, overridable by
`--theme file.css`) and one idea per slide.

| Deck | Slides |
|------|--------|
| `workshop` | Title · problem statement and causes · long-term goal and ceiling of accountability · pathway diagram (whole, then per pathway) · logic model per pathway · key assumptions (if / then / because) · metric definitions · glossary · open questions and decisions |
| `exec` | Headline status · pathway diagram with health · top risks and broken assumptions · what changed since last period · decisions requested · confidence and method footnote |
| `learning-review` | Period summary · surprises · assumption changes · theme counts and selected stories (consented) · questions for discussion · actions |
| `story` | Audience need · three to five headline metrics with caveats · one or two consented stories · call to action |

Diagrams: the library embeds Mermaid source in fenced blocks and, when the
CLI is asked to export, `cmd/toc` renders each to SVG with `mmdc` and rewrites
the slide to reference the SVG; then runs `marp --pdf|--pptx|--html`. Missing
tools produce exit code 3 with install guidance; the Markdown deck is still
written. Same layout conventions as prism-maturity's Marp generator.

### 6.14 Judging (`pkg/judge`)

- **Bundle** (`toc judge show`): JSON with `rubric` (core or development
  `RubricSet`), `document` (canonical ToC JSON), `narrative` (generated
  `narrative` report Markdown), `lint` (SummaryReport), `instructions`
  (judge instructions from `spec/`), and `bundleHash` (SHA-256 of the
  inputs). The judge (an agent session, visionspec, or visionstudio) reads
  the bundle; this module makes no model calls.
- **Verdict** (`toc judge record`): a structured-evaluation evaluation report
  whose `bundleHash` matches; stored as
  `<toc-dir>/.toc/judgements/<bundleHash>/<judge>.json`. Mismatched hashes are
  rejected (the verdict was for a different version).
- **Summary** (`toc judge summary`): per category, verdict distribution and
  inter-rater agreement using structured-evaluation's multi-judge and IRR
  helpers; overall pass requires the rubric's pass criteria under the
  configured aggregation (default: majority, ties → `partial`).

### 6.15 Sharing, review, and approval (`pkg/share`, `pkg/review`)

**Export stamp.** Every export includes
`{tocId, version, contentHash, asOf}`, where `contentHash` = SHA-256 of
canonical JSON (§6.12). Where it goes:

| Format | Location |
|--------|----------|
| Markdown | YAML metadata block |
| DOCX / PPTX | Inherited from the metadata block, plus a footer line |
| CSV | A leading comment row `# toc:<id> version:<v> hash:<h>` |
| Decks | Title slide |

**`pandoc` profile.**

- Mermaid diagrams are rendered to PNG by `mmdc` (CLI only) and referenced
  by relative path.
- Pipe tables only, with no nested lists inside cells.
- `\newpage` markers between sections.
- The basis badge is rendered as text: `[stated]`, `[inferred]`,
  `[hypothesis]`.
- `toc export docx|pptx` runs
  `pandoc -f markdown -t docx|pptx [--reference-doc]` with an argument vector
  (no shell).

**Review packet and sheet.**

- **Packet layout:** one section per element kind (outcomes, links /
  assumptions, interventions, indicators, metrics, actors). Each section has
  one table with columns `ID | Statement | Basis | Sources | Decision |
  Suggested change | Comment`, with the last three empty.
- **Instructions:** the packet asks reviewers to answer in the response
  columns. Comments and suggestions are still accepted and are mapped by
  quoted text.
- **Sheet:** the same rows, one row per element, with a `kind` column.

**`ReviewResponse`** (schema `review-response`):

```go
type ReviewResponse struct {
    Kind        string      `json:"kind"`        // "ReviewResponse"
    TocID       string      `json:"tocId"`
    Version     string      `json:"version"`
    ContentHash string      `json:"contentHash"`
    Reviewers   []string    `json:"reviewers"`   // groups or roles, e.g. "program staff"
    Method      string      `json:"method"`      // document_review | sheet_review | workshop
    Date        string      `json:"date"`
    Rows        []ReviewRow `json:"rows"`
}
type ReviewRow struct {
    ElementID       string `json:"elementId,omitempty"` // empty when only quoted text is known
    Decision        string `json:"decision,omitempty"`  // agree | change | remove | unsure
    SuggestedChange string `json:"suggestedChange,omitempty"`
    Comment         string `json:"comment,omitempty"`
    Origin          string `json:"origin"`              // sheet | table | comment | suggestion
    QuotedText      string `json:"quotedText,omitempty"`
}
```

**Import mapping** (`toc import review`):

| Row | Proposed op |
|-----|-------------|
| `agree` | `replace /…/@id/provenance/basis` → `confirmed` |
| `remove` | basis → `rejected` |
| `change` with a structured `SuggestedChange` (`field: value` lines for `title`, `statement`, `level`, `definition`, …) | `replace` ops |
| `change` with free text, `unsure`, or comment-only | Flagged note for a human, not an op |
| No element ID but `QuotedText` | Resolved by exact match against rendered statements. If ambiguous or not found, flagged |
| Stale hash (response hash ≠ document hash) | Every row flagged; ops still proposed against current IDs |

The import also writes a `Consultation` record (step `whole` unless given;
method; reviewer groups; findings = row counts by decision).

**Approval.** `toc approve` builds a patch that sets:

- `metadata.status = approved`
- `metadata.version`
- appended `Revision{trigger: review, summary, approvers}` (prism-core
  `Approver`)

It applies the patch through `pkg/edit`. It refuses when:

- lint reports any `error` finding;
- any rule TOC-PRV-003 violation remains, unless `--force` with `--reason`
  (recorded in the revision).

## 7. Data Design

### 7.1 Document model (`pkg/toc`)

Go structs are normative; this table summarizes them. All slices are optional
unless marked required.

```go
type Document struct {
    Schema          string           `json:"$schema,omitempty"`
    Kind            string           `json:"kind"`        // "TheoryOfChange"
    SpecVersion     string           `json:"specVersion"` // "0.1"
    ID              string           `json:"id"`          // doc slug
    Metadata        core.Metadata    `json:"metadata"`    // prism-core
    Perspective     *Perspective     `json:"perspective,omitempty"`
    Context         Context          `json:"context"`
    Ceiling         *Ceiling         `json:"ceilingOfAccountability,omitempty"`
    Outcomes        []Outcome        `json:"outcomes"`
    Groups          []PreconditionGroup `json:"preconditionGroups,omitempty"`
    Links           []CausalLink     `json:"links,omitempty"`
    Interventions   []Intervention   `json:"interventions,omitempty"`
    Assumptions     []Assumption     `json:"assumptions,omitempty"`
    Indicators      []Indicator      `json:"indicators,omitempty"`
    Actors          []Actor          `json:"actors,omitempty"`
    ActorRegistry   string           `json:"actorRegistry,omitempty"` // ref to shared registry doc
    Evidence        []Evidence       `json:"evidence,omitempty"`
    ExternalFactors []ExternalFactor `json:"externalFactors,omitempty"`
    Risks           []Risk           `json:"risks,omitempty"`
    Alignments      []Alignment      `json:"alignments,omitempty"`
    Decisions       []Decision       `json:"decisions,omitempty"`
    Revisions       []Revision       `json:"revisions,omitempty"`

    // Measurement-framework design (portfolio, sustainability, story)
    Parent            string               `json:"parent,omitempty"` // toc: ref of parent ToC
    Pathways          []Pathway            `json:"pathways,omitempty"`
    Metrics           []Metric             `json:"metrics,omitempty"`
    SelectionCriteria []SelectionCriterion `json:"selectionCriteria,omitempty"`
    Budget            *MeasurementBudget   `json:"measurementBudget,omitempty"`
    Audiences         []Audience           `json:"audiences,omitempty"`
    Glossary          []GlossaryTerm       `json:"glossary,omitempty"`
    Themes            []Theme              `json:"themes,omitempty"` // story codebook

    // Change analysis and validation (UNDG)
    Causes        []Cause        `json:"causes,omitempty"`        // problem tree
    TargetGroups  []TargetGroup  `json:"targetGroups,omitempty"`  // priority populations
    Consultations []Consultation `json:"consultations,omitempty"` // validation record
}
```

| Type | Key fields | Notes |
|------|-----------|-------|
| `Perspective` | `actorId`, `role` (`proponent`, `funder`, `implementer`, `observer`), `futureRef?` | Whose theory this is. USM uses the actor; UFM uses `futureRef` to group a ToC under a competing future. |
| `Context` | `problemStatement` (req), `rootCauses[]`, `targetPopulation`, `market?` (ExternalRef), `geography?`, `timeHorizon {start,end}` (req), `narrative` | Problem analysis / problem tree inputs. |
| `Ceiling` | `level` (outcome level) or `outcomeIds[]`, `rationale` | Ceiling of accountability; outcomes above it are tracked but not owned. |
| `Outcome` | `id`, `title` (req; a positive condition, not an activity), `description`, `level` (req: `impact`, `long_term`, `intermediate`, `short_term`, `output`), `sphere` (`control`, `influence`, `interest`), `logic` {`type`: `all`/`any`/`at_least`, `k`} over ungrouped incoming links (default `all`), `category` (open string; recommended vocabularies: ORS `impact`/`influence`/`leverage`/`capacity`, COM-B `reach`/`capability`/`opportunity`/`motivation`/`behavior`/`direct_benefit`/`wellbeing`, MSD AAER `adopt`/`adapt`/`expand`/`respond`), `changeType` (`knowledge`, `attitude`, `skill`, `behavior`, `practice`, `policy`, `norm`, `condition`, `investment`, `adoption`), `population`, `actorIds[]` (who changes / co-owners — supports collective outcomes), `responsibility` (`own` default, `shared`, `partner`, `unaddressed`), `deliveredBy[]` (actor IDs), `addressesCauses[]` (cause IDs), `targetGroupIds[]`, `priority?` {`scores[]` {criterionId, score 1–5, note}, `rationale`}, `tags[]` | Graph node. Indicators reference outcomes, not vice versa. `partner` / `unaddressed` outcomes stay in the graph and are monitored (UNDG: monitor the whole solution tree) but are excluded from accountability counts. |
| `Cause` | `id`, `statement` (req), `level` (req: `immediate`, `underlying`, `structural`), `parent?` (cause ID it contributes to), `affectedGroupIds[]`, `evidenceIds[]`, `outOfScope?` {reason} | Problem tree; outcomes map to causes via `addressesCauses`. |
| `TargetGroup` | `id`, `name` (req), `description`, `size?`, `actorIds[]` | Priority population (e.g. under-resourced member organizations, residents of low-access areas). |
| `Consultation` | `id`, `date` (req), `step` (req: `focus`, `change_analysis`, `assumptions_risks`, `actors`, `measurement`, `whole`), `method` (`workshop`, `interview`, `survey`, `document_review`, `evidence_review`, `meeting`), `stakeholderGroups[]` (groups or actor IDs, never personal names required), `count?`, `findings[]`, `changedIds[]` | Validation trail per UNDG principle (a) and the validate-each-step loop. |
| `PreconditionGroup` | `id`, `target` (outcome ID), `members[]` (outcome or group IDs; ≥ 2), `logic` {`all`/`any`/`at_least`, `k`}, `sufficiency` (text) | Explicit AND/OR/k-of-n "cluster" (ActKnowledge). Groups nest, so `(A AND B) OR C` is expressible. Incoming links whose source is a group member are evaluated through the group; the group's result feeds the target's `logic`. |
| `CausalLink` | `id`, `from` (outcome or intervention ID), `to` (outcome ID), `type` (`causes`, `enables`, `contributes_to`, `necessary_condition`), `rationale`, `assumptionIds[]`, `evidenceIds[]`, `interventionIds[]` (interventions placed on the connector, ActKnowledge style), `polarity` (`positive` default, `negative`), `feedback` bool, `loopType` (`reinforcing`, `balancing`; only with `feedback`), `lag?` (ISO 8601 duration) | Directed edge. Intervention → outcome edges MAY be implicit via `Intervention.targets`. |
| `Intervention` | `id`, `title` (req), `type` (`program`, `product`, `policy`, `advocacy`, `partnership`, `standardization`, `investment`, `communication`, `research`, `other`), `targets[]` (req), `outputs[]`, `ownerActorId`, `status`, `capabilities[]` (`core.CapabilityRef`), `initiativeRefs[]` (ExternalRef, e.g. visionstudio `INIT-…`), `resources` | Strategy / activity. |
| `Assumption` | `id`, `statement` (req), `ifThenBecause?` {`if`, `then`, `because`}, `type` (`causal`, `contextual`, `implementation`, `behavioral`, `external`, `partnership`), `appliesTo[]` (link/outcome/intervention IDs), `criticality` (`low`, `medium`, `high`, `critical`), `plausibility` (`low`, `medium`, `high`), `explainsFully?` (bool; UNDG question 1), `needsTesting` (bool; UNDG question 3), `test` (how: monitoring indicator, evaluation, research), `indicatorIds[]` (assumption-monitoring indicators), `evidenceIds[]`, `status`, `ownerActorId` | Static status is the default; observation sets override. `confidence` from v0.1 drafts is replaced by `plausibility` (UNDG question 2). |
| `Indicator` | `id`, `measures` (req; outcome ID, or intervention ID for output indicators, or assumption ID for assumption-monitoring indicators), `name` (req), `definition`, `measurementType` (`quantitative`, `qualitative`, `hybrid`), `unit`, `direction` (`increase`, `decrease`, `maintain`), `tolerance`, `levels[]` (ordered qualitative levels; Outcome Mapping progress markers use `expect`/`like`/`love`), `baseline {value, level, date, source}`, `targets[] {value, level, threshold, date, label, milestone bool}`, `population {description, actorIds[], size}`, `disaggregation[]`, `dataSource`, `method`, `frequency` (ISO 8601 duration), `ownerActorId`, `proxyFor[]` (outcome IDs this indicator stands in for), `externalRefs[]` (e.g. maturity SLI, OKR KR) | ActKnowledge's four components — population, target, threshold, timeline — map to `population`, `targets[].value`, `targets[].threshold`, `targets[].date`. Framework-design extensions: `metricId` (instantiates a shared `Metric`), `pathwayIds[]`, `claim` (req: `attribution` or `contribution`), `caveats[]`, `collection {mode, system, minutesPerPeriod, period, questionsPerRespondent, respondentsPerYear, sampleRate}`, `selection {status: candidate/selected/retired, scores[] {criterionId, score 1–5, note}, rationale}`, `audienceIds[]`, `rollupFrom[]` (indicator IDs whose values roll up into this one via the metric's rule), `themeIds[]` (for `theme_count` indicators derived from stories). `measurementType` gains `theme_count`. |
| `Pathway` | `id`, `title` (req), `kind` (`program`, `strategy`, `theme`, `geography`, `segment`), `description`, `ownerActorId`, `outcomeIds[]`, `interventionIds[]`, `childToc?` (ref to a program-level ToC document) | Program area / strategy grouping. Indicators join via `pathwayIds`. A pathway may delegate detail to a child ToC whose `parent` points back. |
| `Metric` | `id`, `name` (req), `definition` (req), `formula` (req; exact computation in words, e.g. "sum over sessions of attendees × session hours"), `unit`, `direction`, `aggregation` (`sum`, `mean`, `weighted_mean`, `min`, `max`, `latest`, `none`), `weightMetricId?`, `layer` (`common`, `pathway`) | Shared definition so a metric is computed identically across pathways. `none` means values cannot be combined (e.g. distinct counts across overlapping populations) and org-level values must be observed directly. |
| `SelectionCriterion` | `id`, `name`, `question`, `weight`, `appliesTo` (req: `indicator`, `outcome`, `intervention`) | Indicator criteria: decision value, audience value, cost to collect (inverted), consistency, credibility. Outcome / intervention (focus and solution) criteria: equity, comparative advantage, feasibility, evidence of what works. Scores live on the subject (`selection.scores` for indicators, `priority.scores` for outcomes and interventions). |
| `MeasurementBudget` | `staffMinutesPerWeek`, `maxQuestionsPerRespondentPerYear`, `targetSelectedMetrics {min, max}` | Declared sustainability constraints. |
| `Audience` | `id`, `name` (req), `needs` (what they need to see), `channels[]` (e.g. annual report, grant proposal, board deck, newsletter) | Story and reporting target. |
| `GlossaryTerm` | `term` (req), `definition` (req), `examples[]`, `relatedIds[]` | Shared vocabulary (e.g. output vs outcome vs impact with organization-specific examples). |
| `Theme` | `id`, `label` (req), `definition` (req), `includes[]`, `excludes[]`, `examples[]`, `outcomeIds[]` | Codebook for coding stories consistently (human or AI). |
| `Provenance` (embedded as `provenance` on every element type) | `basis` (`stated`, `inferred`, `hypothesis`, `confirmed`, `rejected`), `evidenceIds[]`, `note`, `decidedAt?`, `decidedBy?` (group or role) | Public-source drafting and review lifecycle. `rejected` elements stay in the document for history but are excluded from graph analysis, measurement, and default reports. Sources are ordinary `Evidence` records (URL, title, excerpt, accessedAt, sensitivity). |
| Indicator `dataAvailability` | `public`, `internal_existing`, `internal_new`, `partner_held`, `unknown` | Feeds the assessment report and passive-data-first selection. |
| `Actor` | `id`, `name` (req), `type` (`organization`, `person`, `group`, `community`, `customer_segment`, `standards_body`, `regulator`, `vendor`, `platform`, `investor`, `other`), `roles[]` (`implementer`, `beneficiary`, `boundary_partner`, `funder`, `champion`, `influencer`, `competitor`, `adopter`, `gatekeeper`), `stance` (`champion`, `supporter`, `neutral`, `skeptic`, `opponent`, `unknown`), `influence`/`interest` (`low`/`medium`/`high`), `contact` (`core.Person`), `externalRefs[]` | Shared registry supported via `ActorRegistry` document. |
| `Evidence` | `id`, `title` (req), `type` (`research`, `evaluation`, `monitoring_data`, `expert_opinion`, `stakeholder_testimony`, `case_study`, `internal_metric`, `market_signal`, `outcome_harvest`, `anecdote`), `strength` (`strong`, `moderate`, `weak`, `anecdotal`), `direction` (`supports` default, `contradicts`), `processTracingTest` (`straw_in_the_wind`, `hoop`, `smoking_gun`, `doubly_decisive`; optional), `source`, `url`, `summary`, `excerpt`, `sensitivity` (`public`, `internal`, `restricted`; zero value treated as `restricted`), `accessedAt` | Mirrors prism-roadmap evidence sensitivity semantics. Contradicting evidence lowers link confidence (§6.5). |
| `ExternalFactor` | `id`, `description` (req), `kind` (`enabler`, `constraint`, `trend`, `shock`), `affects[]`, `likelihood`, `trend` (prism-core trend) | Context conditions outside the ToC's control. |
| `Risk` | embeds `core.Risk`, plus `appliesTo[]`, `category` (`environmental_political`, `opportunity`, `design`, `partnership`, `market`, `technical`, `operational`), `relatedAssumptionIds[]` | prism-core reuse; UNDG notes assumptions and risks are often inverse, so the link is explicit. `opportunity` risks are upside. |
| `Alignment` | `elementId`, `target` (`ExternalRef`) | `ExternalRef{framework, ref, url, relation}`; frameworks: `okr`, `bsc`, `v2mom`, `capability`, `maturity_sli`, `market_spec`, `usm`, `ufm`, `sdg`, `vsem` (Vision–Strategy–Execution–Metrics one-pagers), `visionstudio`, `custom`; relations: `supports`, `measures`, `implements`, `contributes_to`, `same_as`. BSC alignments carry `perspective` in `ExternalRef.attributes` (`financial`, `customer`, `internal_process`, `learning_growth`). |
| `Decision` | `id`, `question`, `options[]`, `dueDate`, `ownerActorId`, `status`, `relatedIds[]` | Feeds "decisions requested" in exec reports. |
| `Revision` | `version`, `date`, `summary`, `trigger` (`learning`, `evidence`, `context_change`, `strategy_change`, `review`), `changedIds[]`, `triggerEvidenceIds[]`, `qualityReview?` {`plausible`, `feasible`, `testable`: `pass`/`partial`/`fail`, `notes`} | Adaptive-management log; `qualityReview` records the Aspen Roundtable criteria. |

### 7.2 Observation set and status

```go
type ObservationSet struct {
    Kind        string                 `json:"kind"`        // "ObservationSet"
    SpecVersion string                 `json:"specVersion"`
    TocID       string                 `json:"tocId"`       // doc slug it measures
    Period      string                 `json:"period,omitempty"` // e.g. "2026-Q4"
    Observations []Observation         `json:"observations,omitempty"`
    Assessments  []AssumptionAssessment `json:"assumptionAssessments,omitempty"`
    Stories      []Story               `json:"stories,omitempty"`
    Reviews      []LearningReview      `json:"learningReviews,omitempty"`
    Notes        []Note                `json:"notes,omitempty"`
}
type Story struct {
    ID          string   `json:"id"`
    Date        string   `json:"date"`
    Title       string   `json:"title"`
    Summary     string   `json:"summary"`            // paraphrase, safe to render
    Quote       string   `json:"quote,omitempty"`    // rendered only with consent
    ActorID     string   `json:"actorId,omitempty"`  // who changed (may be anonymized)
    Domain      string   `json:"domain,omitempty"`   // domain of change
    ThemeIDs    []string `json:"themeIds,omitempty"` // must exist in codebook
    OutcomeIDs  []string `json:"outcomeIds,omitempty"`
    AudienceIDs []string `json:"audienceIds,omitempty"`
    Consent     string   `json:"consent"`            // granted | not_granted | unknown
    Sensitivity string   `json:"sensitivity"`        // public | internal | restricted
    CodedBy     string   `json:"codedBy,omitempty"`  // person | model:<name> — provenance of theme coding
    Significance string  `json:"significance,omitempty"` // why it was selected (MSC)
}
type LearningReview struct {
    Period    string   `json:"period"`
    Questions []string `json:"questions,omitempty"`
    Findings  []string `json:"findings,omitempty"`
    Decisions []string `json:"decisions,omitempty"`
    Actions   []string `json:"actions,omitempty"`
    RelatedIDs []string `json:"relatedIds,omitempty"`
}
type Observation struct {
    IndicatorID string     `json:"indicatorId"`
    Date        string     `json:"date"`            // RFC 3339 date
    Value       *float64   `json:"value,omitempty"`
    Level       string     `json:"level,omitempty"` // qualitative
    Source      string     `json:"source,omitempty"`
    EvidenceIDs []string   `json:"evidenceIds,omitempty"`
    Note        string     `json:"note,omitempty"`
    Quality     string     `json:"quality,omitempty"` // verified | reported | estimated
    PathwayID   string     `json:"pathwayId,omitempty"` // when one indicator is observed per pathway
    Disaggregation map[string]string `json:"disaggregation,omitempty"`
}
type AssumptionAssessment struct {
    AssumptionID string `json:"assumptionId"`
    Date         string `json:"date"`
    Status       string `json:"status"`
    Rationale    string `json:"rationale,omitempty"`
    EvidenceIDs  []string `json:"evidenceIds,omitempty"`
}
```

`toc.Status` (kind `TheoryOfChangeStatus`) holds `asOf`, `options`
(weights / tolerance used), and per-element results: `indicators[]`
(progress, expected, status, health, latest observation), `outcomes[]`
(status, pathwayStatus, confidence, assumptionBroken[]), `links[]` (state,
confidence), `assumptions[]` (effective status, source), `overall`
(per impact outcome). Field-level schema is generated; this TRD fixes the
semantics above.

### 7.3 USM / UFM integration contract

- **Identity:** every ToC and element is addressable by §6.7 references.
- **Perspective:** `Perspective.actorId` identifies the market participant
  (USM) and `Perspective.futureRef` (`ufm:<future-slug>`) groups ToCs by
  competing future (UFM).
- **Shared actors:** `ActorRegistry` documents (`kind: ActorRegistry`, a list
  of `Actor`) let several ToCs reference the same actors; actor IDs from a
  registry are resolved before INV-1/INV-4 checks.
- **Alignments:** BSC objectives and measures attach via `Alignment` with
  framework `bsc`; USM computes its scorecard from ToC `Status` plus its own
  BSC data — the ToC never stores BSC data.
- **No back-references:** a ToC never lists the USM/UFM models that consume it.

### 7.4 Data Storage

| Data Type | Storage | Rationale |
|-----------|---------|-----------|
| ToC documents, observation sets, actor registries | Files in git (JSON or YAML) | Reviewable diffs, history, PR workflow; no DB in scope |
| Status snapshots | Files (JSON), optional commit per period | Enables `Diff` across periods |
| Schemas, template, rubric | Embedded in binary and committed | Offline use; single source |

## 8. Operational Guarantees

| Category | Requirement | Measurement | Applicability |
|----------|-------------|-------------|---------------|
| Performance | Validate + lint + status + exec report for 200 outcomes / 400 links / 300 indicators / 5,000 observations in < 1 s on a laptop (Apple M-series) | `go test -bench` benchmarks | Applicable |
| Throughput | Not applicable — single-user CLI/library | — | — |
| Availability | Not applicable — no service | — | — |
| Reliability | Deterministic outputs (INV-3) | Golden tests run twice | Applicable |
| Capacity limits | `PathsToImpact` bounded by `max` (default 1,000) | Unit test | Applicable |
| Degraded behavior | Missing observations yield `no_data`/`stale`, never errors | Unit tests | Applicable |
| Observability | Library returns errors; CLI logs via `slog` to stderr with `--verbose` | Manual | Applicable |
| Auditability | Every reported status cites inputs (INV-5); git history is the audit trail | Report tests | Applicable |

## 9. Security and Privacy

- [x] Authentication mechanism — Not applicable: local CLI and stdio MCP; the MCP client process owns authentication.
- [x] Authorization model — MCP `path` inputs confined to `--root` (§6.3).
- [x] Data encryption at rest — Not applicable: files under the user's control; no storage layer.
- [x] Data encryption in transit — Not applicable: no network I/O (INV-7).
- [x] Audit logging — git history plus revision log.
- [x] Vulnerability scanning — CodeQL (go-sast) and `govulncheck` in CI; dependencies pinned in `go.sum`.

Threat model highlights: (1) untrusted documents — parser limits (max file
size 10 MB, max 10,000 elements) prevent memory exhaustion; (2) user report
templates run `text/template` with no custom functions that touch the
filesystem or network; (3) restricted evidence never rendered (INV-6);
(4) MCP path traversal blocked; (5) personal data — the model is designed
for aggregate observations; stories identify actors by registry ID (which may
be an anonymized label), quotes require `consent: granted`, and documentation
directs users to keep personal records (contacts, sign-ups) in their source
systems and import only aggregates; (6) theme coding provenance is recorded
(`codedBy`) so model-coded stories can be audited; (7) external export
tools are invoked with fixed binary names and argument vectors (no shell),
on files the CLI itself wrote; (8) agent edits are confined to validated
patches (INV-9), so a misbehaving agent cannot leave a document invalid
without `--allow-errors`.

## 10. Compatibility and Migration Guarantees

**Applicability:** Applicable from the first tagged release.

- **Supported versions:** `specVersion` `0.1`; loaders reject unknown majors.
- **Forward / backward compatibility:** v0.x minors are additive; unknown
  fields are rejected by schema (`additionalProperties: false`) to catch typos,
  so a newer document fails on an older binary with a clear error.
- **Mixed-version deployment:** Not applicable — no service.
- **Migration phases:** major bumps ship `toc migrate` converting the previous major.
- **Data validation:** migrated documents must pass `toc validate`.
- **Rollback boundary:** documents are files in git; rollback is a revert.

## 11. Dependencies

| Dependency | Version (verified) | Purpose | Criticality |
|------------|--------------------|---------|-------------|
| Go | 1.26+ | Language (org standard) | High |
| `github.com/grokify/prism-core` | v0.5.0 | Metadata, Person, Risk, health, trend, CapabilityRef | High |
| `github.com/grokify/prism-roadmap` | v0.21.0 | OKR / OutcomeJourney export (adapter only) | Medium |
| `github.com/plexusone/structured-evaluation` | v0.14.0 | `summary.SummaryReport`, `rubric.RubricSet` | High |
| `github.com/invopop/jsonschema` | v0.14.0 | Schema generation | High |
| `github.com/santhosh-tekuri/jsonschema/v6` | v6.0.3 (same as prism-maturity) | Schema validation | High |
| `sigs.k8s.io/yaml` | latest verified at implementation time | YAML → JSON | Medium |
| `github.com/spf13/cobra` | v1.10.2 | CLI | Medium |
| `github.com/modelcontextprotocol/go-sdk` | latest verified at implementation time (≥ v1.8.0) | MCP server | Medium |
| `schemakit` (tool) | installed | Schema lint in CI | Medium |
| `@marp-team/marp-cli` (optional external) | latest at use | Deck export to PDF / PPTX / HTML | Low — Markdown deck works without it |
| `@mermaid-js/mermaid-cli` `mmdc` (optional external) | latest at use | Diagram → SVG for decks, PNG for the `pandoc` profile | Low — Mermaid source works without it |
| `pandoc` (optional external) | 3.x at use | Markdown → DOCX / PPTX for sharing | Low — Markdown export works without it |
| Google Drive tooling (external, used only by the `toc-share-google` skill) | gogoogle CLI with `driveutil` (from `INIT-GOGOOGLE-001`) and mcp-google comments/suggestions tools (from `INIT-MCPGOOGLE-001`) | Upload with conversion; read comments, suggestions, Sheets | Low — manual upload and CSV review sheet are the fallback |

JSON Patch is implemented in-house in `pkg/edit` (RFC 6902 plus ID
addressing); see §14.

Versions are re-verified with `go list -m -versions` when each dependency is
first added (global dependency policy). Fallback: if prism-roadmap's export
types change, `pkg/prismx` is the only package affected.

## 12. Testing Strategy

- Unit tests per package, table-driven; `t.Fatal` on errors.
- Golden files in `testdata/` for lint reports, status snapshots, and reports;
  `-update` flag regenerates.
- Seeded-defect corpus: one fixture per lint rule.
- Property tests: random DAGs for graph invariants (paths end at impact,
  levels monotone).
- Import-boundary test (TRD-014).
- Schema drift test (TRD-003) and embedded-schema load test.
- MCP in-memory transport test (TRD-013).
- Benchmarks for §8 performance.
- Coverage ≥ 80 % on `pkg/...`.

## 13. Deployment

### 13.1 Rollout and Feature Flags

Not applicable — library and CLI released as tagged Go module versions
(`go install …/cmd/toc@vX.Y.Z`). Releases follow the org release process
(CHANGELOG.json via structured-changelog; tag after CI is green).

### 13.2 Rollback Plan

Consumers pin module versions; rollback is pinning the previous tag.
Published tags are immutable, so fixes roll forward.

## 14. Alternatives Considered

| Alternative | Why rejected |
|-------------|--------------|
| Hand-written JSON Schema as the source of truth | Violates org Go-first policy; drifts from code |
| Embed observations inside the ToC document | Couples slow-changing theory with per-period data; noisy diffs; harder agent edits. Rejected for separate `ObservationSet` |
| Model the ToC as a logframe (fixed input → activity → output → outcome → impact rows) | Cannot express multiple pathways, AND/OR preconditions, or feedback; loses the ToC's main value |
| Adjacency stored on outcomes (`preconditions: [ids]`) instead of first-class links | Links need their own rationale, assumptions, evidence, and confidence; first-class edges keep those addressable |
| Reuse `prism-roadmap` `common.Assumption` | Lacks type, criticality, appliesTo, test, and status needed for propagation; ToC-specific type is justified, and prism-core has no assumption type |
| Wrap structured-evaluation `claims.Claim` for evidence (as prism-roadmap assessment does) | Heavier than needed for v0; revisit when evidence verification by LLM judges is in scope |
| Graph database or Dolt storage | Out of scope; git files suffice for document-sized graphs |
| Per-outcome `logic` only (no groups) | Cannot express nested logic such as `(A AND B) OR C`; ActKnowledge "clusters" and Davies' critique both call for explicit joint preconditions |
| Interventions only as nodes with outgoing links (logic-model style) | ActKnowledge places interventions on connectors ("what it takes for A to produce B"); both placements are supported via `Intervention.targets` and `CausalLink.interventionIds` |
| Agents rewrite whole documents (no patch engine) | Large, unreviewable diffs; easy to drop elements silently; no atomic validation. Patches are small, reviewable, and checkable before write |
| Third-party JSON Patch library (e.g. evanphx/json-patch) | Lacks ID-segment addressing and error-delta rejection; wrapping it still needs a custom resolver; the in-house engine over `toc.Document` JSON is small and fully tested |
| Generate PPTX directly from Go | Large dependency and layout effort; Marp gives Markdown-native decks consistent with prism-maturity, with PDF/PPTX export as an optional step |
| Import Google Drive SDK into this module to publish and read reviews | Couples a spec library to one cloud; violates INV-7; the file-based `ReviewResponse` lets any channel (Google, Microsoft 365, email, workshop notes) feed the same import |
| Apply review responses directly | Free-text feedback needs judgment; proposals plus `toc patch` keep a human decision on every change and keep INV-9 |
| Marp PPTX for shared decks | Marp's PPTX is image-based, so Google Slides conversion yields non-editable, unreadable slides; Pandoc PPTX keeps native text |
| Judge inside this module via an LLM SDK | Violates the no-LLM-dependency non-goal; bundle + verdict files let any judge (agent, visionspec, visionstudio) participate and keep results auditable |
| Closed enum for outcome category | Development (ORS, COM-B), market-systems (AAER), and corporate ToCs need different taxonomies; `category` is an open string with documented recommended vocabularies |

### 14.1 Deferred extensions (reserved names)

From the research brief §8.4, these are deferred to a later initiative and
their names are reserved so v0 documents do not collide:
`contributionClaims` (Contribution Analysis), `observedOutcomes` (Outcome
Harvesting with substantiation; an unmapped observed outcome becomes a
revision prompt — Most-Significant-Change-style *stories* are in scope via
`ObservationSet.stories`), and logframe matrix projections (per-pathway
logic-model views are in scope). The UNDG problem tree is in scope as
`Document.causes` (research §9.8).

## 15. Risks and Open Questions

| ID | Risk / Question | Impact | Mitigation / Resolution |
|----|-----------------|--------|-------------------------|
| Q-1 | Should `output` be an outcome level or a separate type? | Model shape | Resolved: outcome level `output`, so outputs can be measured and linked uniformly; interventions also list free-text `outputs` |
| Q-2 | Default roll-up `worst` may look pessimistic to executives | Adoption | Configurable; exec report states the method |
| Q-3 | Confidence weights are opinionated | Trust | Defaults documented, overridable, printed with every score |
| R-1 | schemakit may flag `map[string]string` disaggregation | Lint failure | Replace with `[]KeyValue` if flagged |
| R-2 | prism-core `Metadata` uses `omitzero` time fields; schema generation of `time.Time` | Schema shape | Accept `date-time` format; covered by schema tests |

No open question blocks implementation.

## Appendix

### A. Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1 | 2026-10-29 | John Wang | Initial draft |
| 0.2 | 2026-10-29 | John Wang | Measurement-framework design for small-staff and intermediary organizations: pathways, metric library, claims, selection, burden, audiences, glossary, story codebook, learning reviews, CSV interop, definition reports, agent skills |
| 0.3 | 2026-10-29 | John Wang | UNDG primary-source review: problem tree (causes), outcome responsibility and delivering actors, focus / solution prioritization criteria, target groups, consultation records, partnership assumptions with plausibility and needs-testing, risk categories, narrative and division-of-labour reports, core + development rubric profiles |
| 0.4 | 2026-10-29 | John Wang | Editing (patch engine with ID addressing, element commands, canonical fmt, semantic diff, toc_patch), presentations (Marp decks, optional SVG/PDF/PPTX export), judging (bundles, verdict files, multi-judge agreement), lint profiles |
| 0.5 | 2026-10-29 | John Wang | Public-source drafting with element provenance and data availability, assessment report, sharing exports (pandoc profile, review packet and sheet, export stamps), ReviewResponse import as proposals, approval, toc-draft-from-sources and toc-share-google skills over external Google tooling |
