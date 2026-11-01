# PLAN — Theory of Change Specification — Go-first model, JSON Schema, measurement, and leadership reporting

**Initiative:** `INIT-THEORYOFCHANGESPEC-001`
**Repository:** `github.com/ProductBuildersHQ/impact-model`
**Version:** 0.4
**Status:** Draft — pending tech-lead review (pbhq-lite gate: after PLAN)

This PLAN is non-normative. It sequences the work in [PRD.md](PRD.md) and
[TRD.md](TRD.md) and must not add requirements. Execution state lives in
[ROADMAP.md](ROADMAP.md).

## 1. Approach

**Thin slice first, then deepen.** The first five phases build one complete
path from end to end:

> minimal model → validate → **edit** → pathways, metric library,
> **provenance** → diagram, logic models, metric definitions, glossary,
> **workshop deck**, **draft from public sources** → **assessment, review
> packet and sheet, review import, approval**

That path is the first adopter milestone: draft a theory from public
information, share it as familiar documents, take the organization's
responses back in, and approve the agreed baseline. Later phases add depth
behind interfaces the slice has already proven.

Why this order:

1. **Model, schema, load, validate, edit.** Every consumer binds to the
   model. Every mutation goes through one validated patch engine from day
   one.
2. **Pathways, metric library, glossary, and provenance** are the smallest
   additions a real organization needs. Provenance has to arrive with the
   first draft, because the first draft is typically built from public
   sources and every element must say whether it was stated or inferred.
3. **Workshop pack and source drafting.** Diagram, logic models, metric
   definitions, glossary, CSV, a Marp workshop deck, and the
   draft-from-sources skill. Together they turn public sources into a
   presentable hypothesis theory.
4. **Review and agreement.** The assessment report, the `pandoc` sharing
   profile with export stamps, the review packet and review sheet, review
   import as proposals, and approval. This closes the loop with people who
   never touch the tools. The loop is file-based, so it works with manual
   upload and download before any Google automation exists.
5. **Measurement-framework design.** Metric selection and burden, stories,
   and UNDG change analysis.
6. **Quality.** Lint profiles, the rubric, and judging.
7. **Measurement.**
8. **Reporting, story, and presentations.**
9. **Agents and revisions.** The MCP server, skills, theory diff, and Google
   sharing automation. Google sharing waits on gogoogle and mcp-google
   (`INIT-GOGOOGLE-001`, `INIT-MCPGOOGLE-001`).
10. **Ecosystem integration and release.**

**Major bets:**

- The ToC graph is primary.
- `ObservationSet` is a separate document.
- Shared metric definitions plus deterministic tools.
- All edits go through validated patches.
- Decks reuse the report view.
- Judging is file-based.
- **Provenance lives on every element.**
- **Review responses are a channel-neutral file (`ReviewResponse`) imported
  as proposals.** Google Drive is one channel, reached through external
  tooling; this module stays free of cloud dependencies.

**Deliberately deferred:**

- contribution claims
- substantiated observed outcomes
- logframe matrices
- XLSX
- data connectors and record linkage
- a hosted service
- native PPTX generation in Go

### 1.1 Adopter milestones

The phases are ordered so a first adopter can use the tools at each stage of
a ToC-based impact-measurement engagement. The adopter is a small-staff
organization.

| Adopter milestone | What happens | Phases needed |
|-------------------|--------------|---------------|
| M0 — Hypothesis draft | Draft the ToC from public sources; render the workshop pack and assessment | 1–4 |
| M1 — Review, workshop, and agreement (thin slice) | Share the review packet and sheet; workshop with live edits; import responses; approve v1.0 | 5 |
| M1.5 — Collection plan | Score candidate metrics; set burden budget; collection plan | 6 |
| M2 — First measurement cycle | Load the first month's spreadsheet data; baseline status | 8 (plus RMI 024) |
| M3 — Mid-point learning | Learning review; audience stories; decks; rubric review of the revised ToC | 7, 9 |
| M4 — Handoff | Staff run the routine through an AI assistant, including Google sharing | 10 |

## 2. Affected Components

New repository; every component is new.

