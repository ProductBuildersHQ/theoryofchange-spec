# CLAUDE.md — impact-model

Go-first impact model specification: Go structs in `pkg/impact` are the
source of truth; JSON Schemas under `schema/` are generated from them and
embedded. The SDK, the `impact-model` CLI (Cobra), and the MCP server are
thin layers over one service layer. The theory of change, the logic model, the
outcome map, the logframe, and OKRs are projections of one impact model, and a
V2MOM projection is planned; none is part of another. "Projection" is the term
for a method's rendering; `report.View` is the data a report is built from.

## Specs first

Read before implementing:

- `docs/research/theory-of-change.md` — methodology background
- `docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/PRD.md` — what and why
- `docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/TRD.md` — contracts,
  invariants, data model, lint rules, algorithms
- `docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/PLAN.md` — sequencing
- `docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/ROADMAP.md` — RMI ledger

## Conventions

- **Go-first schemas.** Never hand-edit `schema/*.schema.json`; regenerate
  with `go generate ./schema/...` and lint with
  `schemakit lint --property-case camelCase`.
- **JSON property names** are camelCase; enum values are snake_case.
- **Library-first.** Logic lives in `pkg/...`; `cmd/` and the MCP server stay
  thin. No network I/O and no cloud SDKs in library packages.
- **Edits** to impact models go through the patch engine (`pkg/edit`) once it
  exists; nothing else writes modified documents.
- **Determinism.** Same inputs produce byte-identical outputs; never let map
  iteration order reach output.
- **Reuse PRISM.** Use `prism-core` types (Person, Approver, Risk, health,
  trend, CapabilityRef) instead of redefining them.
- **Examples are fictional.** Never commit a real organization's impact
  model, observations, stories, or review responses to this repository.

## VisionStudio

Roadmap items are tracked in VisionStudio under `INIT-THEORYOFCHANGESPEC-001`,
whose home repository is `github.com/ProductBuildersHQ/impact-model`. Claim work
with `visionstudio work claim <RMI-ID>` and carry the item's own ID as the
trailer on every implementing commit: `Refs: <RMI-ID>`.

- **New items** take the repository's slug: `RMI-IMPACTMODEL-<NNN>`, starting at
  001. `visionstudio rmi create --repo github.com/ProductBuildersHQ/impact-model`
  assigns the next one.
- **Items from before the rename** keep their IDs, `RMI-THEORYOFCHANGESPEC-001`
  to `-076`. Use them as they are; never renumber or recreate them under the new
  slug.

## Checks before committing

```bash
gofmt -l .
go build ./...
go test ./...
golangci-lint run
```
