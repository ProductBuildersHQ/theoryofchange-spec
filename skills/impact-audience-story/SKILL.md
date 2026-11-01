---
name: impact-audience-story
description: Draft a short story or update for one audience (funders, board, members) from a Theory of Change's numbers and consented stories. Use when asked to write an annual report section, a funder update, or a board summary from the data.
---

# Draft an audience story

## Purpose

Write something a person will read, using only numbers the tools produced
and stories the people in them agreed to share.

## Inputs

- `impact-model.yaml` and the period's `observations.yaml`.
- The audience, by ID (see `audiences` in the theory).

## Rules

- Judgment is limited to turning the scaffold into prose. Every number in
  the draft must appear in the scaffold, exactly as given. Do not round,
  estimate, or compute a new figure.
- Use the "how to describe it" column: say "contributed to" for a figure
  marked as a contribution. Never claim credit the scaffold does not give.
- Carry the caveats with the numbers they belong to.
- Quote only the words the scaffold shows, and add no name. A story the
  scaffold does not show is not available to you, however much it would help.
- If the scaffold lacks something the audience asked for, say so rather than
  filling the gap.

## Command sequence

1. Produce the scaffold:

   ```bash
   impact-model report story impact-model.yaml -O observations.yaml --audience <id> --out scaffold.md
   ```

2. Write the draft from the scaffold only (see `templates/story-draft.md` for
   the shape): what the audience needs, three to five numbers, one or two
   stories, and a closing ask.

3. Check the draft against the scaffold: every number in the draft is in the
   scaffold, and every quotation is quoted there. Fix anything that is not.

4. For slides, generate the deck from the same data:

   ```bash
   impact-model deck story impact-model.yaml -O observations.yaml --audience <id> --out story-deck.md
   ```

## Checks to report

- Each figure in the draft and where it appears in the scaffold.
- Each quotation and where it appears.
- Anything the audience wanted that the scaffold could not supply.

## Outputs

`scaffold.md`, the draft, and optionally `story-deck.md`.