| Component | Change | Notes |
|-----------|--------|-------|
| Repo scaffold (`go.mod`, CI, `.golangci.yaml`, `CLAUDE.md`, `README.md`, `LICENSE`) | New | Mirrors market-spec / prism conventions |
| `pkg/impact`, `pkg/ref` | New | Model (including provenance) and reference grammar |
| `schema/`, `cmd/generate-schema` | New | Go-first schema pipeline |
| `pkg/impactio`, `pkg/validate`, `pkg/graph` | New | Load, canonical format, check, analyze |
| `pkg/edit`, `pkg/diff` | New | Patch engine, element commands, theory diff |
| `pkg/report`, `pkg/slides` | New | Reports and Marp decks from one view model |
| `pkg/share`, `pkg/review` | New | Sharing exports and stamps; review import and approval |
| `pkg/burden`, `pkg/tabular` | New | Collection burden; CSV import / export |
| `pkg/lint`, `pkg/judge`, `spec/` | New | Rules and profiles; judge bundles; template + rubrics |
| `pkg/measure` | New | Status computation |
| `pkg/service`, `cmd/impact-model`, `pkg/mcp` | New | Service layer; thin Cobra CLI and MCP adapters |
| `skills/` | New | Agent skills over the CLI, including source drafting and Google sharing |
| `pkg/prismx` | New | prism-roadmap adapters |
| `examples/`, `testdata/`, `docs/` | New | Fictional reference ToCs, defect corpus, concept mapping, integration contract |
| gogoogle, mcp-google | Separate initiatives | `INIT-GOGOOGLE-001` (Drive upload / comments, suggestion-aware Docs), `INIT-MCPGOOGLE-001` (MCP tools) |
| specification-workflow-spec | Follow-up (separate repo) | Sync `toc` template + rubrics as a spec type |

## 3. Work Breakdown

Step `P-NN` corresponds to `RMI-THEORYOFCHANGESPEC-0NN`. Phases group steps
by sequence, not by number.

### Phase 1 — Core Model and Schema

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-01 | Scaffold module, CI, lint config, README, CLAUDE.md, LICENSE | — (enabler) | CI green | — |
| P-02 | Concept-to-type mapping doc and glossary from the research brief | PRD G-1 | Mapping covers research §8.4 and §9.8 | P-01 |
| P-03 | `pkg/ref` reference grammar | PRD-FR-604; TRD-015 | Round-trip tests | P-01 |
| P-04 | `pkg/impact` core model with a minimal required set, reusing prism-core | PRD-FR-101…110; TRD-001, 016 | Round-trip, enum, minimal fixture | P-02, P-03 |
| P-05 | `ObservationSet`, `Status`, `StatusDiff`, `ActorRegistry` | PRD-FR-111, 406; TRD-002 | Round-trip tests | P-04 |
| P-06 | Schema pipeline, embed, schemakit lint, drift check | PRD-FR-201, 202; TRD-003, 004 | 0 lint violations; drift check | P-04, P-05 |

### Phase 2 — Load, Validate, and Edit

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-07 | `pkg/impactio` load, limits, merge, canonical `Format` | PRD-FR-203, 704; TRD-005, 031 | YAML/JSON equality; fmt idempotence | P-06 |
| P-08 | `pkg/validate` | PRD-FR-301; TRD-006 | Table tests | P-07 |
| P-09 | `pkg/graph` | PRD-FR-302; TRD-007 | Unit, property, benchmark | P-08 |
| P-46 | `pkg/edit` patch engine | PRD-FR-701, 702, G-14; TRD-029 | Rejection and `@id` tests | P-08 |
| P-11 | `pkg/service` + `cmd/impact-model`: validate, graph, schema, init, fmt, patch | PRD US-1, US-16; TRD §6.1 | CLI tests | P-09, P-46 |
| P-47 | Element commands compiled to patches | PRD-FR-703, US-16; TRD-030 | Equivalence; dangling-reference tests | P-46, P-11 |

### Phase 3 — Pathways, Metric Library, Provenance, and Reference Examples

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-32 | Pathways and parent / child ToC references | PRD-FR-112, US-9; TRD-017 | Round-trip; cross-document tests | P-08, P-47 |
| P-33 | Metric library, glossary, indicator extensions | PRD-FR-113, 114, 117, US-10; TRD-017 | Roll-up declaration validation | P-32 |
| P-52 | Provenance on every element, indicator `dataAvailability`, `Consultation` type; rejected-element exclusion | PRD-FR-124, 125, 123; TRD-036 | Round-trip; exclusion tests | P-33 |
| P-10 | Reference examples (fictional) with provenance: intermediary organization and market adoption | PRD G-1, G-9, US-1 | Both validate | P-33, P-52 |

