---
name: impact-learning-review
description: Prepare and record a learning review for a period of a Theory of Change: what changed, the surprises, assumptions that moved, and questions to discuss. Use when asked to prepare a learning review, a quarterly reflection, or discussion questions from the data.
---

# Prepare a learning review

## Purpose

Make the review a conversation about what was learned, with the numbers
already read. The tools find the surprises; the people decide what they mean.

## Inputs

- `impact-model.yaml` and the period's `observations.yaml`.
- The previous status snapshot, `status-prev.json`.
- A period label, such as `2026-Q3`.

## Rules

- Judgment is limited to wording discussion questions and recording what the
  group decides. Every number comes from a `impact-model` command.
- Do not decide for the group. Draft questions and record the answers it
  gives, in its words.
- Show only stories with granted consent. The report already withholds the
  rest; do not go around it.
- Record decisions and actions in the observation set's `learningReviews`,
  not in the theory.

## Command sequence

1. Compute the status and the changes since last period:

   ```bash
   impact-model status impact-model.yaml -O observations.yaml --prev status-prev.json
   ```

2. Render the review material:

   ```bash
   impact-model report learning-review impact-model.yaml -O observations.yaml \
     --prev status-prev.json --period 2026-Q3 --out review.md
   impact-model deck learning-review impact-model.yaml -O observations.yaml \
     --prev status-prev.json --period 2026-Q3 --out review-deck.md
   ```

3. Read `review.md` with the user. Add questions of your own where the
   report's list misses something, and say they are yours.

4. After the discussion, add the group's findings, decisions, and actions to
   a `learningReviews` entry for the period in `observations.yaml`, then check it:

   ```bash
   impact-model validate observations.yaml --model impact-model.yaml
   ```

## Checks to report

- Surprises listed, ahead of and behind plan.
- Assumptions whose status changed.
- Themes that are new this period.
- Whether the recorded review validates.

## Outputs

`review.md`, `review-deck.md`, and the updated `observations.yaml`.
