# ADR 0001: Rename the project to impact-model

**Status:** Accepted
**Date:** 2026-11-01

## Context

The project began as a specification for a Theory of Change, and was named
`theoryofchange-spec`. It now holds more than that. One document records the
outcomes and causal links of a program, the assumptions and evidence behind them,
the interventions that bring them about, how progress is measured, who needs to
see which numbers, and where each claim came from. The same document is drawn
as a theory of change, a logic model, an outcome map, and a logframe, and it
exports to OKRs.

Two things made the name wrong:

- **Scope.** The name says the document is a Theory of Change, and the other
  methods look like parts of it or exports from it. A logic model is not a part of
  a theory of change. It is a more operational view of the same underlying thing,
  and each view leaves out what another keeps.
- **Direction.** The project is aimed at more than nonprofits. A product team has
  the same need to say why an intervention should move a metric, and to keep what
  it believes apart from what it has observed.

The published releases, v0.1.0 and v0.2.0, had no known users, so renaming cost
little.

## Decision

The project is the **impact model**. The theory of change, the logic model, the
pathway of change, the outcome map, and the logframe are views of it.

| Was | Is |
|-----|----|
| `github.com/ProductBuildersHQ/theoryofchange-spec` | `github.com/ProductBuildersHQ/impact-model` |
| command `theoryofchange` | command `impact-model` |
| packages `pkg/toc`, `pkg/tocio` | `pkg/impact`, `pkg/impactio` |
| type `toc.Document` | `impact.Model` |
| `kind: TheoryOfChange`, `TheoryOfChangeStatus` | `kind: ImpactModel`, `ImpactModelStatus` |
| schemas `theory-of-change`, `theory-of-change-status` | `impact-model`, `impact-model-status` |
| `theory.yaml` | `impact-model.yaml` |
| fields `tocId`, `childToc` | `modelId`, `childModel` |
| cross-document reference `toc:<id>#<element>` | `model:<id>#<element>` |
| lint rule IDs `TOC-STR-001` | `IMP-STR-001` |
| flag `--toc` | flag `--model` |
| MCP tools `toc_*`, skills `toc-*` | `impact_*`, `impact-*` |

## What did not change

- **The method names.** `report pathway`, `logic-model`, `outcome-map`, and
  `logframe` are still the names of those views, and the research brief still
  describes the Theory of Change method.
- **`specVersion`**, which stays `0.1`. The reader accepts any `0.x`.
- **Identifiers of the work.** The initiative `INIT-THEORYOFCHANGESPEC-001` and the
  roadmap items `RMI-THEORYOFCHANGESPEC-*` keep their names, since an identifier is
  not a label.
- **The quality rubric and the specification template** in `spec/`. They are the
  theory of change quality checklist and outline, and will become the rubric of
  that view.
- **The example's own title.** The membership network example is a theory of
  change, and says so.

## Consequences

- **Releases.** v0.1.0 and v0.2.0 were published under the old module path and stay
  there. A module cannot change its path under an existing tag, so the first
  release under the new name is **v0.3.0**.
- **Migration.** A document written for v0.2.0 needs `kind: ImpactModel`, the
  renamed fields, and `level: medium_term` if it predates that change. Nothing else
  in the format moves.
- **Links.** GitHub redirects the old repository address. The documentation site
  moves to a new path.
- **No alias.** The old command name is not kept, since nobody uses it.

## Alternatives considered

| Name | Why not |
|------|---------|
| Keep `theoryofchange-spec` | Understates the scope and ranks the other methods below one of them |
| `impact-model-spec` | Matches the naming of some sibling repositories, but this repository also ships a command, an SDK, and an MCP server |
| `structured-impact-model` | Longer, and "structured" adds nothing a schema does not already say |
| `impact-spec` | Broad enough to mean almost anything |

`impact-model` has one real drawback: the phrase also means financial and
environmental impact assessment. The tagline says what it is, and "impact" is
already the name of the top outcome level.

A leadership framework with a similar name exists in another project. The two are
unrelated and independent, and nothing here depends on it.

## Follow-ups, not decided here

- A `theory-of-change` report kind, so each method is a named view.
- `context.values`, and an export to V2MOM beside the existing export to OKRs, with
  the OKR and V2MOM documents staying separate and linked to the model by ID.
- A report saying which views a given model has enough information to draw.
- Relations between indicators, such as one predicting another, with the evidence
  for each.