### Phase 4 — Workshop Pack and Source Drafting

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-22 | Versioned `report.View`, status optional, basis badges | PRD-FR-501, 502, 505; TRD-010 | View tests | P-09, P-10 |
| P-23 | Mermaid pathway renderer | PRD-FR-501, G-12 | Golden Mermaid | P-22 |
| P-37 | Logic-model, metric-definitions, glossary reports | PRD-FR-505, US-12; TRD-022 | Golden files | P-22 |
| P-49 | Marp engine and workshop deck | PRD-FR-801, G-15, US-17; TRD-033 | Golden deck | P-23, P-37 |
| P-39 | `impact-model report` definition kinds, `definitions` pack, `impact-model export csv` | PRD-FR-204, 505, G-12; TRD-021 | CLI and CSV golden tests | P-23, P-37, P-11 |
| P-53 | `impact-model evidence add`; `impact-draft-from-sources` skill | PRD-FR-606, US-19, G-17; TRD-037 | Skill smoke test over saved fixture pages | P-47, P-52, P-39 |

### Phase 5 — Review and Agreement (→ M1)

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-54 | `assessment` report | PRD-FR-512, US-19; TRD-038 | Golden file | P-22, P-52 |
| P-56 | `pandoc` Markdown profile, PNG diagrams, export stamps, `impact-model export docx\|pptx` | PRD-FR-901, 902; TRD-039 | Golden Markdown; stamp tests; pandoc test (skipped when absent) | P-39, P-49, P-23 |
| P-57 | Review packet and review sheet exports | PRD-FR-903, US-20; TRD-039 | Golden files | P-54, P-56 |
| P-58 | `ReviewResponse` schema; `impact-model import review` (JSON and `--from-csv`) to proposals plus consultation | PRD-FR-904, 905, US-20; TRD-040; INV-10 | Seeded-row tests incl. stale hash and quoted-text resolution | P-57, P-46 |
| P-55 | `impact-model approve` | PRD-FR-707, US-21; TRD-041 | Approval and refusal tests | P-58, P-52 |

### Phase 6 — Measurement Framework Design (→ M1.5)

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-34 | Selection, collection cost, budget, `pkg/burden` | PRD-FR-115, G-10, US-11; TRD-020 | Burden tests | P-33 |
| P-35 | Audiences, theme codebook, `Story`, `LearningReview` | PRD-FR-116…118; TRD-017, 018 | Validation tests | P-33 |
| P-44 | UNDG change analysis: causes, responsibility, priority criteria, target groups; assumption and risk enrichment | PRD-FR-105, 108, 119…122; TRD-025 | Reference validation | P-32, P-34 |
| P-36 | Intermediary example extended; schemas lint-clean | PRD G-9 success metric | Validates; schemakit 0 | P-34, P-35, P-44 |
| P-38 | Metric-selection and collection-plan reports and CSV | PRD-FR-505, US-11; TRD-022 | Golden files | P-22, P-34, P-39 |

### Phase 7 — Quality: Lint, Rubric, and Judging

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-12 | Lint engine with profiles | PRD-FR-303, 304, 804; TRD-008, 035 | Registry, profile tests | P-09 |
| P-13 | Structure and logic rules | PRD-FR-303 | Fixture per rule | P-12 |
| P-14 | Remaining rule families (incl. provenance TOC-PRV-*) | PRD-FR-303, US-10, US-11 | Fixture per rule | P-12, P-36 |
| P-15 | Template, core rubric, development profile | PRD-FR-601; TRD-012, 028 | Parse; checklist coverage | P-04 |
| P-50 | `pkg/judge` show / record / summary | PRD-FR-803, G-16, US-18; TRD-034 | Bundle, hash, aggregation tests | P-15, P-39 |
| P-16 | `impact-model lint` CLI; discrimination eval with ≥ 2 judges | PRD US-2 | Eval record | P-13, P-14, P-50 |

### Phase 8 — Measurement (→ M2)

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-17 | Indicator status incl. `baseline_only` | PRD-FR-401, 402, 509; TRD-009, 019 | Branch tests | P-08, P-33 |
| P-18 | Outcome, pathway, cross-pathway roll-ups | PRD-FR-403, 113; TRD-019 | Logic and aggregation tests | P-09, P-17 |
| P-19 | Assumption propagation; confidence | PRD-FR-404, 405 | US-5 test | P-18 |
| P-20 | `measure.Evaluate`, theme counts, `Diff` | PRD-FR-406; INV-3 | Golden snapshots | P-19, P-35 |
| P-40 | CSV observation import | PRD-FR-204, US-13; TRD-021 | Seeded-bad-row tests | P-07, P-33 |
| P-21 | `impact-model status`, `impact-model import` CLI; benchmarks | PRD US-4, US-13; TRD §8 | CLI; < 1 s | P-20, P-40, P-11 |

