# Rubric discrimination evaluation

Does the core rubric tell a sound theory from a weak one, and do two
independent judges agree? Two judges scored two theories against the core
rubric (12 categories) through `theoryofchange judge`.

## Setup

- **Reference theory:** `examples/intermediary/theory.yaml`, after it was
  extended with measurement design and change analysis. Lint: GO under the
  standard profile, 0 errors, 9 warnings.
- **Weak theory:** `docs/evals/fixtures/weak-theory.yaml`, a fictional theory
  seeded with the usual faults: a vague problem, outcomes that restate
  activities, no rationale or assumption on its one causal link, no evidence,
  no consultation, one unexplained indicator. It validates; lint is GO with 6
  warnings.
- **Judges:** two Claude models run as separate sub-agents from one Claude Code
  session, each with only the bundle and a skeleton to fill in, and told not to
  read anything else. Judge names in the verdict files are `sonnet` and
  `haiku`.
- **Procedure:** `judge show` built a bundle per theory; each judge returned a
  verdict naming the bundle hash; `judge record` checked and stored it;
  `judge summary` combined them (majority; a tie goes to partial).

Verdicts and summaries are in `docs/evals/verdicts/`.

## Results

| | Reference theory | Weak theory |
|-|------------------|-------------|
| Categories failed (sonnet / haiku) | 0 / 0 | 9 / 8 |
| Categories partial (sonnet / haiku) | 5 / 5 | 3 / 4 |
| Categories passed (sonnet / haiku) | 7 / 7 | 0 / 0 |
| Exact agreement on category scores | 83% (10 of 12) | 92% (11 of 12) |
| High-severity findings, consolidated | 0 | 13 |
| Passes the rubric | No | No |

- **Discrimination holds.** The weak theory fails 8 or 9 categories under each
  judge; the reference fails none. The rubric's requirement that a weak theory
  fail at least 3 categories is met by a wide margin.
- **Agreement is good where it matters.** The judges differ on two categories
  for the reference (`evidence`, `plausible`) and one for the weak theory
  (`plausible`), each by one step (pass against partial, or partial against
  fail), never across the pass-fail gap.
- **The reference theory does not pass.** Both judges scored five required
  categories partial: `process` (residents, the least-heard group, and county
  officials are not consulted), `targeting` (outcomes name target groups but
  few indicators break results down by them), `testable`, and
  `structural_integrity` (the nine lint warnings, which the rubric reads as
  "not clean"), plus `evidence` or `plausible` depending on the judge.
  The rubric needs every required category to pass and allows three medium
  findings; the judges raised seven.

## What this means

The rubric separates the two theories clearly and the judges agree closely
enough to trust it. The stated acceptance, that the reference example passes,
is not met as written, and the cause is the example and the pass rule, not a
disagreement between judges:

- The example has real gaps that both judges named independently. Closing them
  (a consultation with residents and with county staff, target-group
  breakdowns on the indicators, clearing the lint warnings) is example work.
- The pass rule, "every required category passes", is strict for a worked
  example. A theory with no failed category and no high findings is judged
  conditional, which is a defensible reading of "sound enough to show".

## Limits

Two judges, two theories, one run each. The judges are not independent of the
people who wrote the rubric and fixtures. The weak theory is deliberately weak,
so this shows the rubric is not blind, not that it grades subtle cases well.
Verdict shape: one judge returned fields in a slightly different shape and its
file was reshaped to the skeleton (category key, metadata) before recording;
no score, reasoning, or finding was changed.
