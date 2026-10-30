# Measuring Progress

A theory says what you expect. Measuring says how far you have come. The
measurement is a pure function of three inputs: the theory, the observations,
and a date. The same inputs always give the same answer.

## Indicators

An **indicator** says how an outcome (or an assumption) is measured. It carries a
baseline, targets with dates, a collection frequency, and the **direction** that
counts as progress (`increase` or `decrease`). A quantitative indicator has
numbers; a qualitative one has ordered levels, such as `Emerging` to `Embedded`.

Indicators can use a shared **metric** from the library, so the same measure is
defined once. A metric says how values combine across pathways: `sum`, `mean`,
`weighted_mean`, `median`, `mode`, `min`, `max`, `latest`, or `none` for a count
that cannot be added up, such as distinct people. An indicator can also **roll
up** from other indicators; a direct observation always beats a roll-up.

An indicator's **claim** says what the organization can say about it:
`attribution` (we caused this) or `contribution` (we contributed to it). The
default is contribution.

## Observations

Measurements go in an **observation set**, a separate file that refers to the
theory by ID:

```yaml
kind: ObservationSet
specVersion: "0.1"
tocId: riverbend-network
period: 2026-Q3
observations:
  - indicatorId: ind-hours-learning
    date: "2026-09-30"
    value: 3650
    source: Attendance sheets
    quality: verified
```

An observation has a date and either a `value` (quantitative) or a `level`
(qualitative), plus a source, a quality (`verified`, `reported`, or `estimated`), and an
optional note. The same file also holds:

- **Assumption assessments**: whether an assumption is still holding.
- **Stories**: a person's account, with the consent that was given and a
  sensitivity. See [consent](reports-and-decks.md#consent-and-sensitivity).
- **Learning reviews**: what changed in a period, the surprises, and decisions.
- **Notes**: context that belongs with the numbers.

Check a set against the theory it measures:

```bash
theoryofchange validate observations-2026-q3.yaml --toc theory.yaml
```

## Load a spreadsheet

Most programs already keep their numbers in a spreadsheet. Describe its columns
once in a **mapping** file, then import:

```yaml
tocId: riverbend-network
period: 2026-Q4
dateFormat: 02/01/2006
columns:
  indicator: Measure
  date: Month end
  value: Value
  pathway: Program
  quality: Quality
  source: Source
indicatorAliases:
  Leadership practice maturity: ind-practice-maturity
metricAliases:
  Attendee hours: met-attendee-hours
```

```bash
theoryofchange import observations --csv q4.csv --map mapping.yaml --toc theory.yaml \
  --out obs-q4.yaml --unmatched q4-unmatched.csv
```

A row can name an indicator directly, through an alias, or through a metric plus
a pathway. **Every row either becomes an observation or is reported as unmatched**
with its row number and the reason; nothing is dropped silently. Unmatched rows
are written to `--unmatched` as a sheet you can fix and import again. The exit
code is 1 when some rows were unmatched (the rest are still written).

## Status

```bash
theoryofchange status theory.yaml -O obs-q3.yaml -O obs-q4.yaml --as-of 2026-12-31 --out status-q4.json
```

`status` prints a summary. `--format json` prints the full snapshot, and `--out`
writes it to a file. The date defaults to the latest observation.

### Indicator status

For each indicator the snapshot gives its progress against what the targets
call for by the date:

| Status | Meaning |
|--------|---------|
| `achieved` | The latest value has reached the target |
| `on_track` | At or within tolerance of the expected progress |
| `at_risk` | Behind expected progress by more than the tolerance |
| `off_track` | Far behind, or moving the wrong way |
| `stale` | The newest observation is older than twice the collection frequency |
| `no_data` | Nothing observed, and the first target date has passed |
| `not_started` | Nothing yet, before the first target date |
| `baseline_only` | No targets to measure against |

`--tolerance` (default 0.10) sets how far behind still counts as on track.

### Outcome and pathway status

An outcome's status combines its indicators (`--rollup worst`, the default, or
`mean`) with the state of what leads to it. **Pathway status** follows the
precondition logic: an `all` group takes its worst member, `any` its best, and
`at_least k` the *k*-th best. A negative link, an inhibitor, caps its target at
`at_risk` while its source is on track.

### Assumptions and confidence

An assumption assessed as at risk caps every link and outcome that depends on
it at `at_risk`; a critical assumption that is invalidated caps them at
`off_track`. Each link gets a rule-based **confidence** score from its evidence,
its assumptions, and whether it has a rationale. Confidence for a route to
impact is derived from its links, so a weak link shows up where it matters.

## What changed

Keep each snapshot and compare:

```bash
theoryofchange status theory.yaml -O obs-q4.yaml --prev status-q3.json
theoryofchange report exec theory.yaml -O obs-q4.yaml --prev status-q3.json
```

The comparison lists the indicators, outcomes, links, and assumptions whose
status or confidence changed, and the elements that were added or removed.

## Next

[Reports and decks](reports-and-decks.md)
