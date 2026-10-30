# Theory of Change Spec

An AI-native, spec-driven way to capture, measure, and report on a
[Theory of Change](https://en.wikipedia.org/wiki/Theory_of_change): a Go-first
data model with generated JSON Schema, plus an SDK, the `theoryofchange` CLI,
and an MCP server.

A Theory of Change says how and why a program expects to bring about a change:
the problem, the long-term goal, the outcomes that lead to it, the
assumptions that must hold, and the evidence for each. This project keeps that
theory as a structured document, so people and AI agents can edit it safely,
check its quality, measure progress against it, and turn it into the reports and
slide decks that different audiences need.

## What you can do with it

- **Capture** a theory as YAML or JSON, with every element recording whether it
  was stated by a source, inferred, or confirmed in review.
- **Check** it: a schema, structural validation, deterministic lint rules, and a
  rubric for LLM-as-a-Judge review.
- **Measure** it: load a spreadsheet of observations and see which indicators,
  outcomes, and pathways are on track, with assumptions and confidence.
- **Report** it: logic models, outcome maps, results frameworks, assumption
  registers, executive and audience reports, and Marp, DOCX, and PPTX decks.
- **Review** it: send a review packet or sheet to reviewers, bring their answers
  back as proposed changes, and record approval.
- **Serve** it to an AI agent over MCP, with the same validated edits as the CLI.

## Start here

| If you want to | Read |
|----------------|------|
| Install it and try it in five minutes | [Getting started](guides/getting-started.md) |
| Write or change a theory | [Authoring a theory](guides/authoring.md) |
| See how far a program has come | [Measuring progress](guides/measurement.md) |
| Produce reports and decks | [Reports and decks](guides/reports-and-decks.md) |
| Send a theory out for review | [Review round trip](guides/review-round-trip.md) |
| Judge the quality of a theory | [Quality and judging](guides/quality-and-judging.md) |
| Connect an AI agent | [MCP server and agent skills](guides/mcp-and-skills.md) |
| Look at a complete theory | [Examples](guides/examples.md) |

The [release notes](releases/v0.1.0.md) list what the first release contains.

## Part of the PRISM ecosystem

Shared primitives come from [prism-core](https://github.com/grokify/prism-core);
OKR and outcome-journey exports target
[prism-roadmap](https://github.com/grokify/prism-roadmap); rubrics and GO/NO-GO
reports use [structured-evaluation](https://github.com/plexusone/structured-evaluation).

## License

[MIT](https://github.com/ProductBuildersHQ/theoryofchange-spec/blob/main/LICENSE)