### Phase 9 — Reporting, Story, and Presentations (→ M3)

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-24 | Status reports | PRD-FR-501, US-3 | Golden; citations | P-20, P-22, P-23 |
| P-41 | Audience story report | PRD-FR-506, US-14; TRD-023 | Golden; absence | P-24, P-35 |
| P-42 | Learning-review and portfolio reports | PRD-FR-507, 508, US-14 | Golden | P-24, P-34 |
| P-45 | Narrative and division-of-labour reports | PRD-FR-510, 511, 120; TRD-026, 027 | Golden | P-22, P-44 |
| P-25 | Custom templates; restricted suppression | PRD-FR-502, 503; TRD-011 | Template; absence | P-24 |
| P-51 | Exec, learning-review, story decks; export | PRD-FR-801, 802, US-17; TRD-033 | Golden; export | P-49, P-24, P-41, P-42 |
| P-26 | `impact-model report` status kinds; benchmark | PRD G-5 | CLI; benchmark | P-25, P-41, P-42, P-45, P-21 |

### Phase 10 — Agents and Revisions (→ M4)

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-27 | MCP server incl. `impact_patch`, `impact_deck`, `impact_judge_bundle` | PRD-FR-602, 706, US-6; TRD-013 | In-memory transport; traversal | P-26, P-46, P-50 |
| P-48 | `pkg/diff`; `impact-model diff --revision` | PRD-FR-705, US-16, US-21; TRD-032 | Golden diffs incl. rejected elements | P-46, P-09 |
| P-43 | Agent skills: author, status refresh, learning review, audience story | PRD-FR-605, US-15, G-13; TRD-024 | Smoke test twice | P-26, P-39, P-47 |
| P-59 | `toc-share-google` skill with manual fallback | PRD-FR-607, US-20, G-18; TRD-042 | Smoke test with recorded tool outputs | P-56, P-57, P-58, `RMI-GOGOOGLE-006`, `RMI-MCPGOOGLE-004`, `RMI-MCPGOOGLE-005` |

### Phase 11 — Ecosystem Integration and Release

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-28 | `pkg/prismx` exports; import boundary | PRD-FR-603, US-7; TRD-014 | prism-roadmap validation | P-20 |
| P-29 | Actor registry, cross-document refs, competing-futures fixture | PRD G-8, US-8; TRD §7.3 | 0 schema changes | P-08, P-32 |
| P-30 | USM/UFM integration contract | PRD G-8 | Owner review | P-29 |
| P-31 | Docs, CHANGELOG, v0.1.0; spec-type sync follow-up | All | CI green; publication scan clean | P-27, P-43, P-48, P-59, P-28, P-30, P-51 |

## 4. Sequencing and Dependencies

```
Phase 1:  P-01 → {P-02, P-03} → P-04 → P-05 → P-06
Phase 2:  P-06 → P-07 → P-08 → {P-09, P-46} → P-11 → P-47
Phase 3:  P-47 → P-32 → P-33 → P-52 → P-10
Phase 4:  P-10 → P-22 → {P-23, P-37} → {P-49, P-39} → P-53          ← M0
Phase 5:  P-52 → P-54;  {P-39, P-49} → P-56;  {P-54, P-56} → P-57 → P-58 → P-55   ← M1
Phase 6:  P-33 → {P-34, P-35} → P-44 → P-36;  P-34 → P-38            ← M1.5
Phase 7:  P-09 → P-12 → {P-13, P-14};  P-04 → P-15 → P-50 → P-16
Phase 8:  P-33 → P-17 → … → P-20 ─┐  P-40 ─┴→ P-21                      ← M2 (with P-24)
Phase 9:  P-20 → P-24 → {P-41, P-42, P-45, P-25} → {P-51, P-26}     ← M3
Phase 10: P-26 → {P-27, P-43};  P-46 → P-48;  {P-56..P-58, gogoogle, mcp-google} → P-59  ← M4
Phase 11: P-20 → P-28;  P-32 → P-29 → P-30;  all → P-31
```

**Critical path to M1 (27 RMIs):** P-01 → P-04 → P-06 → P-07 → P-08 → P-46 →
P-11 → P-47 → P-32 → P-33 → P-52 → P-10 → P-22 → P-37 → P-49 → P-56 → P-57 →
P-58 → P-55.

