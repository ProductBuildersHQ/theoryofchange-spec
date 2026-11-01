---
name: impact-draft-from-sources
description: Draft a Theory of Change for an organization from its public pages and reports, recording each source as evidence and labelling every element as stated, inferred, or hypothesis. Use when asked to draft, bootstrap, or sketch a Theory of Change from websites, reports, or other documents.
---

# Draft a Theory of Change from public sources

## Purpose

Produce a credible, honest first draft that the organization can correct. The
draft is a conversation starter, not a finding. Every element says whether
the organization stated it or we inferred it, and every stated element cites
a source.

## Inputs

- One or more source URLs or documents, already fetched and readable.
- An organization name and a short identifier (a lowercase kebab-case slug).

## Rules

- Judgment is limited to reading sources, drafting text, and asking the
  user. Every check and every number comes from a `impact-model` command.
- Change the document only with `impact-model evidence add` and
  `impact-model patch`. Never edit the file by hand.
- Copy excerpts verbatim from the source. Keep them short.
- Basis, one per element:
  - `stated`: a source says it. Cite it in `provenance.evidenceIds`.
  - `inferred`: we read it between the lines. Say why in `provenance.note`.
  - `hypothesis`: a guess to test. Say what would settle it.
- Do not use `confirmed` or `rejected`. Only the organization's own review
  decides those.
- Record only what the source supports. Do not invent indicators, targets,
  or numbers.
- Public sources only. Set `--sensitivity public` unless told otherwise.

## Command sequence

1. Start the document:

   ```bash
   impact-model init --id <slug> --name "<Organization>" --out impact-model.yaml
   ```

2. Record each source. Repeat for every page or report. The ID is derived
   from the title (`ev-<title-slug>`) when `--id` is omitted:

   ```bash
   impact-model evidence add impact-model.yaml \
     --title "<page title>" --url <url> --accessed <YYYY-MM-DD> \
     --sensitivity public --excerpt "<verbatim quote>"
   ```

3. Draft the elements as one JSON Patch (see `templates/draft.patch.json`),
   each element with a `provenance` object, then apply it:

   ```bash
   impact-model patch impact-model.yaml draft.patch.json
   ```

   If the patch is rejected, read the reported issues, fix the patch, and
   apply it again. Nothing is written on rejection.

4. Check the document:

   ```bash
   impact-model validate impact-model.yaml
   ```

5. Render the workshop pack for the organization to review:

   ```bash
   impact-model report definitions impact-model.yaml --out definitions.md
   ```

## Checks to report

- Sources recorded.
- Elements by basis (stated, inferred, hypothesis).
- `stated` elements without a citation. This must be 0.
- Validation result.

## Outputs

`impact-model.yaml` and `definitions.md`. Tell the user which elements are
inferred or hypothesis: those are the first to ask the organization about.
