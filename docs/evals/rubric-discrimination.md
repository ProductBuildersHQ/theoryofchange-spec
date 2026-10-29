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

Verdicts and summaries are in `docs/evals/verdicts/`: `run1/` holds the first
run of both theories, `run2/` the second run of the reference theory after it
was improved (below).

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

## Second run: after improving the example

The first run named concrete gaps in the reference theory, and they were fixed:
residents and county staff were consulted (two new consultation records), the
two high-stakes assumptions without evidence now cite it, the indicator for the
low-access residents is broken down by that group, four indicators gained
baselines and targets, the retention rate gained the weight it needed, and the
assessment report now counts indicators that stand in for an outcome. The
example lints with 0 errors and 0 warnings under the standard profile. The two
judges then scored the new bundle.

| Reference theory | Sonnet | Haiku | Combined (majority) |
|------------------|--------|-------|---------------------|
| Pass / partial / fail | 10 / 2 / 0 | 6 / 6 / 0 | 6 / 6 / 0 |
| Decision | conditional | conditional | does not pass |

- Every category that was partial for process, targeting, testable, and
  structural integrity in run 1 improved: Sonnet now passes all four, Haiku
  passes two. No category fails under either judge.
- The remaining partials are different ones: Sonnet scores `evidence` and
  `plausible` partial; Haiku scores `logic_flow`, `assumptions_risks`,
  `targeting`, `plausible`, `feasible`, and `testable` partial, with new
  findings (combination logic at two outcomes, two long-term outcomes measured
  only by candidate indicators, an assumption test that cites a retired
  indicator).
- Exact agreement fell to 50%, from 83% in run 1. The stricter judge (Haiku)
  raised new, mostly medium, findings each time; the gap between the two judges
  is now larger than the gap between the example and a pass.
- `plausible` is partial under both judges, for a reason that is true: an
  eight-person network running three pathways, with resident-level change
  expected within three years.

Under the rule "every required category passes", the reference theory does
not pass in either run. Under a rule of "no failed category and no high
findings", it passes in both, and both judges call it conditional.

## Acceptance

The bar for sending the reference example to human review is that no judge
raises a high or critical finding on it. Both runs meet it: run 1 raised no
high or critical findings under either judge, and neither did run 2. The
rubric's own pass rule (every required category passes, at most three medium
findings) stays as it is; it is the bar for approving a theory, not for
sharing an example for review.

## What this means

The rubric separates the two theories clearly and the judges agree closely
enough to trust it. The stated acceptance, that the reference example passes,
is not met as written, and the cause is the example and the pass rule, not a
disagreement between judges:

- The example had real gaps that both judges named independently in run 1.
  They have been closed. A second run shows the example is better, but not
  that it can be made to pass: a judge asked to find what is missing finds
  something each time.
- The pass rule, "every required category passes", is strict for a worked
  example. A theory with no failed category and no high findings is judged
  conditional, which is a defensible reading of "sound enough to show".

## Limits

Two judges, two theories, one run each for the weak theory and two for the reference. The judges are not independent of the
people who wrote the rubric and fixtures. The weak theory is deliberately weak,
so this shows the rubric is not blind, not that it grades subtle cases well.
Verdict shape: one judge returned fields in a slightly different shape and its
file was reshaped to the skeleton (category key, metadata) before recording;
no score, reasoning, or finding was changed.