**Cross-repository dependency.** P-59 waits on `INIT-GOGOOGLE-001`
(`RMI-GOGOOGLE-006`, the CLI) and `INIT-MCPGOOGLE-001`
(`RMI-MCPGOOGLE-004/005`, suggestions and comments tools). The M1 review loop
does **not** wait on them: it runs with manual upload, a CSV review sheet,
and `impact-model import review --from-csv`.

**Parallel streams:**

- P-15 (rubric) as soon as P-04 is done.
- Phases 7 and 8 alongside each other.
- P-40 alongside P-17–P-20.
- Phase 11 alongside Phases 9–10.
- The gogoogle and mcp-google initiatives alongside Phases 1–9.

## 5. Data and Migration

Not applicable for v0.1. Pre-release documents are migrated with `impact-model fmt`
and `impact-model validate`. The model only gets additive changes after Phase 6.

## 6. Testing Strategy

As in TRD §12, plus:

- **Provenance:** rejected-element exclusion tests.
- **Review import:** seeded rows covering every origin (sheet, table,
  comment, suggestion) and stale hashes.
- **Export stamps:** checked for every format.
- **`pandoc` and `mmdc`:** export tests skipped when the tools are absent.
- **Skills:** smoke tests for `impact-draft-from-sources` (saved fixture pages)
  and `toc-share-google` (recorded gogoogle / mcp-google outputs).

Coverage target ≥ 80 % on `pkg/...`.

## 7. Rollout and Rollback

- **Phase pre-releases.** Each phase ends in `v0.1.0-alpha.N`. Adopter
  milestones pin the pre-release they used.
- **Final release.** Phase 11 ends in `v0.1.0`.
- **Tagging.** Tag only after CI is green.
- **Rollback.** Pin the previous tag.
- **Adopter data stays out of this repository.** Adopter documents, sources,
  evidence, observations, stories, review responses, decks, and verdicts live
  in the adopter's own repository or drive. This repo contains only fictional
  examples.

## 8. Risks and Open Questions

| ID | Risk / Question | Likelihood / Impact | Mitigation / Owner |
|----|-----------------|---------------------|--------------------|
| PR-1 | Model churn after Phase 6 | Medium / Medium | Thin slice exercises the core early; additive-only after Phase 6. Owner: initiative lead |
| PR-2 | schemakit flags generated shapes | Medium / Low | Key/value slices. Owner: implementer |
| PR-3 | prism-roadmap API changes | Low / Low | Adapter isolation. Owner: implementer |
| PR-4 | Rubric does not discriminate, or judges disagree | Medium / Medium | Eval with ≥ 2 judges. Owner: initiative lead |
| PR-5 | USM/UFM needs surface late | Medium / High | Contract review and fixture before v0.1.0. Owner: initiative lead |
| PR-6 | Scope creep | Medium / Medium | Reserved names; connectors upstream. Owner: initiative lead |
| PR-7 | Adopter schedule outpaces the phases | Medium / High | M0 needs Phases 1–4; M1 adds Phase 5 (27 RMIs in all). Use pre-release tags and the manual review fallback. Owner: initiative lead |
| PR-8 | Adopter data leaks into this public-ready repo | Low / High | Fictional examples; adopter files elsewhere; publication scan. Owner: initiative lead |
| PR-9 | Non-technical adopters cannot run a CLI | High / Medium | Skills, edit commands, CSV and review-sheet paths. Owner: initiative lead |
| PR-10 | Patch engine edge cases | Medium / High | Validation after every patch; property tests. Owner: implementer |
| PR-11 | Marp, Mermaid, or Pandoc CLI changes or absence | Medium / Low | Markdown is the contract; exports optional. Owner: implementer |
| PR-12 | Public-source drafts read as claims about the organization | Medium / High | Basis badges in every artifact; assessment report frames inferred items as questions; provenance lint. Owner: initiative lead |
| PR-13 | Reviewers use comments and suggestions that tooling cannot yet read | High / Medium | Response columns requested in the packet; review sheet; gogoogle and mcp-google initiatives add comment and suggestion reading. Owner: initiative lead |
| PR-14 | Free-text review rows are misinterpreted when turned into patches | Medium / Medium | Free text becomes flagged notes, not ops; a human applies proposals with `impact-model patch`. Owner: implementer |

No open question blocks Phase 1.
