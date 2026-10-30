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

> **Status:** the first release, v0.1.0, is prepared; see the
> [release notes](docs/releases/v0.1.0.md).

## Quick start

```bash
go install github.com/ProductBuildersHQ/theoryofchange-spec/cmd/theoryofchange@latest

theoryofchange init --id my-program --name "My program" --out theory.yaml
theoryofchange validate theory.yaml
theoryofchange lint theory.yaml
theoryofchange report definitions theory.yaml --out definitions.md   # pathway, logic model, metrics, glossary
theoryofchange status theory.yaml -O observations.yaml                # how far it has come
theoryofchange mcp --root .                                          # serve the tools to an AI agent
```

Fictional, complete examples are in [`examples/`](examples/): a membership
network, a market-adoption theory, and three competing futures that share
one actor registry.

## Documentation

The user guides are in [`docs/guides/`](docs/guides/index.md) and are published
as a documentation site (`mkdocs serve` runs it locally):

| To | Read |
|----|------|
| Install and try it | [Getting started](docs/guides/getting-started.md) |
| Write or change a theory | [Authoring a theory](docs/guides/authoring.md) |
| Measure progress | [Measuring progress](docs/guides/measurement.md) |
| Produce reports and decks | [Reports and decks](docs/guides/reports-and-decks.md) |
| Send a theory out for review | [Review round trip](docs/guides/review-round-trip.md) |
| Check quality and run a judge | [Quality and judging](docs/guides/quality-and-judging.md) |
| Connect an AI agent | [MCP server and agent skills](docs/guides/mcp-and-skills.md) |
| Look up a command | [CLI reference](docs/guides/cli.md) |

Background: the [research brief](docs/research/theory-of-change.md), the
[concepts-to-types map](docs/concepts.md), and the specifications
([PRD](docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/PRD.md),
[TRD](docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/TRD.md),
[Plan](docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/PLAN.md),
[Roadmap](docs/specs/initiatives/INIT-THEORYOFCHANGESPEC-001/ROADMAP.md)).

## Part of the PRISM ecosystem

Shared primitives come from
[prism-core](https://github.com/grokify/prism-core); OKR and outcome-journey
exports target [prism-roadmap](https://github.com/grokify/prism-roadmap);
rubrics and GO/NO-GO reports use
[structured-evaluation](https://github.com/plexusone/structured-evaluation).

## License

[MIT](LICENSE)
