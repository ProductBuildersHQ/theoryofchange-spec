# CLAUDE.md — theoryofchange-spec

Go-first Theory of Change specification: Go structs in `pkg/toc` are the
source of truth; JSON Schemas under `schema/` are generated from them and
embedded. The SDK, the `theoryofchange` CLI (Cobra), and the MCP server are
thin layers over one service layer.

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
- **Edits** to ToC documents go through the patch engine (`pkg/edit`) once it
  exists; nothing else writes modified documents.
- **Determinism.** Same inputs produce byte-identical outputs; never let map
  iteration order reach output.
- **Reuse PRISM.** Use `prism-core` types (Person, Approver, Risk, health,
  trend, CapabilityRef) instead of redefining them.
- **Examples are fictional.** Never commit a real organization's ToC,
  observations, stories, or review responses to this repository.

## VisionStudio

Roadmap items are tracked in VisionStudio under `INIT-THEORYOFCHANGESPEC-001`.
Claim work with `visionstudio work claim <RMI-ID>` and carry the trailer
`Refs: RMI-THEORYOFCHANGESPEC-<NNN>` on every implementing commit.

## Checks before committing

```bash
gofmt -l .
go build ./...
go test ./...
golangci-lint run
```
