# Getting Started

## Install

```bash
go install github.com/ProductBuildersHQ/theoryofchange-spec/cmd/theoryofchange@latest
theoryofchange --version
```

The tool is a single Go binary. Three optional programs add output formats; the
command that needs one tells you if it is missing and exits with code 3
without writing anything:

| Program | Needed for |
|---------|-----------|
| [Marp](https://marp.app) (`marp`) | Rendering decks to PDF, PPTX, or HTML |
| [Mermaid CLI](https://github.com/mermaid-js/mermaid-cli) (`mmdc`) | Diagrams as images in decks and Word documents |
| [Pandoc](https://pandoc.org) (`pandoc`) | Word and PowerPoint export |

Markdown reports, CSV tables, Marp Markdown, and everything that checks or
measures a theory need none of them.

## Create a theory

```bash
theoryofchange init --id my-program --name "My program" --out theory.yaml
```

This writes a small starter theory that already passes validation, with `TODO`
placeholders to replace. A theory is a YAML or JSON file; the format follows the
file extension. Open it in any editor, or change it through the commands in
[Authoring a theory](authoring.md).

## Check it

```bash
theoryofchange validate theory.yaml   # schema and structure
theoryofchange lint theory.yaml       # is the theory sound?
theoryofchange fmt theory.yaml        # write it in canonical form
```

`validate` says the document is well formed: the schema, duplicate IDs,
references that do not resolve, invalid values, bad dates. `lint` says whether
the theory is sound. A new theory is GO under `--profile draft`; the default
`standard` profile also reports warnings.

`fmt` rewrites a file in canonical form: a stable field order, two-space
indentation, and a trailing newline, so diffs stay readable. Formatting is
idempotent. YAML comments are not kept, because tool-written files are
canonical; put notes in `description` and `narrative` fields.

## Ask questions of the causal graph

```bash
theoryofchange graph theory.yaml                      # a summary
theoryofchange graph theory.yaml --orphans            # outcomes nothing leads to
theoryofchange graph theory.yaml --dead-ends          # outcomes that lead nowhere
theoryofchange graph theory.yaml --upstream --from oc-adoption
theoryofchange graph theory.yaml --format dot | dot -Tsvg > graph.svg
```

## Get a first report

```bash
theoryofchange report definitions theory.yaml --out definitions.md
```

This renders the pathways, logic models, metric definitions, and glossary as one
Markdown document. It needs no observations, so a workshop can review it before
any data exists.

## Measure progress

Once you have numbers, put them in an observation set (see
[Measuring progress](measurement.md)) and ask how far the theory has come:

```bash
theoryofchange status theory.yaml -O observations.yaml
theoryofchange report exec theory.yaml -O observations.yaml -o exec.md
```

## Working in a Git repository

Specifications are meant to be diffed and reviewed. Two habits help:

- Run `theoryofchange fmt --check *.yaml` in CI. It exits with code 1 if any
  file is not in canonical form.
- Add a `.gitattributes` line so Windows checkouts keep LF line endings. The
  canonical form is byte-exact LF, and a checkout that converts to CRLF will
  look non-canonical.

  ```text
  * text=auto eol=lf
  *.csv -text
  ```

  CSV files exported by the tool follow RFC 4180 and keep their CRLF, so they
  are left alone.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | The input is invalid, or a change was rejected, or a check did not pass |
| 2 | A usage or I/O error |
| 3 | An optional external tool (`marp`, `mmdc`, `pandoc`) is missing |

## Next

Read [Authoring a theory](authoring.md), or open a complete theory in
[Examples](examples.md).
