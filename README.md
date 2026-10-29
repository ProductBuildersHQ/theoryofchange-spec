# Theory of Change Spec

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/ProductBuildersHQ/theoryofchange-spec/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/ProductBuildersHQ/theoryofchange-spec/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/ProductBuildersHQ/theoryofchange-spec/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/ProductBuildersHQ/theoryofchange-spec/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/ProductBuildersHQ/theoryofchange-spec/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/ProductBuildersHQ/theoryofchange-spec/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/ProductBuildersHQ/theoryofchange-spec
 [docs-godoc-url]: https://pkg.go.dev/github.com/ProductBuildersHQ/theoryofchange-spec
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/ProductBuildersHQ/theoryofchange-spec/blob/main/LICENSE

An AI-native, spec-driven way to capture, measure, and report on a
[Theory of Change](https://en.wikipedia.org/wiki/Theory_of_change):
a Go-first data model with generated JSON Schema, plus an SDK, the
`theoryofchange` CLI, and an MCP server.

> **Status:** early development. The specification is complete; the
> implementation is proceeding phase by phase (see the roadmap below).

## What it covers

- **Model.** Long-term goals, outcomes and precondition groups (AND / OR /
  k-of-n), causal links with rationales, interventions, assumptions,
  indicators, actors, evidence, pathways, and a shared metric library.
  Every element records whether it was stated by a source, inferred, or
  confirmed in review.
- **Schema.** JSON Schemas generated from the Go types and embedded for
  editors, agents, and validation.
- **Editing.** One validated patch engine behind the CLI, SDK, and MCP tools.
- **Quality.** Causal-graph analysis, deterministic lint rules, and a
  structured-evaluation rubric for LLM-as-a-Judge review.
- **Measurement.** Indicator, outcome, and pathway status with assumption
  propagation and rule-based confidence.
- **Artifacts.** Leadership reports, logic models, results frameworks,
  assumption registers, and slide decks, plus review packets for sharing and
  a review-response import.

## Documentation

- [Research brief](docs/research/theory-of-change.md) — concepts, process,
  measurement, artifacts, and the recommended entity model
- [PRD](docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/PRD.md) ·
  [TRD](docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/TRD.md) ·
  [Plan](docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/PLAN.md) ·
  [Roadmap](docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/ROADMAP.md)

## Part of the PRISM ecosystem

Shared primitives come from
[prism-core](https://github.com/grokify/prism-core); OKR and outcome-journey
exports target [prism-roadmap](https://github.com/grokify/prism-roadmap);
rubrics and GO/NO-GO reports use
[structured-evaluation](https://github.com/plexusone/structured-evaluation).

## License

[MIT](LICENSE)
