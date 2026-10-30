---
name: toc-author
description: Create or revise a Theory of Change with validated edits. Use when asked to add, change, or remove outcomes, links, assumptions, indicators, actors, or other elements of a Theory of Change, or to check and tidy one.
---

# Author and revise a Theory of Change

## Purpose

Change a theory without ever leaving the document in a broken state. Every
edit is checked before it is written, and every edit is small enough to read.

## Rules

- Judgment is limited to deciding what the theory should say. Every check
  comes from a `theoryofchange` command.
- Change the document only with the element commands (`theoryofchange
  outcome add`, `link set`, `assumption rm --cascade`, and so on) or with
  `theoryofchange patch`. Never edit the file by hand.
- Run an edit with `--dry-run` first when it is not obviously safe. A
  rejected edit writes nothing, and the command says which issues it would
  have added.
- Outcomes are states, not activities: "Leaders gain coaching skills", not
  "Run coaching". Interventions are the activities.
- Give every link a reason (`--rationale`) and name the assumptions it rests
  on. A link into an intermediate, long-term, or impact outcome without one
  is a lint warning.
- Do not mark an element `confirmed` or `rejected` unless the people it
  concerns said so. Use `--basis stated|inferred|hypothesis`, and say why in
  `--basis-note`.
- Removing an element that others refer to needs `--cascade`; read what it
  would remove with `--dry-run` first.

## Command sequence

1. Check the starting point:

   ```bash
   theoryofchange validate theory.yaml
   ```

2. Preview each edit, then make it:

   ```bash
   theoryofchange outcome set theory.yaml <id> --title "<new title>" --dry-run
   theoryofchange outcome set theory.yaml <id> --title "<new title>"
   theoryofchange link add theory.yaml --id <id> --from <outcome> --to <outcome> --rationale "<why>"
   theoryofchange assumption add theory.yaml --id <id> --statement "<what must be true>" --applies-to <link-id>
   ```

   For a change several elements share, write one patch and apply it as a unit:

   ```bash
   theoryofchange patch theory.yaml change.patch.json --dry-run
   theoryofchange patch theory.yaml change.patch.json
   ```

3. Check the result and the quality of the theory:

   ```bash
   theoryofchange validate theory.yaml
   theoryofchange lint theory.yaml
   ```

4. Show what changed since the last version, if one is saved:

   ```bash
   theoryofchange diff previous.yaml theory.yaml
   ```

## Checks to report

- Validation result.
- Lint: errors, warnings, and notes, with the rule IDs of anything new.
- The diff: elements added, changed, and removed.

## Outputs

The edited `theory.yaml`. Tell the user which warnings remain and offer to
fix them.
